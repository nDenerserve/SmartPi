package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/gommon/log"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	configRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
)

func (c Controller) ReadSmartPiConfig(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// configuration := context.Get(r, "Config")
		if err := json.NewEncoder(w).Encode(conf); err != nil {
			panic(err)
		}
	}

}

func (c Controller) WriteSmartPiConfig(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")

		var wc models.Writeconfiguration
		var errorM models.Error

		b, _ := io.ReadAll(r.Body)

		if err := json.Unmarshal(b, &wc); err != nil {
			log.Error(err)
		}
		fmt.Println(wc)
		configRepo := configRepository.ConfigRepository{}

		err := configRepo.PrepareConfig(wc, conf)
		if err != nil {
			errorM.Message = "Malformed dateformat"
			serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
			return
		}

		conf.SaveParameterToFile()
		fmt.Println(conf)

		if err := json.NewEncoder(w).Encode(conf); err != nil {
			panic(err)
		}
	}

}
