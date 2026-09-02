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
	"github.com/nDenerserve/SmartPi/smartpi/config"

	smartpiacConnectivity "github.com/nDenerserve/SmartPi/smartpiac/connectivity"
	smartpiacDatabase "github.com/nDenerserve/SmartPi/smartpiac/database"
	smartpiacDevice "github.com/nDenerserve/SmartPi/smartpiac/device"
	smartpiacFile "github.com/nDenerserve/SmartPi/smartpiac/file"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/io/i2c"

	"github.com/fsnotify/fsnotify"

	//import the Paho Go MQTT library

	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

// makeReadout returns an empty readout with all reading maps allocated. A fresh
// readout is built for every sample so that a phase which is not measured in
// this round simply has no entry instead of carrying over a stale value.
func makeReadout() (r models.ADE7878Readout) {
	r.Current = make(models.Readings)
	r.Voltage = make(models.Readings)
	r.ActiveWatts = make(models.Readings)
	r.CosPhi = make(models.Readings)
	r.Frequency = make(models.Readings)
	r.ApparentPower = make(models.Readings)
	r.ReactivePower = make(models.Readings)
	r.PowerFactor = make(models.Readings)
	r.ActiveEnergy = make(models.Readings)
	r.Energyconsumption = make(models.Readings)
	r.Energyproduction = make(models.Readings)
	return r
}

// pollSmartPi is the measurement loop of the readout daemon. It never returns
// and is meant to be run in its own goroutine.
//
// The loop is driven by a ticker derived from the configured samplerate. Every
// tick all phases are read from the ADE7878 and the resulting readout is fed to
// the various sinks, each of which runs on its own schedule:
//
//   - Prometheus metrics and the shared file are updated on every sample.
//   - MQTT and SmartPicloud publish on every sample, or - if a publication
//     interval is configured - once per interval from aggregated values.
//   - InfluxDB either stores every sample (StoreSamples) or one aggregated
//     record per minute.
//   - The persistent energy counter files are updated once per minute.
//
// Aggregation is done by models.ReadoutAggregator, which averages instantaneous
// quantities and sums up energy, so that reducing the publication rate does not
// lose any energy.
func pollSmartPi(acConfig *config.SmartPiACConfig, config *config.SmartPiConfig, device *i2c.Device) {
	var mqttclient mqtt.Client
	var smartpicloudMQTTclient mqtt.Client
	// wattHourBalanced holds the balanced energy (sum over all phases, signed)
	// of the current sample. It is reset after every sample, once all per-sample
	// sinks have consumed it.
	var wattHourBalanced, consumedWattHourBalanced60s, producedWattHourBalanced60s float64
	var p models.SmartPiPhase
	// Energy counters read back from the persistent counter files.
	var consumedCounter, producedCounter float64
	var measureFrequency bool = true

	consumerCounterFile := filepath.Join(acConfig.CounterDir, "consumecounter")
	producerCounterFile := filepath.Join(acConfig.CounterDir, "producecounter")

	if config.MQTTenabled {
		mqttclient = smartpiacConnectivity.NewMQTTClient(config)
	}

	if config.SmartpicloudEnabled {
		smartpicloudMQTTclient = smartpiacConnectivity.NewSmartPicloudMQTTClient(config)
	}

	// i counts the samples within the current minute and wraps around at
	// 60*samplerate. It drives everything that happens once per minute.
	i := 0

	// Minute values for the database and the energy counters are aggregated
	// over the whole 60 second window.
	influxAggregator := models.NewReadoutAggregator()

	// Readouts are published on every sample unless a publication interval is
	// configured. In that case the samples of a window are aggregated and only
	// the condensed readout is published. Both sinks keep their own aggregator
	// and their own window, so they can be throttled independently - typically
	// the cloud connection is published less often than the local broker.
	mqttAggregator := models.NewReadoutAggregator()
	smartpicloudAggregator := models.NewReadoutAggregator()
	mqttWindowStart := time.Now()
	smartpicloudWindowStart := mqttWindowStart

	tick := time.Tick(time.Duration(1000/acConfig.Samplerate) * time.Millisecond)

	// Measuring the frequency costs ~70ms per phase because the ADE7878 needs to
	// capture several full cycles. At more than 4 samples per second that no
	// longer fits into a sampling interval, so frequency measurement is dropped.
	// disable measuring frequency if samplerate higher than 4 samples/second
	if acConfig.Samplerate > 4 {
		measureFrequency = false
	}

	for {
		readouts := makeReadout()
		// Restart the accumulator loop every 60 seconds. Normally the minute
		// block below has already emptied the aggregator at this point; the
		// reset only matters if the samplerate was changed at runtime and the
		// counter wrapped without the minute block ever firing. In that case
		// the partial window is discarded.
		if i > (60*acConfig.Samplerate - 1) {
			i = 0
			influxAggregator.Reset()
		}

		startTime := time.Now()

		// Update readouts and the accumlator.
		// The neutral conductor only carries a current reading, all other
		// quantities are read per main phase.
		smartpiacDevice.ReadPhase(device, acConfig, models.PhaseN, measureFrequency, &readouts)
		for _, p = range smartpiacDevice.MainPhases {
			smartpiacDevice.ReadPhase(device, acConfig, p, measureFrequency, &readouts)

			// Split the active power of this sample into consumed and produced
			// energy. Dividing by 3600*samplerate converts the power in watts
			// into the watt hours accumulated during one sampling interval.
			// Only the matching direction is written, the other one stays unset
			// so that summing up the two never mixes both directions.
			if readouts.ActiveWatts[p] >= 0 {
				readouts.Energyconsumption[p] = math.Abs(readouts.ActiveWatts[p]) / (3600.0 * float64(acConfig.Samplerate))
			} else {
				readouts.Energyproduction[p] = math.Abs(readouts.ActiveWatts[p]) / (3600.0 * float64(acConfig.Samplerate))
			}
			// The balanced energy keeps its sign, so consumption on one phase
			// and production on another cancel each other out.
			wattHourBalanced += readouts.ActiveWatts[p] / (3600.0 * float64(acConfig.Samplerate))
		}
		// Feed the completed sample into the minute aggregator. This has to
		// happen after the phase loop, because only then the readout carries the
		// derived energy values as well.
		influxAggregator.Add(&readouts, wattHourBalanced)

		// Update metrics endpoint.
		smartpiacDatabase.UpdatePrometheusMetrics(&readouts, acConfig)

		// Every sample
		if i%1 == 0 {

			if config.SharedFileEnabled {
				smartpiacFile.WriteSharedFile(config, &readouts, wattHourBalanced)
			}

			// Publish readouts to MQTT.
			// Without a publication interval every sample is sent as it is.
			// Otherwise the samples are collected and one condensed readout is
			// published as soon as the window has elapsed.
			if config.MQTTenabled {
				if config.MQTTinterval <= 0 {
					smartpiacConnectivity.PublishMQTTReadouts(config, mqttclient, &readouts, wattHourBalanced)
				} else {
					mqttAggregator.Add(&readouts, wattHourBalanced)
					if time.Since(mqttWindowStart) >= time.Duration(config.MQTTinterval)*time.Second {
						aggregatedReadouts, aggregatedWattHourBalanced := mqttAggregator.Snapshot()
						smartpiacConnectivity.PublishMQTTReadouts(config, mqttclient, &aggregatedReadouts, aggregatedWattHourBalanced)
						mqttAggregator.Reset()
						mqttWindowStart = time.Now()
					}
				}
			}
			// Publish readouts to SmartPicloud via MQTT.
			if config.SmartpicloudEnabled {
				if config.SmartpicloudMQTTinterval <= 0 {
					smartpiacConnectivity.PublishSmartPicloudMQTTReadouts(config, smartpicloudMQTTclient, &readouts, wattHourBalanced)
				} else {
					smartpicloudAggregator.Add(&readouts, wattHourBalanced)
					if time.Since(smartpicloudWindowStart) >= time.Duration(config.SmartpicloudMQTTinterval)*time.Second {
						aggregatedReadouts, aggregatedWattHourBalanced := smartpicloudAggregator.Snapshot()
						smartpiacConnectivity.PublishSmartPicloudMQTTReadouts(config, smartpicloudMQTTclient, &aggregatedReadouts, aggregatedWattHourBalanced)
						smartpicloudAggregator.Reset()
						smartpicloudWindowStart = time.Now()
					}
				}
			}

			// Update InfluxDB (FastMeasurement) database.
			// if samplerate > 4 and safe to Database enabled.
			// Only I1-I4, U1-U3 and P1-P3
			if config.DatabaseEnabled && acConfig.StoreSamples {
				smartpiacDatabase.UpdateSampleInfluxDatabase(config, &readouts, wattHourBalanced)
			}
			// All per-sample sinks have seen this sample, start over. The minute
			// aggregator has its own copy of the value.
			wattHourBalanced = 0
		}

		// Every 60 seconds.
		// Energymeasurement is only enabled if samplerate < 4
		if i == (60*acConfig.Samplerate - 1) {

			// Condense the last minute into a single readout and immediately
			// start the next window. minuteWattHourBalanced is the signed energy
			// balance of the whole minute.
			minuteReadouts, minuteWattHourBalanced := influxAggregator.Snapshot()
			influxAggregator.Reset()

			// balanced value
			// Split the signed balance into the two unsigned directions that the
			// database and the counter files expect.
			consumedWattHourBalanced60s = 0.0
			producedWattHourBalanced60s = 0.0

			if minuteWattHourBalanced >= 0 {
				consumedWattHourBalanced60s = math.Abs(minuteWattHourBalanced)
			} else {
				producedWattHourBalanced60s = math.Abs(minuteWattHourBalanced)
			}

			// Update InfluxDB database.
			// When every sample is stored anyway, only the calculated minute
			// energies are added on top of the samples already written above.
			if config.DatabaseEnabled && !acConfig.StoreSamples {
				smartpiacDatabase.UpdateInfluxDatabase(config, &minuteReadouts, consumedWattHourBalanced60s, producedWattHourBalanced60s)
			} else if config.DatabaseEnabled && acConfig.StoreSamples {
				smartpiacDatabase.UpdateCalculatedInfluxDatabase(config, consumedWattHourBalanced60s, producedWattHourBalanced60s)
			}

			consumedCounter = 0.0
			producedCounter = 0.0

			// Update persistent counter files and read Values from not updated files
			// Only the counter matching the direction of this minute is
			// incremented, the other one is read back unchanged so that both
			// totals can be published together.
			if acConfig.CounterEnabled {
				if minuteWattHourBalanced >= 0 {
					consumedCounter = smartpiacFile.UpdateCounterFile(config, consumerCounterFile, math.Abs(minuteWattHourBalanced))
					producedCounter = smartpiacFile.ReadCounterFile(config, producerCounterFile)
				} else {
					producedCounter = smartpiacFile.UpdateCounterFile(config, producerCounterFile, math.Abs(minuteWattHourBalanced))
					consumedCounter = smartpiacFile.ReadCounterFile(config, consumerCounterFile)
				}
			}
			// The calculated minute values are always published once per minute,
			// independent of the readout publication interval.
			if config.MQTTenabled {
				smartpiacConnectivity.PublishMQTTCalculations(config, mqttclient, consumedWattHourBalanced60s, producedWattHourBalanced60s, consumedCounter, producedCounter)
			}
			// Publish readouts to SmartPicloud via MQTT.
			if config.SmartpicloudEnabled {
				smartpiacConnectivity.PublishSmartPicloudMQTTCalculations(config, smartpicloudMQTTclient, consumedWattHourBalanced60s, producedWattHourBalanced60s, consumedCounter, producedCounter)
			}
		}

		// Reading all phases must fit into one sampling interval. If it does
		// not, samples are effectively dropped and the loop drifts, so this is
		// logged as an error.
		delay := time.Since(startTime) - (time.Duration(1000/acConfig.Samplerate) * time.Millisecond)
		if int64(delay) > 0 {
			log.Errorf("Readout delayed: %s", delay)
		}
		<-tick
		i++
	}
}

// appVersion is set at build time via -ldflags.
var appVersion = "No Version Provided"

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	version.Version = appVersion
	prometheus.MustRegister(versioncollector.NewCollector("smartpi"))
}

// main wires up the readout daemon: it loads both configuration files, starts
// the measurement loop in the background and then serves the Prometheus metrics
// endpoint in the foreground.
func main() {

	smartpiConfig := config.NewSmartPiConfig()
	smartpiACConfig := config.NewSmartPiACConfig()

	// Both configuration files are watched so that changes take effect without
	// restarting the daemon.
	go configWatcher(smartpiConfig)
	go acConfigWatcher(smartpiACConfig)

	versionFlag := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	log.SetLevel(smartpiConfig.LogLevel)

	// smartpi.CheckDatabase(smartpiConfig.DatabaseDir)

	listenAddress := smartpiConfig.MetricsListenAddress

	log.Debug("Start SmartPi readout")

	device, _ := smartpiacDevice.InitADE7878(smartpiACConfig)

	// The measurement loop runs forever in its own goroutine.
	go pollSmartPi(smartpiACConfig, smartpiConfig, device)

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

// configWatcher reloads /etc/smartpi whenever the file is written to, so that
// settings such as the MQTT publication interval can be changed at runtime.
// It blocks forever and is meant to be run in its own goroutine.
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

// acConfigWatcher does the same as configWatcher for the AC measurement
// configuration in /etc/smartpiAC. It blocks forever and is meant to be run in
// its own goroutine.
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
