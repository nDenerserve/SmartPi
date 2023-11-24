package main

import (
	"encoding/csv"
	"os"
	"strings"
	"time"

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

	log.Debug(moduleconfig.LoraWANSendInterval)

	tick := time.Tick(time.Duration(moduleconfig.LoraWANSendInterval) * time.Second)

	for ; ; <-tick {
		log.Info(time.Now())
		callback := func(port uint8, data []byte) {
			log.Debugf("Received message on port %v: %s", port, string(data))
		}

		for i, sharedfile := range moduleconfig.LoraWANSharedDirs {

			elements := strings.Split(moduleconfig.LoraWANSharedFilesElements[i], ":")
			log.Debug(elements)
			file, err := os.Open(sharedfile)
			utils.Checklog(err)
			defer file.Close()
			reader := csv.NewReader(file)
			reader.Comma = ';'
			records, err := reader.Read()
			log.Debug(records)
			utils.Checklog(err)
		}

		data := []byte("Hallo Welt")
		log.Debug(rn2483.MacGetStatus())

		err := rn2483.MacTx(false, uint8(1), data, callback)
		if err != nil {
			log.Error("FEHLER: " + err.Error())
		}
	}

}
