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

	log "github.com/Sirupsen/logrus"
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

	// [device]
	I2CDevice            string
	PowerFrequency       float64
	CTType               map[string]string
	CTTypePrimaryCurrent map[string]int
	CurrentDirection     map[string]bool
	MeasureCurrent       map[string]bool
	MeasureVoltage       map[string]bool
	Voltage              map[string]float64

	// [ftp]
	FTPupload bool
	FTPserver string
	FTPuser   string
	FTPpass   string
	FTPpath   string

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

	// [mobile]
	MobileEnabled bool
	MobileAPN     string
	MobilePIN     string
	MobileUser    string
	MobilePass    string
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

	// [device]
	p.I2CDevice = cfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	p.PowerFrequency = cfg.Section("device").Key("power_frequency").MustFloat64(50)
	p.CTType = make(map[string]string)
	p.CTType["A"] = cfg.Section("device").Key("ct_type_1").MustString("YHDC_SCT013")
	p.CTType["B"] = cfg.Section("device").Key("ct_type_2").MustString("YHDC_SCT013")
	p.CTType["C"] = cfg.Section("device").Key("ct_type_3").MustString("YHDC_SCT013")
	p.CTType["N"] = cfg.Section("device").Key("ct_type_4").MustString("YHDC_SCT013")
	p.CTTypePrimaryCurrent = make(map[string]int)
	p.CTTypePrimaryCurrent["A"] = cfg.Section("device").Key("ct_type_1_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent["B"] = cfg.Section("device").Key("ct_type_2_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent["C"] = cfg.Section("device").Key("ct_type_3_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent["N"] = cfg.Section("device").Key("ct_type_4_primary_current").MustInt(100)
	p.CurrentDirection = make(map[string]bool)
	p.CurrentDirection["A"] = cfg.Section("device").Key("change_current_direction_1").MustBool(false)
	p.CurrentDirection["B"] = cfg.Section("device").Key("change_current_direction_2").MustBool(false)
	p.CurrentDirection["C"] = cfg.Section("device").Key("change_current_direction_3").MustBool(false)
	p.MeasureCurrent = make(map[string]bool)
	p.MeasureCurrent["A"] = cfg.Section("device").Key("measure_current_1").MustBool(true)
	p.MeasureCurrent["B"] = cfg.Section("device").Key("measure_current_2").MustBool(true)
	p.MeasureCurrent["C"] = cfg.Section("device").Key("measure_current_3").MustBool(true)
	p.MeasureCurrent["N"] = true // Always measure Neutral.
	p.MeasureVoltage = make(map[string]bool)
	p.MeasureVoltage["A"] = cfg.Section("device").Key("measure_voltage_1").MustBool(true)
	p.MeasureVoltage["B"] = cfg.Section("device").Key("measure_voltage_2").MustBool(true)
	p.MeasureVoltage["C"] = cfg.Section("device").Key("measure_voltage_3").MustBool(true)
	p.Voltage = make(map[string]float64)
	p.Voltage["A"] = cfg.Section("device").Key("voltage_1").MustFloat64(230)
	p.Voltage["B"] = cfg.Section("device").Key("voltage_2").MustFloat64(230)
	p.Voltage["C"] = cfg.Section("device").Key("voltage_3").MustFloat64(230)

	// [ftp]
	p.FTPupload = cfg.Section("ftp").Key("ftp_upload").MustBool(false)
	p.FTPserver = cfg.Section("ftp").Key("ftp_server").String()
	p.FTPuser = cfg.Section("ftp").Key("ftp_user").String()
	p.FTPpass = cfg.Section("ftp").Key("ftp_pass").String()
	p.FTPpath = cfg.Section("ftp").Key("ftp_path").String()

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

	// [mobile]
	p.MobileEnabled = cfg.Section("umts").Key("umts").MustBool(false)
	p.MobileAPN = cfg.Section("umts").Key("umts_apn").String()
	p.MobilePIN = cfg.Section("umts").Key("umts_pin").String()
	p.MobileUser = cfg.Section("umts").Key("umts_username").String()
	p.MobilePass = cfg.Section("umts").Key("umts_password").String()

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

	// [device]
	_, err = cfg.Section("device").NewKey("i2c_device", p.I2CDevice)
	_, err = cfg.Section("device").NewKey("power_frequency", strconv.FormatInt(int64(p.PowerFrequency), 10))
	_, err = cfg.Section("device").NewKey("ct_type_1", p.CTType["A"])
	_, err = cfg.Section("device").NewKey("ct_type_2", p.CTType["B"])
	_, err = cfg.Section("device").NewKey("ct_type_3", p.CTType["C"])
	_, err = cfg.Section("device").NewKey("ct_type_4", p.CTType["N"])

	_, err = cfg.Section("device").NewKey("ct_type_1_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent["A"]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_2_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent["B"]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_3_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent["C"]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_4_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent["N"]), 10))

	_, err = cfg.Section("device").NewKey("change_current_direction_1", strconv.FormatBool(p.CurrentDirection["A"]))
	_, err = cfg.Section("device").NewKey("change_current_direction_2", strconv.FormatBool(p.CurrentDirection["B"]))
	_, err = cfg.Section("device").NewKey("change_current_direction_3", strconv.FormatBool(p.CurrentDirection["C"]))

	_, err = cfg.Section("device").NewKey("measure_current_1", strconv.FormatBool(p.MeasureCurrent["A"]))
	_, err = cfg.Section("device").NewKey("measure_current_2", strconv.FormatBool(p.MeasureCurrent["B"]))
	_, err = cfg.Section("device").NewKey("measure_current_3", strconv.FormatBool(p.MeasureCurrent["C"]))

	_, err = cfg.Section("device").NewKey("measure_voltage_1", strconv.FormatBool(p.MeasureVoltage["A"]))
	_, err = cfg.Section("device").NewKey("measure_voltage_2", strconv.FormatBool(p.MeasureVoltage["B"]))
	_, err = cfg.Section("device").NewKey("measure_voltage_3", strconv.FormatBool(p.MeasureVoltage["C"]))

	_, err = cfg.Section("device").NewKey("voltage_1", strconv.FormatFloat(p.Voltage["A"], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_2", strconv.FormatFloat(p.Voltage["B"], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_3", strconv.FormatFloat(p.Voltage["C"], 'f', -1, 64))

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

	// [mobile]
	_, err = cfg.Section("umts").NewKey("umts", strconv.FormatBool(p.MobileEnabled))
	_, err = cfg.Section("umts").NewKey("umts_apn", p.MobileAPN)
	_, err = cfg.Section("umts").NewKey("umts_pin", p.MobilePIN)
	_, err = cfg.Section("umts").NewKey("umts_username", p.MobileUser)
	_, err = cfg.Section("umts").NewKey("umts_password", p.MobilePass)

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
