package modulesRepository

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func (m ModulesRepository) Analog0to10FastWrite(port string, channelA float64, channelB float64, channelC float64, channelD float64) (err error) {

	cFactor := 0.316202

	chanA := uint16(channelA * cFactor)
	chanB := uint16(channelB * cFactor)
	chanC := uint16(channelC * cFactor)
	chanD := uint16(channelD * cFactor)

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open(port)
	if err != nil {
		return err
	}
	defer bus.Close()
	// Address the device with address 0x76 on the I²C bus:
	dev := i2c.Dev{Bus: bus, Addr: 0x60}
	// This is now a point-to-point connection and implements conn.Conn:
	// var _ conn.Conn = &dev

	var output_buffer []byte

	output_buffer = append(output_buffer, 0xCF)
	_, err = dev.Write(output_buffer)
	if err != nil {
		return err
	}

	output_buffer = nil

	output_buffer = append(output_buffer, byte(chanB>>8))
	output_buffer = append(output_buffer, byte(chanB&0xFF))

	output_buffer = append(output_buffer, byte(chanA>>8))
	output_buffer = append(output_buffer, byte(chanA&0xFF))

	output_buffer = append(output_buffer, byte(chanD>>8))
	output_buffer = append(output_buffer, byte(chanD&0xFF))

	output_buffer = append(output_buffer, byte(chanC>>8))
	output_buffer = append(output_buffer, byte(chanC&0xFF))

	_, err = dev.Write(output_buffer)
	if err != nil {
		return err
	}

	return nil

}
