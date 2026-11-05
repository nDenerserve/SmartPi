package smartpiacDevice

import (
	"math"
	"testing"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
)

func ctConfig(ct string, primary int, secondaryMA float64) *config.SmartPiACConfig {
	return &config.SmartPiACConfig{
		CTType:                 map[models.SmartPiPhase]string{models.PhaseA: ct},
		CTTypePrimaryCurrent:   map[models.SmartPiPhase]int{models.PhaseA: primary},
		CTTypeRogowskiVoltage:  map[models.SmartPiPhase]int{models.PhaseA: 100},
		CTTypeSecondaryCurrent: map[models.SmartPiPhase]float64{models.PhaseA: secondaryMA},
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9*math.Max(1, math.Abs(b)) }

func factors(t *testing.T, c *config.SmartPiACConfig) (ccf, pcf float64) {
	t.Helper()
	ccf, ok1 := currentClampFactor(c, models.PhaseA)
	pcf, ok2 := powerScaleFactor(c, models.PhaseA)
	if !ok1 || !ok2 {
		t.Fatalf("%s %dA/%vmA rejected", c.CTType[models.PhaseA], c.CTTypePrimaryCurrent[models.PhaseA], c.CTTypeSecondaryCurrent[models.PhaseA])
	}
	return ccf, pcf
}

// X/mA must reproduce the fixed current-output types it generalises.
func TestCustomCTMatchesFixedTypes(t *testing.T) {
	for _, c := range []struct {
		ref string
		a   int
		ma  float64
	}{
		{"YHDC_SCT013", 100, 50},           // 100A/50mA
		{"YHDC_SCT4333QL", 100, 100.0 / 6}, // 1/60 A per 100A
	} {
		rc, rp := factors(t, ctConfig(c.ref, 0, 0))
		cc, cp := factors(t, ctConfig("X/mA", c.a, c.ma))
		if !near(cc, rc) || !near(cp, rp) {
			t.Errorf("X/mA %dA/%vmA: ccf %v pcf %v, %s has %v %v", c.a, c.ma, cc, cp, c.ref, rc, rp)
		}
	}
	r, x := CTTypes["YHDC_SCT013"], CTTypes["X/mA"]
	if r.CurrentResistor != x.CurrentResistor || r.CurrentRmsOffset != x.CurrentRmsOffset || r.PowerCorrectionFactor != x.PowerCorrectionFactor {
		t.Errorf("X/mA factors differ from YHDC_SCT013: %+v vs %+v", x, r)
	}
}

// Current and power scale with A per mA: for the same secondary current
// 200A/50mA reads twice the current of 100A/50mA and of 200A/100mA.
func TestCustomCTScaling(t *testing.T) {
	c1, p1 := factors(t, ctConfig("X/mA", 100, 50))
	c2, p2 := factors(t, ctConfig("X/mA", 200, 50))
	c3, p3 := factors(t, ctConfig("X/mA", 200, 100))
	// current = k / ccf, power = k / pcf
	if !near(c1/c2, 2) || !near(c3/c2, 2) {
		t.Errorf("clamp factors %v %v %v", c1, c2, c3)
	}
	if !near(p1/p2, 2) || !near(p3/p2, 2) {
		t.Errorf("power scales %v %v %v", p1, p2, p3)
	}
}

func TestCustomCTInvalid(t *testing.T) {
	for _, c := range []struct {
		a  int
		ma float64
	}{{0, 50}, {100, 0}, {100, -5}, {100, 150}} {
		cfg := ctConfig("X/mA", c.a, c.ma)
		if _, ok := currentClampFactor(cfg, models.PhaseA); ok {
			t.Errorf("%dA/%vmA accepted for current", c.a, c.ma)
		}
		if _, ok := powerScaleFactor(cfg, models.PhaseA); ok {
			t.Errorf("%dA/%vmA accepted for power", c.a, c.ma)
		}
	}
}

// The refactoring into currentClampFactor/powerScaleFactor must not change
// the existing types.
func TestExistingCTFactorsUnchanged(t *testing.T) {
	for ct, want := range map[string][2]float64{
		"YHDC_SCT013":    {0.05, 1},
		"YHDC_SCT4333QL": {1.0 / 60, 1.0 / 3},
		"400A/033V":      {0.08325, 0.5},
		"600A/100mA":     {0.1 / 6, 1.0 / 3},
		"X/1A":           {1.0 / 2.5, 200.0 / 250}, // primary 250A
		"X/5A":           {5.0 / 2.5, 200.0 / 250},
		"ROGOWSKI":       {1.0 / 100, 100 / 1635.43},
	} {
		c, p := factors(t, ctConfig(ct, 250, 0))
		if !near(c, want[0]) || !near(p, want[1]) {
			t.Errorf("%s: ccf %v pcf %v, want %v", ct, c, p, want)
		}
	}
}
