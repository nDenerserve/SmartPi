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
	"github.com/nDenerserve/mbserver"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var appVersion = "No Version Provided"

const (
	// minSharedFields is the field count of a complete line in the shared
	// values file: the timestamp plus at least 18 readings.
	minSharedFields = 19
	// readAttempts is how often a read of the shared values file is tried
	// before it counts as failed.
	readAttempts = 3
	// readRetryDelay is the pause between two attempts. smartpireadout leaves
	// the file empty only for a fraction of a millisecond between truncating
	// and writing it, so a short pause is enough to find the complete line.
	readRetryDelay = 10 * time.Millisecond
	// maxFailedReads is how many reads of the shared values file in a row may
	// fail, each after all its attempts, before the server gives up.
	maxFailedReads = 10
	// sharedFileTimeout is how long the server waits at startup for
	// smartpireadout to create the shared values file.
	sharedFileTimeout = 60 * time.Second
)

// readRegisterValues reads the shared values file written by smartpireadout
// and returns its readings as float32 bit patterns, in file order.
//
// smartpireadout rewrites the file in place every second, truncating it before
// writing the new line, so a read can find the file empty or the line
// incomplete. Such a read returns an error and leaves the caller's registers
// as they are.
func readRegisterValues(path string) ([]uint32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	records, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	// The first field is the timestamp; the last one is empty because a
	// complete line ends with the separator.
	if len(records) < minSharedFields || records[len(records)-1] != "" {
		return nil, fmt.Errorf("incomplete line in %s: %d fields", path, len(records))
	}

	values := make([]uint32, 0, len(records)-2)
	for i := 1; i < len(records)-1; i++ {
		val, err := strconv.ParseFloat(records[i], 32)
		if err != nil {
			return nil, fmt.Errorf("field %d in %s: %w", i, path, err)
		}
		values = append(values, math.Float32bits(float32(val)))
	}
	return values, nil
}

// readRegisterValuesRetry calls readRegisterValues up to attempts times, with
// delay between the attempts, and returns the first complete read or the last
// error.
func readRegisterValuesRetry(path string, attempts int, delay time.Duration) ([]uint32, error) {
	values, err := readRegisterValues(path)
	for i := 1; i < attempts && err != nil; i++ {
		time.Sleep(delay)
		values, err = readRegisterValues(path)
	}
	return values, err
}

// waitForFile waits until path exists or timeout has passed. smartpireadout
// creates the shared values file with its first reading, which can be a
// moment after this server has started.
func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		_, err := os.Stat(path)
		if err == nil || !os.IsNotExist(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

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

	sharedFile := smartpiConfig.SharedDir + "/smartpi_values"
	failedReads := 0

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Debugf("EVENT! %#v\n", event)
				time.Sleep(1 * time.Second)
				values, err := readRegisterValuesRetry(sharedFile, readAttempts, readRetryDelay)
				if err != nil {
					failedReads++
					if failedReads >= maxFailedReads {
						log.Fatalf("Values not written, %d reads in a row failed: %v", failedReads, err)
					}
					log.Debugf("Skipping incomplete shared values file: %v", err)
					continue
				}
				failedReads = 0
				for i, registervalue := range values {
					serv.HoldingRegisters[2*i] = uint16(registervalue >> 16)
					serv.HoldingRegisters[2*i+1] = uint16(registervalue)
				}

				// watch for errors
			case err := <-watcher.Errors:
				log.Fatal("ERROR", err)
			}
		}
	}()

	if err := waitForFile(sharedFile, sharedFileTimeout); err != nil {
		log.Fatal("ERROR", err)
	}

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(sharedFile); err != nil {
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
