// Package roothelper is the root side of package rootcall: the actions
// smartpiserver may run as root, entered as
//
//	sudo /usr/local/bin/smartpiserver --root-helper <action> <args...>
//
// sudo only checks the command, not the arguments, and anyone who can run
// commands as the smartpi user may call this (smartpiserver, the readout
// daemons, cron jobs). So every argument is checked here, programs are run
// by absolute path, and each action can only do the one thing it is for:
// no action runs a command, writes a file or changes an account chosen by
// the caller beyond what is checked below.
package roothelper

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	osuser "os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/rootcall"
	cronRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/cron"
	linuxtoolsRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/linuxtools"
	"github.com/nDenerserve/SmartPi/smartpi/update"
)

// Programs run by the helper (absolute paths: sudo's PATH is not trusted).
var (
	useraddPath    = "/usr/sbin/useradd"
	chpasswdPath   = "/usr/sbin/chpasswd"
	aptGetPath     = "/usr/bin/apt-get"
	systemdRunPath = "/usr/bin/systemd-run"
)

// Accounts whose password may be set: regular accounts as created by
// useradd (see /etc/login.defs UID_MIN/UID_MAX) - never root or a system
// account.
const (
	minUID = 1000
	maxUID = 59999
)

// versionRE is a Debian version as used in "name=version".
var versionRE = regexp.MustCompile(`^[0-9][A-Za-z0-9.+~:-]{0,99}$`)

// Run runs one action and returns the exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if err := run(args, stdin, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no action")
	}
	action, args := args[0], args[1:]
	want := map[string]int{
		rootcall.UserAdd: 1, rootcall.Passwd: 1, rootcall.CronFTP: 2,
		rootcall.AptUpdate: 0, rootcall.AptInstall: 2, rootcall.AptUpgrade: 1,
	}
	n, ok := want[action]
	if !ok {
		return fmt.Errorf("unknown action %q", action)
	}
	if len(args) != n {
		return fmt.Errorf("%s: %d arguments expected", action, n)
	}

	switch action {
	case rootcall.UserAdd:
		if !linuxtoolsRepository.ValidUsername(args[0]) {
			return fmt.Errorf("invalid user name")
		}
		return runCmd(exec.Command(useraddPath, "-m", "-s", "/bin/bash", "--", args[0]), nil, stdout, stderr)

	case rootcall.Passwd:
		name := args[0]
		if err := checkRegularAccount(name); err != nil {
			return err
		}
		pw, err := readPassword(stdin)
		if err != nil {
			return err
		}
		return runCmd(exec.Command(chpasswdPath), strings.NewReader(name+":"+pw+"\n"), stdout, stderr)

	case rootcall.CronFTP:
		if args[0] != "on" && args[0] != "off" {
			return fmt.Errorf("cron-ftp: on or off expected")
		}
		return cronRepository.WriteFTPUpload(args[0] == "on", args[1])

	case rootcall.AptUpdate:
		cmd := exec.Command(aptGetPath, "update")
		cmd.Env = append(os.Environ(), "LC_ALL=C")
		return runCmd(cmd, nil, stdout, stderr)

	case rootcall.AptInstall:
		if !update.ValidUnit(args[0]) {
			return fmt.Errorf("invalid unit name")
		}
		target, err := checkInstallTarget(args[1])
		if err != nil {
			return err
		}
		return runCmd(exec.Command(systemdRunPath, update.JobCommand(args[0], update.InstallArgs(target))...), nil, stdout, stderr)

	case rootcall.AptUpgrade:
		if !update.ValidUnit(args[0]) {
			return fmt.Errorf("invalid unit name")
		}
		return runCmd(exec.Command(systemdRunPath, update.JobCommand(args[0], update.UpgradeArgs())...), nil, stdout, stderr)
	}
	return nil
}

func runCmd(cmd *exec.Cmd, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	return cmd.Run()
}

// checkRegularAccount accepts only existing regular accounts (not root,
// not system accounts).
func checkRegularAccount(name string) error {
	if !linuxtoolsRepository.ValidUsername(name) {
		return fmt.Errorf("invalid user name")
	}
	u, err := osuser.Lookup(name)
	if err != nil {
		return fmt.Errorf("unknown user")
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil || uid < minUID || uid > maxUID {
		return fmt.Errorf("the password of this account cannot be changed here")
	}
	return nil
}

// readPassword reads one line; a second line or a carriage return would
// let chpasswd set another account's password.
func readPassword(r io.Reader) (string, error) {
	b, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return "", err
	}
	pw := strings.TrimSuffix(string(b), "\n")
	if pw == "" || strings.ContainsAny(pw, "\r\n") {
		return "", fmt.Errorf("invalid password")
	}
	return pw, nil
}

// checkInstallTarget accepts a package from the configured repositories
// ("name" or "name=version") or an uploaded SmartPi package: a regular
// .deb file in the staging directory whose package name is smartpi or
// smartpi-*.
func checkInstallTarget(target string) (string, error) {
	if !strings.HasPrefix(target, "/") {
		name, version, hasVersion := strings.Cut(target, "=")
		if !update.ValidPackageName(name) || (hasVersion && !versionRE.MatchString(version)) {
			return "", fmt.Errorf("invalid package")
		}
		return target, nil
	}
	dir := config.NewSmartPiConfig().UpdateStagingDir
	if dir == "" {
		dir = update.DefaultStagingDir
	}
	clean := filepath.Clean(target)
	if filepath.Dir(clean) != filepath.Clean(dir) || filepath.Ext(clean) != ".deb" {
		return "", fmt.Errorf("the package must be an uploaded .deb file")
	}
	fi, err := os.Lstat(clean)
	if err != nil || !fi.Mode().IsRegular() {
		return "", fmt.Errorf("the package must be an uploaded .deb file")
	}
	name, _, err := update.InspectDeb(clean)
	if err != nil {
		return "", err
	}
	if !update.IsAllowedDebPackage(name) {
		return "", fmt.Errorf("only SmartPi packages can be installed from a file")
	}
	return clean, nil
}
