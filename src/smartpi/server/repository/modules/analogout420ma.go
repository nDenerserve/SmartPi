package modulesRepository

import (
	"errors"
	"math"
	"sync"

	log "github.com/sirupsen/logrus"

	models "github.com/nDenerserve/SmartPi/models/modules"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// MCP4725 commands
const (
	MCP4725_Command_Write_EEPROM = 0x60
	MCP4725_Command_Read_EEPROM  = 0x60
)

// lastValues stores the last set current value for each MCP4725 address
var (
	lastValues     = make(map[uint16]float64)
	lastValuesLock sync.Mutex
)

// AnalogOut420mA represents the MCP4725 4-20mA output module
func (m ModulesRepository) SetAnalogOut420mA(address uint16, current float64, conf *config.Moduleconfig) (models.AnalogOut420mAStatus, error) {
	var moduleStatus models.AnalogOut420mAStatus

	// Validate current range (4-20mA)
	if current < 4.0 || current > 20.0 {
		return moduleStatus, errors.New("current must be between 4.0 and 20.0 mA")
	}

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus
	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Fatalf("failed to open I²C device %s: %v", conf.I2CDevice, err)
		moduleStatus.Error = "Failed to open I2C device: " + err.Error()
		return moduleStatus, err
	}
	defer bus.Close()

	// Create I2C device at the specified address
	dev := i2c.Dev{Bus: bus, Addr: address}
	log.Debugf("Created I2C device for MCP4725 at address 0x%02X on bus %s", address, conf.I2CDevice)

	// Calculate DAC value from current (4-20mA)
	// Standard formula: DAC = (current - 4) / 16 * 4095
	// This maps 4mA -> 0, 20mA -> 4095 (12-bit full scale)
	dacValue := uint16(math.Round((current / 23.0) * 4095.0))
	log.Debugf("Calculated DAC value for %.3f mA: 0x%03X (%d)", current, dacValue, dacValue)

	// Ensure DAC value is within range
	if dacValue > 4095 {
		dacValue = 4095
	}

	// MCP4725 Fast Mode format:
	// Byte 0: 0x40 | (D11-D8)  -> Command + upper 4 bits of DAC
	// Byte 1: D7-D0            -> Lower 8 bits of DAC
	// Based on user testing: i2cset -y 1 0x60 0x44 0x00 0x44 0x00 works
	// For DAC=1024 (0x0400): 0x40 | 0x04 = 0x44, 0x00
	data := make([]byte, 4)
	data[0] = byte(dacValue >> 8)
	data[1] = byte(dacValue & 0xFF)
	data[2] = byte(dacValue >> 8)   // Repeat
	data[3] = byte(dacValue & 0xFF) // Repeat

	log.Debugf("MCP4725: Address=0x%02X, DAC=0x%03X (%d), Data=[0x%02X, 0x%02X, 0x%02X, 0x%02X]",
		address, dacValue, dacValue, data[0], data[1], data[2], data[3])

	// Write to DAC (4 bytes as required by this MCP4725 module)
	err = dev.Tx(data, nil)
	if err != nil {
		log.Errorf("failed to write to MCP4725 at address 0x%02X: %v", address, err)
		moduleStatus.Error = "I2C write failed: " + err.Error()
		return moduleStatus, err
	}

	log.Debugf("Successfully wrote to MCP4725 at address 0x%02X: [0x%02X, 0x%02X, 0x%02X, 0x%02X]", address, data[0], data[1], data[2], data[3])

	// Store the last set value
	lastValuesLock.Lock()
	lastValues[address] = current
	lastValuesLock.Unlock()

	// Set status values
	moduleStatus.SetValue = current
	moduleStatus.CurrentValue = current

	return moduleStatus, nil
}

// ReadAnalogOut420mAStatus reads the current status of the MCP4725 module
// Note: MCP4725 doesn't have direct DAC readback, so we return the last set value
// from memory or attempt to read from EEPROM
func (m ModulesRepository) ReadAnalogOut420mAStatus(address uint16, conf *config.Moduleconfig) (models.AnalogOut420mAStatus, error) {
	var moduleStatus models.AnalogOut420mAStatus

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus
	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
		return moduleStatus, err
	}
	defer bus.Close()

	// Create I2C device at the specified address
	dev := i2c.Dev{Bus: bus, Addr: address}

	// Try to read from EEPROM first
	dacValueFromEEPROM, err := readMCP4725EEPROM(&dev)
	if err == nil {
		// Successfully read from EEPROM, convert to current
		currentFromEEPROM := CalculateCurrentFromDACValue(dacValueFromEEPROM)
		moduleStatus.CurrentValue = currentFromEEPROM
		moduleStatus.SetValue = currentFromEEPROM
		log.Debugf("Read EEPROM value for address 0x%02X: %.3f mA", address, currentFromEEPROM)
	} else {
		// Fall back to last value from memory
		lastValuesLock.Lock()
		if lastValue, ok := lastValues[address]; ok {
			moduleStatus.CurrentValue = lastValue
			moduleStatus.SetValue = lastValue
			log.Debugf("Read last value from memory for address 0x%02X: %.3f mA", address, lastValue)
		} else {
			// No value found
			moduleStatus.CurrentValue = 0.0
			moduleStatus.SetValue = 0.0
			log.Debugf("No value found for address 0x%02X", address)
		}
		lastValuesLock.Unlock()
	}

	return moduleStatus, nil
}

// readMCP4725EEPROM attempts to read the DAC value from MCP4725 EEPROM
func readMCP4725EEPROM(dev *i2c.Dev) (uint16, error) {
	// Try reading 4 bytes (matching the write format)
	readBuf := make([]byte, 5)
	err := dev.Tx(nil, readBuf)
	if err != nil {
		log.Debugf("Failed to read from MCP4725: %v", err)
		return 0, err
	}
	log.Debugf("MCP4725 read raw bytes: [0x%02X, 0x%02X, 0x%02X, 0x%02X]",
		readBuf[0], readBuf[1], readBuf[2], readBuf[3])

	// DAC-Wert aus den Read-Daten extrahieren
	dacValue := uint16(readBuf[1])<<4 | uint16(readBuf[2]>>4)
	log.Debugf("MCP4725 read raw DAC value (before masking): 0x%03X (%d)", dacValue, dacValue)

	dacValue &= 0x0FFF // Mask to 12 bits
	log.Debugf("MCP4725 read DAC value (after 12-bit mask): 0x%03X (%d)", dacValue, dacValue)
	return dacValue, nil
}

// CalculateDACValueFromCurrent converts mA value to MCP4725 DAC value
func CalculateDACValueFromCurrent(current float64) uint16 {
	if current <= 4.0 {
		return 0
	}
	if current >= 20.0 {
		return 4095
	}
	return uint16(math.Round(current / 23.0 * 4095.0))
}

// CalculateCurrentFromDACValue converts DAC value back to mA
// Standard 4-20mA formula: current = (DAC / 4095) * 23
func CalculateCurrentFromDACValue(dacValue uint16) float64 {
	return float64(dacValue) * 23.0 / 4095.0
}
