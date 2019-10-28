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
	"io"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type Config struct {
	// [base]
	Serial               string
	Name                 string
	LogLevel             log.Level
	DebugLevel           int
	MetricsListenAddress string

	// [location]
	Lat float64
	Lng float64

	// [database]
	CounterEnabled  bool
	CounterDir      string
	DatabaseEnabled bool
	DatabaseDir     string
	Influxuser      string
	Influxpassword  string
	Influxdatabase  string

	// [device]
	I2CDevice            string
	PowerFrequency       float64
	Samplerate           int
	Integrator           bool
	CTType               map[Phase]string
	CTTypePrimaryCurrent map[Phase]int
	CurrentDirection     map[Phase]bool
	MeasureCurrent       map[Phase]bool
	MeasureVoltage       map[Phase]bool
	Voltage              map[Phase]float64

	// [ftp]
	FTPupload bool
	FTPserver string
	FTPuser   string
	FTPpass   string
	FTPpath   string
	FTPcsv    bool
	FTPxml    bool

	// [webserver]
	SharedFileEnabled bool
	SharedDir         string
	SharedFile        string
	WebserverPort     int
	DocRoot           string

	// [csv]
	CSVdecimalpoint string
	CSVtimeformat   string

	// [mqtt]
	MQTTenabled    bool
	MQTTbroker     string
	MQTTbrokerport string
	MQTTuser       string
	MQTTpass       string
	MQTTtopic      string

	// [modbus slave]
	ModbusRTUenabled bool
	ModbusTCPenabled bool
	ModbusRTUAddress uint8
	ModbusRTUDevice  string
	ModbusTCPAddress string

	// [mobile]
	MobileEnabled bool
	MobileAPN     string
	MobilePIN     string
	MobileUser    string
	MobilePass    string

	// [calibration]
	CalibrationfactorI map[Phase]float64
	CalibrationfactorU map[Phase]float64
}

var cfg *ini.File
var err error

func (p *Config) ReadParameterFromFile() {

	cfg, err = ini.Load("/etc/smartpi")
	if err != nil {
		panic(err)
	}

	// [base]
	p.Serial = cfg.Section("base").Key("serial").String()
	p.Name = cfg.Section("base").Key("name").MustString("House")
	// Handle logging levels
	p.LogLevel, err = log.ParseLevel(cfg.Section("base").Key("loglevel").MustString("info"))
	if err != nil {
		panic(err)
	}
	// Handle old debuglevel config key as log.Debug.
	p.DebugLevel, err = cfg.Section("base").Key("debuglevel").Int()
	if err == nil && p.DebugLevel > 0 {
		p.LogLevel = log.DebugLevel
		log.Debug("Config option debuglevel is deprecated, use loglevel=debug.")
	} else {
		p.DebugLevel = 0
	}
	p.MetricsListenAddress = cfg.Section("base").Key("metrics_listen_address").MustString(":9246")

	// [location]
	p.Lat = cfg.Section("location").Key("lat").MustFloat64(52.3667)
	p.Lng = cfg.Section("location").Key("lng").MustFloat64(9.7167)

	// [database]
	p.CounterEnabled = cfg.Section("database").Key("counter_enabled").MustBool(true)
	p.CounterDir = cfg.Section("database").Key("counterdir").MustString("/var/smartpi")
	p.DatabaseEnabled = cfg.Section("database").Key("database_enabled").MustBool(true)
	p.DatabaseDir = cfg.Section("database").Key("dir").MustString("/var/smartpi/db")
	p.Influxuser = cfg.Section("database").Key("influxuser").MustString("smartpi")
	p.Influxpassword = cfg.Section("database").Key("influxpassword").MustString("smart4pi")
	p.Influxdatabase = cfg.Section("database").Key("influxdatabase").MustString("http://localhost:8086")

	// [device]
	p.I2CDevice = cfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	p.PowerFrequency = cfg.Section("device").Key("power_frequency").MustFloat64(50)
	p.Samplerate = cfg.Section("device").Key("samplerate").MustInt(1)
	p.Integrator = cfg.Section("device").Key("integrator").MustBool(false)
	p.CTType = make(map[Phase]string)
	p.CTType[PhaseA] = cfg.Section("device").Key("ct_type_1").MustString("YHDC_SCT013")
	p.CTType[PhaseB] = cfg.Section("device").Key("ct_type_2").MustString("YHDC_SCT013")
	p.CTType[PhaseC] = cfg.Section("device").Key("ct_type_3").MustString("YHDC_SCT013")
	p.CTType[PhaseN] = cfg.Section("device").Key("ct_type_4").MustString("YHDC_SCT013")
	p.CTTypePrimaryCurrent = make(map[Phase]int)
	p.CTTypePrimaryCurrent[PhaseA] = cfg.Section("device").Key("ct_type_1_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[PhaseB] = cfg.Section("device").Key("ct_type_2_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[PhaseC] = cfg.Section("device").Key("ct_type_3_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[PhaseN] = cfg.Section("device").Key("ct_type_4_primary_current").MustInt(100)
	p.CurrentDirection = make(map[Phase]bool)
	p.CurrentDirection[PhaseA] = cfg.Section("device").Key("change_current_direction_1").MustBool(false)
	p.CurrentDirection[PhaseB] = cfg.Section("device").Key("change_current_direction_2").MustBool(false)
	p.CurrentDirection[PhaseC] = cfg.Section("device").Key("change_current_direction_3").MustBool(false)
	p.CurrentDirection[PhaseN] = cfg.Section("device").Key("change_current_direction_4").MustBool(false)
	p.MeasureCurrent = make(map[Phase]bool)
	p.MeasureCurrent[PhaseA] = cfg.Section("device").Key("measure_current_1").MustBool(true)
	p.MeasureCurrent[PhaseB] = cfg.Section("device").Key("measure_current_2").MustBool(true)
	p.MeasureCurrent[PhaseC] = cfg.Section("device").Key("measure_current_3").MustBool(true)
	p.MeasureCurrent[PhaseN] = cfg.Section("device").Key("measure_current_4").MustBool(true)
	p.MeasureVoltage = make(map[Phase]bool)
	p.MeasureVoltage[PhaseA] = cfg.Section("device").Key("measure_voltage_1").MustBool(true)
	p.MeasureVoltage[PhaseB] = cfg.Section("device").Key("measure_voltage_2").MustBool(true)
	p.MeasureVoltage[PhaseC] = cfg.Section("device").Key("measure_voltage_3").MustBool(true)
	p.Voltage = make(map[Phase]float64)
	p.Voltage[PhaseA] = cfg.Section("device").Key("voltage_1").MustFloat64(230)
	p.Voltage[PhaseB] = cfg.Section("device").Key("voltage_2").MustFloat64(230)
	p.Voltage[PhaseC] = cfg.Section("device").Key("voltage_3").MustFloat64(230)

	// [ftp]
	p.FTPupload = cfg.Section("ftp").Key("ftp_upload").MustBool(false)
	p.FTPserver = cfg.Section("ftp").Key("ftp_server").String()
	p.FTPuser = cfg.Section("ftp").Key("ftp_user").String()
	p.FTPpass = cfg.Section("ftp").Key("ftp_pass").String()
	p.FTPpath = cfg.Section("ftp").Key("ftp_path").String()
	p.FTPcsv = cfg.Section("ftp").Key("ftp_csv").MustBool(true)
	p.FTPxml = cfg.Section("ftp").Key("ftp_xml").MustBool(true)

	// [webserver]
	p.SharedFileEnabled = cfg.Section("webserver").Key("shared_file_enabled").MustBool(true)
	p.SharedDir = cfg.Section("webserver").Key("shared_dir").MustString("/var/run")
	p.SharedFile = cfg.Section("webserver").Key("shared_file").MustString("smartpi_values")
	p.WebserverPort = cfg.Section("webserver").Key("port").MustInt(1080)
	p.DocRoot = cfg.Section("webserver").Key("docroot").MustString("/var/smartpi/www")

	// [csv]
	p.CSVdecimalpoint = cfg.Section("csv").Key("decimalpoint").String()
	p.CSVtimeformat = cfg.Section("csv").Key("timeformat").String()

	// [mqtt]
	p.MQTTenabled = cfg.Section("mqtt").Key("mqtt_enabled").MustBool(false)
	p.MQTTbroker = cfg.Section("mqtt").Key("mqtt_broker_url").String()
	p.MQTTbrokerport = cfg.Section("mqtt").Key("mqtt_broker_port").String()
	p.MQTTuser = cfg.Section("mqtt").Key("mqtt_username").String()
	p.MQTTpass = cfg.Section("mqtt").Key("mqtt_password").String()
	p.MQTTtopic = cfg.Section("mqtt").Key("mqtt_topic").String()

	// [modbus slave]
	p.ModbusRTUenabled = cfg.Section("modbus").Key("modbus_rtu_enabled").MustBool(false)
	p.ModbusTCPenabled = cfg.Section("modbus").Key("modbus_tcp_enabled").MustBool(false)
	p.ModbusRTUAddress = uint8(cfg.Section("modbus").Key("modbus_rtu_address").MustInt(1))
	p.ModbusRTUDevice = cfg.Section("modbus").Key("modbus_rtu_device_id").MustString("/dev/serial0")
	p.ModbusTCPAddress = cfg.Section("modbus").Key("modbus_tcp_address").MustString(":502")

	// [mobile]
	p.MobileEnabled = cfg.Section("umts").Key("umts").MustBool(false)
	p.MobileAPN = cfg.Section("umts").Key("umts_apn").String()
	p.MobilePIN = cfg.Section("umts").Key("umts_pin").String()
	p.MobileUser = cfg.Section("umts").Key("umts_username").String()
	p.MobilePass = cfg.Section("umts").Key("umts_password").String()

	// [calibration]
	p.CalibrationfactorI = make(map[Phase]float64)
	p.CalibrationfactorI[PhaseA] = cfg.Section("calibration").Key("calibrationfactorI_1").MustFloat64(1)
	p.CalibrationfactorI[PhaseB] = cfg.Section("calibration").Key("calibrationfactorI_2").MustFloat64(1)
	p.CalibrationfactorI[PhaseC] = cfg.Section("calibration").Key("calibrationfactorI_3").MustFloat64(1)
	p.CalibrationfactorI[PhaseN] = cfg.Section("calibration").Key("calibrationfactorI_4").MustFloat64(1)
	p.CalibrationfactorU = make(map[Phase]float64)
	p.CalibrationfactorU[PhaseA] = cfg.Section("calibration").Key("calibrationfactorU_1").MustFloat64(1)
	p.CalibrationfactorU[PhaseB] = cfg.Section("calibration").Key("calibrationfactorU_2").MustFloat64(1)
	p.CalibrationfactorU[PhaseC] = cfg.Section("calibration").Key("calibrationfactorU_3").MustFloat64(1)

}

func (p *Config) SaveParameterToFile() {

	_, err = cfg.Section("base").NewKey("serial", p.Serial)
	_, err = cfg.Section("base").NewKey("name", p.Name)
	_, err = cfg.Section("base").NewKey("loglevel", p.LogLevel.String())
	_, err = cfg.Section("base").NewKey("metrics_listen_address", p.MetricsListenAddress)

	// [location]
	_, err = cfg.Section("location").NewKey("lat", strconv.FormatFloat(p.Lat, 'f', -1, 64))
	_, err = cfg.Section("location").NewKey("lng", strconv.FormatFloat(p.Lng, 'f', -1, 64))

	// [database]
	_, err = cfg.Section("database").NewKey("counter_enabled", strconv.FormatBool(p.CounterEnabled))
	_, err = cfg.Section("database").NewKey("counterdir", p.CounterDir)
	_, err = cfg.Section("database").NewKey("database_enabled", strconv.FormatBool(p.DatabaseEnabled))
	_, err = cfg.Section("database").NewKey("dir", p.DatabaseDir)
	_, err = cfg.Section("database").NewKey("influxuser", p.Influxuser)
	_, err = cfg.Section("database").NewKey("influxpassword", p.Influxpassword)
	_, err = cfg.Section("database").NewKey("influxdatabase", p.Influxdatabase)

	// [device]
	_, err = cfg.Section("device").NewKey("i2c_device", p.I2CDevice)
	_, err = cfg.Section("device").NewKey("power_frequency", strconv.FormatInt(int64(p.PowerFrequency), 10))
	_, err = cfg.Section("device").NewKey("samplerate", strconv.FormatInt(int64(p.Samplerate), 10))
	_, err = cfg.Section("device").NewKey("integrator", strconv.FormatBool(p.Integrator))
	_, err = cfg.Section("device").NewKey("ct_type_1", p.CTType[PhaseA])
	_, err = cfg.Section("device").NewKey("ct_type_2", p.CTType[PhaseB])
	_, err = cfg.Section("device").NewKey("ct_type_3", p.CTType[PhaseC])
	_, err = cfg.Section("device").NewKey("ct_type_4", p.CTType[PhaseN])

	_, err = cfg.Section("device").NewKey("ct_type_1_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[PhaseA]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_2_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[PhaseB]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_3_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[PhaseC]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_4_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[PhaseN]), 10))

	_, err = cfg.Section("device").NewKey("change_current_direction_1", strconv.FormatBool(p.CurrentDirection[PhaseA]))
	_, err = cfg.Section("device").NewKey("change_current_direction_2", strconv.FormatBool(p.CurrentDirection[PhaseB]))
	_, err = cfg.Section("device").NewKey("change_current_direction_3", strconv.FormatBool(p.CurrentDirection[PhaseC]))
	_, err = cfg.Section("device").NewKey("change_current_direction_4", strconv.FormatBool(p.CurrentDirection[PhaseN]))

	_, err = cfg.Section("device").NewKey("measure_current_1", strconv.FormatBool(p.MeasureCurrent[PhaseA]))
	_, err = cfg.Section("device").NewKey("measure_current_2", strconv.FormatBool(p.MeasureCurrent[PhaseB]))
	_, err = cfg.Section("device").NewKey("measure_current_3", strconv.FormatBool(p.MeasureCurrent[PhaseC]))
	_, err = cfg.Section("device").NewKey("measure_current_4", strconv.FormatBool(p.MeasureCurrent[PhaseN]))

	_, err = cfg.Section("device").NewKey("measure_voltage_1", strconv.FormatBool(p.MeasureVoltage[PhaseA]))
	_, err = cfg.Section("device").NewKey("measure_voltage_2", strconv.FormatBool(p.MeasureVoltage[PhaseB]))
	_, err = cfg.Section("device").NewKey("measure_voltage_3", strconv.FormatBool(p.MeasureVoltage[PhaseC]))

	_, err = cfg.Section("device").NewKey("voltage_1", strconv.FormatFloat(p.Voltage[PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_2", strconv.FormatFloat(p.Voltage[PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_3", strconv.FormatFloat(p.Voltage[PhaseC], 'f', -1, 64))

	// [ftp]
	_, err = cfg.Section("ftp").NewKey("ftp_upload", strconv.FormatBool(p.FTPupload))
	_, err = cfg.Section("ftp").NewKey("ftp_server", p.FTPserver)
	_, err = cfg.Section("ftp").NewKey("ftp_user", p.FTPuser)
	_, err = cfg.Section("ftp").NewKey("ftp_pass", p.FTPpass)
	_, err = cfg.Section("ftp").NewKey("ftp_path", p.FTPpath)

	// [webserver]
	_, err = cfg.Section("webserver").NewKey("shared_file_enabled", strconv.FormatBool(p.SharedFileEnabled))
	_, err = cfg.Section("webserver").NewKey("shared_dir", p.SharedDir)
	_, err = cfg.Section("webserver").NewKey("shared_file", p.SharedFile)
	_, err = cfg.Section("webserver").NewKey("port", strconv.FormatInt(int64(p.WebserverPort), 10))
	_, err = cfg.Section("webserver").NewKey("docroot", p.DocRoot)

	// [csv]
	_, err = cfg.Section("csv").NewKey("decimalpoint", p.CSVdecimalpoint)
	_, err = cfg.Section("csv").NewKey("timeformat", p.CSVtimeformat)

	// [mqtt]
	_, err = cfg.Section("mqtt").NewKey("mqtt_enabled", strconv.FormatBool(p.MQTTenabled))
	_, err = cfg.Section("mqtt").NewKey("mqtt_broker_url", p.MQTTbroker)
	_, err = cfg.Section("mqtt").NewKey("mqtt_broker_port", p.MQTTbrokerport)
	_, err = cfg.Section("mqtt").NewKey("mqtt_username", p.MQTTuser)
	_, err = cfg.Section("mqtt").NewKey("mqtt_password", p.MQTTpass)
	_, err = cfg.Section("mqtt").NewKey("mqtt_topic", p.MQTTtopic)

	// [modbus slave]
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_enabled", strconv.FormatBool(p.ModbusRTUenabled))
	_, err = cfg.Section("modbus").NewKey("modbus_tcp_enabled", strconv.FormatBool(p.ModbusTCPenabled))
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_address", strconv.FormatUint(uint64(p.ModbusRTUAddress), 10))
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_device_id", p.ModbusRTUDevice)
	_, err = cfg.Section("modbus").NewKey("modbus_tcp_address", p.ModbusTCPAddress)

	// [mobile]
	_, err = cfg.Section("umts").NewKey("umts", strconv.FormatBool(p.MobileEnabled))
	_, err = cfg.Section("umts").NewKey("umts_apn", p.MobileAPN)
	_, err = cfg.Section("umts").NewKey("umts_pin", p.MobilePIN)
	_, err = cfg.Section("umts").NewKey("umts_username", p.MobileUser)
	_, err = cfg.Section("umts").NewKey("umts_password", p.MobilePass)

	// [calibration]
	_, err = cfg.Section("device").NewKey("calibrationfactorI_1", strconv.FormatFloat(p.CalibrationfactorI[PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_2", strconv.FormatFloat(p.CalibrationfactorI[PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_3", strconv.FormatFloat(p.CalibrationfactorI[PhaseC], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_4", strconv.FormatFloat(p.CalibrationfactorI[PhaseN], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_1", strconv.FormatFloat(p.CalibrationfactorU[PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_2", strconv.FormatFloat(p.CalibrationfactorU[PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_3", strconv.FormatFloat(p.CalibrationfactorU[PhaseC], 'f', -1, 64))

	tmpFile := "/tmp/smartpi"
	err := cfg.SaveTo(tmpFile)
	if err != nil {
		panic(err)
	}

	srcFile, err := os.Open(tmpFile)
	Checklog(err)
	defer srcFile.Close()

	destFile, err := os.Create("/etc/smartpi") // creates if file doesn't exist
	Checklog(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	Checklog(err)

	err = destFile.Sync()
	Checklog(err)

	defer os.Remove(tmpFile)
}

func NewConfig() *Config {

	t := new(Config)
	t.ReadParameterFromFile()
	return t
}
