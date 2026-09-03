package modulesRepository

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	models "github.com/nDenerserve/SmartPi/models/modules"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// mcp3424ChannelConfig holds the MCP3424 configuration-register byte that
// selects each of the four channels, reverse-engineered from a working
// reference implementation (adr[] = {0x10, 0x30, 0x50, 0x70}):
//
//	bit 7    RDY       - don't-care on write in continuous mode
//	bits 6:5 channel   - 00=1, 01=2, 10=3, 11=4
//	bit 4    O/C       - 1 = continuous conversion
//	bits 3:2 resolution - 00 = 12-bit (240 SPS)
//	bits 1:0 gain      - 00 = x1
//
// Writing one of these switches the chip to continuously convert that
// channel; the previous channel's conversions stop. There is no per-request
// one-shot trigger to wait on - continuous mode is what makes the plain
// fixed settle-then-read below workable.
var mcp3424ChannelConfig = [4]byte{0x10, 0x30, 0x50, 0x70}

// mcp3424SettleTime is how long to wait after switching the ADC to a channel
// before its first conversion is ready. 12-bit resolution at 240 SPS needs
// ~4.17ms; the reference implementation this was reverse-engineered from
// slept a flat 1 second per channel instead, which would make one read of
// all four channels take 4+ seconds. This keeps a generous margin over the
// datasheet minimum without that cost.
const mcp3424SettleTime = 20 * time.Millisecond

// analogInCurrentDivisor converts the raw 16-bit ADC code into a current in
// mA. Taken as-is from a working reference implementation - it is not
// derived from the MCP3424's gain/Vref values in this codebase, so do not
// "simplify" it without re-deriving and validating it against real hardware.
const analogInCurrentDivisor = 908.0

// ReadAnalogIn reads all four channels of an MCP3424-based analog input
// module at address, returning both the current (mA) and voltage (V)
// interpretation of each - see AnalogInChannel for why both are always
// computed.
func (m ModulesRepository) ReadAnalogIn(address uint16, conf *config.Moduleconfig) (models.AnalogInStatus, error) {
	var status models.AnalogInStatus

	dev, closeBus, err := openMCP3424(address, conf)
	if err != nil {
		return status, err
	}
	defer closeBus()

	for ch := 0; ch < 4; ch++ {
		current, err := readAnalogInChannel(dev, ch)
		if err != nil {
			return status, err
		}
		status.Channels[ch] = current
	}

	return status, nil
}

// ReadAnalogInChannel reads a single channel (0-3) of an MCP3424-based
// analog input module at address. Switching only the requested channel
// instead of all four is worthwhile because each channel switch costs a
// settle wait.
func (m ModulesRepository) ReadAnalogInChannel(address uint16, channel int, conf *config.Moduleconfig) (models.AnalogInChannelStatus, error) {
	var status models.AnalogInChannelStatus
	status.Channel = channel + 1

	dev, closeBus, err := openMCP3424(address, conf)
	if err != nil {
		return status, err
	}
	defer closeBus()

	reading, err := readAnalogInChannel(dev, channel)
	status.AnalogInChannel = reading
	return status, err
}

func openMCP3424(address uint16, conf *config.Moduleconfig) (*i2c.Dev, func(), error) {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open(conf.I2CDevice)
	if err != nil {
		log.Errorf("failed to open I2C device %s: %v", conf.I2CDevice, err)
		return nil, nil, err
	}

	dev := &i2c.Dev{Bus: bus, Addr: address}
	return dev, func() { bus.Close() }, nil
}

func readAnalogInChannel(dev *i2c.Dev, channel int) (models.AnalogInChannel, error) {
	if channel < 0 || channel > 3 {
		return models.AnalogInChannel{}, fmt.Errorf("channel must be between 0 and 3, got %d", channel)
	}

	if err := dev.Tx([]byte{mcp3424ChannelConfig[channel]}, nil); err != nil {
		return models.AnalogInChannel{}, fmt.Errorf("selecting channel %d: %w", channel+1, err)
	}

	time.Sleep(mcp3424SettleTime)

	// 2 data bytes + 1 status/config-echo byte, matching the 12-bit
	// continuous-mode read format. The status byte is read but, like the
	// reference implementation, never inspected.
	buf := make([]byte, 3)
	if err := dev.Tx(nil, buf); err != nil {
		return models.AnalogInChannel{}, fmt.Errorf("reading channel %d: %w", channel+1, err)
	}

	// Combined as a plain unsigned big-endian value, not sign-extended two's
	// complement - matches the reference implementation exactly
	// (`stemp=((string[0])*256+string[1])` in C, where string[] is unsigned
	// char). For a unipolar 4-20mA/0-10V signal the raw code never goes
	// negative, so this and a signed interpretation agree in practice.
	raw := int(buf[0])<<8 + int(buf[1])

	current := float64(raw) / analogInCurrentDivisor * 20.0
	return models.AnalogInChannel{
		Current: current,
		Voltage: current / 4.0,
	}, nil
}
