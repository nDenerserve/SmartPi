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

package smartpi

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	rpi "github.com/nDenerserve/go-rpigpio"
	"golang.org/x/exp/io/i2c"

	log "github.com/sirupsen/logrus"
)

const (
	ADE7878_ADDR int     = 0x38
	SAMPLES      int     = 100
	ade7878Clock float64 = 256000
	halfCircle   float64 = math.Pi / 180.0
)

var (
	calibrationCurveFactorCurrent = 1.0
	calibrationCurveFactorPower   = 1.0
	rms_factor_current            float64
)

var MainPhases = []models.Phase{models.PhaseA, models.PhaseB, models.PhaseC}

type Readings map[models.Phase]float64

type ADE7878Readout struct {
	Current           Readings
	Voltage           Readings
	ActiveWatts       Readings
	CosPhi            Readings
	Frequency         Readings
	ApparentPower     Readings
	ReactivePower     Readings
	PowerFactor       Readings
	ActiveEnergy      Readings
	Energyconsumption Readings
	Energyproduction  Readings
}

type CTFactors struct {
	CurrentResistor, CurrentClampFactor, CurrentRmsOffset, VoltageRmsOffset, PowerCorrectionFactor float64
}

var (
	CTTypes = map[string]CTFactors{
		"YHDC_SCT013": CTFactors{
			CurrentResistor:       7.5,
			CurrentClampFactor:    0.05,
			CurrentRmsOffset:      1.032,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.019413,
		},
		"YHDC_SCT4333QL": CTFactors{
			CurrentResistor:       7.5,
			CurrentClampFactor:    (1.0 / 60.0),
			CurrentRmsOffset:      1.032,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.019413,
		},
		"400A/033V": CTFactors{
			CurrentResistor:       1.0,
			CurrentClampFactor:    0.08325,
			CurrentRmsOffset:      1.010725941,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.042929856,
		},
		"X/1A": CTFactors{
			CurrentResistor:       0.33,
			CurrentClampFactor:    1.0,
			CurrentRmsOffset:      1.010725941,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.043861,
		},
		"X/5A": CTFactors{
			CurrentResistor:       0.068,
			CurrentClampFactor:    5.0,
			CurrentRmsOffset:      1.010725941,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.043861,
		},
		"600A/100mA": CTFactors{
			CurrentResistor:       3.75,
			CurrentClampFactor:    (0.1 / 6),
			CurrentRmsOffset:      1.032,
			VoltageRmsOffset:      1.0,
			PowerCorrectionFactor: 0.019413,
		},
	}
)

// Fetch a number of bytes from the device and convert it to an int.
func DeviceFetchInt(d *i2c.Device, l int, cmd []byte) int64 {
	startTime := time.Now()
	err := d.Write(cmd)
	if err != nil {
		panic(err)
	}
	data := make([]byte, l)
	err = d.Read(data)
	if err != nil {
		panic(err)
	}
	var result int64
	switch l {
	case 8:
		result = int64(binary.BigEndian.Uint64(data))
	case 4:
		result = int64(int32(binary.BigEndian.Uint32(data)))
	case 2:
		result = int64(int16(binary.BigEndian.Uint16(data)))
	default:
		panic(fmt.Errorf("Invalid byte length for int conversion %d", l))
	}
	log.Debugf("DeviceFetchInt: %s cmd: %x data: %x result: %d", time.Since(startTime), cmd, data, result)
	return result
}

func resetADE7878() {
	println("RESET")
	p, err := rpi.OpenPin(4, rpi.OUT)
	if err != nil {
		panic(err)
	}
	defer p.Close()
	p.Write(rpi.HIGH)
	time.Sleep(3 * time.Second)
	p.Write(rpi.LOW)
	time.Sleep(time.Second)
}

func initPiForADE7878() {
	/*
	   p, err := rpi.OpenPin(4, rpi.OUT)
	   if err != nil {
	       panic(err)
	   }
	   defer p.Close()
	   p.Write(rpi.HIGH)*/

}

func WriteRegister(d *i2c.Device, register string, data ...byte) (err error) {
	return d.Write(append(ADE7878REG[register], data...))
}

func InitADE7878(c *config.Config) (*i2c.Device, error) {

	// **** Opening and closing the i2c fixes an error that incorrect values are displayed on pahse 2 and 3 when the program is restarted.

	d, erre := i2c.Open(&i2c.Devfs{Dev: c.I2CDevice}, ADE7878_ADDR)
	if erre != nil {
		panic(erre)
	}

	time.Sleep(time.Second)

	erre = d.Close()
	if erre != nil {
		panic(erre)
	}

	// **** END of fix

	d, err := i2c.Open(&i2c.Devfs{Dev: c.I2CDevice}, ADE7878_ADDR)
	if err != nil {
		panic(err)
	}

	// 0xEC01 (CONFIG2-REGISTER)
	// 00000010 --> I2C-Lock
	//err = d.Write(append(ADE7878REG["CONFIG2"], 0x02))
	err = WriteRegister(d, "CONFIG2", 0x02)
	if err != nil {
		panic(err)
	}

	// 0xE1
	err = d.Write([]byte{0xEC})
	if err != nil {
		panic(err)
	}

	// Read i2cLock
	i2cLock := make([]byte, 1)
	err = d.Read(i2cLock)
	if err != nil {
		panic(err)
	}

	// 0xE7FE writeprotection
	err = d.Write([]byte{0xE7, 0xFE, 0xAD})
	if err != nil {
		panic(err)
	}

	// 0xE7E3 writeprotection OFF
	err = d.Write([]byte{0xE7, 0xE3, 0x00})
	if err != nil {
		panic(err)
	}

	// // 0x43B6 (HPFDIS-REGISTER)
	// err = d.Write(append(ADE7878REG["HPFDIS"], 0x00, 0x00, 0x00, 0x00})
	// if err != nil {
	//     panic(err)
	// }

	// 0x43B5 (DICOEFF-REGISTER)
	err = WriteRegister(d, "DICOEFF", 0xFF, 0xFF, 0x80, 0x00)
	if err != nil {
		panic(err)
	}

	if c.Integrator == true {

		// 0xE618 (CONFIG-REGISTER)
		err = WriteRegister(d, "CONFIG", 0x00, 0x01)
		if err != nil {
			panic(err)
		}

	}

	// Set the right power frequency to the COMPMODE-REGISTER.
	// 0xE60E (COMPMODE-REGISTER)
	if c.PowerFrequency == 60 {
		// 0x41FF 60Hz
		err = WriteRegister(d, "COMPMODE", 0x41, 0xFF)
	} else {
		// 0x01FF 50Hz
		err = WriteRegister(d, "COMPMODE", 0x01, 0xFF)
	}
	if err != nil {
		panic(err)
	}

	//0x43AB (WTHR1-REGISTER)
	err = WriteRegister(d, "WTHR1", 0x00, 0x00, 0x00, 0x17)
	if err != nil {
		panic(err)
	}

	//0x43AC (WTHR0-REGISTER)
	err = WriteRegister(d, "WTHR0", 0x00, 0x85, 0x60, 0x16)
	if err != nil {
		panic(err)
	}

	// // 0x43AD (VARTHR1-REGISTER)
	// err = d.Write(append(ADE7878REG["VARTHR1"], 0x17, 0x85, 0x60, 0x16))
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43AE (VARTHR0-REGISTER)
	// err = d.Write(append(ADE7878REG["VARTHR0"], 0x17, 0x85, 0x60, 0x16))
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43A9 (VATHR1-REGISTER)
	// err = d.Write(append(ADE7878REG["VATHR1"], 0x17, 0x85, 0x60, 0x16))
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43AA (VATHR0-REGISTER)
	// err = d.Write(append(ADE7878REG["VATHR0"], 0x17, 0x85, 0x60, 0x16))
	// if err != nil {
	//     panic(err)
	// }

	// 0x43B3 (VLEVEL-REGISTER)
	err = WriteRegister(d, "VLEVEL", 0x00, 0x0C, 0xEC, 0x85)
	if err != nil {
		panic(err)
	}

	time.Sleep(875 * time.Millisecond)

	// // 0x4381 (AVGAIN-REGISTER)
	// outcome := DeviceFetchInt(d, 4, ADE7878REG["AVGAIN"])
	// fmt.Printf("AVGAIN-REGISTER VORHER%g   %x %x %x %x \n", outcome, data[0], data[1], data[2], data[3])

	// 0x4381 (AVGAIN-REGISTER)
	err = WriteRegister(d, "AVGAIN", 0xFF, 0xFC, 0x1C, 0xC2)
	if err != nil {
		panic(err)
	}

	// 0x4383 (BVGAIN-REGISTER)
	err = WriteRegister(d, "BVGAIN", 0xFF, 0xFB, 0xCA, 0x60)
	// err = WriteRegister(d, "BVGAIN", 0xFF, 0xFC, 0x1C, 0xC2)
	if err != nil {
		panic(err)
	}

	// 0x4385 (CVGAIN-REGISTER)
	err = WriteRegister(d, "CVGAIN", 0xFF, 0xFC, 0x12, 0xDE)
	// err = WriteRegister(d, "CVGAIN", 0xFF, 0xFC, 0x1C, 0xC2)
	if err != nil {
		panic(err)
	}

	// err = WriteRegister(d, "AIRMSOS", 0x11, 0x47, 0xE9)
	// if err != nil {
	// 	panic(err)
	// }

	err = WriteRegister(d, "AIRMSOS", 0x0F, 0xFD, 0x40, 0xE0)
	if err != nil {
		panic(err)
	}

	err = WriteRegister(d, "BIRMSOS", 0x0F, 0xFD, 0x40, 0xE0)
	if err != nil {
		panic(err)
	}

	err = WriteRegister(d, "CIRMSOS", 0x0F, 0xFD, 0x40, 0xE0)
	if err != nil {
		panic(err)
	}

	err = WriteRegister(d, "NIRMSOS", 0x0F, 0xFD, 0x40, 0xE0)
	if err != nil {
		panic(err)
	}

	// Line cycle mode
	// 0xE702 LCYCMODE
	err = WriteRegister(d, "LCYCMODE", 0x0F)
	if err != nil {
		panic(err)
	}

	// Line cycle mode count
	// 0xE60C LINECYC
	err = WriteRegister(d, "LINECYC", 0xC8)
	if err != nil {
		panic(err)
	}

	// 0xE7FE writeprotection
	err = d.Write([]byte{0xE7, 0xFE, 0xAD})
	if err != nil {
		panic(err)
	}

	// 0xE7E3 writeprotection
	err = d.Write([]byte{0xE7, 0xE3, 0x80})
	if err != nil {
		panic(err)
	}

	// 0xE228 (RUN-Register)
	err = WriteRegister(d, "RUN", 0x00, 0x01)
	if err != nil {
		panic(err)
	}

	return d, nil
}

func ReadCurrent(d *i2c.Device, c *config.Config, phase models.Phase) (current float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = ADE7878REG["AIRMS"] // 0x43C0 (AIRMS; Current rms an A)
	case models.PhaseB:
		command = ADE7878REG["BIRMS"] // 0x43C2 (AIRMS; Current rms an B)
	case models.PhaseC:
		command = ADE7878REG["CIRMS"] // 0x43C4 (AIRMS; Current rms an C)
	case models.PhaseN:
		command = ADE7878REG["NIRMS"] // 0x43C6 (AIRMS; Current rms an N)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	var rmsFactor float64
	switch c.PowerFrequency {
	case 60:
		rmsFactor = 3493258.0 // 60Hz
	case 50:
		rmsFactor = 4191910.0 // 50Hz
	default:
		panic(fmt.Errorf("Invalid frequency %g", c.PowerFrequency))
	}

	if c.MeasureCurrent[phase] {
		outcome := float64(DeviceFetchInt(d, 4, command))
		cr := CTTypes[c.CTType[phase]].CurrentResistor
		var ccf float64
		if c.CTType[phase] == "YHDC_SCT013" || c.CTType[phase] == "YHDC_SCT4333QL" || c.CTType[phase] == "400A/033V" || c.CTType[phase] == "600A/100mA" {
			ccf = CTTypes[c.CTType[phase]].CurrentClampFactor
		} else {
			ccf = CTTypes[c.CTType[phase]].CurrentClampFactor / (float64(c.CTTypePrimaryCurrent[phase]) / 100.0)
		}
		oc := CTTypes[c.CTType[phase]].CurrentRmsOffset
		// outcome = outcome - 7300
		current = (((((outcome * 0.3535) / rmsFactor) / cr) / ccf) * 100.0 * oc * c.CalibrationfactorI[phase])

		//calibration curves
		// if c.CTType[phase] == "YHDC_SCT013" {
		// 	if math.Abs(current) < 70.0 {
		// 		calibrationCurveFactorCurrent = 0.004591484821337*math.Log1p(current-1) + 0.982318924023796
		// 	} else {
		// 		calibrationCurveFactorCurrent = 1.00246306639231
		// 	}
		// } else if c.CTType[phase] == "400A/033V" {
		// 	if math.Abs(current) < 100.0 {
		// 		calibrationCurveFactorCurrent = -0.002946348708069*math.Log1p(current-1) + 1.04752234441994
		// 	} else {
		// 		calibrationCurveFactorCurrent = 1.035
		// 	}
		// } else {
		// 	calibrationCurveFactorCurrent = 1.0
		// }
		current = current * calibrationCurveFactorCurrent
	} else {
		current = 0.0
	}
	return current
}

func ReadVoltage(d *i2c.Device, c *config.Config, phase models.Phase) (voltage float64, measureVoltage bool) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0x43, 0xC1} // 0x43C1 (AVRMS; Voltage RMS phase A)
	case models.PhaseB:
		command = []byte{0x43, 0xC3} // 0x43C3 (BVRMS; Voltage RMS phase B)
	case models.PhaseC:
		command = []byte{0x43, 0xC5} // 0x43C5 (BVRMS; Voltage RMS phase C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}
	// fmt.Println("CalibrationfactorU: ", phase, "  ", c.CalibrationfactorU[phase])
	voltage = (float64(DeviceFetchInt(d, 4, command)) / 1e+4) * c.CalibrationfactorU[phase]

	measureVoltage = true
	if !c.MeasureVoltage[phase] {
		voltage = c.Voltage[phase]
		measureVoltage = false
	}

	return voltage, measureVoltage
}

func ReadActiveWatts(d *i2c.Device, c *config.Config, phase models.Phase) (watts float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE5, 0x13} // 0xE513 (AWATT total active power phase A)
	case models.PhaseB:
		command = []byte{0xE5, 0x14} // 0xE514 (BWATT total active power phase B)
	case models.PhaseC:
		command = []byte{0xE5, 0x15} // 0xE515 (CWATT total active power phase C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	var pcf float64
	if c.CTType[phase] == "YHDC_SCT013" || c.CTType[phase] == "YHDC_SCT4333QL" {
		pcf = 1.0
	} else if c.CTType[phase] == "400A/033V" {
		pcf = 0.5
	} else {
		pcf = 200.0 / (float64(c.CTTypePrimaryCurrent[phase]))
	}

	outcome := float64(DeviceFetchInt(d, 4, command))
	if c.MeasureCurrent[phase] {
		watts = outcome * calibrationCurveFactorCurrent * CTTypes[c.CTType[phase]].PowerCorrectionFactor / pcf
	} else {
		watts = 0.0
	}
	if c.CurrentDirection[phase] {
		watts *= -1
	}

	//calibration curves
	// if c.CTType[phase] == "YHDC_SCT013" {

	// 	if math.Abs(watts) < 18000.0 {
	// 		calibrationCurveFactorPower = 0.001855740652181*math.Log1p(watts-1) + 1.01456039967075
	// 		fmt.Print("CalibrationCurve: ")
	// 		fmt.Println(calibrationCurveFactorPower)
	// 	} else {
	// 		calibrationCurveFactorPower = 0.994460375354525
	// 		fmt.Print("CalibrationCurve2: ")
	// 		fmt.Println(calibrationCurveFactorPower)
	// 	}

	// } else if c.CTType[phase] == "400A/033V" {
	// 	if math.Abs(watts) < 530000.0 {
	// 		calibrationCurveFactorPower = -0.004876896009306*math.Log1p(watts-1) + 1.07274818852909
	// 	} else {
	// 		calibrationCurveFactorPower = 1.01983381
	// 	}
	// } else {
	// 	calibrationCurveFactorPower = 1.0
	// }

	watts = watts * calibrationCurveFactorPower

	return watts
}

func ReadActiveEnergy(d *i2c.Device, c *config.Config, phase models.Phase) (energy float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE4, 0x00} // 0xE4000 (AWATTHR total active energy phase A)
	case models.PhaseB:
		command = []byte{0xE4, 0x00} // 0xE4001 (BWATTHR total active energy phase B)
	case models.PhaseC:
		command = []byte{0xE4, 0x00} // 0xE4002 (CWATTHR total active energy phase C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	var pcf float64
	if c.CTType[phase] == "YHDC_SCT013" || c.CTType[phase] == "YHDC_SCT4333QL" {
		pcf = 1.0
	} else if c.CTType[phase] == "400A/033V" {
		pcf = 0.5
	} else {
		pcf = 200.0 / (float64(c.CTTypePrimaryCurrent[phase]))
	}

	outcome := float64(DeviceFetchInt(d, 4, command))

	energy = outcome * calibrationCurveFactorCurrent * calibrationCurveFactorPower / pcf

	// if c.CurrentDirection[phase] {
	// 	watts *= -1
	// }

	return energy
}

func ReadAngle(d *i2c.Device, c *config.Config, phase models.Phase) (angle float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE6, 0x01} // 0xE601 (ANGLE0 cosphi an A)
	case models.PhaseB:
		command = []byte{0xE6, 0x02} // 0xE602 (ANGLE1 cosphi an B)
	case models.PhaseC:
		command = []byte{0xE6, 0x03} // 0xE603 (ANGLE2 cosphi an C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	if c.MeasureVoltage[phase] {
		outcome := float64(DeviceFetchInt(d, 2, command))
		angle = math.Cos(outcome * 360 * c.PowerFrequency / ade7878Clock * halfCircle)
		if c.CurrentDirection[phase] {
			angle *= -1
		}
	} else {
		angle = 1.0
	}

	return angle
}

func ReadFrequency(d *i2c.Device, c *config.Config, phase models.Phase) (frequency float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE7, 0x00, 0x1C} // 0xE7001C MMODE-Register measure frequency at VA
	case models.PhaseB:
		command = []byte{0xE7, 0x00, 0x1D} // 0xE7001D MMODE-Register measure frequency at VB
	case models.PhaseC:
		command = []byte{0xE7, 0x00, 0x1E} // 0xE7001E MMODE-Register measure frequency at VC
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	err := d.Write(command) // MMODE-Register measure frequency
	if err != nil {
		panic(err)
	}
	// Make sure we capture 3 full cycles at ~50Hz, 4 cycles at ~60Hz.
	time.Sleep(70 * time.Millisecond)
	// 0xE607 (PERIOD)
	outcome := float64(DeviceFetchInt(d, 2, []byte{0xE6, 0x07}))
	frequency = ade7878Clock / (outcome + 1)

	return frequency
}

func ReadApparentPower(d *i2c.Device, c *config.Config, phase models.Phase) float64 {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE5, 0x19} // 0xE519 (AVA total apparent power phase A)
	case models.PhaseB:
		command = []byte{0xE5, 0x1A} // 0xE51A (BVA total apparent power phase B)
	case models.PhaseC:
		command = []byte{0xE5, 0x1B} // 0xE51B (CVA total apparent power phase C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	var pcf float64
	if c.CTType[phase] == "YHDC_SCT013" || c.CTType[phase] == "YHDC_SCT4333QL" {
		pcf = 1.0
	} else if c.CTType[phase] == "400A/033V" {
		pcf = 0.5
	} else {
		pcf = 200.0 / (float64(c.CTTypePrimaryCurrent[phase]))
	}

	if c.MeasureCurrent[phase] {
		outcome := float64(DeviceFetchInt(d, 4, command))
		return outcome * calibrationCurveFactorCurrent * calibrationCurveFactorPower * CTTypes[c.CTType[phase]].PowerCorrectionFactor / pcf * 1.14989234
	} else {
		return 0.0
	}
}

func ReadReactivePower(d *i2c.Device, c *config.Config, phase models.Phase) (rewatts float64) {
	command := make([]byte, 2)
	switch phase {
	case models.PhaseA:
		command = []byte{0xE5, 0x16} // 0xE516 (AVAR total reactive power phase A)
	case models.PhaseB:
		command = []byte{0xE5, 0x17} // 0xE517 (AVAR total reactive power phase B)
	case models.PhaseC:
		command = []byte{0xE5, 0x18} // 0xE518 (AVAR total reactive power phase C)
	default:
		panic(fmt.Errorf("Invalid phase %q", phase))
	}

	var pcf float64
	if c.CTType[phase] == "YHDC_SCT013" || c.CTType[phase] == "YHDC_SCT4333QL" {
		pcf = 1.0
	} else if c.CTType[phase] == "400A/033V" {
		pcf = 0.5
	} else {
		pcf = 200.0 / (float64(c.CTTypePrimaryCurrent[phase]))
	}

	outcome := float64(DeviceFetchInt(d, 4, command))
	if c.MeasureCurrent[phase] {
		rewatts = outcome * calibrationCurveFactorCurrent * calibrationCurveFactorPower * CTTypes[c.CTType[phase]].PowerCorrectionFactor / pcf * 2.560177029
	} else {
		rewatts = 0.0
	}
	if c.CurrentDirection[phase] {
		rewatts *= -1
	}

	return rewatts
}

func CalculatePowerFactor(c *config.Config, phase models.Phase, watts float64, voltAmps float64, voltAmpsReactive float64) float64 {

	powerFactor := watts / CTTypes[c.CTType[phase]].PowerCorrectionFactor / voltAmps

	if c.MeasureCurrent[phase] {
		if math.Signbit(voltAmpsReactive) {
			return powerFactor
		} else {
			return powerFactor * -1
		}
	} else {
		return 0.0
	}
}

func ReadPhase(d *i2c.Device, c *config.Config, p models.Phase, measureFrequency bool, v *ADE7878Readout) {
	startTime := time.Now()

	// Measure current.
	v.Current[p] = ReadCurrent(d, c, p)

	// Neutral phase has no other updates.
	if p == models.PhaseN {
		logLine := fmt.Sprintf("ReadValues: %s phase: %s", time.Since(startTime), p)
		logLine += fmt.Sprintf("I: %g", v.Current[p])
		log.Debug(logLine)
		return
	}

	// Measure voltage.
	var measureVoltage bool
	v.Voltage[p], measureVoltage = ReadVoltage(d, c, p)

	// Measure active watts.
	if measureVoltage {
		v.ActiveWatts[p] = ReadActiveWatts(d, c, p)
	} else {
		v.ActiveWatts[p] = v.Current[p] * v.Voltage[p]
	}

	// Measure cosphi.
	v.CosPhi[p] = ReadAngle(d, c, p)

	// Measure apparent power (volt-amps).
	v.ApparentPower[p] = ReadApparentPower(d, c, p)

	// Measure reactive power (volt-ampere reactive).
	v.ReactivePower[p] = ReadReactivePower(d, c, p)

	// Measure active energy.
	v.ActiveEnergy[p] = ReadActiveEnergy(d, c, p)

	// Measure frequency.
	if measureFrequency {
		v.Frequency[p] = ReadFrequency(d, c, p)
	}

	// Calculate power factor.
	v.PowerFactor[p] = CalculatePowerFactor(c, p, v.ActiveWatts[p], v.ApparentPower[p], v.ReactivePower[p])

	logLine := fmt.Sprintf("ReadValues: %s phase: %s", time.Since(startTime), p)
	logLine += fmt.Sprintf("I: %g  V: %g  P: %g ", v.Current[p], v.Voltage[p], v.ActiveWatts[p])
	logLine += fmt.Sprintf("COS: %g  F: %g  VA: %g  ", v.CosPhi[p], v.Frequency[p], v.ApparentPower[p])
	logLine += fmt.Sprintf("VAR: %g  PF: %g  WATTHR: %g  ", v.ReactivePower[p], v.PowerFactor[p], v.ActiveEnergy[p])
	logLine += fmt.Sprintf("calibrationCurveFactorCurrent: %g", calibrationCurveFactorCurrent)
	log.Debug(logLine)
}
