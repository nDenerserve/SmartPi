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
	"bufio"
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/controllers"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi"
	"github.com/nDenerserve/SmartPi/utils"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/rs/cors"
	// "golang.org/x/net/context"
)

type JSONMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

var responseCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "smartpi",
		Name:      "responses_total",
		Help:      "Total HTTP requests processed by the server, excluding scrapes.",
	},
	[]string{"code", "method"},
)

func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
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

func BasicAuth(realm string, handler http.HandlerFunc, c *config.Config, u *smartpi.User, roles ...string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		u.ReadUser(user, pass)

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
			if err := json.NewEncoder(w).Encode(JSONMessage{Code: 401, Message: "Unauthorized"}); err != nil {
				panic(err)
			}
			return
		}

		context.Set(r, "Config", c)
		context.Set(r, "Username", u)

		handler(w, r)
	}
}

var appVersion = "No Version Provided"

func init() {
	version.Version = appVersion
	prometheus.MustRegister(versioncollector.NewCollector("smartpi"))
}

type Softwareinformations struct {
	Softwareversion string
	Hardwareserial  string
	Hardwaremodel   string
}

func getSoftwareInformations(w http.ResponseWriter, r *http.Request) {

	serial := ""
	model := ""

	file, err := os.Open("/proc/cpuinfo")
	utils.Checklog(err)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Model") {
			substring := strings.Split(line, ": ")
			if len(substring) > 1 {
				model = (substring[len(substring)-1])
			}
		} else if strings.Contains(line, "Serial") {
			substring := strings.Split(line, ": ")
			if len(substring) > 1 {
				serial = (substring[len(substring)-1])
			}
		}
	}

	data := Softwareinformations{Softwareversion: appVersion, Hardwareserial: serial, Hardwaremodel: model}

	// JSON output of request
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}

func main() {

	smartpiconfig := config.NewConfig()
	smartpidcconfig := config.NewDCconfig()
	moduleconfig := config.NewModuleconfig()

	user := smartpi.NewUser()
	controller := controllers.Controller{}

	versionFlag := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/{phaseId}/{valueId}/now", smartpi.ServeMomentaryValues)
	router.HandleFunc("/api/{phaseId}/{valueId}/now/{format}", smartpi.ServeMomentaryValues)
	router.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeChartValues)
	router.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpi.ServeChartValues)
	router.HandleFunc("/api/values/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeChartValues)
	router.HandleFunc("/api/values/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpi.ServeChartValues)
	router.HandleFunc("/api/dayvalues/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeDayValues)
	router.HandleFunc("/api/dayvalues/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpi.ServeDayValues)
	router.HandleFunc("/api/csv/from/{fromDate}/to/{toDate}", smartpi.ServeCSVValues)
	router.HandleFunc("/api/version", getSoftwareInformations)
	router.HandleFunc("/api/config/read", BasicAuth("Please enter your username and password for this site", smartpi.ReadConfig, smartpiconfig, user, "smartpiadmin")).Methods("GET")
	router.HandleFunc("/api/config/write", BasicAuth("Please enter your username and password for this site", smartpi.WriteConfig, smartpiconfig, user, "smartpiadmin")).Methods("POST")
	router.HandleFunc("/api/config/user/read", BasicAuth("Please enter your username and password for this site", smartpi.ReadUserData, smartpiconfig, user, "smartpiadmin")).Methods("GET")
	router.HandleFunc("/api/config/network/scanwifi", BasicAuth("Please enter your username and password for this site", smartpi.WifiList, smartpiconfig, user, "smartpiadmin")).Methods("GET")
	router.HandleFunc("/api/config/network/networkconnections", BasicAuth("Please enter your username and password for this site", smartpi.NetworkConnections, smartpiconfig, user, "smartpiadmin")).Methods("GET")
	router.HandleFunc("/api/config/network/wifi/set", BasicAuth("Please enter your username and password for this site", smartpi.CreateWifi, smartpiconfig, user, "smartpiadmin")).Methods("POST")
	router.HandleFunc("/api/config/network/wifi/set/{name}", BasicAuth("Please enter your username and password for this site", smartpi.RemoveWifi, smartpiconfig, user, "smartpiadmin")).Methods("DELETE")
	// new API
	router.HandleFunc("/api/v1/login", controller.Login()).Methods("POST")
	router.HandleFunc("/api/v1/module/digitalout/{address}/{port}", utils.TokenVerifyMiddleWare(controller.SetDigitalout(moduleconfig))).Methods("PUT")
	router.HandleFunc("/api/v1/module/digitalout/{address}", utils.TokenVerifyMiddleWare(controller.ReadDigitalout(moduleconfig))).Methods("GET")
	router.HandleFunc("/api/v1/smartpidc/values/now/{format}", controller.SmartPiDCLiveValues(smartpidcconfig)).Methods("GET")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(smartpiconfig.DocRoot)))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "DELETE", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Access-Control-Allow-Headers", "Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
		Debug:            false,
	})

	handler := c.Handler(router)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", promhttp.InstrumentHandlerCounter(responseCount, handler))

	log.Print("Starting Smartpi server @Port: " + strconv.Itoa(smartpiconfig.WebserverPort))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(smartpiconfig.WebserverPort), nil))
}
