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

package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/src/smartpi"
	"golang.org/x/net/context"
)

type JSONError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func stringInSlice(list1 []string, list2 []string) bool {
	for _, a := range list1 {
		for _, b := range list2 {
			if b == a {
				return true
			}
		}
	}
	return false
}

func BasicAuth(realm string, handler http.HandlerFunc, c *smartpi.Config, u *smartpi.User, roles ...string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		u.ReadUserFromFile(user)

		roleAllowed := false
		if len(roles) > 0 && stringInSlice(u.Role, roles) {
			roleAllowed = true
		} else if len(roles) == 0 {
			roleAllowed = true
		}

		if !ok || !u.Exist || subtle.ConstantTimeCompare([]byte(user), []byte(u.Name)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(u.Password)) != 1 || !roleAllowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(400)
			// w.Write([]byte("Unauthorised.\n"))
			if err := json.NewEncoder(w).Encode(JSONError{Code: 401, Message: "Unauthorized"}); err != nil {
				panic(err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), "Username", u)
		ctx = context.WithValue(r.Context(), "Config", c)
		handler(w, r.WithContext(ctx))
	}
}

func main() {

	config := smartpi.NewConfig()
	user := smartpi.NewUser()
	fmt.Println("SmartPi server started")

	r := mux.NewRouter()
	r.HandleFunc("/api/{phaseId}/{valueId}/now", smartpi.ServeMomentaryValues)
	r.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeChartValues)
	r.HandleFunc("/api/values/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeChartValues)
	r.HandleFunc("/api/dayvalues/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeDayValues)
	r.HandleFunc("/api/csv/from/{fromDate}/to/{toDate}", smartpi.ServeCSVValues)
	r.HandleFunc("/api/config/read", BasicAuth("Please enter your username and password for this site", smartpi.ReadConfig, config, user, "administrator")).Methods("GET")
	r.HandleFunc("/api/config/write", BasicAuth("Please enter your username and password for this site", smartpi.WriteConfig, config, user, "administrator")).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(config.DocRoot)))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.WebserverPort), nil))
}
