package smartpidc

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/sandertv/go-formula/v2"
	log "github.com/sirupsen/logrus"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (

	// I2C address for MCP3424 - base address for MCP3424
	// MCP3424_ADDRESS uint16 = 0x68
	// fields in configuration register
	MCP342X_GAIN_FIELD   byte    = 0x03 // PGA field
	MCP342X_GAIN_X1      byte    = 0x00 // PGA gain X1
	MCP342X_GAIN_X2      byte    = 0x01 // PGA gain X2
	MCP342X_GAIN_X4      byte    = 0x02 // PGA gain X4
	MCP342X_GAIN_X8      byte    = 0x03 // PGA gain X8
	MCP342X_RES_FIELD    byte    = 0x0C // resolution/rate field
	MCP342X_RES_SHIFT    byte    = 2    // shift to low bits
	MCP342X_12_BIT       byte    = 0x00 // 12-bit 240 SPS
	MCP342X_14_BIT       byte    = 0x04 // 14-bit 60 SPS
	MCP342X_16_BIT       byte    = 0x08 // 16-bit 15 SPS
	MCP342X_18_BIT       byte    = 0x0C // 18-bit 3.75 SPS
	MCP342X_CONTINUOUS   byte    = 0x10 // 1 = continuous, 0 = one-shot
	MCP342X_CHAN_FIELD   byte    = 0x60 // channel field
	MCP342X_CHANNEL_1    byte    = 0x00 // select MUX channel 1
	MCP342X_CHANNEL_2    byte    = 0x20 // select MUX channel 2
	MCP342X_CHANNEL_3    byte    = 0x40 // select MUX channel 3
	MCP342X_CHANNEL_4    byte    = 0x60 // select MUX channel 4
	MCP342X_START        byte    = 0x80 // write: start a conversion
	MCP342X_BUSY         byte    = 0x80 // read: output not ready
	MCP342X_MODE_CONT    byte    = 0x10 // Continuous Conversion Mode
	MCP342X_MODE_ONESHOT byte    = 0x00 // One-Shot Conversion Mode
	MCP342X_READ_CNVRSN  byte    = 0x00 // Read Conversion Result Data
	RESOLUTION_12BIT     float64 = 12
	RESOLUTION_14BIT     float64 = 14
	RESOLUTION_16BIT     float64 = 16
	RESOLUTION_18BIT     float64 = 18
)

type InputFactors struct {
	MeasurementType          string
	SensorType               string
	Calculation              string
	ConversionFactor         float64
	MaxVal                   float64
	CalibrationTargetVoltage float64
	LowerLimit               float64
	UpperLimit               float64
	Unit                     string
}

var (
	InputTypes = map[string]InputFactors{
		"HSTS016L 10A": {
			MeasurementType:          "Current",
			SensorType:               "Hall",
			Calculation:              "((v-2.5)*10)/0.625",
			ConversionFactor:         368.151,
			MaxVal:                   300,
			CalibrationTargetVoltage: 2.5,
			LowerLimit:               1.86,
			UpperLimit:               3.14,
			Unit:                     "A",
		},
		"HSTS016L 30A": {
			MeasurementType:          "Current",
			SensorType:               "Hall",
			Calculation:              "((v-2.5)*30)/0.625",
			ConversionFactor:         368.151,
			MaxVal:                   300,
			CalibrationTargetVoltage: 2.5,
			LowerLimit:               1.86,
			UpperLimit:               3.14,
			Unit:                     "A",
		},
		"HSTS016L 50A": {
			MeasurementType:          "Current",
			SensorType:               "Hall",
			Calculation:              "((v-2.5)*50)/0.625",
			ConversionFactor:         368.151,
			MaxVal:                   300,
			CalibrationTargetVoltage: 2.5,
			LowerLimit:               1.86,
			UpperLimit:               3.14,
			Unit:                     "A",
		},
		"HSTS016L 200A": {
			MeasurementType:          "Current",
			SensorType:               "Hall",
			Calculation:              "((v-2.5)*200)/0.625",
			ConversionFactor:         368.151,
			MaxVal:                   200,
			CalibrationTargetVoltage: 2.5,
			LowerLimit:               1.86,
			UpperLimit:               3.14,
			Unit:                     "A",
		},
		"HSTS016L 300A": {
			MeasurementType:          "Current",
			SensorType:               "Hall",
			Calculation:              "((v-2.5)*300)/0.625",
			ConversionFactor:         368.151,
			MaxVal:                   300,
			CalibrationTargetVoltage: 2.5,
			LowerLimit:               1.86,
			UpperLimit:               3.14,
			Unit:                     "A",
		},
		"Voltage 0-5V": {
			MeasurementType: "Voltage",
			SensorType:      "Direct",
			Calculation:     "v",
			// ConversionFactor:         390.022321429,
			ConversionFactor:         368.151,
			MaxVal:                   5,
			CalibrationTargetVoltage: 0,
			LowerLimit:               0.0,
			UpperLimit:               5.001,
			Unit:                     "V",
		},
		"Voltage 0-60V": {
			MeasurementType:          "Voltage",
			SensorType:               "Direct",
			Calculation:              "v",
			ConversionFactor:         31.97674419,
			MaxVal:                   60,
			CalibrationTargetVoltage: 0,
			LowerLimit:               0.0,
			UpperLimit:               62.0,
			Unit:                     "V",
		},
		"Voltage 0-120V": {
			MeasurementType: "Voltage",
			SensorType:      "Direct",
			Calculation:     "v",
			// ConversionFactor:         16.38703529,
			ConversionFactor:         16.9125,
			MaxVal:                   120,
			CalibrationTargetVoltage: 0,
			LowerLimit:               0.0,
			UpperLimit:               122.0,
			Unit:                     "V",
		},
	}
)

func readoutDevice(i2cBus string, c *config.DCconfig) ([4]float64, error) {

	var ret [4]float64
	var write []byte

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
		return ret, err
	}
	defer bus.Close()

	log.Debug("i2c-address: " + string(c.ADCAddress[0]))

	dev := i2c.Dev{Bus: bus, Addr: uint16(c.ADCAddress[0])}

	for i := 0; i < 4; i++ {

		switch i {
		case 0:
			write = []byte{(MCP342X_CONTINUOUS | MCP342X_CHANNEL_1 | MCP342X_16_BIT | MCP342X_GAIN_X1)}
		case 1:
			write = []byte{(MCP342X_CONTINUOUS | MCP342X_CHANNEL_2 | MCP342X_16_BIT | MCP342X_GAIN_X1)}
		case 2:
			write = []byte{(MCP342X_CONTINUOUS | MCP342X_CHANNEL_3 | MCP342X_16_BIT | MCP342X_GAIN_X1)}
		case 3:
			write = []byte{(MCP342X_CONTINUOUS | MCP342X_CHANNEL_4 | MCP342X_16_BIT | MCP342X_GAIN_X1)}
		}

		// fmt.Printf("%b\n", write)

		read := make([]byte, 3)

		dev.Write(write)

		time.Sleep(100 * time.Millisecond)

		if err := dev.Tx(write, read); err != nil {
			log.Error(err)
			return ret, err
		}

		fmt.Printf("%v\n", read[0:])
		log.Debug("Raw-Input: " + strconv.Itoa(i) + ": " + strconv.FormatFloat(float64(int(read[0])*256+int(read[1]))/RESOLUTION_16BIT, 'f', -1, 64))

		ret[i] = float64(int(read[0])*256+int(read[1])) / RESOLUTION_16BIT

		if math.Abs(ret[i]) > 2050 {
			// ret[i] = math.NaN()
			ret[i] = 0
		}

	}

	return ret, nil
}

func CalibrateInput(c *config.DCconfig, input int) {

	rawValues, err := readoutDevice(c.I2CDevice, c)
	if err != nil {
		log.Error(err)

	}

	for i := 0; i < len(rawValues); i++ {
		if i == (input - 1) {
			log.Info("Calibrate input: " + strconv.Itoa(input))

			log.Debug("--------------------------------------------------------------------------------")
			log.Debug("Input-Type " + strconv.Itoa(i) + ": " + c.InputType[i] + "     Input-Formular " + strconv.Itoa(i) + ": " + InputTypes[c.InputType[i]].Calculation)
			log.Debug("Value " + strconv.Itoa(i) + ": " + strconv.FormatFloat(rawValues[i], 'f', -1, 64))

			calibrationOffset := (rawValues[i] / InputTypes[c.InputType[i]].ConversionFactor) - InputTypes[c.InputType[i]].CalibrationTargetVoltage

			log.Info("CalibrationOffset: " + strconv.FormatFloat(calibrationOffset, 'f', -1, 64))

			c.InputCalibrationOffset[i] = calibrationOffset
			c.SaveDCParameterToFile()
		}
	}

}

func GetValues(c *config.DCconfig) [4]float64 {

	var ret [4]float64
	var calibratedVal float64

	rawValues, err := readoutDevice(c.I2CDevice, c)

	if err != nil {
		log.Error(err)

	}

	for i := 0; i < len(rawValues); i++ {

		rawValues[i] = rawValues[i] / InputTypes[c.InputType[i]].ConversionFactor
		calibratedVal = rawValues[i] - c.InputCalibrationOffset[i]

		log.Debug("--------------------------------------------------------------------------------")
		log.Debug("Input-Type " + strconv.Itoa(i) + ": " + c.InputType[i] + "          Input-Formular " + strconv.Itoa(i) + ": " + InputTypes[c.InputType[i]].Calculation)
		log.Debug("Value " + strconv.Itoa(i) + ": " + strconv.FormatFloat(rawValues[i], 'f', -1, 64) + "          Calibrated Value " + strconv.Itoa(i) + ": " + strconv.FormatFloat(calibratedVal, 'f', -1, 64))

		if calibratedVal >= InputTypes[c.InputType[i]].LowerLimit && calibratedVal <= InputTypes[c.InputType[i]].UpperLimit {

			f, err := formula.New(InputTypes[c.InputType[i]].Calculation)
			if err != nil {
				log.Print(err)

			}
			v := formula.Var("v", calibratedVal)
			ret[i] = f.MustEval(v)

		} else {
			ret[i] = 0
			// ret[i] = math.NaN()
		}

		log.Debug("Formula output " + strconv.Itoa(i) + ": " + strconv.FormatFloat(ret[i], 'f', -1, 64))
		log.Debug("--------------------------------------------------------------------------------")
	}

	return ret

}
