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

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/devicetoken"
	"github.com/nDenerserve/SmartPi/smartpi/server/controllers"
	modulescontrollers "github.com/nDenerserve/SmartPi/smartpi/server/controllers/modules"
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

// appVersion is set at build time via -ldflags, see the makefile.
var appVersion = "No Version Provided"

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {

	smartpiConfig := config.NewSmartPiConfig()
	smartpiACConfig := config.NewSmartPiACConfig()
	moduleConfig := config.NewModuleconfig()
	controller := controllers.Controller{}
	modulesController := modulescontrollers.ModulesController{}

	// Device tokens (see package devicetoken) are kept separate from the ini
	// config above: they are revoked by deleting them from this store, not by
	// reloading a file, and they must keep working across an appkey rotation,
	// which only ever invalidates session tokens.
	deviceTokens, err := devicetoken.NewStore(devicetoken.DefaultPath)
	if err != nil {
		log.Fatalf("Could not load device token store: %s", err)
	}

	log.SetLevel(smartpiConfig.LogLevel)

	go configWatcher(smartpiConfig)
	go acConfigWatcher(smartpiACConfig)
	go moduleConfigWatcher(moduleConfig)

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/signup", signup).Methods("POST")
	router.HandleFunc("/api/v1/login", controller.Login(smartpiConfig)).Methods("POST")
	// router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}", serverutils.CheckConfigForPasswordMiddleWare(controller.SmartPiLiveValues(smartpiConfig), smartpiConfig))
	router.HandleFunc("/api/all/all/now", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET") // for e.manager compatibility
	router.HandleFunc("/api/v1/smartpiac/livepower", controller.SmartPiLivePower(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/{phaseId}/{valueId}/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/value/{valueId}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/smartpiac/livedata/value/{valueId}/{format}", controller.SmartPiLiveValues(smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/listconnections", serverutils.TokenVerifyMiddleWare(controller.ConnectionList(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/addstaticiptoconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.TokenVerifyMiddleWare(controller.AddStaticIpToConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/removestaticipfromconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.TokenVerifyMiddleWare(controller.RemoveStaticIpFromConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/scanwifi", serverutils.TokenVerifyMiddleWare(controller.ScanWifi(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/createconnection", serverutils.TokenVerifyMiddleWare(controller.CreateConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiacconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiACConfig(smartpiACConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigRead)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiacconfiguration", serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiACConfig(smartpiACConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigWrite)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiConfig(smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigRead)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiconfiguration", serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiConfig(smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigWrite)).Methods("POST")

	router.HandleFunc("/api/v1/tokens", serverutils.RequireSessionToken(controller.ListDeviceTokens(deviceTokens), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/tokens", serverutils.RequireSessionToken(controller.CreateDeviceToken(deviceTokens, smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/tokens/{id}", serverutils.RequireSessionToken(controller.DeleteDeviceToken(deviceTokens), smartpiConfig)).Methods("DELETE")
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
	router.HandleFunc("/api/v1/module/digitalout/{address}/{port}", serverutils.TokenVerifyMiddleWare(modulesController.SetDigitalout(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeDigitalOut)).Methods("PUT")
	router.HandleFunc("/api/v1/module/digitalout/{address}", serverutils.TokenVerifyMiddleWare(modulesController.ReadDigitalout(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeDigitalOut)).Methods("GET")

	// 4-20mA analog output module routes (MCP4725)
	router.HandleFunc("/api/v1/module/analogout420ma/{address}/{current}", serverutils.TokenVerifyMiddleWare(modulesController.SetAnalogOut420mA(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeAnalogOut)).Methods("PUT")
	router.HandleFunc("/api/v1/module/analogout420ma/{address}", serverutils.TokenVerifyMiddleWare(modulesController.ReadAnalogOut420mA(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeAnalogOut)).Methods("GET")

	// Analog input module (MCP3424, 4x 4-20mA and/or 0-10V channels).
	router.HandleFunc("/api/v1/module/analogin/{address}/{channel}", serverutils.TokenVerifyMiddleWare(modulesController.ReadAnalogInChannel(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeAnalogIn)).Methods("GET")
	router.HandleFunc("/api/v1/module/analogin/{address}", serverutils.TokenVerifyMiddleWare(modulesController.ReadAnalogIn(moduleConfig, smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeAnalogIn)).Methods("GET")

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

func moduleConfigWatcher(moduleConfig *config.Moduleconfig) {
	log.Debug("Start SmartPi moduleConfig watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Debug("moduleConfig watcher init done 1")
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					moduleConfig.ReadParameterFromFile()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Debug("moduleConfig watcher init done 2")
	err = watcher.Add("/etc/smartpiModules")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	log.Debug("moduleConfig watcher init done 3")
}
