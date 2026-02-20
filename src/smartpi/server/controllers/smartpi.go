package controllers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
	"github.com/nDenerserve/SmartPi/utils"

	"github.com/nDenerserve/SmartPi/smartpi/config"
	configRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/config"
	smartpiRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/smartpi"
)

func (c Controller) SmartPiLivePower(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		format := "json"
		if vars["format"] != "" {
			format = vars["format"]
		}

		smartpiRepo := smartpiRepository.SmartPiRepository{}

		measurement := smartpiRepo.LivePower(conf)

		if format == "xml" {
			// XML output of request
			type response struct {
				models.TLivePower
			}
			if err := xml.NewEncoder(w).Encode(response{measurement}); err != nil {
				panic(err)
			}
		} else {
			// JSON output of request
			if err := json.NewEncoder(w).Encode(measurement); err != nil {
				panic(err)
			}
		}
	}
}

func (c Controller) SmartPiLiveValues(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		valueId := "all"
		if vars["valueId"] != "" {
			valueId = vars["valueId"]
		}

		phaseId := "all"
		if vars["phaseId"] != "" {
			phaseId = vars["phaseId"]
		}

		format := "json"
		if vars["format"] != "" {
			format = vars["format"]
		}

		smartpiRepo := smartpiRepository.SmartPiRepository{}

		measurement := smartpiRepo.LiveValues(phaseId, valueId, conf)

		if format == "xml" {
			// XML output of request
			type response struct {
				models.TMeasurement
			}
			if err := xml.NewEncoder(w).Encode(response{measurement}); err != nil {
				panic(err)
			}
		} else {
			// JSON output of request
			if err := json.NewEncoder(w).Encode(measurement); err != nil {
				panic(err)
			}
		}

	}

}

func (c Controller) SmartPiChartdata(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var errorM models.Error
		var err error
		var barchartdata models.Progressdatalist

		if r.Method == "OPTIONS" {
			log.Debug("Preflight")
		}
		log.Debug("Progressdata invoked")

		// format := "json"
		vars := mux.Vars(r)
		// from := vars["fromDate"]
		// to := vars["toDate"]
		valueId := vars["value"]
		// valueId := vars["valueId"]
		// format = vars["format"]

		starttime := utils.StartOfMonth(time.Now())
		stoptime := time.Now()

		aggregate := "24h"

		if vars["aggregate"] != "" {
			aggregate = vars["aggregate"]
		}

		if (vars["starttime"] != "") && (vars["stoptime"] != "") {
			starttime, err = dateparse.ParseIn(vars["starttime"], time.UTC)
			if err != nil {
				errorM.Message = "Malformed dateformat"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
			stoptime, err = dateparse.ParseIn(vars["stoptime"], time.UTC)
			if err != nil {
				errorM.Message = "Malformed dateformat"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
		}

		if stoptime.Before(starttime) {
			errorM.Message = "Starttime after Stoptime"
			serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
			return
		}

		smartpiRepo := smartpiRepository.SmartPiRepository{}

		// API request for single values
		if valueId == "current" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "mean", []string{"I1", "I2", "I3", "I4"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "voltage" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "mean", []string{"U1", "U2", "U3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "power" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "mean", []string{"P1", "P2", "P3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "cosphi" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "mean", []string{"CosPhi1", "CosPhi2", "CosPhi3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "frequency" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "mean", []string{"F1", "F2", "F3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "energyconsumed" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "sum", []string{"Ec1", "Ec2", "Ec3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "energyproduced" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "sum", []string{"Ep1", "Ep2", "Ep3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "energy" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "sum", []string{"Ec1", "Ec2", "Ec3", "Ep1", "Ep2", "Ep3"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "energybalancedconsumed" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "sum", []string{"bEc"}, conf)
			if err != nil {
				log.Error(err)
			}
		} else if valueId == "energybalancedproduced" {
			barchartdata, err = smartpiRepo.BarChart(starttime, stoptime, aggregate, "sum", []string{"bEp"}, conf)
			if err != nil {
				log.Error(err)
			}

		}

		// valueList := strings.Split(value, ",")

		if err := json.NewEncoder(w).Encode(barchartdata); err != nil {
			log.Error(err)
		}
	}

}

func (c Controller) SmartPiProgressdata(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var errorM models.Error
		var err error
		var progressdata models.Progressdatalist

		if r.Method == "OPTIONS" {
			log.Debug("Preflight")
		}
		log.Debug("Progressdata invoked")

		// format := "json"
		vars := mux.Vars(r)
		// from := vars["fromDate"]
		// to := vars["toDate"]
		value := vars["value"]
		// valueId := vars["valueId"]
		// format = vars["format"]

		starttime := time.Now().Add(-time.Hour * 24)
		stoptime := time.Now()

		aggregate := "300s"

		if vars["aggregate"] != "" {
			aggregate = vars["aggregate"]
		}

		if (vars["starttime"] != "") && (vars["stoptime"] != "") {
			starttime, err = dateparse.ParseIn(vars["starttime"], time.UTC)
			if err != nil {
				errorM.Message = "Malformed dateformat"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
			stoptime, err = dateparse.ParseIn(vars["stoptime"], time.UTC)
			if err != nil {
				errorM.Message = "Malformed dateformat"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
		}

		if stoptime.Before(starttime) {
			errorM.Message = "Starttime after Stoptime"
			serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
			return
		}

		valueList := strings.Split(value, ",")

		smartpiRepo := smartpiRepository.SmartPiRepository{}
		progressdata, err = smartpiRepo.Progressdata(starttime, stoptime, aggregate, valueList, conf)
		if err != nil {
			log.Error(err)
		}

		if err := json.NewEncoder(w).Encode(progressdata); err != nil {
			log.Error(err)
		}
	}

}

func (c Controller) ReadSmartPiACConfig(conf *config.SmartPiACConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// configuration := context.Get(r, "Config")
		fmt.Println(conf)
		if err := json.NewEncoder(w).Encode(conf); err != nil {
			panic(err)
		}
	}

}

func (c Controller) WriteSmartPiACConfig(conf *config.SmartPiACConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")

		var wc models.Writeconfiguration
		var errorM models.Error
		// spew.Dump(r.Body)
		fmt.Println(r.Body)

		b, _ := io.ReadAll(r.Body)

		fmt.Println("WriteConfig")
		// fmt.Println(b)

		// json.NewDecoder(r.Body).Decode(&wc)

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

func (c Controller) SmartPiCsvExport(conf *config.SmartPiConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var errorM models.Error
		var err error
		var csv string

		vars := mux.Vars(r)

		daterange := 1
		stop := time.Now()
		start := stop.Add(time.Duration(24*(-1)) * time.Hour)
		aggregate := ""

		if vars["range"] != "" {
			daterange, err = strconv.Atoi(vars["range"])
			if err != nil {
				errorM.Message = "Malformed rangeformat"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
			start = stop.Add(time.Duration(daterange*24*(-1)) * time.Hour)
			year, month, day := start.Date()
			start = time.Date(year, month, day, 0, 0, 0, 0, start.Location())
		}

		if vars["start"] != "" {
			start, err = utils.ParseTime(utils.DateFormats, vars["start"])
			if err != nil {
				errorM.Message = "Malformed dateformat startdate"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
		}

		if vars["stop"] != "" {
			stop, err = utils.ParseTime(utils.DateFormats, vars["stop"])
			if err != nil {
				errorM.Message = "Malformed dateformat stopdate"
				serverutils.RespondWithError(w, http.StatusBadRequest, errorM)
				return
			}
		}

		smartpiRepo := smartpiRepository.SmartPiRepository{}

		if vars["aggregate"] != "" {
			aggregate = vars["aggregate"]
			log.Debug("Export CSV-Data from " + start.String() + " to " + stop.String() + ". Aggregate: " + aggregate)
			log.Debug("Please wait. It may take a while...")
			// csv, _ = exportCSV(smartpiconfig, start, stop, *decimalpointPtr, *aggregatePtr)
			csv, err = smartpiRepo.ExportCSV(conf, start, stop, conf.CSVdecimalpoint, aggregate)
			if err != nil {
				errorM.Message = "Error creating CSV"
				serverutils.RespondWithError(w, http.StatusInternalServerError, errorM)
				return
			}
		} else {
			aggregate = vars["aggregate"]
			log.Debug("Export CSV-Data from " + start.String() + " to " + stop.String())
			log.Debug("Please wait. It may take a while...")
			// csv, _ = exportCSV(smartpiconfig, start, stop, *decimalpointPtr)
			csv, err = smartpiRepo.ExportCSV(conf, start, stop, conf.CSVdecimalpoint)
			if err != nil {
				errorM.Message = "Error creating CSV"
				serverutils.RespondWithError(w, http.StatusInternalServerError, errorM)
				return
			}
		}

		w.Header().Set("Content-Disposition", "attachment; filename=export.csv")
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Transfer-Encoding", "chunked")
		fmt.Fprintf(w, csv)

	}

}
