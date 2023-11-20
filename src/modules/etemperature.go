package main

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {

	moduleconfig := config.NewModuleconfig()

	log.SetLevel(moduleconfig.LogLevel)

	log.Info("Start")

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
