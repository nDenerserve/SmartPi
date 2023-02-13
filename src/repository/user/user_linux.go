package userRepository

import (
	"github.com/nDenerserve/SmartPi/linuxtools"
	"github.com/nDenerserve/SmartPi/models"
)

func (u UserRepository) ReadUser(username string, password string, user models.User) (models.User, error) {

	if linuxtools.ValidateUser(username, password) {
		user.Role = []string{}
		user.Name = username
		user.Password = password
		user.Role, _ = linuxtools.GetGroupsFromUser(user.Name)
		user.Exist = true
		return user, nil
	} else {
		// u.Role = []string{}
		// u.Name = ""
		// u.Password = "nopassword"
		// u.Role = append(u.Role, "nobody")
		// u.Exist = false
		var errorObject models.Error
		errorObject.Message = "User does not exist or password wrong."
		return user, errorObject
	}
}
