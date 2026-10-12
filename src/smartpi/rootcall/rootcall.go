// Package rootcall runs the few actions smartpiserver needs root for.
//
// smartpiserver runs unprivileged as the smartpi user. Instead of sudo
// rules for apt-get, useradd, chpasswd and friends (which let anyone acting
// as smartpi run them with any arguments, i.e. become root), it re-executes
// itself as root for exactly these actions:
//
//	sudo /usr/local/bin/smartpiserver --root-helper <action> <args...>
//
// The root side (package roothelper) checks every argument; sudo only
// checks the command (see etc/sudoers.d/smartpi).
package rootcall

import "os/exec"

// ServerPath is where smartpiserver is installed; the sudo rules name
// exactly this path.
const ServerPath = "/usr/local/bin/smartpiserver"

// Flag selects the root helper mode of smartpiserver.
const Flag = "--root-helper"

// Actions of the root helper.
const (
	UserAdd    = "useradd"     // useradd <name>
	Passwd     = "passwd"      // passwd <name>, the password on stdin
	CronFTP    = "cron-ftp"    // cron-ftp on|off <hours>
	AptUpdate  = "apt-update"  // apt-update
	AptInstall = "apt-install" // apt-install <unit> <.deb in the staging dir | name[=version]>
	AptUpgrade = "apt-upgrade" // apt-upgrade <unit>
)

// Command returns the command that runs action as root.
func Command(action string, args ...string) *exec.Cmd {
	return exec.Command("sudo", append([]string{ServerPath, Flag, action}, args...)...)
}
