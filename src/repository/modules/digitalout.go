package modulesRepository

import (
	"strconv"

	"github.com/nDenerserve/SmartPi/models"

	log "github.com/sirupsen/logrus"

	"github.com/nDenerserve/SmartPi/repository/config"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/mcp23xxx"
	"periph.io/x/host/v3"
)

func (m ModulesRepository) SetDigitalOut(address uint16, portmap map[int]bool, conf *config.Moduleconfig) (models.DigitalOutStatus, error) {

	var moduleStatus models.DigitalOutStatus

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
		return moduleStatus, err
	}
	defer bus.Close()

	// Create a new I2C IO extender
	extender, err := mcp23xxx.NewI2C(bus, mcp23xxx.MCP23017, address)
	if err != nil {
		log.Error(err)
		return moduleStatus, err
	}

	port := extender.Pins

	for key, value := range portmap {
		pin := port[0][key+3]
		log.Debug("Pin: " + pin.String() + "  Value: " + strconv.FormatBool(value))
		if value {
			err = pin.Out(true)
		} else {
			err = pin.Out(false)
		}
		if err != nil {
			log.Fatalln(err)
			return moduleStatus, err
		}
	}

	for i := 0; i < 4; i++ {
		pin := port[0][i+4]
		moduleStatus.PortStatus[i] = pin.Read().String()
	}

	return moduleStatus, nil

}

func (m ModulesRepository) ReadDigitalOutStatus(address uint16, conf *config.Moduleconfig) (models.DigitalOutStatus, error) {

	var moduleStatus models.DigitalOutStatus

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Create a new I2C IO extender
	extender, err := mcp23xxx.NewI2C(bus, mcp23xxx.MCP23017, address)
	if err != nil {
		log.Fatalln(err)
		return moduleStatus, err
	}

	port := extender.Pins

	for i := 0; i < 4; i++ {
		pin := port[0][i+4]
		moduleStatus.PortStatus[i] = pin.Read().String()
	}

	return moduleStatus, nil

}
