package controllers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	linuxtoolsRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/linuxtools"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	log "github.com/sirupsen/logrus"
)

// parseIPv4AndCidr validates that ipaddress is a dotted-quad IPv4 address and
// cidrsuffix is a valid prefix length (0-32), before either reaches nmcli.
// nmcli is invoked via exec.Command with an argument array (no shell), so
// this is not about injection - it is about rejecting garbage early instead
// of letting a malformed value (e.g. cidrsuffix "999", which silently wraps
// to 231 once cast to uint8) reach NetworkManager as a confusing failure, or
// worse, an accepted-but-wrong config on the connection actively serving the
// caller's own request.
func parseIPv4AndCidr(ipaddress, cidrsuffix string) (string, uint8, error) {
	ip := net.ParseIP(ipaddress)
	if ip == nil || ip.To4() == nil {
		return "", 0, fmt.Errorf("invalid IPv4 address: %q", ipaddress)
	}
	suffix, err := strconv.Atoi(cidrsuffix)
	if err != nil || suffix < 0 || suffix > 32 {
		return "", 0, fmt.Errorf("invalid CIDR suffix: %q", cidrsuffix)
	}
	return ip.String(), uint8(suffix), nil
}

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

		ipaddress, cidrsuffix, err := parseIPv4AndCidr(vars["ipaddress"], vars["cidrsuffix"])
		if err != nil {
			error.Message = err.Error()
			log.Errorf("AddStaticIpToConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		err = linuxtoolsRepo.AddIpAddressToConnection(vars["connection"], ipaddress, cidrsuffix)
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

		ipaddress, cidrsuffix, err := parseIPv4AndCidr(vars["ipaddress"], vars["cidrsuffix"])
		if err != nil {
			error.Message = err.Error()
			log.Errorf("RemoveStaticIpFromConnection: " + error.Message)
			serverutils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		err = linuxtoolsRepo.RemoveIpAddressFromConnection(vars["connection"], ipaddress, cidrsuffix)
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
