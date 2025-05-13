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

package config

import (
	"io"
	"math/rand"
	"os"
	"strconv"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/utils"

	log "github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
)

type SmartPiACConfig struct {

	// [device]
	I2CDevice            string
	PowerFrequency       float64
	Samplerate           int
	Integrator           bool
	StoreSamples         bool
	CTType               map[models.SmartPiPhase]string
	CTTypePrimaryCurrent map[models.SmartPiPhase]int
	CurrentDirection     map[models.SmartPiPhase]bool
	MeasureCurrent       map[models.SmartPiPhase]bool
	MeasureVoltage       map[models.SmartPiPhase]bool
	Voltage              map[models.SmartPiPhase]float64

	// [calibration]
	CalibrationfactorI map[models.SmartPiPhase]float64
	CalibrationfactorU map[models.SmartPiPhase]float64

	// [GUI]
	GUIMaxCurrent map[models.SmartPiPhase]int

	// [emeter]
	EmeterEnabled          bool
	EmeterMulticastAddress string
	EmeterMulticastPort    int
	EmeterSusyID           uint16
	EmeterSerial           uint32

	// [files]
	CounterEnabled bool
	CounterDir     string
}

var accfg *ini.File
var acerr error

func (p *SmartPiACConfig) ReadParameterFromFile() {

	log.Debug("Read AC-Config from file")

	accfg, acerr = ini.LooseLoad("/etc/smartpiAC")
	if acerr != nil {
		log.Error(acerr)
	}

	// [device]
	p.I2CDevice = accfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	p.PowerFrequency = accfg.Section("device").Key("power_frequency").MustFloat64(50)
	p.Samplerate = accfg.Section("device").Key("samplerate").MustInt(1)
	p.Integrator = accfg.Section("device").Key("integrator").MustBool(false)
	p.StoreSamples = accfg.Section("device").Key("storesamples").MustBool(false)
	p.CTType = make(map[models.SmartPiPhase]string)
	p.CTType[models.PhaseA] = accfg.Section("device").Key("ct_type_1").MustString("YHDC_SCT013")
	p.CTType[models.PhaseB] = accfg.Section("device").Key("ct_type_2").MustString("YHDC_SCT013")
	p.CTType[models.PhaseC] = accfg.Section("device").Key("ct_type_3").MustString("YHDC_SCT013")
	p.CTType[models.PhaseN] = accfg.Section("device").Key("ct_type_4").MustString("YHDC_SCT013")
	p.CTTypePrimaryCurrent = make(map[models.SmartPiPhase]int)
	p.CTTypePrimaryCurrent[models.PhaseA] = accfg.Section("device").Key("ct_type_1_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseB] = accfg.Section("device").Key("ct_type_2_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseC] = accfg.Section("device").Key("ct_type_3_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseN] = accfg.Section("device").Key("ct_type_4_primary_current").MustInt(100)
	p.CurrentDirection = make(map[models.SmartPiPhase]bool)
	p.CurrentDirection[models.PhaseA] = accfg.Section("device").Key("change_current_direction_1").MustBool(false)
	p.CurrentDirection[models.PhaseB] = accfg.Section("device").Key("change_current_direction_2").MustBool(false)
	p.CurrentDirection[models.PhaseC] = accfg.Section("device").Key("change_current_direction_3").MustBool(false)
	p.CurrentDirection[models.PhaseN] = accfg.Section("device").Key("change_current_direction_4").MustBool(false)
	p.MeasureCurrent = make(map[models.SmartPiPhase]bool)
	p.MeasureCurrent[models.PhaseA] = accfg.Section("device").Key("measure_current_1").MustBool(true)
	p.MeasureCurrent[models.PhaseB] = accfg.Section("device").Key("measure_current_2").MustBool(true)
	p.MeasureCurrent[models.PhaseC] = accfg.Section("device").Key("measure_current_3").MustBool(true)
	p.MeasureCurrent[models.PhaseN] = accfg.Section("device").Key("measure_current_4").MustBool(true)
	p.MeasureVoltage = make(map[models.SmartPiPhase]bool)
	p.MeasureVoltage[models.PhaseA] = accfg.Section("device").Key("measure_voltage_1").MustBool(true)
	p.MeasureVoltage[models.PhaseB] = accfg.Section("device").Key("measure_voltage_2").MustBool(true)
	p.MeasureVoltage[models.PhaseC] = accfg.Section("device").Key("measure_voltage_3").MustBool(true)
	p.Voltage = make(map[models.SmartPiPhase]float64)
	p.Voltage[models.PhaseA] = accfg.Section("device").Key("voltage_1").MustFloat64(230)
	p.Voltage[models.PhaseB] = accfg.Section("device").Key("voltage_2").MustFloat64(230)
	p.Voltage[models.PhaseC] = accfg.Section("device").Key("voltage_3").MustFloat64(230)

	// [calibration]
	p.CalibrationfactorI = make(map[models.SmartPiPhase]float64)
	p.CalibrationfactorI[models.PhaseA] = accfg.Section("device").Key("calibrationfactorI_1").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseB] = accfg.Section("device").Key("calibrationfactorI_2").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseC] = accfg.Section("device").Key("calibrationfactorI_3").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseN] = accfg.Section("device").Key("calibrationfactorI_4").MustFloat64(1)
	p.CalibrationfactorU = make(map[models.SmartPiPhase]float64)
	p.CalibrationfactorU[models.PhaseA] = accfg.Section("device").Key("calibrationfactorU_1").MustFloat64(1)
	p.CalibrationfactorU[models.PhaseB] = accfg.Section("device").Key("calibrationfactorU_2").MustFloat64(1)
	p.CalibrationfactorU[models.PhaseC] = accfg.Section("device").Key("calibrationfactorU_3").MustFloat64(1)

	// [GUI]
	p.GUIMaxCurrent = make(map[models.SmartPiPhase]int)
	p.GUIMaxCurrent[models.PhaseA] = accfg.Section("gui").Key("gui_max_current_1").MustInt(100)
	p.GUIMaxCurrent[models.PhaseB] = accfg.Section("gui").Key("gui_max_current_2").MustInt(100)
	p.GUIMaxCurrent[models.PhaseC] = accfg.Section("gui").Key("gui_max_current_3").MustInt(100)
	p.GUIMaxCurrent[models.PhaseN] = accfg.Section("gui").Key("gui_max_current_4").MustInt(100)

	// [emeter]
	p.EmeterEnabled = accfg.Section("emeter").Key("emeter_enabled").MustBool(true)
	p.EmeterMulticastAddress = accfg.Section("emeter").Key("emeter_multicast_address").MustString("239.12.255.254")
	p.EmeterMulticastPort = accfg.Section("emeter").Key("emeter_multicast_port").MustInt(9522)
	p.EmeterSusyID = uint16(accfg.Section("emeter").Key("emeter_susy_id").MustUint(270))
	p.EmeterSerial = uint32(accfg.Section("emeter").Key("emeter_serial").MustUint64(uint64(rand.Uint32())))

	// [files]
	p.CounterEnabled = cfg.Section("files").Key("counter_enabled").MustBool(true)
	p.CounterDir = cfg.Section("files").Key("counterdir").MustString("/var/smartpi")
}

func (p *SmartPiACConfig) SaveParameterToFile() {

	log.Debug("Save AC-Config to file")

	// [device]
	_, acerr = accfg.Section("device").NewKey("i2c_device", p.I2CDevice)
	_, acerr = accfg.Section("device").NewKey("power_frequency", strconv.FormatInt(int64(p.PowerFrequency), 10))
	_, acerr = accfg.Section("device").NewKey("samplerate", strconv.FormatInt(int64(p.Samplerate), 10))
	_, acerr = accfg.Section("device").NewKey("integrator", strconv.FormatBool(p.Integrator))
	_, acerr = accfg.Section("device").NewKey("storesamples", strconv.FormatBool(p.StoreSamples))

	_, acerr = accfg.Section("device").NewKey("ct_type_1", p.CTType[models.PhaseA])
	_, acerr = accfg.Section("device").NewKey("ct_type_2", p.CTType[models.PhaseB])
	_, acerr = accfg.Section("device").NewKey("ct_type_3", p.CTType[models.PhaseC])
	_, acerr = accfg.Section("device").NewKey("ct_type_4", p.CTType[models.PhaseN])

	_, acerr = accfg.Section("device").NewKey("ct_type_1_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseA]), 10))
	_, acerr = accfg.Section("device").NewKey("ct_type_2_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseB]), 10))
	_, acerr = accfg.Section("device").NewKey("ct_type_3_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseC]), 10))
	_, acerr = accfg.Section("device").NewKey("ct_type_4_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseN]), 10))

	_, acerr = accfg.Section("device").NewKey("change_current_direction_1", strconv.FormatBool(p.CurrentDirection[models.PhaseA]))
	_, acerr = accfg.Section("device").NewKey("change_current_direction_2", strconv.FormatBool(p.CurrentDirection[models.PhaseB]))
	_, acerr = accfg.Section("device").NewKey("change_current_direction_3", strconv.FormatBool(p.CurrentDirection[models.PhaseC]))
	_, acerr = accfg.Section("device").NewKey("change_current_direction_4", strconv.FormatBool(p.CurrentDirection[models.PhaseN]))

	_, acerr = accfg.Section("device").NewKey("measure_current_1", strconv.FormatBool(p.MeasureCurrent[models.PhaseA]))
	_, acerr = accfg.Section("device").NewKey("measure_current_2", strconv.FormatBool(p.MeasureCurrent[models.PhaseB]))
	_, acerr = accfg.Section("device").NewKey("measure_current_3", strconv.FormatBool(p.MeasureCurrent[models.PhaseC]))
	_, acerr = accfg.Section("device").NewKey("measure_current_4", strconv.FormatBool(p.MeasureCurrent[models.PhaseN]))

	_, acerr = accfg.Section("device").NewKey("measure_voltage_1", strconv.FormatBool(p.MeasureVoltage[models.PhaseA]))
	_, acerr = accfg.Section("device").NewKey("measure_voltage_2", strconv.FormatBool(p.MeasureVoltage[models.PhaseB]))
	_, acerr = accfg.Section("device").NewKey("measure_voltage_3", strconv.FormatBool(p.MeasureVoltage[models.PhaseC]))

	_, acerr = accfg.Section("device").NewKey("voltage_1", strconv.FormatFloat(p.Voltage[models.PhaseA], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("voltage_2", strconv.FormatFloat(p.Voltage[models.PhaseB], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("voltage_3", strconv.FormatFloat(p.Voltage[models.PhaseC], 'f', -1, 64))

	// [calibration]
	_, acerr = accfg.Section("device").NewKey("calibrationfactorI_1", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseA], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorI_2", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseB], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorI_3", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseC], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorI_4", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseN], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorU_1", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseA], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorU_2", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseB], 'f', -1, 64))
	_, acerr = accfg.Section("device").NewKey("calibrationfactorU_3", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseC], 'f', -1, 64))

	_, acerr = accfg.Section("gui").NewKey("gui_max_current_1", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseA]), 10))
	_, acerr = accfg.Section("gui").NewKey("gui_max_current_2", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseB]), 10))
	_, acerr = accfg.Section("gui").NewKey("gui_max_current_3", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseC]), 10))
	_, acerr = accfg.Section("gui").NewKey("gui_max_current_4", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseN]), 10))

	// [emeter]
	_, acerr = accfg.Section("emeter").NewKey("emeter_enabled", strconv.FormatBool(p.EmeterEnabled))
	_, acerr = accfg.Section("emeter").NewKey("emeter_multicast_address", p.EmeterMulticastAddress)
	_, acerr = accfg.Section("emeter").NewKey("emeter_multicast_port", strconv.FormatInt(int64(p.EmeterMulticastPort), 10))
	_, acerr = accfg.Section("emeter").NewKey("emeter_susy_id", strconv.FormatUint(uint64(p.EmeterSusyID), 10))
	_, acerr = accfg.Section("emeter").NewKey("emeter_serial", strconv.FormatUint(uint64(p.EmeterSerial), 10))

	// [files]
	_, err = cfg.Section("files").NewKey("counter_enabled", strconv.FormatBool(p.CounterEnabled))
	_, err = cfg.Section("files").NewKey("counterdir", p.CounterDir)

	tmpFile := "/tmp/smartpiAC"
	acerr := accfg.SaveTo(tmpFile)
	if acerr != nil {
		panic(acerr)
	}

	srcFile, acerr := os.Open(tmpFile)
	utils.Checklog(acerr)
	defer srcFile.Close()

	destFile, acerr := os.Create("/etc/smartpiAC") // creates if file doesn't exist
	utils.Checklog(acerr)
	defer destFile.Close()

	_, acerr = io.Copy(destFile, srcFile)
	utils.Checklog(acerr)

	acerr = destFile.Sync()
	utils.Checklog(acerr)

	defer os.Remove(tmpFile)
}

func NewSmartPiACConfig() *SmartPiACConfig {

	t := new(SmartPiACConfig)
	t.ReadParameterFromFile()
	return t
}
