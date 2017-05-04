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
	"github.com/nathan-osman/go-rpigpio"
	"golang.org/x/exp/io/i2c"
	"math"
	"time"
)

const (
	ADE7878_ADDR              int     = 0x38
	SAMPLES                   int     = 100
	ADE7878_CLOCK             float32 = 256000
	FACTOR_CIRCLE             float32 = 360
	VAL                       float32 = math.Pi / 180.0
	RMS_FACTOR_VOLTAGE        float32 = 2427873
	CURRENT_RESISTOR_A        float32 = 7.07107
	CURRENT_RESISTOR_B        float32 = 7.07107
	CURRENT_RESISTOR_C        float32 = 7.07107
	CURRENT_RESISTOR_N        float32 = 7.07107
	CURRENT_CLAMP_FACTOR_A    float32 = 0.05
	CURRENT_CLAMP_FACTOR_B    float32 = 0.05
	CURRENT_CLAMP_FACTOR_C    float32 = 0.05
	CURRENT_CLAMP_FACTOR_N    float32 = 0.05
	OFFSET_CURRENT_A          float32 = 0.97129167
	OFFSET_CURRENT_B          float32 = 0.97129167
	OFFSET_CURRENT_C          float32 = 0.97129167
	OFFSET_CURRENT_N          float32 = 0.97129167
	OFFSET_VOLTAGE_A          float32 = 1.0
	OFFSET_VOLTAGE_B          float32 = 1.0
	OFFSET_VOLTAGE_C          float32 = 1.0
	POWER_CORRECTION_FACTOR_A float32 = 0.019413
	POWER_CORRECTION_FACTOR_B float32 = 0.019413
	POWER_CORRECTION_FACTOR_C float32 = 0.019413
)

var (
	rms_factor_current float32
)

// Fetch a number of bytes from the device and convert it to an int.
func DeviceFetchInt(d *i2c.Device, l int, cmd []byte) int64 {
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
	// fmt.Printf("DeviceFetchInt: cmd: %x data: %x result: %d\n", cmd, data, result)
	return result
}

func resetADE7878() {
	println("RESET")
	p, err := rpi.OpenPin(4, rpi.OUT)
	if err != nil {
		panic(err)
	}
	defer p.Close()
	p.Write(rpi.LOW)
	time.Sleep(time.Second)
	p.Write(rpi.HIGH)
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

func InitADE7878(c *Config) (*i2c.Device, error) {
	d, err := i2c.Open(&i2c.Devfs{Dev: c.I2CDevice}, ADE7878_ADDR)
	if err != nil {
		panic(err)
	}

	// 0xEC01 (CONFIG2-REGISTER)
	// 00000010 --> Bedeutet I2C-Lock (I2C ist nun die gewählte Übertragungsart)
	err = d.Write([]byte{0xEC, 0x01, 0x02})
	if err != nil {
		panic(err)
	}

	// 0xE1 (CONFIG2-REGISTER)
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
	// err = d.Write([]byte{0x43, 0xB6, 0x00, 0x00, 0x00, 0x00})
	// if err != nil {
	//     panic(err)
	// }

	// Set the right power frequency to the COMPMODE-REGISTER.
	// 0xE60E (COMPMODE-REGISTER)
	if c.PowerFrequency == 60 {
		// 0x41FF 60Hz
		err = d.Write([]byte{0xE6, 0x0E, 0x41, 0xFF})
	} else {
		// 0x01FF 50Hz
		err = d.Write([]byte{0xE6, 0x0E, 0x01, 0xFF})
	}
	if err != nil {
		panic(err)
	}

	// 0x43B5 (DICOEFF-REGISTER)
	err = d.Write([]byte{0x43, 0xB5, 0xFF, 0x80, 0x00})
	if err != nil {
		panic(err)
	}

	//0x43AB (WTHR1-REGISTER)
	err = d.Write([]byte{0x43, 0xAB, 0x00, 0x00, 0x00, 0x17})
	if err != nil {
		panic(err)
	}

	//0x43AC (WTHR0-REGISTER)
	err = d.Write([]byte{0x43, 0xAC, 0x00, 0x85, 0x60, 0x16})
	if err != nil {
		panic(err)
	}

	// // 0x43AD (VARTHR1-REGISTER)
	// err = d.Write([]byte{0x43, 0xAD, 0x17, 0x85, 0x60, 0x16})
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43AE (VARTHR0-REGISTER)
	// err = d.Write([]byte{0x43, 0xAE, 0x17, 0x85, 0x60, 0x16})
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43A9 (VATHR1-REGISTER)
	// err = d.Write([]byte{0x43, 0xA9, 0x17, 0x85, 0x60, 0x16})
	// if err != nil {
	//     panic(err)
	// }
	//
	// // 0x43AA (VARTHR0-REGISTER)
	// err = d.Write([]byte{0x43, 0xAA, 0x17, 0x85, 0x60, 0x16})
	// if err != nil {
	//     panic(err)
	// }

	// 0x43B3 (VLEVEL-REGISTER)
	err = d.Write([]byte{0x43, 0xB3, 0x00, 0x0C, 0xEC, 0x85})
	if err != nil {
		panic(err)
	}

	time.Sleep(875 * time.Millisecond)

	// // 0x4381 (AVGAIN-REGISTER)
	// outcome := DeviceFetchInt(d, 4, []byte{0x43, 0x81})
	// fmt.Printf("AVGAIN-REGISTER VORHER%g   %x %x %x %x \n", outcome, data[0], data[1], data[2], data[3])

	// 0x4381 (AVGAIN-REGISTER)
	err = d.Write([]byte{0x43, 0x81, 0xFF, 0xFC, 0x1C, 0xC2})
	if err != nil {
		panic(err)
	}

	// 0x4383 (BVGAIN-REGISTER)
	err = d.Write([]byte{0x43, 0x83, 0xFF, 0xFB, 0xCA, 0x60})
	if err != nil {
		panic(err)
	}

	// 0x4385 (CVGAIN-REGISTER)
	err = d.Write([]byte{0x43, 0x85, 0xFF, 0xFC, 0x12, 0xDE})
	if err != nil {
		panic(err)
	}

	// Line cycle mode
	// 0xE702 LCYCMODE
	err = d.Write([]byte{0xE7, 0x02, 0x0F})
	if err != nil {
		panic(err)
	}

	// Line cycle mode count
	// 0xE60C LINECYC
	err = d.Write([]byte{0xE6, 0x0C, 0xC8})
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
	err = d.Write([]byte{0xE2, 0x28, 0x00, 0x01})
	if err != nil {
		panic(err)
	}

	return d, nil
}

func ReadoutValues(d *i2c.Device, c *Config) [25]float32 {
	var values [25]float32
	var outcome float32
	var err error

	if c.PowerFrequency == 60 {
		rms_factor_current = float32(3493258) // 60Hz
	} else {
		rms_factor_current = float32(4191910) // 50Hz
	}

	voltage_measure_1 := true
	voltage_measure_2 := true
	voltage_measure_3 := true

	// Current phase A (amps).
	if c.MeasureCurrent1 {
		// 0x43C0 (AIRMS; Current rms an A)
		outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC0}))
		values[0] = ((((outcome * 0.3535) / rms_factor_current) / CURRENT_RESISTOR_A) / CURRENT_CLAMP_FACTOR_A) * 100.0 * OFFSET_CURRENT_A
	} else {
		values[0] = 0.0
	}

	// Current phase B (amps).
	if c.MeasureCurrent2 {
		// 0x43C2 (BIRMS; Current rms an B)
		outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC2}))
		values[1] = ((((outcome * 0.3535) / rms_factor_current) / CURRENT_RESISTOR_B) / CURRENT_CLAMP_FACTOR_B) * 100.0 * OFFSET_CURRENT_B
	} else {
		values[1] = 0.0
	}

	// Current phase C (amps).
	if c.MeasureCurrent3 {
		// 0x43C4 (CIRMS; Current rms an C)
		outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC4}))
		values[2] = ((((outcome * 0.3535) / rms_factor_current) / CURRENT_RESISTOR_C) / CURRENT_CLAMP_FACTOR_C) * 100.0 * OFFSET_CURRENT_C
	} else {
		values[2] = 0.0
	}

	// Current Neutral (amps)
	// 0x43C6 (NIRMS; Current rms neutral conductor)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC6}))
	values[3] = ((((outcome * 0.3535) / rms_factor_current) / CURRENT_RESISTOR_N) / CURRENT_CLAMP_FACTOR_N) * 100.0 * OFFSET_CURRENT_N

	// Voltage phase A (volts)
	// 0x43C1 (AVRMS; Voltage rms an A)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC1}))
	values[4] = float32(outcome / 1e+4)
	voltage_measure_1 = true
	if !c.MeasureVoltage1 || values[4] < 10 {
		values[4] = float32(c.Voltage1)
		voltage_measure_1 = false
	}

	// Voltage phase B (volts)
	// 0x43C3 (BVRMS; Voltage rms an B)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC3}))
	values[5] = float32(outcome / 1e+4)
	voltage_measure_2 = true
	if !c.MeasureVoltage2 || values[5] < 10 {
		values[5] = float32(c.Voltage2)
		voltage_measure_2 = false
	}

	// Voltage phase C (volts)
	// 0x43C5 (BVRMS; Voltage rms an C)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0x43, 0xC5}))
	values[6] = float32(outcome / 1e+4)
	voltage_measure_3 = true
	if !c.MeasureVoltage3 || values[6] < 10 {
		values[6] = float32(c.Voltage3)
		voltage_measure_3 = false
	}

	// Total active power phase A (watts).
	// 0xE513 (AWATT total active power an A)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x13}))
	if c.MeasureCurrent1 {
		values[7] = float32(outcome * POWER_CORRECTION_FACTOR_A)
	} else {
		values[7] = 0.0
	}
	if c.CurrentDirection1 {
		values[7] *= -1
	}
	if !voltage_measure_1 {
		values[7] = values[0] * values[4]
	}

	// Total active power phase B (watts).
	// 0xE514 (AWATT total active power an B)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x14}))
	if c.MeasureCurrent2 {
		values[8] = float32(outcome * POWER_CORRECTION_FACTOR_B)
	} else {
		values[8] = 0.0
	}
	if c.CurrentDirection2 {
		values[8] *= -1
	}
	if !voltage_measure_2 {
		values[8] = values[1] * values[5]
	}

	// Total active power phase C (watts).
	// 0xE515 (AWATT total active power an C)
	outcome = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x15}))
	if c.MeasureCurrent3 {
		values[9] = float32(outcome * POWER_CORRECTION_FACTOR_C)
	} else {
		values[9] = 0.0
	}
	if c.CurrentDirection3 {
		values[9] *= -1
	}
	if !voltage_measure_3 {
		values[9] = values[2] * values[6]
	}

	// 0xE601 (ANGLE0 cosphi an A)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x01}))
	values[10] = float32(math.Cos(float64(outcome * FACTOR_CIRCLE * float32(c.PowerFrequency) / ADE7878_CLOCK * VAL)))
	if c.CurrentDirection1 {
		values[10] *= -1
	}
	if c.MeasureVoltage1 {
		values[10] = 1.0
	}

	// 0xE602 (ANGLE1 cosphi an B)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x02}))
	values[11] = float32(math.Cos(float64(outcome * FACTOR_CIRCLE * float32(c.PowerFrequency) / ADE7878_CLOCK * VAL)))
	if c.CurrentDirection2 {
		values[11] *= -1
	}
	if c.MeasureVoltage2 {
		values[11] = 1.0
	}

	// 0xE603 (ANGLE1 cosphi an C)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x03}))
	values[12] = float32(math.Cos(float64(outcome * FACTOR_CIRCLE * float32(c.PowerFrequency) / ADE7878_CLOCK * VAL)))
	if c.CurrentDirection3 {
		values[12] *= -1
	}
	if c.MeasureVoltage3 {
		values[12] = 1.0
	}

	err = d.Write([]byte{0xE7, 0x00, 0x1C}) // MMODE-Register measure frequency at VA
	if err != nil {
		panic(err)
	}
	time.Sleep(50 * time.Millisecond)
	// 0xE607 (PERIOD)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x07}))
	values[13] = float32(ADE7878_CLOCK / (outcome + 1))

	err = d.Write([]byte{0xE7, 0x00, 0x1D}) // MMODE-Register measure frequency at VB
	if err != nil {
		panic(err)
	}
	time.Sleep(50 * time.Millisecond)
	// 0xE607 (PERIOD)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x07}))
	values[14] = float32(ADE7878_CLOCK / (outcome + 1))

	err = d.Write([]byte{0xE7, 0x00, 0x1E}) // MMODE-Register measure frequency at VC
	if err != nil {
		panic(err)
	}
	time.Sleep(50 * time.Millisecond)
	// 0xE607 (PERIOD)
	outcome = float32(DeviceFetchInt(d, 2, []byte{0xE6, 0x07}))
	values[15] = float32(ADE7878_CLOCK / (outcome + 1))

	// Total apparent power phase A (volt-amps).
	if c.MeasureCurrent1 {
		// 0xE519 (AVA total apparent power an A)
		values[16] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x19}))
	} else {
		values[16] = 0.0
	}

	// Total apparent power phase B (volt-amps).
	if c.MeasureCurrent2 {
		// 0xE51A (BVA total apparent power an B)
		values[17] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x1A}))
	} else {
		values[17] = 0.0
	}

	// Total apparent power phase A (volt-amps).
	if c.MeasureCurrent3 {
		// 0xE51B (CVA total apparent power an C)
		values[18] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x1B}))
	} else {
		values[18] = 0.0
	}

	// Total reactive power phase A (volt-ampere reactive).
	if c.MeasureCurrent1 {
		// 0xE516 (AVAR total reactive power an A)
		values[19] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x16}))
	} else {
		values[19] = 0.0
	}
	if c.CurrentDirection1 {
		values[19] *= -1
	}

	// Total reactive power phase B (volt-ampere reactive).
	if c.MeasureCurrent2 {
		// 0xE517 (BVAR total reactive power an B)
		values[20] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x17}))
	} else {
		values[20] = 0.0
	}
	if c.CurrentDirection2 {
		values[20] *= -1
	}

	// Total reactive power phase C (volt-ampere reactive).
	if c.MeasureCurrent3 {
		// 0xE518 (CVAR total reactive power an C)
		values[21] = float32(DeviceFetchInt(d, 4, []byte{0xE5, 0x18}))
	} else {
		values[21] = 0.0
	}
	if c.CurrentDirection3 {
		values[21] *= -1
	}

	if math.Signbit(float64(values[19])) {
		values[22] = (values[7] / POWER_CORRECTION_FACTOR_A / values[16])
	} else {
		values[22] = -1 * (values[7] / POWER_CORRECTION_FACTOR_A / values[16])
	}
	if c.MeasureCurrent1 {
		values[22] = 0.0
	}

	if math.Signbit(float64(values[20])) {
		values[23] = (values[8] / POWER_CORRECTION_FACTOR_B / values[17])
	} else {
		values[23] = -1 * (values[8] / POWER_CORRECTION_FACTOR_B / values[17])
	}
	if c.MeasureCurrent2 {
		values[23] = 0.0
	}

	if math.Signbit(float64(values[21])) {
		values[24] = (values[9] / POWER_CORRECTION_FACTOR_C / values[18])
	} else {
		values[24] = -1 * (values[9] / POWER_CORRECTION_FACTOR_C / values[18])
	}
	if c.MeasureCurrent3 {
		values[24] = 0.0
	}

	fmt.Printf("I1: %g  I2: %g  I3: %g  I4: %g  ", values[0], values[1], values[2], values[3])
	fmt.Printf("V1: %g  V2: %g  V3: %g  ", values[4], values[5], values[6])
	fmt.Printf("P1: %g  P2: %g  P3: %g  ", values[7], values[8], values[9])
	fmt.Printf("COS1: %g  COS2: %g  COS3: %g  ", values[10], values[11], values[12])
	fmt.Printf("F1: %g  F2: %g  F3: %g  ", values[13], values[14], values[15])
	fmt.Printf("AVA: %g  BVA: %g  CVA: %g  ", values[16], values[17], values[18])
	fmt.Printf("AVAR: %g  BVAR: %g  CVAR: %g  ", values[19], values[20], values[21])
	fmt.Printf("PFA: %g  PFB: %g  PFC: %g  ", values[22], values[23], values[24])
	fmt.Printf("\n")

	return values
}
