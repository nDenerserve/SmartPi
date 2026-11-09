package smartpiacDevice

import (
	"testing"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
)

func TestCalculatePowerFactor(t *testing.T) {
	var phase models.SmartPiPhase = models.PhaseA
	for ct := range CTTypes {
		c := &config.SmartPiACConfig{
			CTType:         map[models.SmartPiPhase]string{phase: ct},
			MeasureCurrent: map[models.SmartPiPhase]bool{phase: true},
		}
		if got := CalculatePowerFactor(c, phase, 800, 1000, -600); !near(got, 0.8) {
			t.Errorf("%s: negative VAR: got %v, want 0.8", ct, got)
		}
		if got := CalculatePowerFactor(c, phase, 800, 1000, 600); !near(got, -0.8) {
			t.Errorf("%s: positive VAR: got %v, want -0.8", ct, got)
		}
		if got := CalculatePowerFactor(c, phase, 0, 0, 0); got != 0 {
			t.Errorf("%s: no apparent power: got %v, want 0", ct, got)
		}
		c.MeasureCurrent[phase] = false
		if got := CalculatePowerFactor(c, phase, 800, 1000, -600); got != 0 {
			t.Errorf("%s: current not measured: got %v, want 0", ct, got)
		}
	}
}
