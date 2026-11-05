package modulesRepository

import (
	"bytes"
	"math"
	"testing"

	"github.com/nDenerserve/SmartPi/smartpi/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

var noCal = [4]config.AnalogOutCalibration{config.IdentityAnalogOutCalibration, config.IdentityAnalogOutCalibration, config.IdentityAnalogOutCalibration, config.IdentityAnalogOutCalibration}

// fakeMCP4728Bus records every write transaction and answers reads with
// reply (truncated to the read length).
type fakeMCP4728Bus struct {
	writes [][]byte
	reply  []byte
}

func (f *fakeMCP4728Bus) String() string { return "fake" }

func (f *fakeMCP4728Bus) Tx(addr uint16, w, r []byte) error {
	if len(w) > 0 {
		f.writes = append(f.writes, append([]byte(nil), w...))
	}
	if r != nil {
		copy(r, f.reply)
	}
	return nil
}

func (f *fakeMCP4728Bus) SetSpeed(physic.Frequency) error { return nil }

func TestAnalogOut010VCodeFromVoltage(t *testing.T) {
	cases := []struct {
		volts float64
		code  uint16
	}{
		{0, 0},
		{-1, 0},
		// 10 / 3.15 / 4.096 * 4096 = 3174.6
		{10, 3175},
		{5, 1587},
		{100, 4095},
	}
	for _, c := range cases {
		if got := AnalogOut010VCodeFromVoltage(c.volts); got != c.code {
			t.Errorf("%v V: got code %d, want %d", c.volts, got, c.code)
		}
	}
}

func TestAnalogOut010VRoundTrip(t *testing.T) {
	for _, v := range []float64{0, 0.5, 2.5, 7.3, 10} {
		got := analogOut010VVoltageFromCode(AnalogOut010VCodeFromVoltage(v), true, 2, config.IdentityAnalogOutCalibration)
		// One LSB at the terminal is 4.096/4096*3.15 = 3.15mV.
		if math.Abs(got-v) > 0.002 {
			t.Errorf("%v V round-tripped to %v V", v, got)
		}
	}
}

func TestWriteAnalogOut010V_MultiWriteEncoding(t *testing.T) {
	bus := &fakeMCP4728Bus{}
	dev := &i2c.Dev{Bus: bus, Addr: 0x60}

	err := writeAnalogOut010V(dev, []AnalogOut010VSetting{
		{Channel: 0, Voltage: 10}, // output 1 = DAC B, code 3175 = 0xC67
		{Channel: 3, Voltage: 0},  // output 4 = DAC C
	}, false, noCal)
	if err != nil {
		t.Fatal(err)
	}
	if len(bus.writes) != 1 {
		t.Fatalf("want one transaction, got %d", len(bus.writes))
	}
	want := []byte{
		0x42, 0x9C, 0x67, // multi-write DAC B, VREF=int, G=x2, code 0xC67
		0x44, 0x90, 0x00, // multi-write DAC C, code 0
	}
	if !bytes.Equal(bus.writes[0], want) {
		t.Fatalf("wrote % X, want % X", bus.writes[0], want)
	}
}

func TestWriteAnalogOut010V_PersistStoresAllChannels(t *testing.T) {
	// Status read: RDY set; A at 0x123 (internal, x2), B VDD/x1 at 0x800
	// (never written by us), C powered down, D at 0xFFF. The fake answers
	// every read with this, including the RDY poll.
	reply := []byte{
		0xC0, 0x91, 0x23, 0xC0, 0x00, 0x00,
		0xD0, 0x08, 0x00, 0xD0, 0x00, 0x00,
		0xE0, 0xB0, 0x00, 0xE0, 0x00, 0x00,
		0xF0, 0x9F, 0xFF, 0xF0, 0x00, 0x00,
	}
	bus := &fakeMCP4728Bus{reply: reply}
	dev := &i2c.Dev{Bus: bus, Addr: 0x60}

	if err := writeAnalogOut010V(dev, []AnalogOut010VSetting{{Channel: 2, Voltage: 10}}, true, noCal); err != nil {
		t.Fatal(err)
	}
	if len(bus.writes) != 2 {
		t.Fatalf("want 2 transactions (set, store), got %d: %X", len(bus.writes), bus.writes)
	}
	// First the new value goes to the input register only ...
	if want := []byte{0x46, 0x9C, 0x67}; !bytes.Equal(bus.writes[0], want) { // output 3 = DAC D
		t.Errorf("set: wrote % X, want % X", bus.writes[0], want)
	}
	// ... then all four live registers go to EEPROM in one sequential
	// write from channel A, untouched channels exactly as read.
	want := []byte{0x50, 0x91, 0x23, 0x08, 0x00, 0xB0, 0x00, 0x9F, 0xFF}
	if !bytes.Equal(bus.writes[1], want) {
		t.Errorf("store: wrote % X, want % X", bus.writes[1], want)
	}
}

func TestWriteAnalogOut010V_EmptySettings(t *testing.T) {
	reply := []byte{
		0xC0, 0x90, 0x01, 0xC0, 0x00, 0x00,
		0xD0, 0x90, 0x02, 0xD0, 0x00, 0x00,
		0xE0, 0x90, 0x03, 0xE0, 0x00, 0x00,
		0xF0, 0x90, 0x04, 0xF0, 0x00, 0x00,
	}
	// Without persist: no bus traffic at all.
	bus := &fakeMCP4728Bus{reply: reply}
	if err := writeAnalogOut010V(&i2c.Dev{Bus: bus, Addr: 0x60}, nil, false, noCal); err != nil {
		t.Fatal(err)
	}
	if len(bus.writes) != 0 {
		t.Errorf("wrote %X", bus.writes)
	}
	// With persist: only the store of the current state.
	if err := writeAnalogOut010V(&i2c.Dev{Bus: bus, Addr: 0x60}, nil, true, noCal); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x50, 0x90, 0x01, 0x90, 0x02, 0x90, 0x03, 0x90, 0x04}
	if len(bus.writes) != 1 || !bytes.Equal(bus.writes[0], want) {
		t.Errorf("wrote %X, want [% X]", bus.writes, want)
	}
}

func TestReadAnalogOut010V_Decoding(t *testing.T) {
	reply := []byte{
		// ch A: register internal/x2 code 0xC67, EEPROM factory default (VDD, x1, 0)
		0xC0, 0x9C, 0x67, 0xC0, 0x00, 0x00,
		// ch B: register VDD/x1 code 0x800 (1.65V at DAC -> 5.198V)
		0xD0, 0x08, 0x00, 0xD0, 0x08, 0x00,
		// ch C: internal/x2 code 0, powered down mode 01
		0xE0, 0xB0, 0x00, 0xE0, 0x90, 0x00,
		// ch D: internal/x1 code 0xFFF (2.048V -> 6.451V)
		0xF0, 0x8F, 0xFF, 0xF0, 0x8F, 0xFF,
	}
	bus := &fakeMCP4728Bus{reply: reply}
	dev := &i2c.Dev{Bus: bus, Addr: 0x60}

	st, err := readAnalogOut010V(dev, noCal)
	if err != nil {
		t.Fatal(err)
	}

	a := st.Channels[1] // DAC A drives output 2
	if a.Channel != 2 || a.DACChannel != "A" || a.DACValue != 0xC67 || a.Vref != "internal" || a.Gain != 2 || math.Abs(a.Voltage-10.0) > 0.002 {
		t.Errorf("channel A decoded wrong: %+v", a)
	}
	if a.EEPROMDACValue != 0 || a.EEPROMVoltage != 0 {
		t.Errorf("channel A EEPROM decoded wrong: %+v", a)
	}
	b := st.Channels[0] // DAC B drives output 1
	if b.DACChannel != "B" || b.Vref != "vdd" || b.Gain != 1 || math.Abs(b.Voltage-5.198) > 0.002 {
		t.Errorf("channel B decoded wrong: %+v", b)
	}
	if c := st.Channels[3]; c.PowerDown != 1 || c.Channel != 4 || c.DACChannel != "C" {
		t.Errorf("channel C decoded wrong: %+v", c)
	}
	if d := st.Channels[2]; d.Channel != 3 || d.DACChannel != "D" || math.Abs(d.Voltage-6.449) > 0.002 {
		t.Errorf("channel D decoded wrong: %+v", d)
	}
}

// The board's output -> DAC routing as measured on hardware; see
// analogOut010VOutputToDAC. Also guards the inverse table.
func TestAnalogOut010VOutputMapping(t *testing.T) {
	want := [4]int{1, 0, 3, 2} // outputs 1-4 -> DAC B, A, D, C
	if analogOut010VOutputToDAC != want {
		t.Fatalf("output->DAC = %v, want %v", analogOut010VOutputToDAC, want)
	}
	for out, dac := range analogOut010VOutputToDAC {
		if analogOut010VDACToOutput[dac] != out {
			t.Errorf("inverse table wrong for DAC %d", dac)
		}
	}
}

// Output 4 of the first board, measured 2026-11-04: nominal 0.99855V gave
// 1.0091V, nominal 8.99955V gave 8.992V -> gain 0.997738, offset 12.81mV.
var measuredOutput4 = config.AnalogOutCalibration{Gain: 0.997738, Offset: 0.01281}

func TestAnalogOut010VCalibratedCode(t *testing.T) {
	// Asking for 9V must pick the code whose *measured* output is 9V:
	// nominal (9 - 0.01281) / 0.997738 = 9.0077V -> code 2860 (uncalibrated 2857).
	if got := analogOut010VCalibratedCode(9, measuredOutput4); got != 2860 {
		t.Errorf("9V calibrated: code %d, want 2860", got)
	}
	// The identity changes nothing.
	if got := analogOut010VCalibratedCode(9, config.IdentityAnalogOutCalibration); got != AnalogOut010VCodeFromVoltage(9) {
		t.Errorf("identity calibration changed the code: %d", got)
	}
	// 0V with a positive offset is unreachable: clamp to code 0, no wrap.
	if got := analogOut010VCalibratedCode(0, measuredOutput4); got != 0 {
		t.Errorf("0V calibrated: code %d, want 0", got)
	}
	// 10V stays reachable on every measured output.
	for _, c := range []config.AnalogOutCalibration{{0.998625, 0.00882}, {1.0034, -0.02114}, {1.00385, 0.01781}, measuredOutput4} {
		code := analogOut010VCalibratedCode(10, c)
		if v := analogOut010VVoltageFromCode(code, true, 2, c); math.Abs(v-10) > 0.003 {
			t.Errorf("%+v: 10V -> code %d -> %v V", c, code, v)
		}
	}
}

func TestAnalogOut010VVoltageFromCode_Calibrated(t *testing.T) {
	// Reproduces the measurement the calibration came from.
	if v := analogOut010VVoltageFromCode(317, true, 2, measuredOutput4); math.Abs(v-1.0091) > 0.0005 {
		t.Errorf("code 317: %v V, measured 1.0091", v)
	}
	if v := analogOut010VVoltageFromCode(2857, true, 2, measuredOutput4); math.Abs(v-8.992) > 0.0005 {
		t.Errorf("code 2857: %v V, measured 8.992", v)
	}
	// A negative offset never reports below 0V.
	if v := analogOut010VVoltageFromCode(0, true, 2, config.AnalogOutCalibration{Gain: 1, Offset: -0.02}); v != 0 {
		t.Errorf("negative offset at code 0: %v V", v)
	}
}

func TestReadAnalogOut010V_CalibratedFlag(t *testing.T) {
	cal := noCal
	cal[3] = measuredOutput4
	// All four DACs at code 0, internal reference x2.
	reply := []byte{
		0xC0, 0x90, 0x00, 0xC0, 0x90, 0x00,
		0xD0, 0x90, 0x00, 0xD0, 0x90, 0x00,
		0xE0, 0x90, 0x00, 0xE0, 0x90, 0x00, // DAC C -> output 4
		0xF0, 0x90, 0x00, 0xF0, 0x90, 0x00,
	}
	dev := &i2c.Dev{Bus: &fakeMCP4728Bus{reply: reply}, Addr: 0x60}
	st, err := readAnalogOut010V(dev, cal)
	if err != nil {
		t.Fatal(err)
	}
	for i, ch := range st.Channels {
		if ch.Calibrated != (i == 3) {
			t.Errorf("output %d calibrated = %v", i+1, ch.Calibrated)
		}
	}
	if v := st.Channels[3].Voltage; math.Abs(v-0.013) > 0.0005 {
		t.Errorf("output 4 at code 0 reports %v V, want the offset 0.013", v)
	}
}
