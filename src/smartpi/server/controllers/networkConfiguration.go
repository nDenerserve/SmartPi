package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	linuxtoolsRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/linuxtools"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"
)

func (c Controller) ConnectionList() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		linuxtoolsRepo := linuxtoolsRepository.LinuxToolsRepository{}

		interfacelist, err := linuxtoolsRepo.ListConnections()
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(interfacelist); err != nil {
			panic(err)
		}

	}

}

func (c Controller) ScanWifi() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		linuxtoolsRepo := linuxtoolsRepository.LinuxToolsRepository{}

		wifilist, err := linuxtoolsRepo.ScanWifiNetworks()
		if err != nil {
			log.Debug(err)
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		if err := json.NewEncoder(w).Encode(wifilist); err != nil {
			panic(err)
		}

	}

}

func (c Controller) CreateConnection() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		linuxtoolsRepo := linuxtoolsRepository.LinuxToolsRepository{}

		interfacelist, err := linuxtoolsRepo.ListConnections()
		if err != nil {
			log.Debug(err)
			error.Message = err.Error()
			serverutils.RespondWithError(w, http.StatusUnauthorized, error)
			return
		}

		if err := json.NewEncoder(w).Encode(interfacelist); err != nil {
			panic(err)
		}

	}

}

func (c Controller) AddStaticIpToConnection() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		vars := mux.Vars(r)

		if vars["ipaddress"] == "" {
			error.Message = "ip is missing."
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if vars["connection"] == "" {
			error.Message = "connection is missing."
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if vars["cidrsuffix"] == "" {
			error.Message = "CIDR-Suffix is missing."
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		linuxtoolsRepo := linuxtoolsRepository.LinuxToolsRepository{}

		cidrsuffix, err := strconv.Atoi(vars["cidrsuffix"])
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		err = linuxtoolsRepo.AddIpAddressToConnection(vars["connection"], vars["ipaddress"], uint8(cidrsuffix))
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		err = linuxtoolsRepo.RestartConnection(vars["connection"])
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		interfacelist, err := linuxtoolsRepo.ListConnections()
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(interfacelist); err != nil {
			panic(err)
		}

	}
}

func (c Controller) RemoveStaticIpFromConnection() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var error models.Error

		vars := mux.Vars(r)

		if vars["ipaddress"] == "" {
			error.Message = "ip is missing."
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if vars["connection"] == "" {
			error.Message = "connection is missing."
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		if vars["cidrsuffix"] == "" {
			error.Message = "CIDR-Suffix is missing."
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		linuxtoolsRepo := linuxtoolsRepository.LinuxToolsRepository{}

		cidrsuffix, err := strconv.Atoi(vars["cidrsuffix"])
		if err != nil {
			log.Error(err)
			error.Message = err.Error()
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		err = linuxtoolsRepo.RemoveIpAddressFromConnection(vars["connection"], vars["ipaddress"], uint8(cidrsuffix))
		if err != nil {
			error.Message = err.Error()
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		err = linuxtoolsRepo.RestartConnection(vars["connection"])
		if err != nil {
			error.Message = err.Error()
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		interfacelist, err := linuxtoolsRepo.ListConnections()
		if err != nil {
			error.Message = err.Error()
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		if err := json.NewEncoder(w).Encode(interfacelist); err != nil {
			panic(err)
		}

	}
}
