package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/fsnotify/fsnotify"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpidc"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

func pollSmartPiDC(config *config.DCconfig) {

	var mqttclient mqtt.Client
	var logline string
	var power []float64
	var energyConsumed []float64
	var energyProduced []float64
	var accumulatorEnergyConsumed []float64
	var accumulatorEnergyProduced []float64
	var accumulatorEnergyBalanced []float64
	var accumulatorInput []float64
	var accumulatorPower []float64
	inputConfiguration := [4]int{models.NotUsed, models.NotUsed, models.NotUsed, models.NotUsed}

	if config.MQTTenabled {
		mqttclient = smartpidc.NewMQTTClient(config)
	}

	i := 0

	// create input configuration of hardware inputs
	// 0 for not connected
	// 1 for voltage
	// 2 for current
	for j := range config.InputType {

		if strings.Contains(config.InputType[j], "Voltage") {
			inputConfiguration[j] = models.Voltage
		} else {
			inputConfiguration[j] = models.Current
		}

	}

	powerCalculationCounter := 0
	for j := range config.Power {
		if len(config.Power[j]) == 2 {
			powerCalculationCounter++
		}
	}
	power = make([]float64, powerCalculationCounter)
	energyConsumed = make([]float64, powerCalculationCounter)
	energyProduced = make([]float64, powerCalculationCounter)
	accumulatorInput = make([]float64, 4)
	accumulatorPower = make([]float64, powerCalculationCounter)
	accumulatorEnergyConsumed = make([]float64, powerCalculationCounter)
	accumulatorEnergyProduced = make([]float64, powerCalculationCounter)
	accumulatorEnergyBalanced = make([]float64, powerCalculationCounter)

	log.Debug("InputConfiguration: " + strings.Join(utils.Int2StringSlice(inputConfiguration[:]), "|"))

	tick := time.Tick(time.Duration(1000/config.Samplerate) * time.Millisecond)

	for {

		if i > (60*config.Samplerate - 1) {
			i = 0
			accumulatorInput = make([]float64, 4)
			accumulatorPower = make([]float64, powerCalculationCounter)
			accumulatorEnergyConsumed = make([]float64, powerCalculationCounter)
			accumulatorEnergyProduced = make([]float64, powerCalculationCounter)
			accumulatorEnergyBalanced = make([]float64, powerCalculationCounter)
		}

		logline = "\n"
		for j := range power {
			power[j] = 0
			energyConsumed[j] = 0
			energyProduced[j] = 0
		}

		startTime := time.Now()

		values := smartpidc.GetValues(config)

		for j := range values {
			// log.Debug("J: " + strconv.Itoa(j))
			accumulatorInput[j] += values[j] / (60.0 * float64(config.Samplerate))
		}

		// calculate power and energy
		calculationCounter := 0
		for j := range config.Power {
			if len(config.Power[j]) == 2 {
				logline = logline + "Calculate Power " + strconv.Itoa(calculationCounter) + ": \n"
				logline = logline + "VoltageInput: " + strconv.Itoa(config.Power[j][0]) + "   CurrentInput: " + strconv.Itoa(config.Power[j][1]) + "\n"
				power[calculationCounter] = values[config.Power[j][0]-1] * values[config.Power[j][1]-1]
				if math.IsNaN(power[calculationCounter]) {
					power[calculationCounter] = 0
				}
				accumulatorPower[calculationCounter] += power[calculationCounter] / (60.0 * float64(config.Samplerate))
				if power[calculationCounter] >= 0 {
					energyConsumed[calculationCounter] = math.Abs(power[calculationCounter]) / (3600.0 * float64(config.Samplerate))
					accumulatorEnergyConsumed[calculationCounter] += energyConsumed[calculationCounter]
				} else {
					energyProduced[calculationCounter] = math.Abs(power[calculationCounter]) / (3600.0 * float64(config.Samplerate))
					accumulatorEnergyProduced[calculationCounter] += energyProduced[calculationCounter]
				}
				accumulatorEnergyBalanced[calculationCounter] = accumulatorEnergyBalanced[calculationCounter] + power[calculationCounter]/(3600.0*float64(config.Samplerate))

				logline = logline + "Power " + strconv.Itoa(calculationCounter) + ": " + fmt.Sprintf("%f", power[calculationCounter]) + "\n"
				logline = logline + "Energy consumed " + strconv.Itoa(calculationCounter) + ": " + fmt.Sprintf("%f", energyConsumed[calculationCounter]) + "\n"
				logline = logline + "Energy produced " + strconv.Itoa(calculationCounter) + ": " + fmt.Sprintf("%f", energyProduced[calculationCounter]) + "\n"
				logline = logline + "Energy balanced " + strconv.Itoa(calculationCounter) + ": " + fmt.Sprintf("%f", accumulatorEnergyBalanced[calculationCounter]) + "\n"
				calculationCounter++
			}
		}

		log.Debug(logline)
		log.Debug("I: " + strconv.Itoa(i))
		log.Debug("Database: " + strconv.FormatBool(config.DatabaseEnabled))
		log.Debug("StoreIntervall: " + config.DatabaseStoreIntervall)
		log.Debug("Samplerate: " + strconv.Itoa(config.Samplerate))

		// Every sample
		if i%1 == 0 {

			// Publish readouts to MQTT.
			if config.MQTTenabled && (config.MQTTpublishintervall == "sample") {
				smartpidc.PublishMQTTReadouts(config, mqttclient, inputConfiguration[:], values[:], power, energyConsumed, energyProduced)
			}

			log.Debug("SharedFileEnable: " + strconv.FormatBool(config.SharedFileEnabled))

			if config.SharedFileEnabled {
				smartpidc.WriteSharedFile(config, inputConfiguration[:], values[:], power, energyConsumed, energyProduced)
			}

			// write Database for FastData
			if config.DatabaseEnabled && (config.DatabaseStoreIntervall == "sample") {
				log.Debug("Write sample data to database")
				smartpidc.InsertInfluxData(config, inputConfiguration[:], values[:], power, energyConsumed, energyProduced)
			}

		}

		if i == (60*config.Samplerate - 1) {

			log.Debug("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!MINUTE!!!!!!!!!!!!!!!!!!!!!!!!!!!")

			// TODO: update persist counter files

			// Publish readouts to MQTT.
			if config.MQTTenabled && (config.MQTTpublishintervall == "minute") {
				smartpidc.PublishMQTTReadouts(config, mqttclient, inputConfiguration[:], values[:], power, accumulatorEnergyConsumed, accumulatorEnergyProduced)
				smartpidc.PublishMQTTCalculations(config, mqttclient, accumulatorEnergyBalanced)
			}

			if config.DatabaseEnabled && (config.DatabaseStoreIntervall == "minute") {
				log.Debug("Write minute data to database")
				smartpidc.InsertInfluxData(config, inputConfiguration[:], accumulatorInput[:], accumulatorPower, accumulatorEnergyConsumed, accumulatorEnergyProduced)
			}

		}

		delay := time.Since(startTime) - (time.Duration(1000/config.Samplerate) * time.Millisecond)
		if int64(delay) > 0 {
			log.Errorf("Readout delayed: %s", delay)
		}
		<-tick
		i++
	}

}

func configWatcher(config *config.DCconfig) {
	log.Debug("Start SmartPiDC watcher")
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
					config.ReadDCParameterFromFile()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Debug("init done 2")
	err = watcher.Add("/etc/smartpidc")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	log.Debug("init done 3")
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

}

var appVersion = "No Version Provided"

func main() {

	log.Info("SmartPiDC started")

	smartpidcconfig := config.NewDCconfig()

	log.SetLevel(smartpidcconfig.LogLevel)

	go configWatcher(smartpidcconfig)

	c := flag.Int("c", 0, "calibrate input #")
	flag.Parse()
	calibration := *c

	if calibration > 0 {
		smartpidc.CalibrateInput(smartpidcconfig, calibration)
		os.Exit(0)
	}

	go pollSmartPiDC(smartpidcconfig)

	for {
	}

}
