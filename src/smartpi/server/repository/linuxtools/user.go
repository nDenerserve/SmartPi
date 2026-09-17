package linuxtoolsRepository

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/msteinert/pam"
)

type LinuxUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Uid      int    `json:"uid"`
	Gid      int    `json:"gid"`
	Comments string `json:"comments"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
}

type LinuxGroup struct {
	Groupname string   `json:"groupname"`
	Password  string   `json:"password"`
	Gid       int      `json:"gid"`
	Users     []string `json:"users"`
}

// minHumanUid/maxHumanUid bound the UID range Debian/Raspbian assigns to
// regular (human-created) accounts - below that are system/service accounts
// (root, daemon, the smartpi service account is itself typically >=1000 so
// it is included), at/above that is the nobody/nogroup range. This is the
// same convention useradd itself uses (see /etc/login.defs UID_MIN/UID_MAX).
const minHumanUid = 1000
const maxHumanUid = 59999

// loginShells excludes accounts that exist for a service rather than a
// person (e.g. an account created with --shell /usr/sbin/nologin) from the
// user list shown in the web UI.
var nonLoginShells = map[string]bool{
	"/usr/sbin/nologin": true,
	"/sbin/nologin":     true,
	"/bin/false":        true,
	"/usr/bin/false":    true,
}

// usernameRe mirrors the pattern useradd itself enforces (see NAME_REGEX in
// /etc/adduser.conf / man 8 useradd): a lowercase letter or underscore,
// followed by lowercase letters, digits, underscores or hyphens.
var usernameRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

// ValidUsername reports whether username is an acceptable Linux account name
// for CreateUser - checked here, before ever shelling out to useradd, so a
// malformed name is rejected as a normal validation error rather than as a
// useradd failure.
func ValidUsername(username string) bool {
	return usernameRe.MatchString(username)
}

// ListUsers returns every local, human-usable account from /etc/passwd (see
// minHumanUid/maxHumanUid and nonLoginShells above for what "human-usable"
// excludes). Passwords are never read from /etc/shadow or included here -
// the web UI only ever offers to overwrite a password, never to see one.
func ListUsers() ([]LinuxUser, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	users := []LinuxUser{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 7 {
			continue
		}
		uid, err := strconv.Atoi(fields[2])
		if err != nil || uid < minHumanUid || uid > maxHumanUid {
			continue
		}
		if nonLoginShells[fields[6]] {
			continue
		}
		gid, _ := strconv.Atoi(fields[3])
		users = append(users, LinuxUser{
			Username: fields[0],
			Uid:      uid,
			Gid:      gid,
			Comments: fields[4],
			Home:     fields[5],
			Shell:    fields[6],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser creates a new local Linux account with a home directory and a
// login shell, then sets its initial password. smartpiserver runs
// unprivileged (see etc/systemd/system/smartpiserver.service), so both steps
// shell out through sudo - see etc/sudoers.d/smartpi-users for the rule that
// allows exactly this without a password prompt.
func CreateUser(username string, password string) error {
	if !ValidUsername(username) {
		return errors.New("Invalid username.")
	}

	out, err := exec.Command("sudo", "useradd", "-m", "-s", "/bin/bash", username).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}

	if _, err := ChangePassword(username, password); err != nil {
		return err
	}

	return nil
}

func ChangePassword(user string, newpassword string) (bool, error) {
	cmd := exec.Command("sudo", "chpasswd")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false, err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, user+":"+newpassword+"\n")
	}()

	// chpasswd prints nothing on success - its exit status is the only
	// signal of success or failure, so unlike a typical CombinedOutput()
	// check, out is only used to enrich the error message, never compared
	// against an expected value.
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}

	return true, nil

}

func GetGroupsFromUser(user string) ([]string, error) {
	out, err := exec.Command("/bin/sh", "-c", `groups `+user).Output()
	if err != nil {
		return nil, err
	}
	tmpstring := string(out)[strings.Index(string(out), ":")+1 : len(string(out))]
	groups := strings.Fields(tmpstring)
	return groups, nil
}

// smartpiServerPath is where smartpiserver installs itself (see readme.md's
// install steps) - ValidateUser re-execs this same binary as root via sudo
// to run CheckPassword below, and etc/sudoers.d/smartpi-users scopes that
// sudo rule to exactly this path and argument.
const smartpiServerPath = "/usr/local/bin/smartpiserver"

// ValidateUser checks username/password for the Login endpoint - both the
// original "smartpi" account and any account created via the settings
// "Users" tab. It re-execs smartpiserver itself as root
// ("smartpiserver --pam-check", see runPamCheck in smartpi/server/server.go)
// rather than calling CheckPassword directly, because smartpiserver.service
// itself runs unprivileged (see etc/systemd/system/smartpiserver.service):
// PAM's unix_chkpwd helper refuses to check any account's password other
// than the caller's own real uid for a non-root caller, so an in-process
// call here would only ever succeed for the "smartpi" account, silently
// rejecting a correct password for every other account. The username and
// password are passed over stdin, not argv, so they never show up in `ps`.
func ValidateUser(username string, password string) bool {
	cmd := exec.Command("sudo", smartpiServerPath, "--pam-check")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, username+"\n"+password+"\n")
	}()

	return cmd.Run() == nil
}

// CheckPassword runs the actual PAM check in the current process - safe to
// call only as root (see ValidateUser above for why), which is exactly what
// runPamCheck in smartpi/server/server.go is re-exec'd as via sudo.
func CheckPassword(username string, password string) bool {
	return pamAuth("passwd", username, password) == nil
}

func pamAuth(serviceName, userName, passwd string) error {
	t, err := pam.StartFunc(serviceName, userName, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("Unrecognized PAM message style")
	})

	if err != nil {
		return err
	}

	if err = t.Authenticate(0); err != nil {
		return err
	}
	if err = t.AcctMgmt(0); err != nil {
		return err
	}
	return nil
}
