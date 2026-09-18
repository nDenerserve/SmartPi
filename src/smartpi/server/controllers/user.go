package controllers

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	userRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/user"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"

	"net/http"
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

// minPasswordLength is a deliberately modest floor - just enough to reject
// an empty or one-character password from the settings "Users" tab, not a
// full password-strength policy.
const minPasswordLength = 4

// createUserRequest is the body of POST /api/v1/users.
type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// changePasswordRequest is the body of POST /api/v1/users/{username}/password.
type changePasswordRequest struct {
	Password string `json:"password"`
}

// ListUsers returns every local Linux account the web UI's settings "Users"
// tab lets you manage - see userRepository.ListUsers for which accounts
// qualify. Route access is restricted to session tokens (RequireSessionToken),
// same as the other account-management endpoints below.
func (c Controller) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRepo := userRepository.UserRepository{}

		users, err := userRepo.ListUsers()
		if err != nil {
			var errorObject models.Error
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		serverutils.ResponseJSON(w, users)
	}
}

// CreateUser creates a new local Linux account - the settings "Users" tab's
// "create user" form. The new account can log in to the web UI immediately
// afterwards, the same as any other local account (see Login above).
func (c Controller) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorObject.Message = "Malformed request body."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		if req.Username == "" {
			errorObject.Message = "Username is missing."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}
		if len(req.Password) < minPasswordLength {
			errorObject.Message = "Password is too short."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		userRepo := userRepository.UserRepository{}
		if err := userRepo.CreateUser(req.Username, req.Password); err != nil {
			log.Error(err)
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// ChangeUserPassword sets a new password for the local Linux account named
// by the {username} route var - the settings "Users" tab's per-account
// "change password" action, used both for a user's own password and, since
// this app has no admin/non-admin distinction (any logged-in session can
// already do everything else the settings page offers), for any other
// account's password too.
func (c Controller) ChangeUserPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		username := mux.Vars(r)["username"]
		if username == "" {
			errorObject.Message = "Username is missing."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		var req changePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorObject.Message = "Malformed request body."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}
		if len(req.Password) < minPasswordLength {
			errorObject.Message = "Password is too short."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		userRepo := userRepository.UserRepository{}
		if err := userRepo.ChangePassword(username, req.Password); err != nil {
			log.Error(err)
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
