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
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nDenerserve/SmartPi/src/smartpi"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/exp/io/i2c"

	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

var readouts = [...]string{
	"I1", "I2", "I3", "I4", "V1", "V2", "V3", "P1", "P2", "P3", "COS1", "COS2", "COS3", "F1", "F2", "F3"}

func pollSmartPi(config *smartpi.Config, device *i2c.Device) {
	var mqttclient MQTT.Client

	consumerCounterFile := filepath.Join(config.CounterDir, "consumecounter")
	producerCounterFile := filepath.Join(config.CounterDir, "producecounter")

	if config.MQTTenabled {
		mqttclient = newMQTTClient(config)
	}

	data := make([]float32, 22)
	i := 0

	for {
		// Restart the accumulator loop every 60 seconds.
		if i > 59 {
			i = 0
			data = make([]float32, 22)
		}

		startTime := time.Now()
		valuesr := smartpi.ReadoutValues(device, config)

		// Update the accumlator.
		for index, _ := range data {
			switch index {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15:
				data[index] += float32(valuesr[index] / 60.0)
			case 16:
				if valuesr[7] >= 0 {
					data[index] += float32(math.Abs(valuesr[7]) / 3600.0)
				}
			case 17:
				if valuesr[8] >= 0 {
					data[index] += float32(math.Abs(valuesr[8]) / 3600.0)
				}
			case 18:
				if valuesr[9] >= 0 {
					data[index] += float32(math.Abs(valuesr[9]) / 3600.0)
				}
			case 19:
				if valuesr[7] < 0 {
					data[index] += float32(math.Abs(valuesr[7]) / 3600.0)
				}
			case 20:
				if valuesr[8] < 0 {
					data[index] += float32(math.Abs(valuesr[8]) / 3600.0)
				}
			case 21:
				if valuesr[9] < 0 {
					data[index] += float32(math.Abs(valuesr[9]) / 3600.0)
				}
			}
		}

		// Update metrics endpoint.
		updatePrometheusMetrics(valuesr)

		// Every 5 seconds
		if i%5 == 0 {
			if config.SharedFileEnabled {
				writeSharedFile(config, valuesr)
			}

			// Publish readouts to MQTT.
			if config.MQTTenabled {
				publishMQTTReadouts(config, mqttclient, valuesr)
			}
		}

		// Every 60 seconds.
		if i == 59 {
			// Update SQLlite database.
			if config.DatabaseEnabled {
				updateSQLiteDatabase(config, data)
			}

			// Update persistent counter files.
			if config.CounterEnabled {
				updateCounterFile(config, consumerCounterFile, float64(data[16]+data[17]+data[18]))
				updateCounterFile(config, producerCounterFile, float64(data[19]+data[20]+data[21]))
			}
		}

		sleepFor := (1000 * time.Millisecond) - time.Since(startTime)
		if int64(sleepFor) <= 0 {
			log.Errorf("Sleep duration negative: %s", sleepFor)
		} else {
			log.Debugf("Sleeping for %s", sleepFor)
			time.Sleep(sleepFor)
		}
		i++
	}
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
	prometheus.MustRegister(version.NewCollector("smartpi"))
}

func main() {
	config := smartpi.NewConfig()
	log.SetLevel(config.LogLevel)

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
