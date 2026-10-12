package utils

import "testing"

func TestFluxParams(t *testing.T) {
	for _, s := range []string{"300s", "15m", "1h", "1h30m", "1d", "1mo"} {
		if err := FluxDuration(s); err != nil {
			t.Errorf("duration %q rejected: %v", s, err)
		}
	}
	for _, s := range []string{"", "1h)", `1h) |> yield(name:"a")`, "1 h", "h", "1h\n"} {
		if FluxDuration(s) == nil {
			t.Errorf("duration %q accepted", s)
		}
	}
	for _, s := range []string{"P1", "CosPhi3", "I4", "all"} {
		if err := FluxField(s); err != nil {
			t.Errorf("field %q rejected", s)
		}
	}
	for _, s := range []string{`P1"`, `P1" or true or "`, "P 1", ""} {
		if FluxField(s) == nil {
			t.Errorf("field %q accepted", s)
		}
	}
	if FluxFunc("mean") != nil || FluxFunc("mean, x: 1") == nil {
		t.Error("function check")
	}
}
