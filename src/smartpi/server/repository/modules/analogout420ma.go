package modulesRepository

import (
	"errors"
	"fmt"
	"math"
	"sync"

	log "github.com/sirupsen/logrus"

	models "github.com/nDenerserve/SmartPi/models/modules"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// lastValues stores the last set current value for each MCP4725 address
var (
	lastValues     = make(map[uint16]float64)
	lastValuesLock sync.Mutex
)

// The module's current output is proportional to the DAC code: full scale
// (code 4095) is 23mA, so 4-20mA is codes 712-3561 and both ends have some
// headroom for a calibration to correct into.
const analogOut420mAFullScale = 23.0

// AnalogOut420mACodeFromCurrent is the DAC code for current (mA) on an
// output with calibration cal (measured = Gain * nominal + Offset).
func AnalogOut420mACodeFromCurrent(current float64, cal config.AnalogOutCalibration) uint16 {
	nominal := (current - cal.Offset) / cal.Gain
	code := math.Round(nominal / analogOut420mAFullScale * 4095.0)
	if code <= 0 || math.IsNaN(code) {
		return 0
	}
	if code > 4095 {
		return 4095
	}
	return uint16(code)
}

// AnalogOut420mACurrentFromCode is the inverse of
// AnalogOut420mACodeFromCurrent, rounded to 1uA.
func AnalogOut420mACurrentFromCode(code uint16, cal config.AnalogOutCalibration) float64 {
	nominal := float64(code) * analogOut420mAFullScale / 4095.0
	return math.Round((cal.Gain*nominal+cal.Offset)*1000) / 1000
}

func openMCP4725(address uint16, conf *config.Moduleconfig) (*i2c.Dev, func(), error) {
	if _, err := host.Init(); err != nil {
		return nil, nil, fmt.Errorf("initializing periph host: %w", err)
	}
	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Errorf("failed to open I2C device %s: %v", conf.I2CDevice, err)
		return nil, nil, fmt.Errorf("opening I2C device %s: %w", conf.I2CDevice, err)
	}
	return &i2c.Dev{Bus: bus, Addr: address}, func() { bus.Close() }, nil
}

// SetAnalogOut420mA sets the output of the MCP4725 4-20mA module at
// address, applying the module's calibration if one is stored.
func (m ModulesRepository) SetAnalogOut420mA(address uint16, current float64, conf *config.Moduleconfig) (models.AnalogOut420mAStatus, error) {
	var moduleStatus models.AnalogOut420mAStatus

	if !(current >= 4.0 && current <= 20.0) {
		return moduleStatus, errors.New("current must be between 4.0 and 20.0 mA")
	}

	dev, closeBus, err := openMCP4725(address, conf)
	if err != nil {
		moduleStatus.Error = err.Error()
		return moduleStatus, err
	}
	defer closeBus()

	cal, calibrated := conf.AnalogOut420mACalibrationFor(address)
	dacValue := AnalogOut420mACodeFromCurrent(current, cal)
	log.Debugf("MCP4725 0x%02X: %.3f mA -> code %d (calibrated: %v)", address, current, dacValue, calibrated)

	// MCP4725 Fast Mode write (C2:C1 = 00, PD = 00, then D11-D0), sent
	// twice as this module has always done.
	data := []byte{byte(dacValue >> 8), byte(dacValue), byte(dacValue >> 8), byte(dacValue)}
	if err := dev.Tx(data, nil); err != nil {
		log.Errorf("failed to write to MCP4725 at address 0x%02X: %v", address, err)
		moduleStatus.Error = "I2C write failed: " + err.Error()
		return moduleStatus, fmt.Errorf("writing to MCP4725 at 0x%02X: %w", address, err)
	}

	lastValuesLock.Lock()
	lastValues[address] = current
	lastValuesLock.Unlock()

	moduleStatus.SetValue = current
	moduleStatus.CurrentValue = AnalogOut420mACurrentFromCode(dacValue, cal)
	moduleStatus.DACValue = dacValue
	moduleStatus.Calibrated = calibrated
	return moduleStatus, nil
}

// ReadAnalogOut420mAStatus reads the DAC register of the MCP4725 module at
// address back. CurrentValue is what that code drives (calibration
// applied), SetValue the last value set through this server - or the same
// as CurrentValue after a restart.
func (m ModulesRepository) ReadAnalogOut420mAStatus(address uint16, conf *config.Moduleconfig) (models.AnalogOut420mAStatus, error) {
	var moduleStatus models.AnalogOut420mAStatus

	dev, closeBus, err := openMCP4725(address, conf)
	if err != nil {
		moduleStatus.Error = err.Error()
		return moduleStatus, err
	}
	defer closeBus()

	cal, calibrated := conf.AnalogOut420mACalibrationFor(address)
	moduleStatus.Calibrated = calibrated

	dacValue, err := readMCP4725DAC(dev)
	if err != nil {
		moduleStatus.Error = err.Error()
		return moduleStatus, fmt.Errorf("reading from MCP4725 at 0x%02X: %w", address, err)
	}
	moduleStatus.DACValue = dacValue
	moduleStatus.CurrentValue = AnalogOut420mACurrentFromCode(dacValue, cal)
	moduleStatus.SetValue = moduleStatus.CurrentValue

	lastValuesLock.Lock()
	if v, ok := lastValues[address]; ok {
		moduleStatus.SetValue = v
	}
	lastValuesLock.Unlock()

	return moduleStatus, nil
}

// readMCP4725DAC reads the MCP4725 status: byte 0 status, bytes 1-2 the
// DAC register (D11-D4, D3-D0 in the upper nibble), bytes 3-4 EEPROM.
func readMCP4725DAC(dev *i2c.Dev) (uint16, error) {
	readBuf := make([]byte, 5)
	if err := dev.Tx(nil, readBuf); err != nil {
		return 0, err
	}
	log.Debugf("MCP4725 0x%02X read: % X", dev.Addr, readBuf)
	return (uint16(readBuf[1])<<4 | uint16(readBuf[2]>>4)) & 0x0FFF, nil
}
