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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/src/smartpi"
	"github.com/nDenerserve/SmartPi/src/smartpi/smartpiapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"github.com/rs/cors"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
)

var appVersion = "No Version Provided"
var User smartpi.User
var config smartpi.Config

type JSONMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// type User struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

type JwtToken struct {
	Token string `json:"token"`
}

type Exception struct {
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

// AuthMiddleware is our middleware to check our token is valid. Returning
// a 401 status to the client if it is not valid.
func AuthMiddleware(next http.Handler) http.Handler {
	if len(config.AppKey) == 0 {
		log.Fatal("HTTP server unable to start, expected an APP_KEY for JWT auth")
	}
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return jwtMiddleware.Handler(next)
}

// func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		authorizationHeader := req.Header.Get("authorization")
// 		if authorizationHeader != "" {
// 			bearerToken := strings.Split(authorizationHeader, " ")
// 			if len(bearerToken) == 2 {
// 				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
// 					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 						return nil, fmt.Errorf("There was an error")
// 					}
// 					return []byte("secret"), nil
// 				})
// 				if error != nil {
// 					json.NewEncoder(w).Encode(Exception{Message: error.Error()})
// 					return
// 				}
// 				if token.Valid {
// 					context.Set(req, "decoded", token.Claims)
// 					next(w, req)
// 				} else {
// 					json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
// 				}
// 			}
// 		} else {
// 			json.NewEncoder(w).Encode(Exception{Message: "An authorization header is required"})
// 		}
// 	})
// }

func init() {
	prometheus.MustRegister(version.NewCollector("smartpi"))
}

type Softwareinformations struct {
	Softwareversion string
}

func getSoftwareInformations(w http.ResponseWriter, r *http.Request) {
	data := Softwareinformations{Softwareversion: appVersion}

	// JSON output of request
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}

func main() {

	config = *smartpi.NewConfig()
	User = *smartpi.NewUser()
	smartpiapi.User = User
	smartpiapi.Config = config

	version := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	fmt.Println("SmartPi server started")

	corsWrapper := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})

	r := mux.NewRouter()
	r.HandleFunc("/api/{phaseId}/{valueId}/now", smartpiapi.ServeMomentaryValues)
	r.HandleFunc("/api/{phaseId}/{valueId}/now/{format}", smartpiapi.ServeMomentaryValues)
	r.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpiapi.ServeChartValues)
	r.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpiapi.ServeChartValues)
	r.HandleFunc("/api/values/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpiapi.ServeChartValues)
	r.HandleFunc("/api/values/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpiapi.ServeChartValues)
	r.HandleFunc("/api/dayvalues/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpiapi.ServeDayValues)
	r.HandleFunc("/api/dayvalues/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}/{format}", smartpiapi.ServeDayValues)
	r.HandleFunc("/api/csv/from/{fromDate}/to/{toDate}", smartpiapi.ServeCSVValues)
	r.HandleFunc("/api/version", getSoftwareInformations)
	r.HandleFunc("/api/login", smartpiapi.Login).Methods("POST")
	r.HandleFunc("/api/config/test", smartpiapi.TestEndpoint).Methods("GET")
	r.Handle("/api/config/read", AuthMiddleware(http.HandlerFunc(smartpiapi.ReadConfig))).Methods("GET")
	// r.HandleFunc("/api/config/write", AuthMiddleware(http.HandlerFunc(smartpiapi.WriteConfig(config)))).Methods("POST")
	// r.HandleFunc("/api/config/user/read", AuthMiddleware(http.HandlerFunc(smartpiapi.ReadUserData(config)))).Methods("GET")
	// r.HandleFunc("/api/config/network/scanwifi", AuthMiddleware(http.HandlerFunc(smartpiapi.WifiList(config)))).Methods("GET")
	// r.HandleFunc("/api/config/network/networkconnections", AuthMiddleware(http.HandlerFunc(smartpiapi.NetworkConnections(config)))).Methods("GET")
	// r.HandleFunc("/api/config/network/wifi/set", AuthMiddleware(http.HandlerFunc(smartpiapi.CreateWifi(config)))).Methods("POST")
	// r.HandleFunc("/api/config/network/wifi/set/{name}", AuthMiddleware(http.HandlerFunc(smartpiapi.RemoveWifi(config)))).Methods("DELETE")
	// r.HandleFunc("/api/config/network/wifi/active/{name}", AuthMiddleware(http.HandlerFunc(smartpiapi.ActivateWifi(config)))).Methods("GET")
	// r.HandleFunc("/api/config/network/wifi/active/{name}", AuthMiddleware(http.HandlerFunc(smartpiapi.DeactivateWifi(config)))).Methods("DELETE")
	// r.HandleFunc("/api/config/network/wifi/security/change/key", AuthMiddleware(http.HandlerFunc(smartpiapi.ChangeWifiKey(config)))).Methods("POST")
	r.HandleFunc("/api/v2/{phaseId}/{valueId}/now", smartpiapi.ServeMomentaryValues)
	r.HandleFunc("/api/v2/{phaseId}/{valueId}/now/{format}", smartpiapi.ServeMomentaryValues)
	r.HandleFunc("/api/v2/version", getSoftwareInformations)
	r.HandleFunc("/api/v2/csv/from/{fromDate}/to/{toDate}", smartpiapi.ServeInfluxCSVValues)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(config.DocRoot)))
	// http.Handle("/metrics", prometheus.Handler())
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", corsWrapper.Handler(r))
	// log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.WebserverPort), nil))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(8910), nil))
}
