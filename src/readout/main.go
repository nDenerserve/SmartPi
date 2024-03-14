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
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/io/i2c"

	"github.com/fsnotify/fsnotify"

	//import the Paho Go MQTT library

	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

func makeReadoutAccumulator() (r smartpi.ReadoutAccumulator) {
	r.Current = make(smartpi.Readings)
	r.Voltage = make(smartpi.Readings)
	r.ActiveWatts = make(smartpi.Readings)
	r.CosPhi = make(smartpi.Readings)
	r.Frequency = make(smartpi.Readings)
	r.WattHoursConsumed = make(smartpi.Readings)
	r.WattHoursProduced = make(smartpi.Readings)
	return r
}

func makeReadout() (r smartpi.ADE7878Readout) {
	r.Current = make(smartpi.Readings)
	r.Voltage = make(smartpi.Readings)
	r.ActiveWatts = make(smartpi.Readings)
	r.CosPhi = make(smartpi.Readings)
	r.Frequency = make(smartpi.Readings)
	r.ApparentPower = make(smartpi.Readings)
	r.ReactivePower = make(smartpi.Readings)
	r.PowerFactor = make(smartpi.Readings)
	r.ActiveEnergy = make(smartpi.Readings)
	r.Energyconsumption = make(smartpi.Readings)
	r.Energyproduction = make(smartpi.Readings)
	return r
}

func pollSmartPi(config *config.Config, device *i2c.Device) {
	var mqttclient mqtt.Client
	var wattHourBalanced, wattHourBalancedAccu, consumedWattHourBalanced60s, producedWattHourBalanced60s float64
	var p models.Phase
	var consumedCounter, producedCounter float64
	var measureFrequency bool = true

	consumerCounterFile := filepath.Join(config.CounterDir, "consumecounter")
	producerCounterFile := filepath.Join(config.CounterDir, "producecounter")

	if config.MQTTenabled {
		mqttclient = newMQTTClient(config)
	}

	accumulator := makeReadoutAccumulator()
	i := 0

	tick := time.Tick(time.Duration(1000/config.Samplerate) * time.Millisecond)

	// disable measuring frequency if samplerate higher than 4 samples/second
	if config.Samplerate > 4 {
		measureFrequency = false
	}

	for {
		readouts := makeReadout()
		// Restart the accumulator loop every 60 seconds.
		if i > (60*config.Samplerate - 1) {
			i = 0
			accumulator = makeReadoutAccumulator()
		}

		startTime := time.Now()

		// Update readouts and the accumlator.
		smartpi.ReadPhase(device, config, models.PhaseN, measureFrequency, &readouts)
		accumulator.Current[models.PhaseN] += readouts.Current[models.PhaseN] / (60.0 * float64(config.Samplerate))
		for _, p = range smartpi.MainPhases {
			smartpi.ReadPhase(device, config, p, measureFrequency, &readouts)
			accumulator.Current[p] += readouts.Current[p] / (60.0 * float64(config.Samplerate))
			accumulator.Voltage[p] += readouts.Voltage[p] / (60.0 * float64(config.Samplerate))
			accumulator.ActiveWatts[p] += readouts.ActiveWatts[p] / (60.0 * float64(config.Samplerate))
			accumulator.CosPhi[p] += readouts.CosPhi[p] / (60.0 * float64(config.Samplerate))
			accumulator.Frequency[p] += readouts.Frequency[p] / (60.0 * float64(config.Samplerate))

			if readouts.ActiveWatts[p] >= 0 {
				readouts.Energyconsumption[p] = math.Abs(readouts.ActiveWatts[p]) / (3600.0 * float64(config.Samplerate))
				accumulator.WattHoursConsumed[p] += readouts.Energyconsumption[p]
			} else {
				readouts.Energyproduction[p] = math.Abs(readouts.ActiveWatts[p]) / (3600.0 * float64(config.Samplerate))
				accumulator.WattHoursProduced[p] += readouts.Energyproduction[p]
			}
			wattHourBalanced += readouts.ActiveWatts[p] / (3600.0 * float64(config.Samplerate))
		}

		// Update metrics endpoint.
		updatePrometheusMetrics(&readouts, config)

		// Every sample
		if i%1 == 0 {

			if config.SharedFileEnabled {
				smartpi.WriteSharedFile(config, &readouts, wattHourBalanced)
			}

			// Publish readouts to MQTT.
			if config.MQTTenabled {
				publishMQTTReadouts(config, mqttclient, &readouts, wattHourBalanced)
			}

			// Update InfluxDB (FastMeasurement) database.
			// if samplerate > 4 and safe to Database enabled.
			// Only I1-I4, U1-U3 and P1-P3
			if config.DatabaseEnabled && (measureFrequency == false) {
				updateInfluxFastdata(config, &readouts)
			}
			wattHourBalancedAccu += wattHourBalanced
			wattHourBalanced = 0
		}

		// Every 60 seconds.
		// Energymeasurement is only enabled if samplerate < 4
		if (i == (60*config.Samplerate - 1)) && (measureFrequency == true) {

			// balanced value
			consumedWattHourBalanced60s = 0.0
			producedWattHourBalanced60s = 0.0

			if wattHourBalancedAccu >= 0 {
				consumedWattHourBalanced60s = math.Abs(wattHourBalancedAccu)
			} else {
				producedWattHourBalanced60s = math.Abs(wattHourBalancedAccu)
			}

			// Update InfluxDB database.
			if config.DatabaseEnabled {
				updateInfluxDatabase(config, accumulator, consumedWattHourBalanced60s, producedWattHourBalanced60s)
			}
			// Update SQLlite database.
			if config.SQLLiteEnabled {
				updateSQLiteDatabase(config, accumulator, consumedWattHourBalanced60s, producedWattHourBalanced60s)
			}

			consumedCounter = 0.0
			producedCounter = 0.0

			// Update persistent counter files and read Values from not updated files
			if config.CounterEnabled {
				if wattHourBalancedAccu >= 0 {
					consumedCounter = smartpi.UpdateCounterFile(config, consumerCounterFile, math.Abs(wattHourBalancedAccu))
					producedCounter = smartpi.ReadCounterFile(config, producerCounterFile)
				} else {
					producedCounter = smartpi.UpdateCounterFile(config, producerCounterFile, math.Abs(wattHourBalancedAccu))
					consumedCounter = smartpi.ReadCounterFile(config, consumerCounterFile)
				}
				wattHourBalancedAccu = 0.0
			}
			if config.MQTTenabled {
				publishMQTTCalculations(config, mqttclient, consumedWattHourBalanced60s, producedWattHourBalanced60s, consumedCounter, producedCounter)
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

func configWatcher(config *config.Config) {
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

var appVersion = "No Version Provided"

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	version.Version = appVersion
	prometheus.MustRegister(versioncollector.NewCollector("smartpi"))
}

func main() {
	smartpiconfig := config.NewConfig()

	go configWatcher(smartpiconfig)

	versionFlag := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	log.SetLevel(smartpiconfig.LogLevel)

	smartpi.CheckDatabase(smartpiconfig.DatabaseDir)

	listenAddress := smartpiconfig.MetricsListenAddress

	log.Debug("Start SmartPi readout")

	device, _ := smartpi.InitADE7878(smartpiconfig)

	go pollSmartPi(smartpiconfig, device)

	//http.Handle("/metrics", prometheus.Handler())
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>SmartPi Readout Metrics Server</title></head>
            <body>
            <h1>SmartPi Readout Metrics Server</h1>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})

	log.Debugf("Listening on %s", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		panic(fmt.Errorf("Error starting HTTP server: %s", err))
	}
}
