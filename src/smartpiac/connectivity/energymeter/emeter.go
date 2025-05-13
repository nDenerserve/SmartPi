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

// https://gist.github.com/fiorix/9664255
// https://en.wikipedia.org/wiki/Multicast_address
// https://support.mcommstv.com/hc/en-us/articles/202306226-Choosing-multicast-addresses-and-ports

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	smartpiacFile "github.com/nDenerserve/SmartPi/smartpiac/file"

	"github.com/nDenerserve/SmartPi/smartpi/network/multicast"
	"github.com/nDenerserve/SmartPi/utils"

	log "github.com/sirupsen/logrus"
)

const (
	address = "239.12.255.254:9522"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var appVersion = "No Version Provided"

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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("ERROR", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	log.Info("Emeter started")

	// go ping(config, watcher)
	// multicast.Listen(address, msgHandler)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Debugf("EVENT! %#v\n", event)
				time.Sleep(100 * time.Millisecond)

				if smartpiACConfig.EmeterEnabled {
					ping(smartpiACConfig, smartpiConfig)
				}

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

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Debug(n, "bytes read from", src)
	log.Debug(hex.Dump(b[:n]))
}

func ping(smartpiACConfig *config.SmartPiACConfig, smartpiConfig *config.SmartPiConfig) {

	var datagram []byte
	// var temp4byte []byte
	// var temp8byte []byte

	sumActivePowerPlus := uint32(0)
	sumActivePowerMinus := uint32(0)
	sumActiveEnergyPlus := uint64(0)
	sumActiveEnergyMinus := uint64(0)
	sumReactivePowerPlus := uint32(0)
	sumReactivePowerMinus := uint32(0)
	sumReactiveEnergyPlus := uint64(0)
	sumReactiveEnergyMinus := uint64(0)
	sumApparentPowerPlus := uint32(0)
	sumApparentPowerMinus := uint32(0)
	sumApparentEnergyPlus := uint64(0)
	sumApparentEnergyMinus := uint64(0)
	sumPowerFactor := uint32(0)

	phase1ActivePowerPlus := uint32(0)
	phase1ActivePowerMinus := uint32(0)
	phase1ActiveEnergyPlus := uint64(0)
	phase1ActiveEnergyMinus := uint64(0)
	phase1ReactivePowerPlus := uint32(0)
	phase1ReactivePowerMinus := uint32(0)
	phase1ReactiveEnergyPlus := uint64(0)
	phase1ReactiveEnergyMinus := uint64(0)
	phase1ApparentPowerPlus := uint32(0)
	phase1ApparentPowerMinus := uint32(0)
	phase1ApparentEnergyPlus := uint64(0)
	phase1ApparentEnergyMinus := uint64(0)
	phase1Current := uint32(0)
	phase1Voltage := uint32(0)
	phase1PowerFactor := uint32(0)

	phase2ActivePowerPlus := uint32(0)
	phase2ActivePowerMinus := uint32(0)
	phase2ActiveEnergyPlus := uint64(0)
	phase2ActiveEnergyMinus := uint64(0)
	phase2ReactivePowerPlus := uint32(0)
	phase2ReactivePowerMinus := uint32(0)
	phase2ReactiveEnergyPlus := uint64(0)
	phase2ReactiveEnergyMinus := uint64(0)
	phase2ApparentPowerPlus := uint32(0)
	phase2ApparentPowerMinus := uint32(0)
	phase2ApparentEnergyPlus := uint64(0)
	phase2ApparentEnergyMinus := uint64(0)
	phase2Current := uint32(0)
	phase2Voltage := uint32(0)
	phase2PowerFactor := uint32(0)

	phase3ActivePowerPlus := uint32(0)
	phase3ActivePowerMinus := uint32(0)
	phase3ActiveEnergyPlus := uint64(0)
	phase3ActiveEnergyMinus := uint64(0)
	phase3ReactivePowerPlus := uint32(0)
	phase3ReactivePowerMinus := uint32(0)
	phase3ReactiveEnergyPlus := uint64(0)
	phase3ReactiveEnergyMinus := uint64(0)
	phase3ApparentPowerPlus := uint32(0)
	phase3ApparentPowerMinus := uint32(0)
	phase3ApparentEnergyPlus := uint64(0)
	phase3ApparentEnergyMinus := uint64(0)
	phase3Current := uint32(0)
	phase3Voltage := uint32(0)
	phase3PowerFactor := uint32(0)

	softwareversion := uint32(111)

	file, err := os.Open(smartpiConfig.SharedDir + "/smartpi_values")
	utils.Checklog(err)
	defer file.Close()
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	records, err := reader.Read()
	utils.Checklog(err)
	if len(records) >= 19 {

		tmpValSum := 0.0
		tmpVal1 := 0.0
		tmpVal2 := 0.0
		tmpVal3 := 0.0

		// Active power
		tmpVal1, _ = strconv.ParseFloat(records[8], 64)
		tmpVal2, _ = strconv.ParseFloat(records[9], 64)
		tmpVal3, _ = strconv.ParseFloat(records[10], 64)

		// Active power sum
		tmpValSum = tmpVal1 + tmpVal2 + tmpVal3

		if (tmpValSum) >= 0.0 {
			sumActivePowerPlus = uint32(math.Round(tmpValSum*100) / 10)
			sumActivePowerMinus = uint32(0)
		} else {
			sumActivePowerMinus = uint32(math.Abs(math.Round(tmpValSum*100) / 10))
			sumActivePowerPlus = uint32(0)
		}
		// Active power phase 1
		if (tmpVal1) >= 0.0 {
			phase1ActivePowerPlus = uint32(math.Round(tmpVal1*100) / 10)
			phase1ActivePowerMinus = uint32(0)
		} else {
			phase1ActivePowerMinus = uint32(math.Abs(math.Round(tmpVal1*100) / 10))
			phase1ActivePowerPlus = uint32(0)
		}
		// Active power phase 2
		if (tmpVal2) >= 0.0 {
			phase2ActivePowerPlus = uint32(math.Round(tmpVal2*100) / 10)
			phase2ActivePowerMinus = uint32(0)
		} else {
			phase2ActivePowerMinus = uint32(math.Abs(math.Round(tmpVal2*100) / 10))
			phase2ActivePowerPlus = uint32(0)
		}
		// Active power phase 3
		if (tmpVal3) >= 0.0 {
			phase3ActivePowerPlus = uint32(math.Round(tmpVal3*100) / 10)
			phase3ActivePowerMinus = uint32(0)
		} else {
			phase3ActivePowerMinus = uint32(math.Abs(math.Round(tmpVal3*100) / 10))
			phase3ActivePowerPlus = uint32(0)
		}

		// Current
		tmpVal1, _ = strconv.ParseFloat(records[1], 64)
		tmpVal2, _ = strconv.ParseFloat(records[2], 64)
		tmpVal3, _ = strconv.ParseFloat(records[3], 64)
		phase1Current = uint32(math.Abs(math.Round(tmpVal1*10000) / 10))
		phase2Current = uint32(math.Abs(math.Round(tmpVal2*10000) / 10))
		phase3Current = uint32(math.Abs(math.Round(tmpVal3*10000) / 10))

		// Voltage
		tmpVal1, _ = strconv.ParseFloat(records[5], 64)
		tmpVal2, _ = strconv.ParseFloat(records[6], 64)
		tmpVal3, _ = strconv.ParseFloat(records[7], 64)
		phase1Voltage = uint32(math.Abs(math.Round(tmpVal1*10000) / 10))
		phase2Voltage = uint32(math.Abs(math.Round(tmpVal2*10000) / 10))
		phase3Voltage = uint32(math.Abs(math.Round(tmpVal3*10000) / 10))

		// Power factor
		tmpVal1, _ = strconv.ParseFloat(records[23], 64)
		tmpVal2, _ = strconv.ParseFloat(records[24], 64)
		tmpVal3, _ = strconv.ParseFloat(records[25], 64)
		phase1PowerFactor = uint32(math.Abs(math.Round(tmpVal1*10000) / 10))
		phase2PowerFactor = uint32(math.Abs(math.Round(tmpVal2*10000) / 10))
		phase3PowerFactor = uint32(math.Abs(math.Round(tmpVal3*10000) / 10))

		// Energy counter
		consumerCounterFile := filepath.Join(smartpiACConfig.CounterDir, "consumecounter")
		producerCounterFile := filepath.Join(smartpiACConfig.CounterDir, "producecounter")

		consumedCounter := smartpiacFile.ReadCounterFile(smartpiConfig, consumerCounterFile)
		sumActiveEnergyPlus = uint64(math.Abs(math.Round(consumedCounter*36000) / 10))
		producedCounter := smartpiacFile.ReadCounterFile(smartpiConfig, producerCounterFile)
		sumActiveEnergyMinus = uint64(math.Abs(math.Round(producedCounter*36000) / 10))

	} else {
		log.Fatal("Values not written")
	}

	datagram = []byte("SMA")
	datagram = append(datagram, []byte{0x00}...)
	datagram = append(datagram, []byte{0x00, 0x04}...)             // Data length: 4 byte (0x00000004)
	datagram = append(datagram, []byte{0x02, 0xA0}...)             // Tag: "Tag0" (42), version 0
	datagram = append(datagram, []byte{0x00, 0x00, 0x00, 0x01}...) // Group1 (default group)
	datagram = append(datagram, []byte{0x02, 0x44}...)             // Data length: 44 byte (variable)
	datagram = append(datagram, []byte{0x00, 0x10}...)             // Tag: "SMA Net 2", version 0
	datagram = append(datagram, []byte{0x60, 0x69}...)             // Protocol ID (energy meter protocol), Data length: 2 byte

	// Energy meter identifier Data length: 6 byte Susy-ID + Serial
	susyId := make([]byte, 2)
	binary.BigEndian.PutUint16(susyId, smartpiACConfig.EmeterSusyID)
	datagram = append(datagram, susyId...)

	// fmt.Printf("%x", config.EmeterSerial)

	tmpserial := utils.Uint32ToByteArrayBigEndian(smartpiACConfig.EmeterSerial)
	// datagram = append(datagram, config.EmeterSerial...)
	datagram = append(datagram, tmpserial[3])
	datagram = append(datagram, tmpserial[2])
	datagram = append(datagram, tmpserial[1])
	datagram = append(datagram, tmpserial[0])

	// Ticker measuring time in ms (with overflow)
	tickerMeasuringTimeByte := make([]byte, 4)
	binary.BigEndian.PutUint32(tickerMeasuringTimeByte, uint32(time.Now().UnixMilli()))

	datagram = append(datagram, tickerMeasuringTimeByte...)

	// Sum active power+
	append4ByteDatagram(&datagram, CurrentAverage[1], sumActivePowerPlus)
	// Sum active energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[1], sumActiveEnergyPlus)
	// Sum active power-
	append4ByteDatagram(&datagram, CurrentAverage[2], sumActivePowerMinus)
	// Sum active energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[2], sumActiveEnergyMinus)
	// Sum reactive power+
	append4ByteDatagram(&datagram, CurrentAverage[3], sumReactivePowerPlus)
	// Sum reactive energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[3], sumReactiveEnergyPlus)
	// Sum reactive power-
	append4ByteDatagram(&datagram, CurrentAverage[4], sumReactivePowerMinus)
	// Sum reactive energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[4], sumReactiveEnergyMinus)
	// Sum apparent power+
	append4ByteDatagram(&datagram, CurrentAverage[9], sumApparentPowerPlus)
	// Sum apparent energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[9], sumApparentEnergyPlus)
	// Sum apparent power-
	append4ByteDatagram(&datagram, CurrentAverage[10], sumApparentPowerMinus)
	// Sum apparent energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[10], sumApparentEnergyMinus)
	// Power Factor
	append4ByteDatagram(&datagram, CurrentAverage[13], sumPowerFactor)

	// Phase1 active power+
	append4ByteDatagram(&datagram, CurrentAverage[21], phase1ActivePowerPlus)
	// Phase1 active energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[21], phase1ActiveEnergyPlus)
	// Phase1 active power-
	append4ByteDatagram(&datagram, CurrentAverage[22], phase1ActivePowerMinus)
	// Phase1 active energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[22], phase1ActiveEnergyMinus)
	// Phase1 reactive power+
	append4ByteDatagram(&datagram, CurrentAverage[23], phase1ReactivePowerPlus)
	// Phase1 reactive energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[23], phase1ReactiveEnergyPlus)
	// Phase1 reactive power-
	append4ByteDatagram(&datagram, CurrentAverage[24], phase1ReactivePowerMinus)
	// Phase1 reactive energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[24], phase1ReactiveEnergyMinus)
	// Phase1 apparent power+
	append4ByteDatagram(&datagram, CurrentAverage[29], phase1ApparentPowerPlus)
	// Phase1 apparent energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[29], phase1ApparentEnergyPlus)
	// Phase1 apparent power-
	append4ByteDatagram(&datagram, CurrentAverage[30], phase1ApparentPowerMinus)
	// Phase1 apparent energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[30], phase1ApparentEnergyMinus)
	// Phase1 current
	append4ByteDatagram(&datagram, CurrentAverage[31], phase1Current)
	// Phase1 current
	append4ByteDatagram(&datagram, CurrentAverage[32], phase1Voltage)
	// Phase1 power Factor
	append4ByteDatagram(&datagram, CurrentAverage[33], phase1PowerFactor)

	// Phase2 active power+
	append4ByteDatagram(&datagram, CurrentAverage[41], phase2ActivePowerPlus)
	// Phase2 active energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[41], phase2ActiveEnergyPlus)
	// Phase2 active power-
	append4ByteDatagram(&datagram, CurrentAverage[42], phase2ActivePowerMinus)
	// Phase2 active energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[42], phase2ActiveEnergyMinus)
	// Phase2 reactive power+
	append4ByteDatagram(&datagram, CurrentAverage[43], phase2ReactivePowerPlus)
	// Phase2 reactive energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[43], phase2ReactiveEnergyPlus)
	// Phase2 reactive power-
	append4ByteDatagram(&datagram, CurrentAverage[44], phase2ReactivePowerMinus)
	// Phase2 reactive energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[44], phase2ReactiveEnergyMinus)
	// Phase2 apparent power+
	append4ByteDatagram(&datagram, CurrentAverage[49], phase2ApparentPowerPlus)
	// Phase2 apparent energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[49], phase2ApparentEnergyPlus)
	// Phase2 apparent power-
	append4ByteDatagram(&datagram, CurrentAverage[50], phase2ApparentPowerMinus)
	// Phase2 apparent energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[50], phase2ApparentEnergyMinus)
	// Phase2 current
	append4ByteDatagram(&datagram, CurrentAverage[51], phase2Current)
	// Phase2 current
	append4ByteDatagram(&datagram, CurrentAverage[52], phase2Voltage)
	// Phase2 power Factor
	append4ByteDatagram(&datagram, CurrentAverage[53], phase2PowerFactor)

	// Phase3 active power+
	append4ByteDatagram(&datagram, CurrentAverage[41], phase3ActivePowerPlus)
	// Phase3 active energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[41], phase3ActiveEnergyPlus)
	// Phase3 active power-
	append4ByteDatagram(&datagram, CurrentAverage[42], phase3ActivePowerMinus)
	// Phase3 active energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[42], phase3ActiveEnergyMinus)
	// Phase3 reactive power+
	append4ByteDatagram(&datagram, CurrentAverage[43], phase3ReactivePowerPlus)
	// Phase3 reactive energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[43], phase3ReactiveEnergyPlus)
	// Phase3 reactive power-
	append4ByteDatagram(&datagram, CurrentAverage[44], phase3ReactivePowerMinus)
	// Phase3 reactive energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[44], phase3ReactiveEnergyMinus)
	// Phase3 apparent power+
	append4ByteDatagram(&datagram, CurrentAverage[49], phase3ApparentPowerPlus)
	// Phase3 apparent energy+
	append8ByteDatagram(&datagram, EnergyDatapoint[49], phase3ApparentEnergyPlus)
	// Phase3 apparent power-
	append4ByteDatagram(&datagram, CurrentAverage[50], phase3ApparentPowerMinus)
	// Phase3 apparent energy-
	append8ByteDatagram(&datagram, EnergyDatapoint[50], phase3ApparentEnergyMinus)
	// Phase3 current
	append4ByteDatagram(&datagram, CurrentAverage[51], phase3Current)
	// Phase3 current
	append4ByteDatagram(&datagram, CurrentAverage[52], phase3Voltage)
	// Phase3 power Factor
	append4ByteDatagram(&datagram, CurrentAverage[73], phase3PowerFactor)

	// Sofwareversion
	append4ByteDatagram(&datagram, CurrentAverage[127], softwareversion)

	datagram = append(datagram, []byte{0x00, 0x00, 0x00, 0x00}...) //End

	log.Debugf("% x \n", datagram)

	conn, err := multicast.NewBroadcaster(smartpiACConfig.EmeterMulticastAddress + ":" + strconv.Itoa(smartpiACConfig.EmeterMulticastPort))

	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte(datagram))

	file.Close()

}

func append4ByteDatagram(datagram *[]byte, measurementType []byte, value uint32) {
	temp4byte := make([]byte, 4)
	*datagram = append(*datagram, measurementType...)
	binary.BigEndian.PutUint32(temp4byte, uint32(value))
	*datagram = append(*datagram, temp4byte...)
}

func append8ByteDatagram(datagram *[]byte, measurementType []byte, value uint64) {
	temp8byte := make([]byte, 8)
	*datagram = append(*datagram, measurementType...)
	binary.BigEndian.PutUint64(temp8byte, uint64(value))
	*datagram = append(*datagram, temp8byte...)
}

func IntToBytes(i int) []byte {
	if i > 0 {
		return append(big.NewInt(int64(i)).Bytes(), byte(1))
	}
	return append(big.NewInt(int64(i)).Bytes(), byte(0))
}
func BytesToInt(b []byte) int {
	if b[len(b)-1] == 0 {
		return -int(big.NewInt(0).SetBytes(b[:len(b)-1]).Int64())
	}
	return int(big.NewInt(0).SetBytes(b[:len(b)-1]).Int64())
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
