package linuxtoolsRepository

import (
	"os"
	osuser "os/user"
	"path/filepath"
	"testing"
)

// The account name from the login form must never reach a shell: a name
// with line breaks used to pass the PAM check line by line and then run the
// remaining lines in `sh -c "groups "+name`.
func TestGetGroupsFromUserRunsNoShell(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "marker")
	name := "nobody\ntouch " + marker
	GetGroupsFromUser(name) // an unknown account: an error, nothing else
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the account name was executed as a command")
	}
}

func TestGetGroupsFromUserCurrentAccount(t *testing.T) {
	me, err := osuser.Current()
	if err != nil {
		t.Skip(err)
	}
	groups, err := GetGroupsFromUser(me.Username)
	if err != nil || len(groups) == 0 {
		t.Fatalf("groups of %s: %v %v", me.Username, groups, err)
	}
}

func TestLineBreaksRejected(t *testing.T) {
	// rejected before sudo is run, so these never reach PAM or chpasswd
	if ValidateUser("bob\nbobpw\nid", "x") {
		t.Error("line break in the login name accepted")
	}
	if ValidateUser("bob", "pw\nroot") {
		t.Error("line break in the password accepted")
	}
	if _, err := ChangePassword("bob", "x\nroot:pw"); err == nil {
		t.Error("line break in a new password accepted (would change root's password)")
	}
	if _, err := ChangePassword("bob\nroot", "pw"); err == nil {
		t.Error("invalid account name accepted")
	}
}
