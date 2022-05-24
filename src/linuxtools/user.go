/*
    Copyright (C) Jens Ramhorst
  	This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/
/*
File: etcpasswd.go
Description: Interface to read a standard Unix passwd, shadow and group file-format
*/

package linuxtools

import (
	"errors"
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
	_, err := exec.Command("/bin/sh", "-c", `echo "echo '`+user+`:`+newpassword+`' | chpasswd" >> /tmp/changepass.sh`).Output()
	if err != nil {
		return false, err
	}
	_, err = exec.Command("/bin/sh", "-c", `chmod +x /tmp/changepass.sh`).Output()
	if err != nil {
		return false, err
	}
	_, err = exec.Command("/bin/sh", "-c", `sudo /tmp/changepass.sh`).Output()
	if err != nil {
		return false, err
	}
	_, err = exec.Command("/bin/sh", "-c", `rm /tmp/changepass.sh`).Output()
	if err != nil {
		return false, err
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
