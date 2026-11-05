package modulesRepository

import (
	"math"
	"testing"

	"github.com/nDenerserve/SmartPi/smartpi/config"
	"periph.io/x/conn/v3/i2c"
)

func TestAnalogOut420mACode(t *testing.T) {
	id := config.IdentityAnalogOutCalibration
	// Uncalibrated: same formula as before (current / 23 * 4095).
	for _, c := range []struct {
		mA   float64
		code uint16
	}{{4, 712}, {12, 2137}, {20, 3561}} {
		if got := AnalogOut420mACodeFromCurrent(c.mA, id); got != c.code {
			t.Errorf("%v mA: code %d, want %d", c.mA, got, c.code)
		}
	}
	// Calibrated: a board that delivers 2% more plus 0.1mA needs a lower code,
	// and reading that code back reports the asked-for current.
	cal := config.AnalogOutCalibration{Gain: 1.02, Offset: 0.1}
	for _, mA := range []float64{4, 12, 20} {
		code := AnalogOut420mACodeFromCurrent(mA, cal)
		if code >= AnalogOut420mACodeFromCurrent(mA, id) {
			t.Errorf("%v mA: calibrated code %d not lower", mA, code)
		}
		if got := AnalogOut420mACurrentFromCode(code, cal); math.Abs(got-mA) > 0.006 { // 1 LSB = 5.6uA
			t.Errorf("%v mA -> code %d -> %v mA", mA, code, got)
		}
	}
}

func TestReadMCP4725DAC(t *testing.T) {
	// status, DAC register 0xDE9 (D11-D4 = 0xDE, D3-D0 = 0x9 in the upper nibble), EEPROM
	bus := &fakeMCP4728Bus{reply: []byte{0xC0, 0xDE, 0x90, 0x0D, 0xE9}}
	code, err := readMCP4725DAC(&i2c.Dev{Bus: bus, Addr: 0x60})
	if err != nil || code != 0xDE9 {
		t.Fatalf("code 0x%X, err %v", code, err)
	}
}
