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
File: apihandlerscsv.go
Description: Handels API requests
*/

package smartpiapi

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/src/smartpi"
)

func ServeCSVValues(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	from := vars["fromDate"]
	to := vars["toDate"]

	w.Header().Set("Content-Type", "application/text")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	location := time.Now().Location()

	end, err := time.ParseInLocation(time.RFC3339, to, location)
	if err != nil {
		log.Println(err)
	}
	end = end.In(location)
	start, err := time.ParseInLocation(time.RFC3339, from, location)
	if err != nil {
		log.Println(err)
	}
	start = start.In(location)

	if end.Before(start) {
		start = start.AddDate(0, 0, -1)
	}

	csvfile := smartpi.CreateCSV(start, end)

	fmt.Fprintf(w, csvfile)

}

func ServeInfluxCSVValues(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	from := vars["fromDate"]
	to := vars["toDate"]

	config := smartpi.NewConfig()

	w.Header().Set("Content-Type", "application/text")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	location := time.Now().Location()

	end, err := time.ParseInLocation(time.RFC3339, to, location)
	if err != nil {
		log.Println(err)
	}
	end = end.In(location)
	start, err := time.ParseInLocation(time.RFC3339, from, location)
	if err != nil {
		log.Println(err)
	}
	start = start.In(location)

	if end.Before(start) {
		start = start.AddDate(0, 0, -1)
	}

	csvfile := smartpi.ReadCSVData(config, start, end)

	fmt.Fprintf(w, csvfile)

}
