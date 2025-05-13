package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/goburrow/serial"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/utils"
	"github.com/nDenerserve/mbserver"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var appVersion = "No Version Provided"

// main
func main() {

	smartpiConfig := config.NewSmartPiConfig()
	smartpiACConfig := config.NewSmartPiACConfig()
	go configWatcher(smartpiConfig)
	go acConfigWatcher(smartpiACConfig)

	version := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	log.SetLevel(smartpiConfig.LogLevel)

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("ERROR", err)
	}
	defer watcher.Close()

	//
	done := make(chan bool)

	serv := mbserver.NewServer()

	if smartpiConfig.ModbusTCPenabled {
		err := serv.ListenTCP(smartpiConfig.ModbusTCPAddress)
		if err != nil {
			log.Fatalf("%v\n", err)
		} else {
			log.Info("Modbus TCP started on: ")
			log.Info(smartpiConfig.ModbusTCPAddress)
		}
	}

	if smartpiConfig.ModbusRTUenabled {
		log.Info("Device: ", smartpiConfig.ModbusRTUDevice, "  Address: ", smartpiConfig.ModbusRTUAddress)
		err := serv.ListenRTU(&serial.Config{
			Address:  smartpiConfig.ModbusRTUDevice,
			BaudRate: 19200,
			DataBits: 8,
			StopBits: 1,
			Parity:   "N"}, smartpiConfig.ModbusRTUAddress)
		if err != nil {
			log.Fatalf("failed to listen, got %v\n", err)
		} else {
			log.Info("Modbus RTU started on address: ")
			log.Info(smartpiConfig.ModbusRTUAddress)
		}
	}
	defer serv.Close()

	var registervalue uint32

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Debugf("EVENT! %#v\n", event)
				time.Sleep(1 * time.Second)
				file, err := os.Open(smartpiConfig.SharedDir + "/smartpi_values")
				utils.Checklog(err)
				defer file.Close()
				reader := csv.NewReader(bufio.NewReader(file))
				reader.Comma = ';'
				records, err := reader.Read()
				log.Debugf("%v", records)
				utils.Checklog(err)
				if len(records) >= 19 {
					for i := 1; i < len(records)-1; i++ {
						registervalue = 0
						val, err := strconv.ParseFloat(records[i], 32)
						if err != nil {
							log.Error("error converting value", err)
							val = 0.0
						} else {
							registervalue = math.Float32bits(float32(val))
						}
						serv.HoldingRegisters[2*i-2] = uint16(registervalue >> 16)
						serv.HoldingRegisters[2*i-1] = uint16(registervalue)
					}
				} else {
					log.Fatal("Values not written")
				}

				file.Close()

				// watch for errors
			case err := <-watcher.Errors:
				log.Fatal("ERROR", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(smartpiConfig.SharedDir + "/smartpi_values"); err != nil {
		log.Fatal("ERROR", err)

	}

	<-done
}

func configWatcher(config *config.SmartPiConfig) {
	log.Debug("Start SmartPi watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Debug("init done 1")
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

	log.Debug("init done 2")
	err = watcher.Add("/etc/smartpi")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	log.Debug("init done 3")
}

func acConfigWatcher(acConfig *config.SmartPiACConfig) {
	log.Debug("Start SmartPi watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Debug("init done 1")
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

	log.Debug("init done 2")
	err = watcher.Add("/etc/smartpiAC")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	log.Debug("init done 3")
}
