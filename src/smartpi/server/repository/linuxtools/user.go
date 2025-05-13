package linuxtoolsRepository

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
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

func GetUsers() {

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

	out, err := cmd.CombinedOutput()
	mess := fmt.Sprintf("%s", out)
	if err != nil {
		return false, err
	}

	if mess != "0" {
		return false, errors.New(mess)
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

func ValidateUser(username string, password string) bool {

	err := pamAuth("passwd", username, password)
	if err != nil {
		return false
	} else {
		return true
	}
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
