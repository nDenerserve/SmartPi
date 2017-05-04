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
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"smartpi"
	"strconv"
	"strings"
	"time"

	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var readouts = [...]string{
	"I1", "I2", "I3", "I4", "V1", "V2", "V3", "P1", "P2", "P3", "COS1", "COS2", "COS3", "F1", "F2", "F3"}

func writeSharedFile(c *smartpi.Config, values [25]float32) {
	var f *os.File
	var err error
	s := make([]string, 16)
	for i, v := range values[0:16] {
		s[i] = fmt.Sprintf("%g", v)
	}
	t := time.Now()
	timeStamp := t.Format("2006-01-02 15:04:05")
	if c.Debuglevel > 0 {
		fmt.Println(t.Format("## Shared File Update ##"))
		fmt.Println(timeStamp)
		fmt.Printf("I1: %s  I2: %s  I3: %s  I4: %s  ", s[0], s[1], s[2], s[3])
		fmt.Printf("V1: %s  V2: %s  V3: %s  ", s[4], s[5], s[6])
		fmt.Printf("P1: %s  P2: %s  P3: %s  ", s[7], s[8], s[9])
		fmt.Printf("COS1: %s  COS2: %s  COS3: %s  ", s[10], s[11], s[12])
		fmt.Printf("F1: %s  F2: %s  F3: %s  ", s[13], s[14], s[15])
		fmt.Printf("\n")
	}
	sharedFile := filepath.Join(c.Shareddir, c.Sharedfile)
	if _, err = os.Stat(sharedFile); os.IsNotExist(err) {
		os.MkdirAll(c.Shareddir, 0777)
		f, err = os.Create(sharedFile)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(sharedFile, os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()
	_, err = f.WriteString(timeStamp + ";" + strings.Join(s, ";"))
	if err != nil {
		panic(err)
	}
	f.Sync()
	f.Close()
}

func updateCounterFile(c *smartpi.Config, f string, v float64) {
	t := time.Now()
	var counter float64
	counterFile, err := ioutil.ReadFile(f)
	if err == nil {
		counter, err = strconv.ParseFloat(string(counterFile), 64)
		if err != nil {
			counter = 0.0
			log.Fatal(err)
		}
	} else {
		counter = 0.0
	}

	if c.Debuglevel > 0 {
		fmt.Println("## Persistent counter file update ##")
		fmt.Println(t.Format("2006-01-02 15:04:05"))
		fmt.Printf("File: %q  Current: %g  Increment: %g \n ", f, counter, v)
	}

	err = ioutil.WriteFile(f, []byte(strconv.FormatFloat(counter+v, 'f', 8, 64)), 0644)
	if err != nil {
		panic(err)
	}
}

func updateSQLiteDatabase(c *smartpi.Config, data []float32) {
	t := time.Now()
	if c.Debuglevel > 0 {
		fmt.Println("## SQLITE Database Update ##")
		fmt.Println(t.Format("2006-01-02 15:04:05"))
		fmt.Printf("I1: %g  I2: %g  I3: %g  I4: %g  ", data[0], data[1], data[2], data[3])
		fmt.Printf("V1: %g  V2: %g  V3: %g  ", data[4], data[5], data[6])
		fmt.Printf("P1: %g  P2: %g  P3: %g  ", data[7], data[8], data[9])
		fmt.Printf("COS1: %g  COS2: %g  COS3: %g  ", data[10], data[11], data[12])
		fmt.Printf("F1: %g  F2: %g  F3: %g  ", data[13], data[14], data[15])
		fmt.Printf("EB1: %g  EB2: %g  EB3: %g  ", data[16], data[17], data[18])
		fmt.Printf("EL1: %g  EL2: %g  EL3: %g", data[19], data[20], data[21])
		fmt.Printf("\n")
	}

	dbFileName := "smartpi_logdata_" + t.Format("200601") + ".db"
	if _, err := os.Stat(filepath.Join(c.Databasedir, dbFileName)); os.IsNotExist(err) {
		if c.Debuglevel > 0 {
			fmt.Printf("Creating new database file.")
		}
		smartpi.CreateSQlDatabase(c.Databasedir, t)
	}
	smartpi.InsertData(c.Databasedir, t, data)
}

func publishReadouts(c *smartpi.Config, mqttclient MQTT.Client, values [25]float32) {
	//[basetopic]/[node]/[keyname]
	if c.MQTTenabled == 1 {
		if mqttclient.IsConnected() {
			if c.Debuglevel > 0 {
				fmt.Println("Publishing readoputs via MQTT...")
			}
			for i := 0; i < len(readouts); i++ {
				//fmt.Printf(config.MQTTtopic + "/" + readouts[i] + "\n")
				topic := c.MQTTtopic + "/" + readouts[i]
				if token := mqttclient.Publish(topic, 1, false, strconv.FormatFloat(float64(values[i]), 'f', 2, 32)); token.Wait() && token.Error() != nil {
					fmt.Println(token.Error())
				}
			}
		}
	}
}

func main() {
	config := smartpi.NewConfig()
	consumerCounterFile := filepath.Join(config.Counterdir, "consumecounter")
	producerCounterFile := filepath.Join(config.Counterdir, "producecounter")

	var mqttclient MQTT.Client

	if config.Debuglevel > 0 {
		fmt.Printf("Start SmartPi readout\n")
	}

	if config.MQTTenabled == 1 {
		if config.Debuglevel > 0 {
			fmt.Printf("Connecting to MQTT broker at %s\n", (config.MQTTbroker + ":" + config.MQTTbrokerport))
		}
		//create a MQTTClientOptions struct setting the broker address, clientid, user and password
		opts := MQTT.NewClientOptions().AddBroker("tcp://" + config.MQTTbroker + ":" + config.MQTTbrokerport)
		opts.SetClientID("SmartPi")
		opts.SetUsername(config.MQTTuser)
		opts.SetPassword(config.MQTTpass)
		//create and start a client using the above ClientOptions
		mqttclient = MQTT.NewClient(opts)
		if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
			panic(mqtttoken.Error())
		}
	}

	device, _ := smartpi.InitADE7878(config)

	for {
		data := make([]float32, 22)

		for i := 0; i < 12; i++ {
			valuesr := smartpi.ReadoutValues(device, config)

			writeSharedFile(config, valuesr)

			//Publish readouts to MQTT
			publishReadouts(config, mqttclient, valuesr)

			for index, _ := range data {
				switch index {
				case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15:
					data[index] += valuesr[index] / 12.0
				case 16:
					if valuesr[7] >= 0 {
						data[index] += float32(math.Abs(float64(valuesr[7]))) / 720.0
					}
				case 17:
					if valuesr[8] >= 0 {
						data[index] += float32(math.Abs(float64(valuesr[8]))) / 720.0
					}
				case 18:
					if valuesr[9] >= 0 {
						data[index] += float32(math.Abs(float64(valuesr[9]))) / 720.0
					}
				case 19:
					if valuesr[7] < 0 {
						data[index] += float32(math.Abs(float64(valuesr[7]))) / 720.0
					}
				case 20:
					if valuesr[8] < 0 {
						data[index] += float32(math.Abs(float64(valuesr[8]))) / 720.0
					}
				case 21:
					if valuesr[9] < 0 {
						data[index] += float32(math.Abs(float64(valuesr[9]))) / 720.0
					}
				}
			}
			time.Sleep(5000 * time.Millisecond)
		}

		// Update SQLlite database.
		updateSQLiteDatabase(config, data)

		// Update persistent counter files.
		updateCounterFile(config, consumerCounterFile, float64(data[16]+data[17]+data[18]))
		updateCounterFile(config, producerCounterFile, float64(data[19]+data[20]+data[21]))
	}
}
