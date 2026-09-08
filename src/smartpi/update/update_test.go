package update

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestCommandInC_ForcesCLocale(t *testing.T) {
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	t.Setenv("LANGUAGE", "de_DE:de")

	cmd := commandInC("echo", "hi")

	var lcAll string
	for _, kv := range cmd.Env {
		if strings.HasPrefix(kv, "LANGUAGE=") {
			t.Fatalf("LANGUAGE should have been stripped, got %q", kv)
		}
		if name, value, ok := strings.Cut(kv, "="); ok && name == "LC_ALL" {
			lcAll = value
		}
	}
	if lcAll != "C" {
		t.Fatalf("LC_ALL = %q, want \"C\"", lcAll)
	}
}

func TestIsAllowedDebPackage(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"smartpi", true},
		{"smartpi-modules", true},
		{"smartpi-", true},
		{"smartpiextra", false},
		{"nodered", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsAllowedDebPackage(tt.name); got != tt.want {
			t.Errorf("IsAllowedDebPackage(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestValidPackageName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"smartpi", true},
		{"grafana", true},
		{"lib32-foo.bar+baz", true},
		{"", false},
		{"-leadingdash", false},
		{"Uppercase", false},
		{"has space", false},
		{"semi;colon", false},
	}
	for _, tt := range tests {
		if got := ValidPackageName(tt.name); got != tt.want {
			t.Errorf("ValidPackageName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseControlFields(t *testing.T) {
	// dpkg-deb -f <archive> Package Version echoes the field name on every
	// line ("Package: smartpi"), not just the bare value - unlike, say,
	// dpkg-query -W -f='${Version}'.
	out := "Package: smartpi\nVersion: 2026.09.04-trixie\n"

	got := parseControlFields(out)
	want := map[string]string{"Package": "smartpi", "Version": "2026.09.04-trixie"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseSearchOutput(t *testing.T) {
	out := "smartpi - SmartPi energy monitor\n" +
		"smartpi-modules - SmartPi optional hardware modules\n" +
		"nodered - Node-RED\n"

	got := parseSearchOutput(out)
	want := []PackageSummary{
		{Name: "smartpi", Description: "SmartPi energy monitor"},
		{Name: "smartpi-modules", Description: "SmartPi optional hardware modules"},
		{Name: "nodered", Description: "Node-RED"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseSearchOutput_Empty(t *testing.T) {
	if got := parseSearchOutput(""); len(got) != 0 {
		t.Fatalf("got %+v, want empty", got)
	}
}

func TestParsePolicyOutput(t *testing.T) {
	out := `smartpi:
  Installed: 1.2.3
  Candidate: 1.2.4
  Version table:
     1.2.4 500
        500 https://repo.enerserve.eu stable/main armhf Packages
 *** 1.2.3 100
        100 /var/lib/dpkg/status
`
	got := parsePolicyOutput("smartpi", out)
	want := PackageInfo{Name: "smartpi", InstalledVersion: "1.2.3", CandidateVersion: "1.2.4"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParsePolicyOutput_NotInstalled(t *testing.T) {
	out := `newpackage:
  Installed: (none)
  Candidate: 2.0.0
  Version table:
     2.0.0 500
        500 https://repo.enerserve.eu stable/main armhf Packages
`
	got := parsePolicyOutput("newpackage", out)
	want := PackageInfo{Name: "newpackage", CandidateVersion: "2.0.0"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseUpgradableOutput(t *testing.T) {
	out := `Listing... Done
smartpi/stable 1.2.4 armhf [upgradable from: 1.2.3]
grafana/stable 11.0.0 armhf [upgradable from: 10.9.0]
`
	got := parseUpgradableOutput(out)
	want := []UpgradablePackage{
		{Name: "smartpi", CurrentVersion: "1.2.3", NewVersion: "1.2.4"},
		{Name: "grafana", CurrentVersion: "10.9.0", NewVersion: "11.0.0"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseUpgradableOutput_NoneUpgradable(t *testing.T) {
	if got := parseUpgradableOutput("Listing... Done\n"); len(got) != 0 {
		t.Fatalf("got %+v, want empty", got)
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"simple", "'simple'"},
		{"has space", "'has space'"},
		{"pkg=1.2.3", "'pkg=1.2.3'"},
		{"it's", `'it'\''s'`},
	}
	for _, tt := range tests {
		if got := shellQuote(tt.in); got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAptGetScript_ValidShell(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	script := aptGetScript([]string{"install", "-y", "smartpi's-package"})
	cmd := exec.Command("sh", "-n")
	cmd.Stdin = strings.NewReader(script)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated script is not valid sh: %v\n%s\n---\n%s", err, out, script)
	}
}

func TestAptGetScript_QuotesArgsAndRunsBothBranches(t *testing.T) {
	script := aptGetScript([]string{"install", "-y", "pkg=1.2.3"})

	want := "apt-get 'install' '-y' 'pkg=1.2.3'"
	if n := strings.Count(script, want); n != 2 {
		t.Fatalf("expected the quoted apt-get command twice (tmpfs and non-tmpfs branch), got %d occurrences in:\n%s", n, script)
	}
	if !strings.Contains(script, varTmpPath) || !strings.Contains(script, varTmpEnlargedSize) {
		t.Fatalf("script does not remount %s to %s:\n%s", varTmpPath, varTmpEnlargedSize, script)
	}
}
