package modulesRepository

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	models "github.com/nDenerserve/SmartPi/models/modules"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// Hardware of the 0-10V output module (schematic "0-10V Output"):
//
//   - U3 MCP4728: 4-channel 12-bit DAC, VDD = 3.3V, ~LDAC pulled to GND via
//     R3 (optionally jumpered to D0 via JP1), so outputs update as soon as
//     a write is acknowledged - no LDAC pulse needed.
//   - U2 MCP6H04: one non-inverting amplifier per channel, supplied from
//     12V, gain = 1 + R9/R5 = 1 + 21k5/10k = 3.15.
//
// 10V at the terminal therefore needs 10/3.15 = 3.175V at the DAC. The DAC
// is driven from its internal 2.048V reference with gain x2 (4.096V full
// scale): that is below VDD for every code we use (code 3175 = 3.175V), so
// the DAC output never clips, and the bandgap is far more stable than the
// 3.3V buck output a VDD reference would depend on.
const (
	analogOut010VAmpGain   = 1.0 + 21.5/10.0 // 3.15
	analogOut010VMaxVolts  = 10.0
	mcp4728InternalVref    = 2.048
	mcp4728VDD             = 3.3
	mcp4728FullScale       = 4096.0
	mcp4728MaxCode         = 4095
	mcp4728EEPROMWriteWait = 100 * time.Millisecond
)

// MCP4728 command bytes (datasheet DS22187, section 5.6). The low bits
// carry DAC1:DAC0 (channel, bits 2:1) and UDAC (bit 0, 0 = update output on
// ACK).
const (
	mcp4728CmdMultiWrite  = 0x40 // 0100 0xxx: write DAC input register(s)
	mcp4728CmdSingleWrite = 0x58 // 0101 1xxx: write DAC input register + EEPROM
)

// analogOut010VOutputToDAC maps the outputs as labelled on the board
// (index 0-3 = output 1-4, left to right) to the MCP4728 channel (0-3 =
// A-D) that drives them. The layout routes them in pairs swapped - measured
// on hardware 2026-11-04 with A/B/C/D set to 1/2/3/4 V: output 1 read 2.01 V
// (B), 2 -> 0.98 V (A), 3 -> 4.03 V (D), 4 -> 3.01 V (C). The schematic's
// VOUTA..D -> U2 -> J4/J5 netlist doesn't tell which terminal is "left",
// so this table, not the schematic, is the reference.
var analogOut010VOutputToDAC = [4]int{1, 0, 3, 2}

// analogOut010VDACToOutput is the inverse of analogOut010VOutputToDAC.
var analogOut010VDACToOutput = func() (m [4]int) {
	for out, dac := range analogOut010VOutputToDAC {
		m[dac] = out
	}
	return m
}()

// mcp4728ConfigBits is the upper nibble of the second command byte:
// VREF=1 (internal), PD1:PD0=00 (normal), Gx=1 (gain x2).
const mcp4728ConfigBits = 0x90

// AnalogOut010VCodeFromVoltage converts a terminal voltage (0-10V) into the
// 12-bit DAC code for the internal-reference/gain-x2 configuration.
func AnalogOut010VCodeFromVoltage(voltage float64) uint16 {
	if voltage <= 0 {
		return 0
	}
	code := math.Round(voltage / analogOut010VAmpGain / (2 * mcp4728InternalVref) * mcp4728FullScale)
	if code > mcp4728MaxCode {
		return mcp4728MaxCode
	}
	return uint16(code)
}

// analogOut010VCalibratedCode is the DAC code that puts voltage on an
// output with calibration cal (see config.AnalogOutCalibration): the
// nominal voltage to ask for is (voltage - Offset) / Gain. Below what the
// output can reach (a positive offset at 0V) the code clamps to 0.
func analogOut010VCalibratedCode(voltage float64, cal config.AnalogOutCalibration) uint16 {
	return AnalogOut010VCodeFromVoltage((voltage - cal.Offset) / cal.Gain)
}

// analogOut010VVoltageFromCode converts a DAC code back into the terminal
// voltage, honouring whatever reference/gain the channel is actually set to
// (a factory-fresh MCP4728 powers up with the VDD reference until we first
// write to it) and the output's calibration.
func analogOut010VVoltageFromCode(code uint16, internalVref bool, gain int, cal config.AnalogOutCalibration) float64 {
	vref := mcp4728VDD
	if internalVref {
		vref = mcp4728InternalVref * float64(gain)
	}
	dacVolts := float64(code) / mcp4728FullScale * vref
	// The DAC output cannot exceed its supply rail.
	if dacVolts > mcp4728VDD {
		dacVolts = mcp4728VDD
	}
	// The amplifier runs from a single supply, it cannot go below 0V
	// whatever a negative offset says.
	v := math.Max(0, cal.Gain*dacVolts*analogOut010VAmpGain+cal.Offset)
	return math.Round(v*1000) / 1000
}

// mcp4728WriteBlock builds the 3-byte write block for one channel (0-3).
func mcp4728WriteBlock(cmd byte, channel int, code uint16) []byte {
	return []byte{
		cmd | byte(channel)<<1,
		mcp4728ConfigBits | byte(code>>8)&0x0F,
		byte(code),
	}
}

// AnalogOut010VSetting is one output to set and its terminal voltage.
// Channel is the output as labelled on the board, 0-3 for outputs 1-4 -
// not the DAC channel, see analogOut010VOutputToDAC.
type AnalogOut010VSetting struct {
	Channel int
	Voltage float64
}

// SetAnalogOut010V sets one or more channels of the MCP4728 0-10V module at
// address and returns the module's state as read back afterwards. With
// persist, each value is also written to the DAC's EEPROM so it survives a
// power cycle (EEPROM writes are slow and wear the cell - use sparingly).
func (m ModulesRepository) SetAnalogOut010V(address uint16, settings []AnalogOut010VSetting, persist bool, conf *config.Moduleconfig) (models.AnalogOut010VStatus, error) {
	var status models.AnalogOut010VStatus

	for _, s := range settings {
		if s.Channel < 0 || s.Channel > 3 {
			return status, fmt.Errorf("channel must be between 1 and 4, got %d", s.Channel+1)
		}
		if math.IsNaN(s.Voltage) || s.Voltage < 0 || s.Voltage > analogOut010VMaxVolts {
			return status, fmt.Errorf("voltage must be between 0 and %.0f V, got %v", analogOut010VMaxVolts, s.Voltage)
		}
	}

	dev, closeBus, err := openMCP4728(address, conf)
	if err != nil {
		return status, err
	}
	defer closeBus()

	cal, _ := conf.AnalogOut010VCalibrationFor(address)
	if err := writeAnalogOut010V(dev, settings, persist, cal); err != nil {
		status.Error = err.Error()
		return status, err
	}

	return readAnalogOut010V(dev, cal)
}

// ReadAnalogOut010V reads back all four channels of the MCP4728 0-10V
// module at address.
func (m ModulesRepository) ReadAnalogOut010V(address uint16, conf *config.Moduleconfig) (models.AnalogOut010VStatus, error) {
	dev, closeBus, err := openMCP4728(address, conf)
	if err != nil {
		return models.AnalogOut010VStatus{}, err
	}
	defer closeBus()

	cal, _ := conf.AnalogOut010VCalibrationFor(address)
	return readAnalogOut010V(dev, cal)
}

func openMCP4728(address uint16, conf *config.Moduleconfig) (*i2c.Dev, func(), error) {
	if _, err := host.Init(); err != nil {
		return nil, nil, fmt.Errorf("initializing periph host: %w", err)
	}

	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Errorf("failed to open I2C device %s: %v", conf.I2CDevice, err)
		return nil, nil, fmt.Errorf("opening I2C device %s: %w", conf.I2CDevice, err)
	}

	dev := &i2c.Dev{Bus: bus, Addr: address}
	return dev, func() { bus.Close() }, nil
}

// writeAnalogOut010V writes settings, applying cal (indexed by output).
func writeAnalogOut010V(dev *i2c.Dev, settings []AnalogOut010VSetting, persist bool, cal [4]config.AnalogOutCalibration) error {
	if len(settings) == 0 {
		return nil
	}

	if !persist {
		// Multi-Write: several channel blocks may follow each other in one
		// transaction, so a multi-channel update lands (almost) at once.
		buf := make([]byte, 0, 3*len(settings))
		for _, s := range settings {
			code := analogOut010VCalibratedCode(s.Voltage, cal[s.Channel])
			dac := analogOut010VOutputToDAC[s.Channel]
			log.Debugf("MCP4728 0x%02X: output %d (DAC %c) -> %.3f V (code %d)", dev.Addr, s.Channel+1, 'A'+dac, s.Voltage, code)
			buf = append(buf, mcp4728WriteBlock(mcp4728CmdMultiWrite, dac, code)...)
		}
		if err := dev.Tx(buf, nil); err != nil {
			return fmt.Errorf("writing to MCP4728 at 0x%02X: %w", dev.Addr, err)
		}
		return nil
	}

	// Single Write (register + EEPROM) takes one channel per transaction
	// and keeps the chip busy for up to ~50ms afterwards, during which it
	// NACKs further EEPROM writes - wait for RDY between channels.
	for _, s := range settings {
		code := analogOut010VCalibratedCode(s.Voltage, cal[s.Channel])
		dac := analogOut010VOutputToDAC[s.Channel]
		log.Debugf("MCP4728 0x%02X: output %d (DAC %c) -> %.3f V (code %d, persist)", dev.Addr, s.Channel+1, 'A'+dac, s.Voltage, code)
		if err := dev.Tx(mcp4728WriteBlock(mcp4728CmdSingleWrite, dac, code), nil); err != nil {
			return fmt.Errorf("writing output %d to MCP4728 at 0x%02X: %w", s.Channel+1, dev.Addr, err)
		}
		if err := waitMCP4728Ready(dev); err != nil {
			return err
		}
	}
	return nil
}

// waitMCP4728Ready polls the RDY/~BSY bit (bit 7 of the first read byte)
// until the EEPROM write has finished.
func waitMCP4728Ready(dev *i2c.Dev) error {
	deadline := time.Now().Add(mcp4728EEPROMWriteWait)
	buf := make([]byte, 1)
	for {
		if err := dev.Tx(nil, buf); err == nil && buf[0]&0x80 != 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("MCP4728 at 0x%02X: EEPROM write did not finish within %v", dev.Addr, mcp4728EEPROMWriteWait)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// readAnalogOut010V reads the MCP4728's 24-byte status: for each channel
// A-D, 3 bytes of live input register followed by 3 bytes of EEPROM:
//
//	byte 0: RDY/~BSY POR DAC1 DAC0 0 A2 A1 A0
//	byte 1: VREF PD1 PD0 Gx D11 D10 D9 D8
//	byte 2: D7 ... D0
func readAnalogOut010V(dev *i2c.Dev, cal [4]config.AnalogOutCalibration) (models.AnalogOut010VStatus, error) {
	var status models.AnalogOut010VStatus

	buf := make([]byte, 24)
	if err := dev.Tx(nil, buf); err != nil {
		status.Error = err.Error()
		return status, fmt.Errorf("reading from MCP4728 at 0x%02X: %w", dev.Addr, err)
	}
	log.Debugf("MCP4728 0x%02X read: % X", dev.Addr, buf)

	for i := 0; i < 4; i++ {
		reg := buf[i*6 : i*6+3]
		eep := buf[i*6+3 : i*6+6]

		ch := int(reg[0]>>4) & 0x03
		internal := reg[1]&0x80 != 0
		gain := 1
		if reg[1]&0x10 != 0 {
			gain = 2
		}
		code := uint16(reg[1]&0x0F)<<8 | uint16(reg[2])

		eepInternal := eep[1]&0x80 != 0
		eepGain := 1
		if eep[1]&0x10 != 0 {
			eepGain = 2
		}
		eepCode := uint16(eep[1]&0x0F)<<8 | uint16(eep[2])

		vref := "vdd"
		if internal {
			vref = "internal"
		}

		out := analogOut010VDACToOutput[ch]
		status.Channels[out] = models.AnalogOut010VChannel{
			Channel:        out + 1,
			DACChannel:     string(rune('A' + ch)),
			Voltage:        analogOut010VVoltageFromCode(code, internal, gain, cal[out]),
			DACValue:       code,
			EEPROMVoltage:  analogOut010VVoltageFromCode(eepCode, eepInternal, eepGain, cal[out]),
			Calibrated:     cal[out] != config.IdentityAnalogOutCalibration,
			EEPROMDACValue: eepCode,
			Vref:           vref,
			Gain:           gain,
			PowerDown:      int(reg[1]>>5) & 0x03,
		}
	}

	return status, nil
}
