package modulescontrollers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	config "github.com/nDenerserve/SmartPi/smartpi/config"
	modulesRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/modules"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// ReadAnalogIn reads all four channels of an MCP3424-based analog input
// module. Each channel is reported both as a current (mA, 4-20mA scale) and
// a voltage (V, 0-10V scale) - see models.AnalogInChannel for why both are
// always returned.
func (c ModulesController) ReadAnalogIn(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var error models.Error

		// TokenVerifyMiddleWare already required the analogin scope for a
		// device token, same reasoning as the other module endpoints - see
		// digitalout.go.
		if !serverutils.IsDeviceToken(r) {
			user, err := serverutils.DecryptUserdataFromToken(r, conf)
			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
			if !slices.Contains(mconf.AllowedAnalogInUser, user.Name) {
				log.Warnf("User %s not allowed to read analogin (allowed: %v)", user.Name, mconf.AllowedAnalogInUser)
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		addressStr := vars["address"]

		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			error.Message = "Invalid address: " + err.Error() + ". Must be hex, e.g. 0x68"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.ReadAnalogIn(uint16(address), mconf)
		status.Moduleaddress = addressStr
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			panic(err)
		}
	}
}

// ReadAnalogInChannel reads a single channel (1-4) of an MCP3424-based
// analog input module - cheaper than ReadAnalogIn when only one channel is
// needed, since only that channel's settle time is paid.
func (c ModulesController) ReadAnalogInChannel(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var error models.Error

		if !serverutils.IsDeviceToken(r) {
			user, err := serverutils.DecryptUserdataFromToken(r, conf)
			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
			if !slices.Contains(mconf.AllowedAnalogInUser, user.Name) {
				log.Warnf("User %s not allowed to read analogin (allowed: %v)", user.Name, mconf.AllowedAnalogInUser)
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		addressStr := vars["address"]
		channelStr := vars["channel"]

		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			error.Message = "Invalid address: " + err.Error() + ". Must be hex, e.g. 0x68"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		// Channels are addressed 1-4 in the URL (the physical labelling on
		// the module), 0-3 internally.
		channel, err := strconv.Atoi(channelStr)
		if err != nil || channel < 1 || channel > 4 {
			error.Message = "Invalid channel: must be a number between 1 and 4"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		moduleRepo := modulesRepository.ModulesRepository{}
		status, err := moduleRepo.ReadAnalogInChannel(uint16(address), channel-1, mconf)
		status.Moduleaddress = addressStr
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			panic(err)
		}
	}
}
