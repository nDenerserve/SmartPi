package controllers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	valuesRepository "github.com/nDenerserve/SmartPi/repository/values"
	log "github.com/sirupsen/logrus"
)

func (c Controller) SmartPiDCLiveValues(config *config.DCconfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		format := "json"
		vars := mux.Vars(r)
		format = vars["format"]

		valuesRepo := valuesRepository.ValuesRepository{}
		livevalues, err := valuesRepo.DCLivevalues(config)

		if err != nil {
			log.Fatal(err)
		}

		if format == "xml" {
			// XML output of request
			type response struct {
				models.Livedata
			}
			if err := xml.NewEncoder(w).Encode(response{livevalues}); err != nil {
				panic(err)
			}
		} else {
			// JSON output of request
			if err := json.NewEncoder(w).Encode(livevalues); err != nil {
				panic(err)
			}
		}

	}
}
