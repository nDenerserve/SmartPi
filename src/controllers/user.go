package controllers

import (
	"encoding/json"

	"github.com/nDenerserve/SmartPi/models"
	userRepository "github.com/nDenerserve/SmartPi/repository/user"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"

	"net/http"
)

func (c Controller) Login() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var user models.User
		var credentials models.Credentials
		var jwt models.JWT
		var error models.Error
		// spew.Dump(r)
		json.NewDecoder(r.Body).Decode(&credentials)

		if credentials.Username == "" {
			error.Message = "Username is missing."
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if credentials.Password == "" {
			error.Message = "Password is missing."
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		userRepo := userRepository.UserRepository{}

		user, err := userRepo.ReadUser(credentials.Username, credentials.Password, user)

		if err != nil {
			error.Message = err.Error()
			utils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		// spew.Dump(user)

		token, err := utils.GenerateToken(user)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		jwt.Token = token

		utils.ResponseJSON(w, jwt)
	}
}
