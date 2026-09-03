package modulesRepository

import (
	"testing"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

// fakeMCP3424Bus stands in for the real I2C bus. writes records every config
// byte written (one per channel-select), and reply is returned verbatim for
// every subsequent read, regardless of which channel it was meant for -
// tests set it before exercising a channel.
type fakeMCP3424Bus struct {
	writes []byte
	reply  []byte
	txErr  error
}

func (f *fakeMCP3424Bus) String() string { return "fake" }

func (f *fakeMCP3424Bus) Tx(addr uint16, w, r []byte) error {
	if f.txErr != nil {
		return f.txErr
	}
	if len(w) == 1 {
		f.writes = append(f.writes, w[0])
	}
	if r != nil {
		copy(r, f.reply)
	}
	return nil
}

func (f *fakeMCP3424Bus) SetSpeed(physic.Frequency) error { return nil }

func TestReadAnalogInChannel_ConfigByteSelectsRightChannel(t *testing.T) {
	bus := &fakeMCP3424Bus{reply: []byte{0x00, 0x00, 0x10}}
	dev := &i2c.Dev{Bus: bus, Addr: 0x68}

	for ch := 0; ch < 4; ch++ {
		bus.writes = nil
		if _, err := readAnalogInChannel(dev, ch); err != nil {
			t.Fatalf("channel %d: %v", ch, err)
		}
		if len(bus.writes) != 1 || bus.writes[0] != mcp3424ChannelConfig[ch] {
			t.Fatalf("channel %d: wrote %v, want [0x%02X]", ch, bus.writes, mcp3424ChannelConfig[ch])
		}
	}
}

func TestReadAnalogInChannel_ConversionFormula(t *testing.T) {
	// raw = 0x0908 = 2312. current = 2312/908*20 = 50.9188... mA.
	// Chosen so the divisor is unambiguous in the expected value if the
	// wrong constant (e.g. 980) were used by mistake.
	bus := &fakeMCP3424Bus{reply: []byte{0x09, 0x08, 0x00}}
	dev := &i2c.Dev{Bus: bus, Addr: 0x68}

	got, err := readAnalogInChannel(dev, 0)
	if err != nil {
		t.Fatal(err)
	}

	wantCurrent := float64(0x0908) / 908.0 * 20.0
	if got.Current != wantCurrent {
		t.Fatalf("current = %v, want %v", got.Current, wantCurrent)
	}
	wantVoltage := wantCurrent / 4.0
	if got.Voltage != wantVoltage {
		t.Fatalf("voltage = %v, want %v", got.Voltage, wantVoltage)
	}
}

func TestReadAnalogInChannel_RawCombinationIsUnsignedBigEndian(t *testing.T) {
	// 0xFF in the high byte must contribute +65280, not a negative
	// two's-complement value - matches the reference implementation, which
	// combined the two bytes without any sign extension.
	bus := &fakeMCP3424Bus{reply: []byte{0xFF, 0x00, 0x00}}
	dev := &i2c.Dev{Bus: bus, Addr: 0x68}

	got, err := readAnalogInChannel(dev, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Current <= 0 {
		t.Fatalf("current = %v, want a large positive value (unsigned combination), not negative/zero", got.Current)
	}
}

func TestReadAnalogInChannel_RejectsOutOfRangeChannel(t *testing.T) {
	bus := &fakeMCP3424Bus{reply: []byte{0x00, 0x00, 0x00}}
	dev := &i2c.Dev{Bus: bus, Addr: 0x68}

	for _, ch := range []int{-1, 4, 99} {
		if _, err := readAnalogInChannel(dev, ch); err == nil {
			t.Fatalf("channel %d: expected an error, got none", ch)
		}
	}
}

func TestReadAnalogInChannel_PropagatesWriteError(t *testing.T) {
	wantErr := &fakeTxError{"bus unavailable"}
	bus := &fakeMCP3424Bus{txErr: wantErr}
	dev := &i2c.Dev{Bus: bus, Addr: 0x68}

	if _, err := readAnalogInChannel(dev, 0); err == nil {
		t.Fatal("expected an error when the bus Tx fails, got none")
	}
}

type fakeTxError struct{ msg string }

func (e *fakeTxError) Error() string { return e.msg }
