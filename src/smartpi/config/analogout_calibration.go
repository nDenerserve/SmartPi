package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// modulesConfigPath is the module configuration file. A variable so tests
// can point the calibration writer at a temporary file.
var modulesConfigPath = "/etc/smartpiModules"

// AnalogOutCalibration corrects one analog output for its amplifier's
// tolerances: the voltage actually measured at the terminal is
//
//	measured = Gain * nominal + Offset
//
// with nominal the value the uncalibrated formula expects and Offset in the
// output's unit: volts for the 0-10V module, milliamps for the 4-20mA
// module. The identity is {Gain: 1, Offset: 0}. Determined per output from
// two measurements, e.g. at 1V and 9V or at 4mA and 20mA.
type AnalogOutCalibration struct {
	Gain   float64 `json:"gain"`
	Offset float64 `json:"offset"`
}

// IdentityAnalogOutCalibration leaves the outputs uncorrected.
var IdentityAnalogOutCalibration = AnalogOutCalibration{Gain: 1}

// Plausibility limits: the amplifier's resistors are 1% parts, so a real
// calibration is within a few percent and some tens of millivolts. Anything
// beyond these is a typo or a wrong measurement, not a correction.
const (
	analogOutCalibrationMaxGainError = 0.1 // +-10%
	analogOutCalibrationMaxOffset    = 0.5 // +-0.5V (0-10V module)
	analogOut420mACalibrationMaxOff  = 1.0 // +-1mA (4-20mA module)
)

// Validate rejects calibrations of a 0-10V output that cannot be a real
// correction.
func (c AnalogOutCalibration) Validate() error {
	return c.validate(analogOutCalibrationMaxOffset, "V")
}

// ValidateCurrent is Validate for a 4-20mA output (offset in mA).
func (c AnalogOutCalibration) ValidateCurrent() error {
	return c.validate(analogOut420mACalibrationMaxOff, "mA")
}

func (c AnalogOutCalibration) validate(maxOffset float64, unit string) error {
	if math.IsNaN(c.Gain) || math.Abs(c.Gain-1) > analogOutCalibrationMaxGainError {
		return fmt.Errorf("gain %v outside %.2f..%.2f", c.Gain, 1-analogOutCalibrationMaxGainError, 1+analogOutCalibrationMaxGainError)
	}
	if math.IsNaN(c.Offset) || math.Abs(c.Offset) > maxOffset {
		return fmt.Errorf("offset %v %s outside +-%.1f %s", c.Offset, unit, maxOffset, unit)
	}
	return nil
}

const analogOut010VCalibrationPrefix = "calibration_"

func analogOut010VCalibrationKey(address uint16) string {
	return fmt.Sprintf("%s0x%02x", analogOut010VCalibrationPrefix, address)
}

// The 4-20mA module uses the same key scheme in its own section.
var analogOut420mACalibrationKey = analogOut010VCalibrationKey

// formatCalibration renders one output as "gain:offset".
func formatCalibration(c AnalogOutCalibration) string {
	return strconv.FormatFloat(c.Gain, 'f', -1, 64) + ":" + strconv.FormatFloat(c.Offset, 'f', -1, 64)
}

// parseCalibration is the inverse of formatCalibration, without validation.
func parseCalibration(s string) (AnalogOutCalibration, error) {
	gs, offs, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok {
		return AnalogOutCalibration{}, fmt.Errorf("%q is not gain:offset", s)
	}
	g, err := strconv.ParseFloat(strings.TrimSpace(gs), 64)
	if err != nil {
		return AnalogOutCalibration{}, fmt.Errorf("gain: %w", err)
	}
	o, err := strconv.ParseFloat(strings.TrimSpace(offs), 64)
	if err != nil {
		return AnalogOutCalibration{}, fmt.Errorf("offset: %w", err)
	}
	return AnalogOutCalibration{Gain: g, Offset: o}, nil
}

// calibrationKeys calls fn for every calibration_<address> key of sec, with
// the parsed address; keys with an invalid address are logged and skipped.
func calibrationKeys(sec *ini.Section, fn func(addr uint16, key *ini.Key)) {
	for _, key := range sec.Keys() {
		name := key.Name()
		if !strings.HasPrefix(name, analogOut010VCalibrationPrefix) {
			continue
		}
		addrStr := strings.TrimPrefix(name, analogOut010VCalibrationPrefix)
		addr, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(addrStr), "0x"), 16, 7)
		if err != nil {
			log.Warnf("/etc/smartpiModules [%s] %s: invalid I2C address %q, ignored", sec.Name(), name, addrStr)
			continue
		}
		fn(uint16(addr), key)
	}
}

// saveCalibrationKey sets (value != "") or removes (value == "") one key of
// /etc/smartpiModules. Only that key is touched: the file is re-read and
// written back as it is otherwise, so a hand edit made meanwhile is kept
// (the file watcher re-reads it once more afterwards, harmlessly).
func saveCalibrationKey(section, key, value string) error {
	f, err := ini.LooseLoad(modulesConfigPath)
	if err != nil {
		return err
	}
	sec := f.Section(section)
	if value != "" {
		sec.Key(key).SetValue(value)
	} else {
		sec.DeleteKey(key)
	}
	return writeConfigFile(f, modulesConfigPath)
}

// FormatAnalogOutCalibration renders the four outputs as
// "gain:offset, gain:offset, gain:offset, gain:offset" (offset in V),
// outputs 1-4 in board order.
func FormatAnalogOutCalibration(cal [4]AnalogOutCalibration) string {
	parts := make([]string, 4)
	for i, c := range cal {
		parts[i] = formatCalibration(c)
	}
	return strings.Join(parts, ", ")
}

// ParseAnalogOutCalibration is the inverse of FormatAnalogOutCalibration.
func ParseAnalogOutCalibration(s string) ([4]AnalogOutCalibration, error) {
	var cal [4]AnalogOutCalibration
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return cal, fmt.Errorf("want 4 entries gain:offset, got %d", len(parts))
	}
	for i, part := range parts {
		c, err := parseCalibration(part)
		if err != nil {
			return cal, fmt.Errorf("output %d: %w", i+1, err)
		}
		cal[i] = c
		if err := cal[i].Validate(); err != nil {
			return cal, fmt.Errorf("output %d: %w", i+1, err)
		}
	}
	return cal, nil
}

// readAnalogOut010VCalibrations collects every calibration_<address> key of
// the [analogout010v] section. A broken entry is logged and skipped - that
// module then runs uncalibrated instead of the whole config failing.
func readAnalogOut010VCalibrations(sec *ini.Section) map[uint16][4]AnalogOutCalibration {
	cals := make(map[uint16][4]AnalogOutCalibration)
	calibrationKeys(sec, func(addr uint16, key *ini.Key) {
		cal, err := ParseAnalogOutCalibration(key.String())
		if err != nil {
			log.Warnf("/etc/smartpiModules [analogout010v] %s: %v, module runs uncalibrated", key.Name(), err)
			return
		}
		cals[addr] = cal
	})
	return cals
}

// AnalogOut010VCalibrationFor returns the calibration of the module at
// address, or the identity for all outputs if it has none.
func (p *Moduleconfig) AnalogOut010VCalibrationFor(address uint16) ([4]AnalogOutCalibration, bool) {
	if cal, ok := p.AnalogOut010VCalibration[address]; ok {
		return cal, true
	}
	return [4]AnalogOutCalibration{IdentityAnalogOutCalibration, IdentityAnalogOutCalibration, IdentityAnalogOutCalibration, IdentityAnalogOutCalibration}, false
}

// SaveAnalogOut010VCalibration stores (cal != nil) or removes (cal == nil)
// the calibration of the module at address in /etc/smartpiModules and
// applies it right away - see saveCalibrationKey.
func (p *Moduleconfig) SaveAnalogOut010VCalibration(address uint16, cal *[4]AnalogOutCalibration) error {
	if cal != nil {
		for i, c := range cal {
			if err := c.Validate(); err != nil {
				return fmt.Errorf("output %d: %w", i+1, err)
			}
		}
	}

	value := ""
	if cal != nil {
		value = FormatAnalogOutCalibration(*cal)
	}
	if err := saveCalibrationKey("analogout010v", analogOut010VCalibrationKey(address), value); err != nil {
		return err
	}

	cals := make(map[uint16][4]AnalogOutCalibration, len(p.AnalogOut010VCalibration)+1)
	for a, c := range p.AnalogOut010VCalibration {
		cals[a] = c
	}
	if cal != nil {
		cals[address] = *cal
	} else {
		delete(cals, address)
	}
	p.AnalogOut010VCalibration = cals
	return nil
}

// readAnalogOut420mACalibrations collects every calibration_<address> key
// ("gain:offset", offset in mA) of the [analogout420ma] section; broken
// entries are logged and skipped like for the 0-10V module.
func readAnalogOut420mACalibrations(sec *ini.Section) map[uint16]AnalogOutCalibration {
	cals := make(map[uint16]AnalogOutCalibration)
	calibrationKeys(sec, func(addr uint16, key *ini.Key) {
		cal, err := parseCalibration(key.String())
		if err == nil {
			err = cal.ValidateCurrent()
		}
		if err != nil {
			log.Warnf("/etc/smartpiModules [analogout420ma] %s: %v, module runs uncalibrated", key.Name(), err)
			return
		}
		cals[addr] = cal
	})
	return cals
}

// AnalogOut420mACalibrationFor returns the calibration of the 4-20mA module
// at address, or the identity if it has none.
func (p *Moduleconfig) AnalogOut420mACalibrationFor(address uint16) (AnalogOutCalibration, bool) {
	if cal, ok := p.AnalogOut420mACalibration[address]; ok {
		return cal, true
	}
	return IdentityAnalogOutCalibration, false
}

// SaveAnalogOut420mACalibration stores (cal != nil) or removes (cal == nil)
// the calibration of the 4-20mA module at address in /etc/smartpiModules
// and applies it right away.
func (p *Moduleconfig) SaveAnalogOut420mACalibration(address uint16, cal *AnalogOutCalibration) error {
	value := ""
	if cal != nil {
		if err := cal.ValidateCurrent(); err != nil {
			return err
		}
		value = formatCalibration(*cal)
	}
	if err := saveCalibrationKey("analogout420ma", analogOut420mACalibrationKey(address), value); err != nil {
		return err
	}

	cals := make(map[uint16]AnalogOutCalibration, len(p.AnalogOut420mACalibration)+1)
	for a, c := range p.AnalogOut420mACalibration {
		cals[a] = c
	}
	if cal != nil {
		cals[address] = *cal
	} else {
		delete(cals, address)
	}
	p.AnalogOut420mACalibration = cals
	return nil
}
