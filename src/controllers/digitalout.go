package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	modulesRepository "github.com/nDenerserve/SmartPi/repository/modules"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

func (c Controller) SetDigitalout(conf *config.Moduleconfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		user, err := utils.DecryptUserdataFromToken(r)

		if err != nil {
			error.Message = err.Error()
			utils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		vars := mux.Vars(r)
		address := utils.Reverse(vars["address"])
		portstring := vars["port"]

		addr, err := strconv.ParseUint(address, 2, 8)
		if err != nil {
			error.Message = err.Error()
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		a := ^uint8(addr + 0xD8)

		if utils.CheckAllowedUser(user.Role, conf.AllowedDigitalOutUser) {

			log.Debug("Address: " + fmt.Sprintf("%02X", a) + "    Port: " + portstring)
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
					utils.RespondWithError(w, http.StatusBadRequest, error)
					return
				}
				portmap[k] = v
			}
			moduleRepo := modulesRepository.ModulesRepository{}

			status, err := moduleRepo.SetDigitalOut(uint16(a), portmap, conf)
			status.Moduleaddress = address
			if err != nil {
				error.Message = err.Error()
				utils.RespondWithError(w, http.StatusInternalServerError, error)
				return
			}

			if err := json.NewEncoder(w).Encode(status); err != nil {
				panic(err)
			}

		} else {
			error.Message = "User not allowed"
			utils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

	}
}

func (c Controller) ReadDigitalout(conf *config.Moduleconfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		user, err := utils.DecryptUserdataFromToken(r)

		if err != nil {
			error.Message = err.Error()
			utils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		vars := mux.Vars(r)
		address := vars["address"]

		addr, err := strconv.ParseUint(address, 2, 8)

		if err != nil {
			error.Message = err.Error()
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		a := ^uint8(addr + 0xD8)

		if utils.CheckAllowedUser(user.Role, conf.AllowedDigitalOutUser) {

			moduleRepo := modulesRepository.ModulesRepository{}

			status, err := moduleRepo.ReadDigitalOutStatus(uint16(a), conf)
			status.Moduleaddress = address
			if err != nil {
				error.Message = err.Error()
				utils.RespondWithError(w, http.StatusInternalServerError, error)
				return
			}

			if err := json.NewEncoder(w).Encode(status); err != nil {
				panic(err)
			}

		} else {
			error.Message = "User not allowed"
			utils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

	}
}
