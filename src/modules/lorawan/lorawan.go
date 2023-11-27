package main

import (
	"encoding/binary"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/x448/float16"

	rn2483 "github.com/nDenerserve/RN2483"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

func main() {

	moduleconfig := config.NewModuleconfig()

	log.SetLevel(moduleconfig.LogLevel)

	log.Info("Start lorawan")

	rn2483.SetName("/dev/ttySC0")
	rn2483.SetBaud(57600)
	rn2483.SetTimeout(time.Millisecond * 500)

	// Connect the RN2483 via serial
	rn2483.Connect()

	// Make sure the app closes the connection at the end the free the resource
	defer rn2483.Disconnect()

	err := rn2483.MacSetApplicationEUI(moduleconfig.LoraWANApplicationEUI)
	if err != nil {
		log.Error(err)
	}
	err = rn2483.MacSetApplicationKey(moduleconfig.LoraWANApplicationKey)
	if err != nil {
		log.Error(err)
	}
	log.Debug(rn2483.MacGetStatus())

	joined := rn2483.MacJoin(rn2483.OTAA)

	log.Info(joined)
	log.Debug(rn2483.MacGetStatus())

	go sendData(moduleconfig)

	select {}
}

func sendData(moduleconfig *config.Moduleconfig) {

	var data []byte

	tick := time.Tick(time.Duration(moduleconfig.LoraWANSendInterval) * time.Second)

	for ; ; <-tick {
		log.Info(time.Now())
		callback := func(port uint8, data []byte) {
			log.Debugf("Received message on port %v: %s", port, string(data))
		}

		for i, sharedfile := range moduleconfig.LoraWANSharedDirs {

			file, err := os.Open(sharedfile)
			utils.Checklog(err)
			defer file.Close()
			reader := csv.NewReader(file)
			reader.Comma = ';'
			records, err := reader.Read()
			log.Debug("!!!!!")
			log.Debug(records)
			utils.Checklog(err)

			for j, element := range strings.Split(moduleconfig.LoraWANSharedFilesElements[i], "|") {

				log.Debug(element)

				tmpElements := strings.Split(string(element), ":")
				log.Debug(tmpElements)
				log.Debug(len(tmpElements))
				elementNumber, err := strconv.Atoi(tmpElements[0])
				utils.Checklog(err)
				numberLength, err := strconv.Atoi(tmpElements[1][:1])
				utils.Checklog(err)
				numberFormat := tmpElements[1][1:]
				numberFactor, err := strconv.ParseFloat(tmpElements[2], 64)
				utils.Checklog(err)

				if elementNumber >= 0 && elementNumber < len(records) {
					if strings.EqualFold(numberFormat, "f") {

						tmpValue, err := strconv.ParseFloat(records[elementNumber], 32)
						log.Debug("Value: " + strconv.FormatFloat(tmpValue, 'f', -1, 32))
						utils.Checklog(err)
						if numberLength == 2 {
							data = binary.BigEndian.AppendUint16(data, float16.Fromfloat32(float32(tmpValue*numberFactor)).Bits())
						} else if numberLength == 4 {
							data = binary.BigEndian.AppendUint32(data, uint32(float32(tmpValue*numberFactor)))
						}

					}

				}

			}
		}

		log.Debug(rn2483.MacGetStatus())

		err := rn2483.MacTx(false, uint8(1), data, callback)
		if err != nil {
			log.Error("FEHLER: " + err.Error())
		}
	}

}
