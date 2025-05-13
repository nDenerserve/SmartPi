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
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/utils"

	log "github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
)

type SmartPiConfig struct {
	// [base]
	Serial   string
	Name     string
	LogLevel log.Level
	// DebugLevel           int
	MetricsListenAddress string

	// [location]
	Lat float64
	Lng float64

	// [database]
	DatabaseEnabled bool
	Influxversion   string
	Influxuser      string
	Influxpassword  string
	Influxdatabase  string
	InfluxAPIToken  string
	InfluxOrg       string
	InfluxBucket    string

	// [ftp]
	FTPupload    bool
	FTPserver    string
	FTPuser      string
	FTPpass      string
	FTPpath      string
	FTPcsv       bool
	FTPxml       bool
	FTPsendtimes [24]bool

	// [webserver]
	SharedFileEnabled bool
	SharedDir         string
	WebserverPort     int
	DocRoot           string
	AppKey            string
	Dashboard         string
	SecureValues      bool

	// [csv]
	CSVdecimalpoint string
	CSVtimeformat   string

	// [mqtt]
	MQTTenabled      bool
	MQTTbrokerscheme string
	MQTTbroker       string
	MQTTbrokerport   string
	MQTTuser         string
	MQTTpass         string
	MQTTtopic        string
	MQTTQoS          uint8

	// [modbus slave]
	ModbusRTUenabled bool
	ModbusTCPenabled bool
	ModbusRTUAddress uint8
	ModbusRTUDevice  string
	ModbusTCPAddress string
}

var cfg *ini.File
var err error

func (p *SmartPiConfig) ReadHardwareInfos() (string, string) {
	serial := "0000000000000000"
	model := "unknown model"

	file, err := os.Open("/proc/cpuinfo")
	utils.Checklog(err)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Model") {
			substring := strings.Split(line, ": ")
			if len(substring) > 1 {
				model = (substring[len(substring)-1])
			}
		} else if strings.Contains(line, "Serial") {
			substring := strings.Split(line, ": ")
			if len(substring) > 1 {
				serial = (substring[len(substring)-1])
			}
		}
	}
	return serial, model
}

func (p *SmartPiConfig) ReadParameterFromFile() {

	log.Debug("Read SmartPi-Config from file")

	serial, _ := p.ReadHardwareInfos()

	cfg, err = ini.LooseLoad("/etc/smartpi")
	if err != nil {
		panic(err)
	}

	// [base]
	// p.Serial = cfg.Section("base").Key("serial").String()
	p.Serial = serial
	p.Name = cfg.Section("base").Key("name").MustString("SmartPi " + serial)
	// Handle logging levels
	p.LogLevel, err = log.ParseLevel(cfg.Section("base").Key("loglevel").MustString("info"))
	if err != nil {
		panic(err)
	}
	// Handle old debuglevel config key as log.Debug.
	// p.DebugLevel, err = cfg.Section("base").Key("debuglevel").Int()
	// if err == nil && p.DebugLevel > 0 {
	// 	p.LogLevel = log.DebugLevel
	// 	log.Debug("Config option debuglevel is deprecated, use loglevel=debug.")
	// } else {
	// 	p.DebugLevel = 0
	// }
	p.MetricsListenAddress = cfg.Section("base").Key("metrics_listen_address").MustString(":9246")

	// [location]
	p.Lat = cfg.Section("location").Key("lat").MustFloat64(52.3192358)
	p.Lng = cfg.Section("location").Key("lng").MustFloat64(9.7189549)

	// [database]
	p.DatabaseEnabled = cfg.Section("database").Key("database_enabled").MustBool(true)
	p.Influxversion = cfg.Section("database").Key("influxversion").MustString("2")
	p.Influxuser = cfg.Section("database").Key("influxuser").MustString("smartpi")
	p.Influxpassword = cfg.Section("database").Key("influxpassword").MustString("smart4pi")
	p.Influxdatabase = cfg.Section("database").Key("influxdatabase").MustString("http://localhost:8086")
	p.InfluxAPIToken = cfg.Section("database").Key("influxapitoken").MustString("cg_gCGlRKeox4XiD7ti55gZKIhwlfknH7HKJo_hczmhjh_Dkutz291oAF82GHkEG8HfVGAQWKwZIuXJGwtdtLw==")
	p.InfluxOrg = cfg.Section("database").Key("influxorg").MustString("smartpi")
	p.InfluxBucket = cfg.Section("database").Key("influxbucket").MustString("meteringdata")

	// [ftp]
	p.FTPupload = cfg.Section("ftp").Key("ftp_upload").MustBool(false)
	p.FTPserver = cfg.Section("ftp").Key("ftp_server").String()
	p.FTPuser = cfg.Section("ftp").Key("ftp_user").String()
	p.FTPpass = cfg.Section("ftp").Key("ftp_pass").String()
	p.FTPpath = cfg.Section("ftp").Key("ftp_path").String()
	p.FTPcsv = cfg.Section("ftp").Key("ftp_csv").MustBool(true)
	p.FTPxml = cfg.Section("ftp").Key("ftp_xml").MustBool(true)
	sendtimes := strings.Split(cfg.Section("ftp").Key("ftp_sendtimes").MustString("1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"), ",")
	for i, r := range sendtimes {
		p.FTPsendtimes[i], _ = strconv.ParseBool(r)
	}

	// [webserver]
	p.SharedFileEnabled = cfg.Section("webserver").Key("shared_file_enabled").MustBool(true)
	p.SharedDir = cfg.Section("webserver").Key("shared_dir").MustString("/var/run/")
	p.WebserverPort = cfg.Section("webserver").Key("port").MustInt(1080)
	p.DocRoot = cfg.Section("webserver").Key("docroot").MustString("/var/smartpi/www")
	p.AppKey = cfg.Section("webserver").Key("appkey").MustString("ew980723j35h97fqw4!234490#t33465")
	p.SecureValues = cfg.Section("webserver").Key("secure_values").MustBool(false)

	// [csv]
	p.CSVdecimalpoint = cfg.Section("csv").Key("decimalpoint").String()
	p.CSVtimeformat = cfg.Section("csv").Key("timeformat").String()

	// [mqtt]
	p.MQTTenabled = cfg.Section("mqtt").Key("mqtt_enabled").MustBool(false)
	p.MQTTbrokerscheme = cfg.Section("mqtt").Key("mqtt_broker_scheme").MustString("tcp://")
	p.MQTTbroker = cfg.Section("mqtt").Key("mqtt_broker_url").String()
	p.MQTTbrokerport = cfg.Section("mqtt").Key("mqtt_broker_port").String()
	p.MQTTuser = cfg.Section("mqtt").Key("mqtt_username").String()
	p.MQTTpass = cfg.Section("mqtt").Key("mqtt_password").String()
	p.MQTTtopic = cfg.Section("mqtt").Key("mqtt_topic").String()
	p.MQTTQoS = uint8(cfg.Section("mqtt").Key("mqtt_qos").MustUint(0))

	// [modbus slave]
	p.ModbusRTUenabled = cfg.Section("modbus").Key("modbus_rtu_enabled").MustBool(false)
	p.ModbusTCPenabled = cfg.Section("modbus").Key("modbus_tcp_enabled").MustBool(false)
	p.ModbusRTUAddress = uint8(cfg.Section("modbus").Key("modbus_rtu_address").MustInt(1))
	p.ModbusRTUDevice = cfg.Section("modbus").Key("modbus_rtu_device_id").MustString("/dev/serial0")
	p.ModbusTCPAddress = cfg.Section("modbus").Key("modbus_tcp_address").MustString(":502")

}

func (p *SmartPiConfig) SaveParameterToFile() {

	log.Debug("Write SmartPi-Config to file")

	var tempSendFTP [24]string

	// _, err = cfg.Section("base").NewKey("serial", p.Serial)
	_, err = cfg.Section("base").NewKey("name", p.Name)
	_, err = cfg.Section("base").NewKey("loglevel", p.LogLevel.String())
	_, err = cfg.Section("base").NewKey("metrics_listen_address", p.MetricsListenAddress)

	// [location]
	_, err = cfg.Section("location").NewKey("lat", strconv.FormatFloat(p.Lat, 'f', -1, 64))
	_, err = cfg.Section("location").NewKey("lng", strconv.FormatFloat(p.Lng, 'f', -1, 64))

	// [database]
	_, err = cfg.Section("database").NewKey("database_enabled", strconv.FormatBool(p.DatabaseEnabled))
	_, err = cfg.Section("database").NewKey("influxversion", p.Influxversion)
	_, err = cfg.Section("database").NewKey("influxuser", p.Influxuser)
	_, err = cfg.Section("database").NewKey("influxpassword", p.Influxpassword)
	_, err = cfg.Section("database").NewKey("influxdatabase", p.Influxdatabase)
	_, err = cfg.Section("database").NewKey("influxapitoken", p.InfluxAPIToken)
	_, err = cfg.Section("database").NewKey("influxorg", p.InfluxOrg)
	_, err = cfg.Section("database").NewKey("influxbucket", p.InfluxBucket)

	// [ftp]
	_, err = cfg.Section("ftp").NewKey("ftp_upload", strconv.FormatBool(p.FTPupload))
	_, err = cfg.Section("ftp").NewKey("ftp_server", p.FTPserver)
	_, err = cfg.Section("ftp").NewKey("ftp_user", p.FTPuser)
	_, err = cfg.Section("ftp").NewKey("ftp_pass", p.FTPpass)
	_, err = cfg.Section("ftp").NewKey("ftp_path", p.FTPpath)
	for i, r := range p.FTPsendtimes {
		tempSendFTP[i] = strconv.FormatBool(r)
	}
	_, err = cfg.Section("ftp").NewKey("ftp_sendtimes", strings.Join(tempSendFTP[:], ","))

	// [webserver]
	_, err = cfg.Section("webserver").NewKey("shared_file_enabled", strconv.FormatBool(p.SharedFileEnabled))
	_, err = cfg.Section("webserver").NewKey("shared_dir", p.SharedDir)
	_, err = cfg.Section("webserver").NewKey("port", strconv.FormatInt(int64(p.WebserverPort), 10))
	_, err = cfg.Section("webserver").NewKey("docroot", p.DocRoot)
	_, err = cfg.Section("webserver").NewKey("appkey", p.AppKey)
	_, err = cfg.Section("webserver").NewKey("secure_values", strconv.FormatBool(p.SecureValues))

	// [csv]
	_, err = cfg.Section("csv").NewKey("decimalpoint", p.CSVdecimalpoint)
	_, err = cfg.Section("csv").NewKey("timeformat", p.CSVtimeformat)

	// [mqtt]
	_, err = cfg.Section("mqtt").NewKey("mqtt_enabled", strconv.FormatBool(p.MQTTenabled))
	_, err = cfg.Section("mqtt").NewKey("mqtt_broker_scheme", p.MQTTbrokerscheme)
	_, err = cfg.Section("mqtt").NewKey("mqtt_broker_url", p.MQTTbroker)
	_, err = cfg.Section("mqtt").NewKey("mqtt_broker_port", p.MQTTbrokerport)
	_, err = cfg.Section("mqtt").NewKey("mqtt_username", p.MQTTuser)
	_, err = cfg.Section("mqtt").NewKey("mqtt_password", p.MQTTpass)
	_, err = cfg.Section("mqtt").NewKey("mqtt_topic", p.MQTTtopic)
	_, err = cfg.Section("mqtt").NewKey("mqtt_qos", strconv.FormatUint(uint64(p.MQTTQoS), 10))

	// [modbus slave]
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_enabled", strconv.FormatBool(p.ModbusRTUenabled))
	_, err = cfg.Section("modbus").NewKey("modbus_tcp_enabled", strconv.FormatBool(p.ModbusTCPenabled))
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_address", strconv.FormatUint(uint64(p.ModbusRTUAddress), 10))
	_, err = cfg.Section("modbus").NewKey("modbus_rtu_device_id", p.ModbusRTUDevice)
	_, err = cfg.Section("modbus").NewKey("modbus_tcp_address", p.ModbusTCPAddress)

	tmpFile := "/tmp/smartpi_test"
	err := cfg.SaveTo(tmpFile)
	if err != nil {
		panic(err)
	}

	srcFile, err := os.Open(tmpFile)
	utils.Checklog(err)
	defer srcFile.Close()

	destFile, err := os.Create("/etc/smartpi") // creates if file doesn't exist
	utils.Checklog(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	utils.Checklog(err)

	err = destFile.Sync()
	utils.Checklog(err)

	defer os.Remove(tmpFile)
}

func NewSmartPiConfig() *SmartPiConfig {

	t := new(SmartPiConfig)
	t.ReadParameterFromFile()
	return t
}
