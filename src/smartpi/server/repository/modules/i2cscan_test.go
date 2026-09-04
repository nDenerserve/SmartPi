package modulesRepository

import (
	"fmt"
	"strings"
	"testing"
)

// buildI2CDetectRow mimics the fixed-width row format i2cdetect prints:
// "%02x:" followed by one 3-character cell per column (fewer than 16 for
// row 0x70's 8-column tail). An empty string is a not-probed address
// (printed as blank), "--" is no device, "UU" is in-use, anything else is
// treated as a detected address byte.
func buildI2CDetectRow(rowBase int, cells []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%02x:", rowBase)
	for _, c := range cells {
		if c == "" {
			b.WriteString("   ")
		} else {
			fmt.Fprintf(&b, " %s", c)
		}
	}
	return b.String()
}

func TestParseI2CDetectOutput_DetectsKnownChipRangesAndInUseAddresses(t *testing.T) {
	lines := []string{
		"     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f",
		buildI2CDetectRow(0x00, []string{"", "", "", "--", "--", "--", "--", "--", "--", "--", "--", "--", "--", "--", "--", "--"}),
		buildI2CDetectRow(0x20, []string{"--", "--", "--", "--", "--", "--", "--", "27", "--", "--", "--", "--", "--", "--", "--", "--"}),
		buildI2CDetectRow(0x30, []string{"--", "--", "--", "--", "--", "--", "--", "--", "38", "--", "--", "--", "--", "--", "--", "--"}),
		buildI2CDetectRow(0x50, []string{"--", "--", "--", "--", "--", "--", "--", "--", "--", "--", "UU", "--", "--", "--", "--", "--"}),
		buildI2CDetectRow(0x60, []string{"60", "--", "--", "--", "--", "--", "--", "--", "68", "--", "--", "--", "--", "--", "--", "--"}),
		buildI2CDetectRow(0x70, []string{"--", "--", "--", "--", "--", "--", "--", "77"}),
	}
	output := strings.Join(lines, "\n")

	entries := parseI2CDetectOutput(output)

	// 0x00-0x02 are reserved (blank cells, never printed as "--") and must
	// not show up as detected; 0x5a is "UU" - claimed by a kernel driver,
	// so it's reported as in_use with no chip hint; 0x77 exercises the
	// short last row without a hint, since it falls outside every known
	// chip's address range.
	want := map[string]struct{ status, hint string }{
		"0x27": {"detected", "MCP23017"},
		"0x38": {"detected", "ADE7878"},
		"0x5a": {"in_use", ""},
		"0x60": {"detected", "MCP4725"},
		"0x68": {"detected", "MCP3424"},
		"0x77": {"detected", ""},
	}

	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for _, e := range entries {
		w, ok := want[e.Address]
		if !ok {
			t.Fatalf("unexpected address %s in result: %+v", e.Address, entries)
		}
		if e.Status != w.status || e.Hint != w.hint {
			t.Fatalf("%s: got {%s %s}, want {%s %s}", e.Address, e.Status, e.Hint, w.status, w.hint)
		}
	}
}

func TestI2CBusNumber(t *testing.T) {
	tests := []struct {
		device  string
		want    string
		wantErr bool
	}{
		{"/dev/i2c-1", "1", false},
		{"/dev/i2c-21", "21", false},
		{"/dev/ttyS0", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		got, err := i2cBusNumber(tt.device)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("device %q: expected an error, got none", tt.device)
			}
			continue
		}
		if err != nil {
			t.Fatalf("device %q: unexpected error: %v", tt.device, err)
		}
		if got != tt.want {
			t.Fatalf("device %q: got %q, want %q", tt.device, got, tt.want)
		}
	}
}
