package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/version"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/server/controllers"
	"github.com/nDenerserve/SmartPi/smartpi/server/serverutils"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWT struct {
	Token string `json:"token"`
}

type Error struct {
	Message string `json:"message"`
}

var responseCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "smartpi",
		Name:      "responses_total",
		Help:      "Total HTTP requests processed by the server, excluding scrapes.",
	},
	[]string{"code", "method"},
)

var appVersion = "No Version Provided"

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	version.Version = appVersion
}

func main() {

	smartpiConfig := config.NewSmartPiConfig()
	smartpiACConfig := config.NewSmartPiACConfig()
	// moduleConfig := config.NewModuleconfig()
	controller := controllers.Controller{}

	log.SetLevel(smartpiConfig.LogLevel)

	go configWatcher(smartpiConfig)
	go acConfigWatcher(smartpiACConfig)

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/signup", signup).Methods("POST")
	router.HandleFunc("/api/v1/login", controller.Login(smartpiConfig)).Methods("POST")
	// router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}", serverutils.CheckConfigForPasswordMiddleWare(controller.SmartPiLiveValues(smartpiConfig), smartpiConfig))
	router.HandleFunc("/api/all/all/now", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET") // for e.manager compatibility
	router.HandleFunc("/api/v1/smartpiac/livedata", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/value/{valueId}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/value/{valueId}/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/listconnections", serverutils.TokenVerifyMiddleWare(controller.ConnectionList(), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/addstaticiptoconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.TokenVerifyMiddleWare(controller.AddStaticIpToConnection(), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/removestaticipfromconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.TokenVerifyMiddleWare(controller.RemoveStaticIpFromConnection(), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/scanwifi", serverutils.TokenVerifyMiddleWare(controller.ScanWifi(), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/createconnection", serverutils.TokenVerifyMiddleWare(controller.CreateConnection(), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiacconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiACConfig(smartpiACConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiacconfiguration", serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiACConfig(smartpiACConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiConfig(smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiconfiguration", serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiConfig(smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}/starttime/{starttime}/stoptime/{stoptime}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}/starttime/{starttime}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}/aggregate/{aggregate}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}/aggregate/{aggregate}/starttime/{starttime}/stoptime/{stoptime}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/progressdata/value/{value}/aggregate/{aggregate}/starttime/{starttime}", controller.SmartPiProgressdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/barchart/value/{value}", controller.SmartPiChartdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/barchart/value/{value}/aggregate/{aggregate}", controller.SmartPiChartdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/barchart/value/{value}/aggregate/{aggregate}/starttime/{starttime}", controller.SmartPiChartdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/barchart/value/{value}/aggregate/{aggregate}/starttime/{starttime}/stoptime/{stoptime}", controller.SmartPiChartdata(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/range/{range}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/range/{range}/aggregate/{aggregate}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/start/{start}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/start/{start}/aggregate/{aggregate}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/start/{start}/stop/{stop}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/csvexport/start/{start}/stop/{stop}/aggregate/{aggregate}", controller.SmartPiCsvExport(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/value/{valueId}/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")

	router.PathPrefix("/assets").Handler(http.FileServer(http.Dir(smartpiConfig.DocRoot + "/")))
	// Catch-all: Serve our JavaScript application's entry-point (index.html).
	router.PathPrefix("/").HandlerFunc(IndexHandler(smartpiConfig.DocRoot + "/index.html"))

	// router.PathPrefix("/").Handler(http.FileServer(http.Dir(smartpiConfig.DocRoot)))

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

	log.Print("Starting Smartpi server @Port: " + strconv.Itoa(smartpiConfig.WebserverPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(smartpiConfig.WebserverPort), nil))

}

func IndexHandler(entrypoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entrypoint)
	}
	return http.HandlerFunc(fn)
}

func signup(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("successfully called signup"))
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protected endpoint invoked")
}

func configWatcher(config *config.SmartPiConfig) {
	log.Debug("Start SmartPi config watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Debug("config watcher init done 1")
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					config.ReadParameterFromFile()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Debug("config watcher init done 2")
	err = watcher.Add("/etc/smartpi")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func acConfigWatcher(acConfig *config.SmartPiACConfig) {
	log.Debug("Start SmartPi acConfig watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Debug("acConfig watcher init done 1")
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					acConfig.ReadParameterFromFile()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Debug("acConfig watcher init done 2")
	err = watcher.Add("/etc/smartpiAC")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	log.Debug("acConfig watcher init done 3")
}
