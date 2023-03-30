package modulesRepository

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	MCP4728_I2CADDR_DEFAULT  uint8 = 0x60 ///< MCP4728 default i2c address
	MCP4728_MULTI_IR_CMD     uint8 = 0x40 ///< Command to write to the input register only
	MCP4728_MULTI_EEPROM_CMD uint8 = 0x40 ///< Command to write to the input register and EEPROM
	MCP4728_FAST_WRITE_CMD   uint8 = 0xC0 ///< Command to write all channels at once with

)

const (
	MCP4728_PD_MODE_NORMAL   uint16 = iota ///< Normal; the channel outputs the given value as normal.
	MCP4728_PD_MODE_GND_1K                 ///< VOUT is loaded with 1 kΩ resistor to ground. Most of the channel circuits are powered off.
	MCP4728_PD_MODE_GND_100K               ///< VOUT is loaded with 100 kΩ resistor to ground. Most of the channel circuits are powered off.
	MCP4728_PD_MODE_GND_500K               ///< VOUT is loaded with 500 kΩ resistor to ground. Most of the channel circuits are powered off.
)

const (
	MCP4728_GAIN_1X uint16 = iota
	MCP4728_GAIN_2X
)

const (
	MCP4728_VREF_VDD uint16 = iota
	MCP4728_VREF_INTERNAL
)

const (
	MCP4728_CHANNEL_B uint16 = iota
	MCP4728_CHANNEL_A
	MCP4728_CHANNEL_D
	MCP4728_CHANNEL_C
)

func (m ModulesRepository) Analog0to10ChannelWrite(port string, channel uint8, value float64) (err error) {

	var b2uint = map[bool]uint8{false: 0, true: 1}

	var output_buffer []byte

	cFactor := 0.316202
	val := uint16(value * cFactor)

	udac := false
	new_vref := MCP4728_VREF_INTERNAL
	new_pd_mode := MCP4728_PD_MODE_NORMAL
	new_gain := MCP4728_GAIN_2X

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

	sequential_write_cmd := MCP4728_MULTI_IR_CMD
	sequential_write_cmd |= (channel << 1)
	sequential_write_cmd |= b2uint[udac]

	output_buffer = nil

	output_buffer = append(output_buffer, byte(sequential_write_cmd))
	// VREF PD1 PD0 Gx D11 D10 D9 D8 [A] D7 D6 D5 D4 D3 D2 D1 D0 [A]
	val |= (new_vref << 15)
	val |= (new_pd_mode << 13)
	val |= (new_gain << 12)

	output_buffer = append(output_buffer, byte(val>>8))
	output_buffer = append(output_buffer, byte(val&0xFF))

	_, err = dev.Write(output_buffer)
	if err != nil {
		return err
	}

	return nil

}

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

// Not tested yet
func (m ModulesRepository) Analog0to10SaveToEEPROM(port string) (err error) {

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

	dummy_buffer := make([]byte, 0)

	input_buffer := make([]byte, 24)
	output_buffer := make([]byte, 9)

	dev.Tx(dummy_buffer, input_buffer)

	// build header byte 0 1 0 1 0 DAC1 DAC0 UDAC [A]
	eeprom_write_cmd := MCP4728_MULTI_EEPROM_CMD // 0 1 0 1 0 xxx
	eeprom_write_cmd |= uint8(0)                 // DAC1 DAC0, start at channel A obvs
	eeprom_write_cmd |= uint8(0)                 // UDAC ; yes, latch please
	// First byte is the write command+options

	output_buffer[0] = eeprom_write_cmd

	// copy the incoming input register bytes to the outgoing buffer
	// Channel A
	output_buffer[1] = input_buffer[1]
	output_buffer[2] = input_buffer[2]
	// Channel B
	output_buffer[3] = input_buffer[7]
	output_buffer[4] = input_buffer[8]
	// Channel C
	output_buffer[5] = input_buffer[13]
	output_buffer[6] = input_buffer[14]
	// Channel D
	output_buffer[7] = input_buffer[19]
	output_buffer[8] = input_buffer[20]

	_, err = dev.Write(output_buffer)
	if err != nil {
		return err
	}

	return nil
}
