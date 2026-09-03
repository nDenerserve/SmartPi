package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/devicetoken"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
)

// createDeviceTokenRequest is the body of POST /api/v1/tokens.
type createDeviceTokenRequest struct {
	Label  string   `json:"label"`
	Scopes []string `json:"scopes"`
}

// ListDeviceTokens returns every device token, secret included - the web UI
// this is served to is already behind a session login, which is the whole
// reason a token can be shown again instead of only once at creation.
func (c Controller) ListDeviceTokens(tokens *devicetoken.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverutils.ResponseJSON(w, tokens.List())
	}
}

// CreateDeviceToken issues a new device token with the requested label and
// scopes. Route access itself is already restricted to session tokens by
// serverutils.RequireSessionToken, so the createdBy attribution below is
// always a logged-in human, never another device token.
func (c Controller) CreateDeviceToken(tokens *devicetoken.Store, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		var req createDeviceTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorObject.Message = "Malformed request body."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}

		if req.Label == "" {
			errorObject.Message = "Label is missing."
			serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
			return
		}
		for _, scope := range req.Scopes {
			if !devicetoken.ValidScope(scope) {
				errorObject.Message = "Unknown scope: " + scope
				serverutils.RespondWithError(w, http.StatusBadRequest, errorObject)
				return
			}
		}

		// RequireSessionToken already guaranteed this is a session token, so
		// the only error DecryptUserdataFromToken could still return here is
		// an internal inconsistency, not an auth failure - the request was
		// already authenticated once to reach this handler at all.
		user, err := serverutils.DecryptUserdataFromToken(r, conf)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		token, err := tokens.Create(req.Label, req.Scopes, user.Name)
		if err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		w.WriteHeader(http.StatusCreated)
		serverutils.ResponseJSON(w, token)
	}
}

// DeleteDeviceToken revokes a token by id. This is the actual revocation
// mechanism the whole design rests on: TokenVerifyMiddleWare checks the
// store on every request, so a token deleted here stops working on the very
// next request against it, without needing to expire or wait for anything.
func (c Controller) DeleteDeviceToken(tokens *devicetoken.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		id := mux.Vars(r)["id"]
		if err := tokens.Delete(id); err != nil {
			errorObject.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, errorObject)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
