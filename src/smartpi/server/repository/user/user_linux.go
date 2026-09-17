package userRepository

import (
	"github.com/nDenerserve/SmartPi/models"
	linuxtoolsRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/linuxtools"
)

// ListUsers returns the local Linux accounts usable to log in to the web UI.
func (u UserRepository) ListUsers() ([]linuxtoolsRepository.LinuxUser, error) {
	return linuxtoolsRepository.ListUsers()
}

// CreateUser creates a new local Linux account with the given password.
func (u UserRepository) CreateUser(username string, password string) error {
	return linuxtoolsRepository.CreateUser(username, password)
}

// ChangePassword sets a new password for an existing local Linux account.
func (u UserRepository) ChangePassword(username string, password string) error {
	_, err := linuxtoolsRepository.ChangePassword(username, password)
	return err
}

func (u UserRepository) ReadUser(username string, password string, user models.User) (models.User, error) {

	if linuxtoolsRepository.ValidateUser(username, password) {
		user.Role = []string{}
		user.Name = username
		user.Password = password
		user.Role, _ = linuxtoolsRepository.GetGroupsFromUser(user.Name)
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
