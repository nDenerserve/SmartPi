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

	"github.com/nDenerserve/SmartPi/src/smartpi"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/io/i2c"

	"github.com/fsnotify/fsnotify"

	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/prometheus/client_golang/prometheus"
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
	return r
}

func pollSmartPi(config *smartpi.Config, device *i2c.Device) {
	var mqttclient MQTT.Client
	var consumed, produced, wattHourBalanced5s, consumedWattHourBalanced60s, producedWattHourBalanced60s float64
	var p smartpi.Phase

	consumerCounterFile := filepath.Join(config.CounterDir, "consumecounter")
	producerCounterFile := filepath.Join(config.CounterDir, "producecounter")

	if config.MQTTenabled {
		mqttclient = newMQTTClient(config)
	}

	accumulator := makeReadoutAccumulator()
	i := 0

	tick := time.Tick(time.Second)

	for {
		readouts := makeReadout()
		// Restart the accumulator loop every 60 seconds.
		if i > 59 {
			i = 0
			accumulator = makeReadoutAccumulator()
		}

		startTime := time.Now()

		// Update readouts and the accumlator.
		smartpi.ReadPhase(device, config, smartpi.PhaseN, &readouts)
		accumulator.Current[smartpi.PhaseN] += readouts.Current[smartpi.PhaseN] / 60.0
		for _, p = range smartpi.MainPhases {
			smartpi.ReadPhase(device, config, p, &readouts)
			accumulator.Current[p] += readouts.Current[p] / 60.0
			accumulator.Voltage[p] += readouts.Voltage[p] / 60.0
			accumulator.ActiveWatts[p] += readouts.ActiveWatts[p] / 60.0
			accumulator.CosPhi[p] += readouts.CosPhi[p] / 60.0
			accumulator.Frequency[p] += readouts.Frequency[p] / 60.0

			if readouts.ActiveWatts[p] >= 0 {
				accumulator.WattHoursConsumed[p] += math.Abs(readouts.ActiveWatts[p]) / 3600.0
			} else {
				accumulator.WattHoursProduced[p] += math.Abs(readouts.ActiveWatts[p]) / 3600.0
			}
			wattHourBalanced5s += readouts.ActiveWatts[p] / 3600.0
		}

		// Update metrics endpoint.
		updatePrometheusMetrics(&readouts)

		// Every 5 seconds
		if i%5 == 0 {
			if config.SharedFileEnabled {
				writeSharedFile(config, &readouts, wattHourBalanced5s)
			}

			// Publish readouts to MQTT.
			if config.MQTTenabled {
				publishMQTTReadouts(config, mqttclient, &readouts)
			}

			wattHourBalanced5s = 0
		}

		// Every 60 seconds.
		if i == 59 {

			// balanced value
			var wattHourBalanced60s float64
			consumedWattHourBalanced60s = 0.0
			producedWattHourBalanced60s = 0.0

			for _, p = range smartpi.MainPhases {
				wattHourBalanced60s += accumulator.WattHoursConsumed[p]
				wattHourBalanced60s -= accumulator.WattHoursProduced[p]
			}
			if wattHourBalanced60s >= 0 {
				consumedWattHourBalanced60s = wattHourBalanced60s
			} else {
				producedWattHourBalanced60s = wattHourBalanced60s
			}

			// Update SQLlite database.
			if config.DatabaseEnabled {
				updateSQLiteDatabase(config, accumulator, consumedWattHourBalanced60s, producedWattHourBalanced60s)
			}

			// Update persistent counter files.
			if config.CounterEnabled {
				consumed = 0.0
				for _, p = range smartpi.MainPhases {
					consumed += accumulator.WattHoursConsumed[p]
				}
				updateCounterFile(config, consumerCounterFile, consumed)
				produced = 0.0
				for _, p = range smartpi.MainPhases {
					produced += accumulator.WattHoursProduced[p]
				}
				updateCounterFile(config, producerCounterFile, produced)
			}
		}

		delay := time.Since(startTime) - (1000 * time.Millisecond)
		if int64(delay) > 0 {
			log.Errorf("Readout delayed: %s", delay)
		}
		<-tick
		i++
	}
}

func configWatcher(config *smartpi.Config) {
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

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	prometheus.MustRegister(currentMetric)
	prometheus.MustRegister(voltageMetric)
	prometheus.MustRegister(activePowerMetirc)
	prometheus.MustRegister(cosphiMetric)
	prometheus.MustRegister(frequencyMetric)
	prometheus.MustRegister(apparentPowerMetric)
	prometheus.MustRegister(reactivePowerMetric)
	prometheus.MustRegister(powerFactorMetric)
	prometheus.MustRegister(version.NewCollector("smartpi"))
}

var appVersion = "No Version Provided"

func main() {
	config := smartpi.NewConfig()

	go configWatcher(config)

	version := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	log.SetLevel(config.LogLevel)

	smartpi.CheckDatabase(config.DatabaseDir)

	listenAddress := config.MetricsListenAddress

	log.Debug("Start SmartPi readout")

	device, _ := smartpi.InitADE7878(config)

	go pollSmartPi(config, device)

	http.Handle("/metrics", prometheus.Handler())
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
