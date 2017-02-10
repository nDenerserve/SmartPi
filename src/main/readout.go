
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
	  "smartpi"
		"fmt"
		"os"
		"time"
		"math"
		"io/ioutil"
		"strconv"
		"log"

		//import the Paho Go MQTT library
		MQTT "github.com/eclipse/paho.mqtt.golang"
)

var readouts = [...]string {
 "I1", "I2", "I3", "I4", "V1", "V2", "V3", "P1", "P2", "P3", "COS1", "COS2", "COS3", "F1", "F2", "F3",}

func main() {
	config := smartpi.NewConfig()
	var counter float64

	var mqttclient MQTT.Client

	if (config.Debuglevel > 0){
		fmt.Printf("Start SmartPi readout\n")
	}


	if (config.MQTTenabled == 1){
		if (config.Debuglevel > 0){
			fmt.Printf("Connecting to MQTT broker at %s\n", (config.MQTTbroker + ":" + config.MQTTbrokerport) )
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

		for i:=0; i<12; i++ {
			valuesr := smartpi.ReadoutValues(device,config)

			t := time.Now()
			if (config.Debuglevel > 0){
				fmt.Println(t.Format("## Actuals File Update ##"))
				fmt.Println(t.Format("2006-01-02 15:04:05"))
				fmt.Printf("I1: %g  I2: %g  I3: %g  I4: %g  V1: %g  V2: %g  V3: %g  P1: %g  P2: %g  P3: %g  COS1: %g  COS2: %g  COS3: %g  F1: %g  F2: %g  F3: %g  \n",valuesr[0],valuesr[1],valuesr[2],valuesr[3],valuesr[4],valuesr[5],valuesr[6],valuesr[7],valuesr[8],valuesr[9],valuesr[10],valuesr[11],valuesr[12],valuesr[13],valuesr[14],valuesr[15]);
			}
			var f *os.File
			var err error
			if _, err = os.Stat(config.Shareddir+"/"+config.Sharedfile); os.IsNotExist(err) {
				os.MkdirAll(config.Shareddir, 0777)
				f, err = os.Create(config.Shareddir+"/"+config.Sharedfile)
				if err != nil {
			      panic(err)
			  }
			} else {
				f, err = os.OpenFile(config.Shareddir+"/"+config.Sharedfile,os.O_WRONLY, 0666)
				if err != nil {
			      panic(err)
			  }
			}
			defer f.Close()
			_, err = f.WriteString(t.Format("2006-01-02 15:04:05")+fmt.Sprintf(";%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g;%g",valuesr[0],valuesr[1],valuesr[2],valuesr[3],valuesr[4],valuesr[5],valuesr[6],valuesr[7],valuesr[8],valuesr[9],valuesr[10],valuesr[11],valuesr[12],valuesr[13],valuesr[14],valuesr[15]))
			if err != nil {
					panic(err)
			}
			f.Sync()
			f.Close()

			//Publish readouts to MQTT
			//[basetopic]/[node]/[keyname]
			if (config.MQTTenabled == 1){
				if (mqttclient.IsConnected()){
					if (config.Debuglevel > 0){
						fmt.Println("Publishing readoputs via MQTT...")
					}
					for i:=0; i<len(readouts); i++ {
						//fmt.Printf(config.MQTTtopic + "/" + readouts[i] + "\n")
						topic := config.MQTTtopic + "/" + readouts[i]
						if token := mqttclient.Publish(topic , 1, false, strconv.FormatFloat(float64(valuesr[i]), 'f', 2, 32)); token.Wait() && token.Error() != nil {
							fmt.Println(token.Error())
						}
					}
				}
			}

			for index, _ := range data {

				switch (index) {

				case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15:
						data[index] = data[index] + valuesr[index] / 12.0
					/*	if index==7 || index==8 || index==9 {
							fmt.Printf("Index: %g,  Valuesr: %g, Data: %g \n", index, valuesr[index], data[index])
						  }*/

				case 16:
					if valuesr[7] >= 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[7]))) / 720.0
					}
				case 17:
					if valuesr[8] >= 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[8]))) / 720.0
					}
				case 18:
					if valuesr[9] >= 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[9]))) / 720.0
					}
				case 19:
					if valuesr[7] < 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[7]))) / 720.0
					}
				case 20:
					if valuesr[8] < 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[8]))) / 720.0
					}
				case 21:
					if valuesr[9] < 0 {
						data[index] = data[index] + float32(math.Abs(float64(valuesr[9]))) / 720.0
					}



				}

			}
			time.Sleep(5000 * time.Millisecond)

		}

		t := time.Now()

		if (config.Debuglevel > 0){
			fmt.Println("## SQLITE Database Update ##")
			fmt.Println(t.Format("2006-01-02 15:04:05"))
			fmt.Printf("I1: %g  I2: %g  I3: %g  I4: %g  V1: %g  V2: %g  V3: %g  P1: %g  P2: %g  P3: %g  COS1: %g  COS2: %g  COS3: %g  F1: %g  F2: %g  F3: %g  EB1: %g  EB2: %g  EB3: %g  EL1: %g  EL2: %g  EL3: %g \n",data[0],data[1],data[2],data[3],data[4],data[5],data[6],data[7],data[8],data[9],data[10],data[11],data[12],data[13],data[14],data[15],data[16],data[17],data[18],data[19],data[20],data[21]);
		}
		if _, err := os.Stat(config.Databasedir+"/smartpi_logdata_"+t.Format("200601")+".db"); os.IsNotExist(err) {
			if (config.Debuglevel > 0){
				fmt.Printf("Databasefile doesn't exist. Create.")
			}
			smartpi.CreateSQlDatabase(config.Databasedir, t);
		}
		smartpi.InsertData(config.Databasedir,t , data)





		consumecounter, err := ioutil.ReadFile(config.Counterdir + "/" + "consumecounter")
	  if err == nil {
			counter, err = strconv.ParseFloat(string(consumecounter), 64)
			if err != nil {
				counter = 0.0
				log.Fatal(err)
			}

	  } else {
			counter = 0.0
		}

		counter = counter + float64(data[16]+data[17]+data[18])

		err = ioutil.WriteFile(config.Counterdir + "/" + "consumecounter", []byte(strconv.FormatFloat(counter, 'f', 8, 64)), 0644)
	  if err != nil {
	     panic(err)
	  }


		producecounter, err := ioutil.ReadFile(config.Counterdir + "/" + "producecounter")
	  if err == nil {
			counter, err = strconv.ParseFloat(string(producecounter), 64)
			if err != nil {
				counter = 0.0
				log.Fatal(err)
			}

	  } else {
			counter = 0.0
		}

		counter = counter + float64(data[19]+data[20]+data[21])

		err = ioutil.WriteFile(config.Counterdir + "/" + "producecounter", []byte(strconv.FormatFloat(counter, 'f', 8, 64)), 0644)
	  if err != nil {
	     panic(err)
	  }
	}
}
