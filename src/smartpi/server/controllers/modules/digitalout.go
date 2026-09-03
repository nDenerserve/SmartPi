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
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// parseModuleAddress resolves the bus address of a digital-out module from
// the {address} path segment. A 0x-prefixed value (case-insensitive) is used
// directly as the module's hexadecimal I2C bus address. Otherwise the value
// is read as the on/off pattern of the module's three jumper switches (e.g.
// "111" for all three on) - a binary number that the module's hardware
// wiring maps onto its actual bus address via one's complement, which is
// what the +0xD8 and bitwise-not below reproduce.
func parseModuleAddress(address string) (uint8, error) {
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		addr, err := strconv.ParseUint(address[2:], 16, 8)
		if err != nil {
			return 0, err
		}
		return uint8(addr), nil
	}

	jumpers, err := strconv.ParseUint(address, 2, 8)
	if err != nil {
		return 0, err
	}
	return ^uint8(jumpers + 0xD8), nil
}

func (c ModulesController) SetDigitalout(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		// TokenVerifyMiddleWare already required the digitalout scope for a
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
			if !slices.Contains(mconf.AllowedDigitalOutUser, user.Name) {
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		address := vars["address"]
		portstring := vars["port"]
		// The jumper encoding's bit order only makes sense reversed (see
		// parseModuleAddress); a 0x-prefixed hex address is used exactly as
		// given, reversing it would turn it into a different address.
		if !strings.HasPrefix(address, "0x") && !strings.HasPrefix(address, "0X") {
			address = utils.Reverse(address)
		}
		log.Debug("SetDigitalout: Vars: ", vars, " Address: ", address, " Portstring: ", portstring)

		a, err := parseModuleAddress(address)
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		portstring = strings.TrimSpace(portstring)
		if portstring[len(portstring)-1:] == ";" {
			portstring = portstring[:len(portstring)-1]
		}
		ports := strings.Split(portstring, ";")
		portmap := make(map[int]bool)
		for _, e := range ports {
			parts := strings.Split(e, "=")
			k, err := strconv.Atoi(parts[0])
			v, err := strconv.ParseBool(parts[1])
			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusBadRequest, error)
				return
			}
			portmap[k] = v
		}
		moduleRepo := modulesRepository.ModulesRepository{}

		status, err := moduleRepo.SetDigitalOut(uint16(a), portmap, mconf)
		status.Moduleaddress = address
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

func (c ModulesController) ReadDigitalout(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		// See SetDigitalout above: a device token already carries the
		// digitalout scope required by TokenVerifyMiddleWare, so the
		// username allowlist below only applies to session tokens.
		if !serverutils.IsDeviceToken(r) {
			user, err := serverutils.DecryptUserdataFromToken(r, conf)

			log.Debug("ReadDigitalout: user: ", user, " mconf.AllowedDigitalOutUser: ", mconf.AllowedDigitalOutUser)

			if err != nil {
				error.Message = err.Error()
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
			if !slices.Contains(mconf.AllowedDigitalOutUser, user.Name) {
				error.Message = "User not allowed"
				serverutils.RespondWithError(w, http.StatusUnauthorized, error)
				return
			}
		}

		vars := mux.Vars(r)
		address := vars["address"]

		log.Debug("SetDigitalout: Vars: ", vars, " Address: ", address)

		a, err := parseModuleAddress(address)
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		moduleRepo := modulesRepository.ModulesRepository{}

		status, err := moduleRepo.ReadDigitalOutStatus(uint16(a), mconf)
		status.Moduleaddress = address
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
