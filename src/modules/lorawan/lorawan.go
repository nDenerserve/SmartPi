package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/x448/float16"

	rn2483 "github.com/nDenerserve/RN2483"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

var appVersion = "No Version Provided"
var serialName = "/dev/ttySC0"

func main() {

	moduleconfig := config.NewModuleconfig()

	version := flag.Bool("v", false, "prints current version information")
	devEUI := flag.Bool("deveui", false, "prints the devEUI of the LoRaWAN module")
	moduleReset := flag.Bool("reset", false, "hardwarereset of the LoRaWAN module")
	flag.Parse()
	if *version {
		log.Debug(appVersion)
		os.Exit(0)
	} else if *devEUI {
		rn2483.SetName(serialName)
		rn2483.SetBaud(57600)
		rn2483.SetTimeout(time.Millisecond * 500)
		rn2483.Connect()
		defer rn2483.Disconnect()
		log.Debug(rn2483.MacGetDeviceEUI())
		os.Exit(0)
	} else if *moduleReset {
		loraHardwareReset()
		os.Exit(0)
	}

	log.SetLevel(moduleconfig.LogLevel)

	loraHardwareReset()

	log.Info("starting lorawan")
	rn2483.SetName(serialName)
	rn2483.SetBaud(57600)
	rn2483.SetTimeout(time.Millisecond * 500)

	// Connect the RN2483 via serial
	rn2483.Connect()
	rn2483.MacSetDataRate(5)

	// Make sure the app closes the connection at the end the free the resource
	defer rn2483.Disconnect()

	err := rn2483.MacSetApplicationEUI(moduleconfig.LoRaWANApplicationEUI)
	if err != nil {
		log.Error(err)
	}
	err = rn2483.MacSetApplicationKey(moduleconfig.LoRaWANApplicationKey)
	if err != nil {
		log.Error(err)
	}
	log.Debug(rn2483.MacGetStatus())

	joined := rn2483.MacJoin(rn2483.OTAA)

	log.Debug(joined)

	if !joined {
		joinlora()
	}

	log.Info(joined)
	log.Debug(rn2483.MacGetStatus())

	go sendData(moduleconfig)

	select {}
}

func sendData(moduleconfig *config.Moduleconfig) {

	var data []byte

	tick := time.Tick(time.Duration(moduleconfig.LoRaWANSendInterval) * time.Second)

	for ; ; <-tick {
		log.Info(time.Now())
		callback := func(port uint8, data []byte) {
			log.Debugf("Received message on port %v: %s", port, string(data))
		}

		for i, sharedfile := range moduleconfig.LoRaWANSharedDirs {

			file, err := os.Open(sharedfile)
			utils.Checklog(err)
			defer file.Close()
			reader := csv.NewReader(file)
			reader.Comma = ';'
			records, err := reader.Read()
			// log.Debug("!!!!!")
			// log.Debug(records)
			utils.Checklog(err)

			for _, element := range strings.Split(moduleconfig.LoRaWANSharedFilesElements[i], "|") {

				// log.Debug(element)

				tmpElements := strings.Split(string(element), ":")
				// log.Debug(tmpElements)
				// log.Debug(len(tmpElements))
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

	Send:
		if len(data) > 0 {
			err := rn2483.MacTx(false, uint8(1), data, callback)
			if err != nil {

				log.Error("FEHLER: " + err.Error())

				if err == rn2483.ErrNotJoined {
					log.Info("Try to join")
					if joinlora() {
						goto Send
					}
				} else {
					log.Error("Error quit LoRaWAN-Service")
					os.Exit(1)
					// loraHardwareReset()
					// if joinlora() {
					// 	goto Send
					// }
				}

			}
			data = []byte{}
		} else {
			log.Info("Datalength = 0. No send required.")
		}

	}
}

// system("echo 496 > /sys/class/gpio/export");
// system("echo 497 > /sys/class/gpio/export");
// system("echo 498 > /sys/class/gpio/export");
// system("echo 499 > /sys/class/gpio/export");
// echo "out" > /sys/class/gpio/gpio496/direction
// echo 1 > /sys/class/gpio/gpio496/value
func loraHardwareReset() {
	sh := os.Getenv("SHELL")
	log.Info("lora module hardware reset")
	cmd := exec.Command(sh, "-c", "echo 496 > /sys/class/gpio/export")
	log.Debug(cmd)
	out, _ := cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	cmd = exec.Command(sh, "-c", "echo \"out\" > /sys/class/gpio/gpio496/direction")
	log.Debug(cmd)
	out, _ = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	cmd = exec.Command(sh, "-c", "echo 0 > /sys/class/gpio/gpio496/value")
	log.Debug(cmd)
	out, _ = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	time.Sleep(500 * time.Millisecond)
	cmd = exec.Command(sh, "-c", "echo 1 > /sys/class/gpio/gpio496/value")
	log.Debug(cmd)
	out, _ = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	time.Sleep(500 * time.Millisecond)
	log.Info("hardware reset done")
}

func joinlora() bool {

	jointry := 0
	resetcounter := 0

	for {

		log.Debug("wait for next join...")
		time.Sleep(60 * time.Second)
		log.Debug("try to join")
		joined := rn2483.MacJoin(rn2483.OTAA)
		log.Debug(joined)
		jointry++
		log.Debug("trial no: " + fmt.Sprintf("%v", jointry))

		if joined {
			break
		} else {

			if jointry == 10 {
				log.Debug("call reset")
				loraHardwareReset()
				jointry = 0
				resetcounter++
			}

			if resetcounter == 10 {
				log.Info("Sleep for a long time ...")
				time.Sleep(3600 * time.Second)
				log.Info("Wake up ...")
				loraHardwareReset()
				resetcounter = 0
			}

		}

	}
	return true
}

func getSc16is752GpioNumber() (int, error) {

	var gpionumber = 0
	var err error

	sh := os.Getenv("SHELL") //fetch default shell
	log.Debug(sh)
	cmd := exec.Command(sh, "-c", "ls -la /sys/class/gpio | grep 3f804000.i2c")
	log.Debug(cmd)
	out, _ := cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 {
			log.Debug(line)
			gpio := strings.Split(line, "gpio/gpio")
			log.Debug(gpio[len(gpio)-1])
			re := regexp.MustCompile("[0-9]+")
			gpionumberstring := re.FindString(gpio[len(gpio)-1])
			log.Debug(gpionumberstring)
			gpionumber, err = strconv.Atoi(gpionumberstring)
			if err != nil {
				return 0, err
			}
		}
	}
	return gpionumber, nil
}
