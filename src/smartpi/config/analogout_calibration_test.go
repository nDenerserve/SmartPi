package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func TestAnalogOutCalibrationRoundTrip(t *testing.T) {
	cal := [4]AnalogOutCalibration{{0.998625, 0.00882}, {1.0034, -0.02114}, {1.00385, 0.01781}, {0.997738, 0.01281}}
	got, err := ParseAnalogOutCalibration(FormatAnalogOutCalibration(cal))
	if err != nil {
		t.Fatal(err)
	}
	if got != cal {
		t.Fatalf("round trip: got %v, want %v", got, cal)
	}
}

func TestParseAnalogOutCalibrationRejects(t *testing.T) {
	for _, s := range []string{
		"",
		"1:0, 1:0, 1:0",        // 3 entries
		"1:0, 1:0, 1:0, 1",     // no offset
		"1:0, 1:0, 1:0, x:0",   // not a number
		"1:0, 1:0, 1:0, 1.5:0", // gain off by 50%
		"1:0, 1:0, 1:0, 1:2",   // 2V offset
		"1:0, 1:0, 1:0, NaN:0", // NaN
	} {
		if _, err := ParseAnalogOutCalibration(s); err == nil {
			t.Errorf("%q accepted", s)
		}
	}
}

func TestReadAnalogOut010VCalibrations(t *testing.T) {
	f, err := ini.Load([]byte(`[analogout010v]
allowed_user = root
calibration_0x60 = 0.99:0.01, 1:0, 1:0, 1.01:-0.02
calibration_0x61 = broken
calibration_zz = 1:0, 1:0, 1:0, 1:0
`))
	if err != nil {
		t.Fatal(err)
	}
	cals := readAnalogOut010VCalibrations(f.Section("analogout010v"))
	if len(cals) != 1 {
		t.Fatalf("want only the valid 0x60 entry, got %v", cals)
	}
	if c := cals[0x60]; c[0] != (AnalogOutCalibration{0.99, 0.01}) || c[3] != (AnalogOutCalibration{1.01, -0.02}) {
		t.Fatalf("0x60 parsed wrong: %v", c)
	}
}

func TestSaveAnalogOut010VCalibration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpiModules")
	if err := os.WriteFile(path, []byte("[base]\ni2c_device = /dev/i2c-1\n\n[analogout010v]\nallowed_user = root\n"), 0640); err != nil {
		t.Fatal(err)
	}
	old := modulesConfigPath
	modulesConfigPath = path
	defer func() { modulesConfigPath = old }()

	p := &Moduleconfig{}
	cal := [4]AnalogOutCalibration{{1, 0.01}, {1, 0}, {1, 0}, {0.99, 0}}
	if err := p.SaveAnalogOut010VCalibration(0x60, &cal); err != nil {
		t.Fatal(err)
	}
	if got, ok := p.AnalogOut010VCalibrationFor(0x60); !ok || got != cal {
		t.Fatalf("not applied in memory: %v %v", got, ok)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "calibration_0x60") || !strings.Contains(string(data), "i2c_device") || !strings.Contains(string(data), "allowed_user") {
		t.Fatalf("file content wrong:\n%s", data)
	}

	bad := cal
	bad[1].Gain = 2
	if err := p.SaveAnalogOut010VCalibration(0x60, &bad); err == nil {
		t.Fatal("implausible calibration saved")
	}

	if err := p.SaveAnalogOut010VCalibration(0x60, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.AnalogOut010VCalibrationFor(0x60); ok {
		t.Fatal("calibration still active after delete")
	}
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), "calibration_0x60") {
		t.Fatalf("key not removed:\n%s", data)
	}
}

func TestReadAnalogOut420mACalibrations(t *testing.T) {
	f, err := ini.Load([]byte(`[analogout420ma]
allowed_user = root
calibration_0x60 = 1.02:-0.05
calibration_0x61 = 1:3
calibration_0x62 = broken
`))
	if err != nil {
		t.Fatal(err)
	}
	cals := readAnalogOut420mACalibrations(f.Section("analogout420ma"))
	if len(cals) != 1 || cals[0x60] != (AnalogOutCalibration{1.02, -0.05}) {
		t.Fatalf("want only 0x60 = 1.02:-0.05 (3mA offset is implausible), got %v", cals)
	}
}

func TestSaveAnalogOut420mACalibration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpiModules")
	if err := os.WriteFile(path, []byte("[analogout420ma]\nallowed_user = root\n\n[analogout010v]\ncalibration_0x60 = 1:0, 1:0, 1:0, 1:0\n"), 0640); err != nil {
		t.Fatal(err)
	}
	old := modulesConfigPath
	modulesConfigPath = path
	defer func() { modulesConfigPath = old }()

	p := &Moduleconfig{}
	cal := AnalogOutCalibration{Gain: 0.995, Offset: 0.08}
	if err := p.SaveAnalogOut420mACalibration(0x60, &cal); err != nil {
		t.Fatal(err)
	}
	if got, ok := p.AnalogOut420mACalibrationFor(0x60); !ok || got != cal {
		t.Fatalf("not applied in memory: %v %v", got, ok)
	}
	f, err := ini.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if v := f.Section("analogout420ma").Key("calibration_0x60").String(); v != "0.995:0.08" {
		t.Fatalf("stored %q", v)
	}
	// The 0-10V module's key with the same address is a different entry.
	if !f.Section("analogout010v").HasKey("calibration_0x60") || !f.Section("analogout420ma").HasKey("allowed_user") {
		t.Fatalf("other keys lost:\n%v", f)
	}

	if err := p.SaveAnalogOut420mACalibration(0x60, &AnalogOutCalibration{Gain: 1, Offset: 2}); err == nil {
		t.Fatal("2mA offset saved")
	}

	if err := p.SaveAnalogOut420mACalibration(0x60, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.AnalogOut420mACalibrationFor(0x60); ok {
		t.Fatal("calibration still active after delete")
	}
	f, _ = ini.Load(path)
	if f.Section("analogout420ma").HasKey("calibration_0x60") || !f.Section("analogout010v").HasKey("calibration_0x60") {
		t.Fatalf("wrong key removed:\n%v", f)
	}
}
