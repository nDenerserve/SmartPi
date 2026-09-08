package update

import (
	"os"
	"os/exec"
	"path/filepath"
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
	script := aptGetScript([]string{"install", "-y", "smartpi's-package"}, "/var/smartpi/update-logs/unit.exitcode")
	cmd := exec.Command("sh", "-n")
	cmd.Stdin = strings.NewReader(script)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated script is not valid sh: %v\n%s\n---\n%s", err, out, script)
	}
}

func TestAptGetScript_QuotesArgsAndRecordsExitCode(t *testing.T) {
	script := aptGetScript([]string{"install", "-y", "pkg=1.2.3"}, "/var/smartpi/update-logs/unit.exitcode")

	want := "apt-get 'install' '-y' 'pkg=1.2.3'"
	if n := strings.Count(script, want); n != 1 {
		t.Fatalf("expected the quoted apt-get command exactly once, got %d occurrences in:\n%s", n, script)
	}
	if !strings.Contains(script, varTmpPath) || !strings.Contains(script, varTmpEnlargedSize) {
		t.Fatalf("script does not remount %s to %s:\n%s", varTmpPath, varTmpEnlargedSize, script)
	}
	if !strings.Contains(script, "echo $ec > '/var/smartpi/update-logs/unit.exitcode'") {
		t.Fatalf("script does not record apt-get's exit code:\n%s", script)
	}
}

// TestAptGetScript_RecordsRealExitCode actually runs the generated script
// (against /bin/true and /bin/false standing in for apt-get, via a PATH
// override) and checks the exit-code file it writes matches - this is what
// jobOutcome relies on to tell a genuinely finished job apart from one
// systemd hasn't started running yet, see jobOutcome's doc comment.
func TestAptGetScript_RecordsRealExitCode(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	for _, tt := range []struct {
		binary   string
		wantCode string
	}{
		{"true", "0"},
		{"false", "1"},
	} {
		t.Run(tt.binary, func(t *testing.T) {
			dir := t.TempDir()
			// aptGetScript always invokes "apt-get" by name; put a same-named
			// stand-in on PATH ahead of the real one instead of teaching the
			// function to run something else.
			aptGetStub := filepath.Join(dir, "apt-get")
			if err := os.Symlink(mustLookPath(t, tt.binary), aptGetStub); err != nil {
				t.Fatal(err)
			}
			exitCodePath := filepath.Join(dir, "unit.exitcode")

			cmd := exec.Command("sh", "-c", aptGetScript(nil, exitCodePath))
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"))
			out, _ := cmd.CombinedOutput()

			got, err := os.ReadFile(exitCodePath)
			if err != nil {
				t.Fatalf("exit-code file was not written: %v\nscript output:\n%s", err, out)
			}
			if strings.TrimSpace(string(got)) != tt.wantCode {
				t.Fatalf("exit-code file = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

// TestUnitState_UnknownUnitReportsUnknown guards against the exact bug
// jobOutcome's doc comment describes: `systemctl show` on a unit it has
// never heard of is indistinguishable, from ActiveState alone, from a unit
// that already finished successfully and was collected (both report
// ActiveState "inactive"), or from one that was requested but hasn't
// started yet. It must not be read as "succeeded" - nor as "running",
// which would leave a genuinely stale job (e.g. one started by an older
// smartpiserver build that predates aptGetScript's exit-code file) stuck
// looking active forever.
func TestUnitState_UnknownUnitReportsUnknown(t *testing.T) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}
	state, exitCode := unitState("smartpi-update-test-unit-that-was-never-started")
	if state != "unknown" {
		t.Fatalf(`unitState on a never-started unit = %q (exitCode %d), want "unknown"`, state, exitCode)
	}
}

func mustLookPath(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
	return path
}
