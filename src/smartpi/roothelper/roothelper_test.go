package roothelper

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nDenerserve/SmartPi/smartpi/rootcall"
)

// fakePrograms replaces the programs the helper runs with a script that
// records its arguments, so nothing is changed on the test machine.
func fakePrograms(t *testing.T) (logFile string) {
	dir := t.TempDir()
	logFile = filepath.Join(dir, "calls")
	script := filepath.Join(dir, "fake")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$0 $*\" >> "+logFile+"\ncat >> "+logFile+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	old := []string{useraddPath, chpasswdPath, aptGetPath, systemdRunPath}
	useraddPath, chpasswdPath, aptGetPath, systemdRunPath = script, script, script, script
	t.Cleanup(func() { useraddPath, chpasswdPath, aptGetPath, systemdRunPath = old[0], old[1], old[2], old[3] })
	return logFile
}

func call(stdin string, args ...string) (int, string) {
	var out bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &out, &out)
	return code, out.String()
}

func TestRejectedBeforeRunning(t *testing.T) {
	logFile := fakePrograms(t)
	for _, c := range []struct {
		stdin string
		args  []string
	}{
		{"", nil},
		{"", []string{"sh", "-c", "id"}},                      // no such action
		{"", []string{rootcall.UserAdd}},                      // missing argument
		{"", []string{rootcall.UserAdd, "-o"}},                // option instead of a name
		{"", []string{rootcall.UserAdd, "bob", "-u", "0"}},    // extra arguments
		{"x\n", []string{rootcall.Passwd, "root"}},            // root
		{"x\n", []string{rootcall.Passwd, "daemon"}},          // system account
		{"x\nroot:pw\n", []string{rootcall.Passwd, "nobody"}}, // nobody is not a regular account either
		{"", []string{rootcall.CronFTP, "maybe", "1"}},        // not on/off
		{"", []string{rootcall.CronFTP, "on", "1 * * * root id"}},
		{"", []string{rootcall.AptInstall, "smartpi-update-1", "/etc/passwd"}},
		{"", []string{rootcall.AptInstall, "smartpi-update-1", "/var/smartpi/update-uploads/../../../tmp/x.deb"}},
		{"", []string{rootcall.AptInstall, "smartpi-update-1", "-o=APT::Update::Pre-Invoke::=sh"}},
		{"", []string{rootcall.AptInstall, "smartpi-update-1", "pkg=1.0;id"}},
		{"", []string{rootcall.AptInstall, "x;id", "smartpi"}},      // unit name
		{"", []string{rootcall.AptUpgrade, "--property=User=root"}}, // unit name
	} {
		if code, out := call(c.stdin, c.args...); code == 0 {
			t.Errorf("%q accepted: %s", c.args, out)
		}
	}
	if b, _ := os.ReadFile(logFile); len(b) > 0 {
		t.Errorf("a program was run for a rejected call:\n%s", b)
	}
}

func TestAllowedCalls(t *testing.T) {
	logFile := fakePrograms(t)
	if code, out := call("", rootcall.UserAdd, "bob"); code != 0 {
		t.Fatalf("useradd: %s", out)
	}
	if code, out := call("", rootcall.AptInstall, "smartpi-update-1", "smartpi=2026.10.09"); code != 0 {
		t.Fatalf("apt-install: %s", out)
	}
	if code, out := call("", rootcall.AptUpgrade, "smartpi-update-2"); code != 0 {
		t.Fatalf("apt-upgrade: %s", out)
	}
	b, _ := os.ReadFile(logFile)
	calls := string(b)
	for _, want := range []string{"-m -s /bin/bash -- bob", "--unit=smartpi-update-1", "'smartpi=2026.10.09'", "--unit=smartpi-update-2", "dist-upgrade"} {
		if !strings.Contains(calls, want) {
			t.Errorf("calls do not contain %q:\n%s", want, calls)
		}
	}
}

func TestReadPassword(t *testing.T) {
	if pw, err := readPassword(strings.NewReader("secret\n")); err != nil || pw != "secret" {
		t.Errorf("got %q %v", pw, err)
	}
	for _, in := range []string{"", "\n", "a\nroot:pw\n", "a\rb\n"} {
		if _, err := readPassword(strings.NewReader(in)); err == nil {
			t.Errorf("%q accepted", in)
		}
	}
}
