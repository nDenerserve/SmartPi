package controllers

import (
	"encoding/json"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	userRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/user"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/davecgh/go-spew/spew"
)

func (c Controller) Login(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var user models.User
		var credentials models.Credentials
		var jwt models.JWT
		var error models.Error
		// spew.Dump(r.Body)
		// fmt.Println(r.Body)
		json.NewDecoder(r.Body).Decode(&credentials)

		if credentials.Username == "" {
			error.Message = "Username is missing."
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if credentials.Password == "" {
			error.Message = "Password is missing."
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		userRepo := userRepository.UserRepository{}

		user, err := userRepo.ReadUser(credentials.Username, credentials.Password, user)

		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		spew.Dump(user)

		token, err := serverutils.GenerateToken(user, conf)

		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		w.WriteHeader(http.StatusOK)
		jwt.Token = token

		serverutils.ResponseJSON(w, jwt)
	}
}
