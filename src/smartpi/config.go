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
	"gopkg.in/ini.v1"
	// "path/filepath"
	// "os"
)

type Config struct {
	// [base]
	Serial     string
	Name       string
	DebugLevel int

	// [location]
	Lat float64
	Lng float64

	// [database]
	CounterDir  string
	DatabaseDir string

	// [device]
	I2CDevice         string
	SharedDir         string
	SharedFile        string
	PowerFrequency    int
	CTType            map[string]string
	MeasureCurrent    map[string]bool
	MeasureVoltage1   bool
	MeasureVoltage2   bool
	MeasureVoltage3   bool
	Voltage1          float64
	Voltage2          float64
	Voltage3          float64
	CurrentDirection1 bool
	CurrentDirection2 bool
	CurrentDirection3 bool

	// [ftp]
	FTPupload bool
	FTPserver string
	FTPuser   string
	FTPpass   string
	FTPpath   string

	// [webserver]
	WebserverPort   int
	DocRoot         string
	CSVdecimalpoint string
	CSVtimeformat   string

	// [mqtt]
	MQTTenabled    bool
	MQTTbroker     string
	MQTTbrokerport string
	MQTTuser       string
	MQTTpass       string
	MQTTtopic      string
}

func (p *Config) ReadParameterFromFile() {

	cfg, err := ini.Load("/etc/smartpi")
	if err != nil {
		panic(err)
	}

	// [base]
	p.Serial = cfg.Section("base").Key("serial").String()
	p.Name = cfg.Section("base").Key("name").String()
	p.DebugLevel = cfg.Section("base").Key("debuglevel").MustInt(0)

	// [location]
	p.Lat = cfg.Section("location").Key("lat").MustFloat64(52.3667)
	p.Lng = cfg.Section("location").Key("lng").MustFloat64(9.7167)

	// [database]
	p.CounterDir = cfg.Section("database").Key("counterdir").MustString("/var/smartpi")
	p.DatabaseDir = cfg.Section("database").Key("dir").MustString("/var/smartpi/db")

	// [device]
	p.I2CDevice = cfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	p.SharedDir = cfg.Section("device").Key("shared_dir").MustString("/var/tmp/smartpi")
	p.SharedFile = cfg.Section("device").Key("shared_file").MustString("values")
	p.PowerFrequency = cfg.Section("device").Key("power_frequency").MustInt(50)
	p.CTType = make(map[string]string)
	p.CTType["A"] = cfg.Section("device").Key("ct_type_1").MustString("YHDC_SCT013")
	p.CTType["B"] = cfg.Section("device").Key("ct_type_2").MustString("YHDC_SCT013")
	p.CTType["C"] = cfg.Section("device").Key("ct_type_3").MustString("YHDC_SCT013")
	p.CTType["N"] = cfg.Section("device").Key("ct_type_4").MustString("YHDC_SCT013")
	p.MeasureVoltage1 = cfg.Section("device").Key("measure_voltage_1").MustBool(true)
	p.MeasureVoltage2 = cfg.Section("device").Key("measure_voltage_2").MustBool(true)
	p.MeasureVoltage3 = cfg.Section("device").Key("measure_voltage_3").MustBool(true)
	p.MeasureCurrent = make(map[string]bool)
	p.MeasureCurrent["A"] = cfg.Section("device").Key("measure_current_1").MustBool(true)
	p.MeasureCurrent["B"] = cfg.Section("device").Key("measure_current_2").MustBool(true)
	p.MeasureCurrent["C"] = cfg.Section("device").Key("measure_current_3").MustBool(true)
	p.MeasureCurrent["N"] = true // Always measure Neutral.
	p.Voltage1 = cfg.Section("device").Key("voltage_1").MustFloat64(230)
	p.Voltage2 = cfg.Section("device").Key("voltage_2").MustFloat64(230)
	p.Voltage3 = cfg.Section("device").Key("voltage_3").MustFloat64(230)
	p.CurrentDirection1 = cfg.Section("device").Key("change_current_direction_1").MustBool(false)
	p.CurrentDirection2 = cfg.Section("device").Key("change_current_direction_2").MustBool(false)
	p.CurrentDirection3 = cfg.Section("device").Key("change_current_direction_3").MustBool(false)

	// [ftp]
	p.FTPupload = cfg.Section("ftp").Key("ftp_upload").MustBool(false)
	p.FTPserver = cfg.Section("ftp").Key("ftp_server").String()
	p.FTPuser = cfg.Section("ftp").Key("ftp_user").String()
	p.FTPpass = cfg.Section("ftp").Key("ftp_pass").String()
	p.FTPpath = cfg.Section("ftp").Key("ftp_path").String()

	// [webserver]
	p.WebserverPort = cfg.Section("webserver").Key("port").MustInt(1080)
	p.DocRoot = cfg.Section("webserver").Key("docroot").MustString("/var/smartpi/www")
	p.CSVdecimalpoint = cfg.Section("csv").Key("decimalpoint").String()
	p.CSVtimeformat = cfg.Section("csv").Key("timeformat").String()

	// [mqtt]
	p.MQTTenabled = cfg.Section("mqtt").Key("mqtt_enabled").MustBool(false)
	p.MQTTbroker = cfg.Section("mqtt").Key("mqtt_broker_url").String()
	p.MQTTbrokerport = cfg.Section("mqtt").Key("mqtt_broker_port").String()
	p.MQTTuser = cfg.Section("mqtt").Key("mqtt_username").String()
	p.MQTTpass = cfg.Section("mqtt").Key("mqtt_password").String()
	p.MQTTtopic = cfg.Section("mqtt").Key("mqtt_topic").String()
}

func NewConfig() *Config {

	t := new(Config)
	t.ReadParameterFromFile()
	return t
}
