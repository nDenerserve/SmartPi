package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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
	cronRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/cron"
	linuxtoolsRepository "github.com/nDenerserve/SmartPi/smartpi/server/repository/linuxtools"
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

// pamCheckFlag is smartpiserver's hidden re-exec entry point - see
// runPamCheck below and linuxtoolsRepository.ValidateUser for why this
// exists instead of a normal in-process PAM check.
const pamCheckFlag = "--pam-check"

func main() {

	if len(os.Args) > 1 && os.Args[1] == pamCheckFlag {
		os.Exit(runPamCheck())
	}

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

	// Reconcile /etc/cron.d/smartpi with whatever [ftp] settings are
	// currently in /etc/smartpi - covers a hand-edited config file or a
	// cron.d/smartpi that was only ever freshly installed (still commented
	// out, see etc/cron.d/smartpi), without waiting for the next settings
	// save to fix it up.
	cronRepo := cronRepository.CronRepository{}
	if err := cronRepo.SyncFTPUpload(smartpiConfig.FTPupload, smartpiConfig.FTPsendtimes); err != nil {
		log.Error(err)
	}

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
	// Network settings and config:write are gated to the smartpiadmin group
	// for session (human) logins, on top of the existing scope check - see
	// RequireAdminGroup. config:read stays open to any authenticated
	// session, same as the hardware I/O module routes further down.
	router.HandleFunc("/api/v1/config/network/listconnections", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.ConnectionList(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork), smartpiConfig)).Methods("GET")
	// POST, not GET: both mutate the connection's address list (and restart
	// it), which a GET must never do - see parseIPv4AndCidr's callers.
	router.HandleFunc("/api/v1/config/network/addstaticiptoconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.AddStaticIpToConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/network/removestaticipfromconnection/ip/{ipaddress}/cidrsuffix/{cidrsuffix}/connection/{connection}", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.RemoveStaticIpFromConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/network/scanwifi", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.ScanWifi(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/config/network/createconnection", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.CreateConnection(), smartpiConfig, deviceTokens, devicetoken.ScopeNetwork), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiacconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiACConfig(smartpiACConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigRead)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiacconfiguration", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiACConfig(smartpiACConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigWrite), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/config/readsmartpiconfiguration", serverutils.TokenVerifyMiddleWare(controller.ReadSmartPiConfig(smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigRead)).Methods("GET")
	router.HandleFunc("/api/v1/config/writesmartpiconfiguration", serverutils.RequireAdminGroup(serverutils.TokenVerifyMiddleWare(controller.WriteSmartPiConfig(smartpiConfig), smartpiConfig, deviceTokens, devicetoken.ScopeConfigWrite), smartpiConfig)).Methods("POST")

	// Token/user management and the self-update endpoints below are already
	// session-only (RequireSessionToken); RequireAdminGroup additionally
	// restricts them to smartpiadmin, since a non-admin session could
	// otherwise mint itself a network-scoped device token (or a new user
	// account) and route straight around every other check on this list.
	router.HandleFunc("/api/v1/tokens", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.ListDeviceTokens(deviceTokens), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/tokens", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.CreateDeviceToken(deviceTokens, smartpiConfig), smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/tokens/{id}", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.DeleteDeviceToken(deviceTokens), smartpiConfig), smartpiConfig)).Methods("DELETE")

	// Settings "Users" tab: local Linux accounts (the same accounts Login
	// authenticates against via PAM). Session-only, like the token and
	// update endpoints above.
	router.HandleFunc("/api/v1/users", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.ListUsers(), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/users", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.CreateUser(), smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/users/{username}/password", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.ChangeUserPassword(), smartpiConfig), smartpiConfig)).Methods("POST")
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

	// I2C bus scan: reports occupied addresses on the configured bus (see
	// repository/modules/i2cscan.go for why this shells out to i2cdetect
	// instead of probing directly).
	router.HandleFunc("/api/v1/i2c/scan", serverutils.TokenVerifyMiddleWare(modulesController.ScanI2C(moduleConfig), smartpiConfig, deviceTokens, devicetoken.ScopeI2CScan)).Methods("GET")

	// Self-update (settings "Update" tab): upload+install a smartpi .deb, and
	// search/install/upgrade packages from the repositories already
	// configured on the device. Session-only, like the token management
	// endpoints above - these are at least as privileged as minting a
	// config:write device token, so a device token must never reach them.
	router.HandleFunc("/api/v1/update/version", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.GetUpdateVersion(appVersion), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/update/package", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.UploadUpdatePackage(smartpiConfig), smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/update/status", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.GetUpdateStatus(), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/apt/refresh", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.RefreshAptCache(), smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/apt/search", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.SearchAptPackages(), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/apt/upgradable", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.ListUpgradablePackages(), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/apt/package/{name}", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.GetAptPackageInfo(), smartpiConfig), smartpiConfig)).Methods("GET")
	router.HandleFunc("/api/v1/apt/install", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.InstallAptPackage(), smartpiConfig), smartpiConfig)).Methods("POST")
	router.HandleFunc("/api/v1/apt/upgrade-all", serverutils.RequireAdminGroup(serverutils.RequireSessionToken(controller.UpgradeAllPackages(), smartpiConfig), smartpiConfig)).Methods("POST")

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

// runPamCheck is smartpiserver re-exec'd as "smartpiserver --pam-check" via
// sudo (see etc/sudoers.d/smartpi-users and linuxtoolsRepository.ValidateUser)
// so that the PAM check inside linuxtoolsRepository.CheckPassword runs as
// root - smartpiserver.service itself runs unprivileged, and PAM's
// unix_chkpwd helper refuses to check any account's password other than the
// caller's own real uid for a non-root caller, which would otherwise make
// every account besides "smartpi" (including ones created via the settings
// "Users" tab) permanently unable to log in, regardless of password.
//
// Username and password are read from stdin as two newline-terminated
// lines, never argv, so they don't show up in `ps`. Returns 0 for a valid
// password, 1 otherwise; nothing is printed, since a caller only checks the
// exit code.
func runPamCheck() int {
	reader := bufio.NewReader(os.Stdin)

	username, err := reader.ReadString('\n')
	if err != nil {
		return 1
	}
	password, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return 1
	}

	if linuxtoolsRepository.CheckPassword(strings.TrimSuffix(username, "\n"), strings.TrimSuffix(password, "\n")) {
		return 0
	}
	return 1
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
