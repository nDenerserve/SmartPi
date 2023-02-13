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

package smartpi

import (
	"github.com/nDenerserve/SmartPi/linuxtools"
)

// "path/filepath"
// "os"

type User struct {
	Name     string
	Password string
	Role     []string
	Exist    bool
}

func (u *User) ReadUser(username string, password string) {

	if linuxtools.ValidateUser(username, password) {
		u.Role = []string{}
		u.Name = username
		u.Password = password
		u.Role, _ = linuxtools.GetGroupsFromUser(u.Name)
		u.Exist = true
		return
	} else {
		u.Role = []string{}
		u.Name = username
		u.Password = "nopassword"
		u.Role = append(u.Role, "nobody")
		u.Exist = false
		return
	}
}

func NewUser() *User {
	t := new(User)
	return t
}
