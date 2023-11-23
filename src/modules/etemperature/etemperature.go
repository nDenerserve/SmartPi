package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {

	moduleconfig := config.NewModuleconfig()

	log.SetLevel(moduleconfig.LogLevel)

	log.Info("Start etemperature")

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open(moduleconfig.I2CDevice)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	dev := i2c.Dev{Bus: bus, Addr: moduleconfig.EtemperatureI2CAddress}

	pollTemperature(dev, moduleconfig)

}

func pollTemperature(dev i2c.Dev, moduleconfig *config.Moduleconfig) {

	maxRetry := 3
	hasError := false
	temperatures := make([]float64, 16)

	tick := time.Tick(time.Duration(60000/moduleconfig.EtemperatureSamplerate) * time.Millisecond)

	log.Debugf("Tick: %v", tick)

	for {

		hasError = false

		startTime := time.Now()

		for i := 1; i <= 16; i++ {
			retry := 0

		READTEMP:
			time.Sleep(10 * time.Millisecond)
			temp, err := readTemperature(uint8(i), dev)
			if err != nil && retry < maxRetry {
				log.Debug(err)
				retry++
				time.Sleep(50 * time.Millisecond)
				log.Debugf("Retry: %v", retry)
				goto READTEMP
			} else if err != nil {
				hasError = true
				log.Debugf("Out: %v", hasError)
				resetDevice(dev)
			}
			if (temp < -500.0) || (temp > 10000.0) || (temp == 0.000000) {
				temp = math.NaN()
			}
			temperatures[i-1] = temp
		}

		log.Debugf("Temperaturen: %f", temperatures)

		if hasError == false {
			log.Debug("FILE WRITEING")
			if moduleconfig.EtemperatureSharedFileEnabled {
				writeSharedTemperaturefile(moduleconfig, temperatures)
			}
		} else {
			hasError = false
		}

		delay := time.Since(startTime) - (time.Duration(60000/moduleconfig.EtemperatureSamplerate) * time.Millisecond)
		if int64(delay) > 0 {
			log.Errorf("Readout delayed: %s", delay)
		}
		<-tick

	}
}

func resetDevice(dev i2c.Dev) {

	output_buffer := make([]byte, 1)
	output_buffer[0] = byte(255)

	_, err := dev.Write(output_buffer)
	if err != nil {
		log.Debug(err)
	}
	time.Sleep(300 * time.Millisecond)
}

func readTemperature(input uint8, dev i2c.Dev) (float64, error) {

	output_buffer := make([]byte, 1)
	raw := make([]byte, 4)

	output_buffer[0] = byte(input)

	err := dev.Tx(output_buffer, raw)
	if err != nil {
		log.Debug(err)
		return 0.0, err
	}

	temperature := math.Float32frombits(binary.LittleEndian.Uint32(raw))

	log.Debugf("Temperature input %v: %v", input, temperature)

	return float64(temperature), nil
}

func writeSharedTemperaturefile(c *config.Moduleconfig, values []float64) {
	var f *os.File
	var err error

	t := time.Now()
	timeStamp := t.Format("2006-01-02 15:04:05")
	logLine := "## Shared File Update ## "
	logLine += fmt.Sprintf(timeStamp)
	logLine += fmt.Sprintf(" T1: %f  T2: %f  T3: %f  T4: %f  T5: %f  T6: %f  T7: %f  T8: %f  T9: %f  T10: %f  T11: %f  T12: %f  T13: %f  T14: %f  T15: %f  T16: %f  ", values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7], values[8], values[9], values[10], values[11], values[12], values[13], values[14], values[15])
	log.Info(logLine)
	sharedFile := filepath.Join(c.EtemperatureSharedDir, c.EtemperatureSharedFile)
	if _, err = os.Stat(sharedFile); os.IsNotExist(err) {
		os.MkdirAll(c.EtemperatureSharedDir, 0777)
		f, err = os.Create(sharedFile)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(sharedFile, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()
	_, err = f.WriteString(timeStamp + ";" + utils.Float64ArrayToString(values, ";") + ";")
	if err != nil {
		panic(err)
	}
	f.Close()
}
