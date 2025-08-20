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

func (c ModulesController) SetDigitalout(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		user, err := serverutils.DecryptUserdataFromToken(r, conf)

		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		vars := mux.Vars(r)
		address := utils.Reverse(vars["address"])
		portstring := vars["port"]
		log.Debug("SetDigitalout: Vars: ", vars, " Address: ", address, " Portstring: ", portstring)

		addr, err := strconv.ParseUint(address, 2, 8)
		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		a := ^uint8(addr + 0xD8)

		if slices.Contains(mconf.AllowedDigitalOutUser, user.Name) {

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

		} else {
			error.Message = "User not allowed"
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

	}
}

func (c ModulesController) ReadDigitalout(mconf *config.Moduleconfig, conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		user, err := serverutils.DecryptUserdataFromToken(r, conf)

		log.Debug("ReadDigitalout: user: ", user, " mconf.AllowedDigitalOutUser: ", mconf.AllowedDigitalOutUser)

		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		vars := mux.Vars(r)
		address := vars["address"]

		log.Debug("SetDigitalout: Vars: ", vars, " Address: ", address)

		addr, err := strconv.ParseUint(address, 2, 8)

		if err != nil {
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		a := ^uint8(addr + 0xD8)

		if slices.Contains(mconf.AllowedDigitalOutUser, user.Name) {

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

		} else {
			error.Message = "User not allowed"
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

	}
}
