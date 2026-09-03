package modulescontrollers

import "testing"

func TestParseModuleAddressJumperEncoding(t *testing.T) {
	// All three jumpers on ("111") is the example from the module's
	// documentation: it must resolve to bus address 0x20.
	got, err := parseModuleAddress("111")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x20 {
		t.Fatalf("jumpers \"111\" = 0x%02X, want 0x20", got)
	}
}

func TestParseModuleAddressHexBypass(t *testing.T) {
	cases := []struct {
		in   string
		want uint8
	}{
		{"0x20", 0x20},
		{"0X20", 0x20},
		{"0x00", 0x00},
		{"0xff", 0xFF},
		{"0xFF", 0xFF},
		{"0x1a", 0x1A},
	}
	for _, c := range cases {
		got, err := parseModuleAddress(c.in)
		if err != nil {
			t.Errorf("parseModuleAddress(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseModuleAddress(%q) = 0x%02X, want 0x%02X", c.in, got, c.want)
		}
	}
}

func TestParseModuleAddressInvalid(t *testing.T) {
	cases := []string{
		"",
		"2",         // not a binary digit
		"0x",        // hex prefix with nothing after it
		"0xzz",      // not hex digits
		"0x100",     // out of range for uint8
		"111111111", // too many jumper bits for uint8 after +0xD8 is meaningless, but base parse itself should still succeed or fail consistently
	}
	for _, c := range cases {
		if _, err := parseModuleAddress(c); err == nil {
			t.Errorf("parseModuleAddress(%q) succeeded, want an error", c)
		}
	}
}

func TestParseModuleAddressHexDoesNotGoThroughJumperTransform(t *testing.T) {
	// A hex address is the bus address itself - it must not also be run
	// through the one's-complement jumper transform, which would silently
	// address the wrong module.
	got, err := parseModuleAddress("0x20")
	if err != nil {
		t.Fatal(err)
	}
	transformed := ^uint8(0x20 + 0xD8)
	if got == transformed {
		t.Fatalf("0x20 accidentally matched the jumper transform's output 0x%02X - hex bypass is not being taken", transformed)
	}
	if got != 0x20 {
		t.Fatalf("got 0x%02X, want 0x20 unchanged", got)
	}
}
