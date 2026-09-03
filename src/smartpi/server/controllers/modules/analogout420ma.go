package modulescontrollers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	config "github.com/nDenerserve/SmartPi/smartpi/config"
	modulesRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/modules"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// SetAnalogOut420mA sets the 4-20mA output current value for a specific MCP4725 module
// Address is provided as hex (e.g., "0x60") or decimal (e.g., "96") and current is a float value between 4.0 and 20.0
func (c ModulesController) SetAnalogOut420mA(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var error models.Error

		// TokenVerifyMiddleWare already required the analogout scope for a
		// device token - a stronger, per-token grant made explicitly by an
		// operator, and device tokens have no OS user to check here anyway.
		// The username allowlist below only applies to session tokens, same
		// as before device tokens existed.
		if !serverutils.IsDeviceToken(r) {
			user, err := serverutils.DecryptUserdataFromToken(r, conf)
			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}

			log.Debugf("User: %s, Allowed users: %v", user.Name, mconf.AllowedAnalogOut420mAUser)
			if !slices.Contains(mconf.AllowedAnalogOut420mAUser, user.Name) {
				log.Warnf("User %s not allowed to set analogout420ma (allowed: %v)", user.Name, mconf.AllowedAnalogOut420mAUser)
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		addressStr := vars["address"]
		currentStr := vars["current"]
		log.Debug("SetAnalogOut420mA: Address: ", addressStr, " Current: ", currentStr)

		// Parse address as hex or decimal
		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			error.Message = "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		// Parse current value
		current, err := strconv.ParseFloat(currentStr, 64)
		if err != nil {
			error.Message = "Invalid current value: " + err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		// Validate current range (4-20mA)
		if current < 4.0 || current > 20.0 {
			error.Message = "Current must be between 4.0 and 20.0 mA"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		log.Debug("User is allowed, calling repository")
		moduleRepo := modulesRepository.ModulesRepository{}

		status, err := moduleRepo.SetAnalogOut420mA(uint16(address), current, mconf)
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

// ReadAnalogOut420mA reads the current status of a specific MCP4725 module
func (c ModulesController) ReadAnalogOut420mA(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var error models.Error

		// See SetAnalogOut420mA above: a device token already carries the
		// analogout scope required by TokenVerifyMiddleWare, so the username
		// allowlist below only applies to session tokens.
		if !serverutils.IsDeviceToken(r) {
			user, err := serverutils.DecryptUserdataFromToken(r, conf)
			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}

			log.Debugf("User: %s, Allowed users: %v", user.Name, mconf.AllowedAnalogOut420mAUser)
			if !slices.Contains(mconf.AllowedAnalogOut420mAUser, user.Name) {
				log.Warnf("User %s not allowed to read analogout420ma (allowed: %v)", user.Name, mconf.AllowedAnalogOut420mAUser)
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		addressStr := vars["address"]

		log.Debug("ReadAnalogOut420mA: Address: ", addressStr)

		// Parse address as hex or decimal
		address, err := parseHexOrDecimal(addressStr)
		if err != nil {
			error.Message = "Invalid address: " + err.Error() + ". Must be hex (e.g., 0x60) or decimal (e.g., 96)"
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		log.Debug("User is allowed, calling repository")
		moduleRepo := modulesRepository.ModulesRepository{}

		status, err := moduleRepo.ReadAnalogOut420mAStatus(uint16(address), mconf)
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

// parseHexOrDecimal parses a string as hex (with 0x prefix) or decimal
func parseHexOrDecimal(s string) (uint8, error) {
	// Check if it's hex format
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 8)
		return uint8(val), err
	}
	// Otherwise parse as decimal
	val, err := strconv.ParseUint(s, 10, 8)
	return uint8(val), err
}
