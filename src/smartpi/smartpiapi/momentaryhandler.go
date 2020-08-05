/*
    Copyright (C) Jens Ramhorst
  	This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/
/*
File: apihandlersmomentary.go
Description: Handels API requests
*/

package smartpiapi

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/src/smartpi"
	"github.com/nDenerserve/SmartPi/src/smartpi/network"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func ServeMomentaryValues(w http.ResponseWriter, r *http.Request) {
	var phases = []*smartpi.TPhase{}
	var datasets = []*smartpi.TDataset{}
	var tempVal *smartpi.TValue
	var tempPhase *smartpi.TPhase
	var tempDataset *smartpi.TDataset

	w.Header().Set("Content-Type", "application/json")
	// w.Header().Add("Access-Control-Allow-Origin", "*")

	format := "json"
	vars := mux.Vars(r)
	phaseId := vars["phaseId"]
	valueId := vars["valueId"]
	format = vars["format"]

	config := smartpi.NewConfig()
	file, err := os.Open(config.SharedDir + "/" + config.SharedFile)
	smartpi.Checklog(err)
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.Read()
	smartpi.Checklog(err)

	t := time.Now()

	// API request for single values
	if valueId == "current" || valueId == "voltage" || valueId == "power" || valueId == "cosphi" || valueId == "frequency" {

		// request for one of the phases
		if (phaseId == "1") || (phaseId == "2") || (phaseId == "3") || (phaseId == "4") {

			var val float64
			var err error
			var info string
			values := []*smartpi.TValue{}
			id, err := strconv.Atoi(phaseId)
			smartpi.Checklog(err)
			if valueId == "current" && id < 5 {
				val, err = strconv.ParseFloat(records[id], 32)
			} else if valueId == "voltage" && id < 4 {
				val, err = strconv.ParseFloat(records[id+4], 32)
			} else if valueId == "power" && id < 4 {
				val, err = strconv.ParseFloat(records[id+7], 32)
			} else if valueId == "cosphi" && id < 4 {
				val, err = strconv.ParseFloat(records[id+10], 32)
			} else if valueId == "frequency" && id < 4 {
				val, err = strconv.ParseFloat(records[id+13], 32)
			} else {
				val = 0.0
				info = "error: not allowed"
			}

			if err != nil {
				val = 0.0
				info = "warning: parse error. set value to 0.0"
			}

			if math.IsNaN(val) {
				val = 0.0
				info = "warning: value NaN. set value to 0.0"
			}

			if math.IsInf(val, 0) {
				val = 0.0
				info = "warning: value infinity. set value to 0.0"
			}

			tempVal = new(smartpi.TValue)
			if valueId == "current" {
				tempVal.Type = "current"
				tempVal.Unity = "A"
			} else if valueId == "voltage" {
				tempVal.Type = "voltage"
				tempVal.Unity = "V"
			} else if valueId == "power" {
				tempVal.Type = "power"
				tempVal.Unity = "W"
			} else if valueId == "cosphi" {
				tempVal.Type = "cosphi"
				tempVal.Unity = ""
			} else if valueId == "frequency" {
				tempVal.Type = "frequency"
				tempVal.Unity = "Hz"
			}

			tempVal.Info = info
			tempVal.Data = float32(val)

			values = append(values, tempVal)

			tempPhase = new(smartpi.TPhase)
			tempPhase.Phase, _ = strconv.Atoi(phaseId)
			tempPhase.Name = "phase " + phaseId
			tempPhase.MaxGUICurrent = config.GUIMaxCurrent[smartpi.PhaseNameFromNumber(phaseId)]
			tempPhase.Values = values

			phases = append(phases, tempPhase)

			// request for all phases
		} else if phaseId == "all" {

			for i := 0; i <= 3; i++ {

				var val float64
				var err error
				var info string
				values := []*smartpi.TValue{}
				if valueId == "current" {
					val, err = strconv.ParseFloat(records[i+1], 32)
				} else if valueId == "voltage" && i < 3 {
					val, err = strconv.ParseFloat(records[i+5], 32)
				} else if valueId == "power" && i < 3 {
					val, err = strconv.ParseFloat(records[i+8], 32)
				} else if valueId == "cosphi" && i < 3 {
					val, err = strconv.ParseFloat(records[i+11], 32)
				} else if valueId == "frequency" && i < 3 {
					val, err = strconv.ParseFloat(records[i+14], 32)
				}

				if err != nil {
					val = 0.0
					info = "warning: parse error. set value to 0.0"
				}

				if math.IsNaN(val) {
					val = 0.0
					info = "warning: value NaN. set value to 0.0"
				}

				if math.IsInf(val, 0) {
					val = 0.0
					info = "warning: value infinity. set value to 0.0"
				}

				tempVal = new(smartpi.TValue)
				if valueId == "current" {
					tempVal.Type = "current"
					tempVal.Unity = "A"
				} else if valueId == "voltage" {
					tempVal.Type = "voltage"
					tempVal.Unity = "V"
				} else if valueId == "power" {
					tempVal.Type = "power"
					tempVal.Unity = "W"
				} else if valueId == "cosphi" {
					tempVal.Type = "cosphi"
					tempVal.Unity = ""
				} else if valueId == "frequency" {
					tempVal.Type = "frequency"
					tempVal.Unity = "Hz"
				}
				tempVal.Info = info
				tempVal.Data = float32(val)

				values = append(values, tempVal)

				tempPhase = new(smartpi.TPhase)
				tempPhase.Phase = i + 1
				tempPhase.Name = "phase " + strconv.Itoa(i+1)
				tempPhase.MaxGUICurrent = config.GUIMaxCurrent[smartpi.PhaseNameFromNumber(strconv.Itoa(i+1))]
				tempPhase.Values = values

				phases = append(phases, tempPhase)
			}
		}

	} else if valueId == "all" {
		// request for one of the phases
		if (phaseId == "1") || (phaseId == "2") || (phaseId == "3") || (phaseId == "4") {

			var val float64
			var err error
			var info string
			values := []*smartpi.TValue{}
			id, err := strconv.Atoi(phaseId)
			smartpi.Checklog(err)

			for i := 1; i <= 5; i++ {

				if i == 1 && id < 5 {
					val, err = strconv.ParseFloat(records[id], 32)
				} else if i == 2 && id < 4 {
					val, err = strconv.ParseFloat(records[id+4], 32)
				} else if i == 3 && id < 4 {
					val, err = strconv.ParseFloat(records[id+7], 32)
				} else if i == 4 && id < 4 {
					val, err = strconv.ParseFloat(records[id+10], 32)
				} else if i == 5 && id < 4 {
					val, err = strconv.ParseFloat(records[id+13], 32)
				}

				if err != nil {
					val = 0.0
					info = "warning: parse error. set value to 0.0"
				}

				if math.IsNaN(val) {
					val = 0.0
					info = "warning: value NaN. set value to 0.0"
				}

				if math.IsInf(val, 0) {
					val = 0.0
					info = "warning: value infinity. set value to 0.0"
				}

				tempVal = new(smartpi.TValue)
				if i == 1 {
					tempVal.Type = "current"
					tempVal.Unity = "A"
				} else if i == 2 {
					tempVal.Type = "voltage"
					tempVal.Unity = "V"
				} else if i == 3 {
					tempVal.Type = "power"
					tempVal.Unity = "W"
				} else if i == 4 {
					tempVal.Type = "cosphi"
					tempVal.Unity = ""
				} else if i == 5 {
					tempVal.Type = "frequency"
					tempVal.Unity = "Hz"
				}

				if (i == 1 && id < 5) || (i >= 2 && id < 4) {
					tempVal.Info = info
					tempVal.Data = float32(val)

					values = append(values, tempVal)
				}

			}

			tempPhase = new(smartpi.TPhase)
			tempPhase.Phase, _ = strconv.Atoi(phaseId)
			tempPhase.Name = "phase " + phaseId
			tempPhase.MaxGUICurrent = config.GUIMaxCurrent[smartpi.PhaseNameFromNumber(phaseId)]
			tempPhase.Values = values

			phases = append(phases, tempPhase)

			// request for all phases
		} else if phaseId == "all" {

			for i := 0; i <= 3; i++ {

				var val float64
				var err error
				var info string
				values := []*smartpi.TValue{}

				for j := 1; j <= 5; j++ {

					if j == 1 && i < 4 {
						val, err = strconv.ParseFloat(records[i+1], 32)
					} else if j == 2 && i < 3 {
						val, err = strconv.ParseFloat(records[i+5], 32)
					} else if j == 3 && i < 3 {
						val, err = strconv.ParseFloat(records[i+8], 32)
					} else if j == 4 && i < 3 {
						val, err = strconv.ParseFloat(records[i+11], 32)
					} else if j == 5 && i < 3 {
						val, err = strconv.ParseFloat(records[i+14], 32)
					}

					if err != nil {
						val = 0.0
						info = "warning: parse error. set value to 0.0"
					}

					if math.IsNaN(val) {
						val = 0.0
						info = "warning: value NaN. set value to 0.0"
					}

					if math.IsInf(val, 0) {
						val = 0.0
						info = "warning: value infinity. set value to 0.0"
					}

					tempVal = new(smartpi.TValue)
					if j == 1 && i < 4 {
						tempVal.Type = "current"
						tempVal.Unity = "A"
					} else if j == 2 && i < 3 {
						tempVal.Type = "voltage"
						tempVal.Unity = "V"
					} else if j == 3 && i < 3 {
						tempVal.Type = "power"
						tempVal.Unity = "W"
					} else if j == 4 && i < 3 {
						tempVal.Type = "cosphi"
						tempVal.Unity = ""
					} else if j == 5 && i < 3 {
						tempVal.Type = "frequency"
						tempVal.Unity = "Hz"
					}

					if (j == 1 && i < 4) || (j >= 2 && i < 3) {
						tempVal.Info = info
						tempVal.Data = float32(val)

						values = append(values, tempVal)
					}

				}

				tempPhase = new(smartpi.TPhase)
				tempPhase.Phase = i + 1
				tempPhase.Name = "phase " + strconv.Itoa(i+1)
				tempPhase.MaxGUICurrent = config.GUIMaxCurrent[smartpi.PhaseNameFromNumber(strconv.Itoa(i+1))]
				tempPhase.Values = values

				phases = append(phases, tempPhase)

			}

		}

	}

	// create dataset with actual timestamp
	// for actual values there are only one dataset
	tempDataset = new(smartpi.TDataset)
	tempDataset.Time = records[0]
	tempDataset.Phases = phases

	datasets = append(datasets, tempDataset)

	measurement := smartpi.TMeasurement{
		Serial:          config.Serial,
		Name:            config.Name,
		Lat:             config.Lat,
		Lng:             config.Lng,
		Time:            t.Format("2006-01-02 15:04:05"),
		Softwareversion: "",
		Ipaddress:       network.GetLocalIP(),
		Datasets:        datasets,
	}
	if format == "xml" {
		// XML output of request
		type response struct {
			smartpi.TMeasurement
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
