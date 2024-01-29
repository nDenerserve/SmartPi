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
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/utils"

	log "github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
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
	SQLLiteEnabled  bool
	DatabaseDir     string
	Influxversion   string
	Influxuser      string
	Influxpassword  string
	Influxdatabase  string
	InfluxAPIToken  string
	InfluxOrg       string
	InfluxBucket    string

	// [device]
	I2CDevice            string
	PowerFrequency       float64
	Samplerate           int
	Integrator           bool
	CTType               map[models.Phase]string
	CTTypePrimaryCurrent map[models.Phase]int
	CurrentDirection     map[models.Phase]bool
	MeasureCurrent       map[models.Phase]bool
	MeasureVoltage       map[models.Phase]bool
	Voltage              map[models.Phase]float64

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
	SharedFileEnabled    bool
	SharedDir            string
	SharedFile           string
	SharedEnergyFile     string
	SharedCalculatedFile string
	WebserverPort        int
	DocRoot              string
	AppKey               string

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
	CalibrationfactorI map[models.Phase]float64
	CalibrationfactorU map[models.Phase]float64

	// [GUI]
	GUIMaxCurrent map[models.Phase]int

	// [emeter]
	EmeterEnabled          bool
	EmeterMulticastAddress string
	EmeterMulticastPort    int
	EmeterSusyID           uint16
	EmeterSerial           []byte
}

var cfg *ini.File
var err error

func (p *Config) ReadHardwareInfos() (string, string) {
	serial := ""
	model := ""

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

func (p *Config) ReadParameterFromFile() {

	serial, _ := p.ReadHardwareInfos()

	cfg, err = ini.Load("/etc/smartpi")
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
	p.SQLLiteEnabled = cfg.Section("database").Key("sqlite_enabled").MustBool(true)
	p.DatabaseDir = cfg.Section("database").Key("sqlite_dir").MustString("/var/smartpi/db")
	p.Influxversion = cfg.Section("database").Key("influxversion").MustString("2")
	p.Influxuser = cfg.Section("database").Key("influxuser").MustString("smartpi")
	p.Influxpassword = cfg.Section("database").Key("influxpassword").MustString("smart4pi")
	p.Influxdatabase = cfg.Section("database").Key("influxdatabase").MustString("http://localhost:8086")
	p.InfluxAPIToken = cfg.Section("database").Key("influxapitoken").MustString("847583öjkhldkjfg9er)/(&jljh)")
	p.InfluxOrg = cfg.Section("database").Key("influxorg").MustString("smartpi")
	p.InfluxBucket = cfg.Section("database").Key("influxbucket").MustString("meteringdata")

	// [device]
	p.I2CDevice = cfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	p.PowerFrequency = cfg.Section("device").Key("power_frequency").MustFloat64(50)
	p.Samplerate = cfg.Section("device").Key("samplerate").MustInt(1)
	p.Integrator = cfg.Section("device").Key("integrator").MustBool(false)
	p.CTType = make(map[models.Phase]string)
	p.CTType[models.PhaseA] = cfg.Section("device").Key("ct_type_1").MustString("YHDC_SCT013")
	p.CTType[models.PhaseB] = cfg.Section("device").Key("ct_type_2").MustString("YHDC_SCT013")
	p.CTType[models.PhaseC] = cfg.Section("device").Key("ct_type_3").MustString("YHDC_SCT013")
	p.CTType[models.PhaseN] = cfg.Section("device").Key("ct_type_4").MustString("YHDC_SCT013")
	p.CTTypePrimaryCurrent = make(map[models.Phase]int)
	p.CTTypePrimaryCurrent[models.PhaseA] = cfg.Section("device").Key("ct_type_1_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseB] = cfg.Section("device").Key("ct_type_2_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseC] = cfg.Section("device").Key("ct_type_3_primary_current").MustInt(100)
	p.CTTypePrimaryCurrent[models.PhaseN] = cfg.Section("device").Key("ct_type_4_primary_current").MustInt(100)
	p.CurrentDirection = make(map[models.Phase]bool)
	p.CurrentDirection[models.PhaseA] = cfg.Section("device").Key("change_current_direction_1").MustBool(false)
	p.CurrentDirection[models.PhaseB] = cfg.Section("device").Key("change_current_direction_2").MustBool(false)
	p.CurrentDirection[models.PhaseC] = cfg.Section("device").Key("change_current_direction_3").MustBool(false)
	p.CurrentDirection[models.PhaseN] = cfg.Section("device").Key("change_current_direction_4").MustBool(false)
	p.MeasureCurrent = make(map[models.Phase]bool)
	p.MeasureCurrent[models.PhaseA] = cfg.Section("device").Key("measure_current_1").MustBool(true)
	p.MeasureCurrent[models.PhaseB] = cfg.Section("device").Key("measure_current_2").MustBool(true)
	p.MeasureCurrent[models.PhaseC] = cfg.Section("device").Key("measure_current_3").MustBool(true)
	p.MeasureCurrent[models.PhaseN] = cfg.Section("device").Key("measure_current_4").MustBool(true)
	p.MeasureVoltage = make(map[models.Phase]bool)
	p.MeasureVoltage[models.PhaseA] = cfg.Section("device").Key("measure_voltage_1").MustBool(true)
	p.MeasureVoltage[models.PhaseB] = cfg.Section("device").Key("measure_voltage_2").MustBool(true)
	p.MeasureVoltage[models.PhaseC] = cfg.Section("device").Key("measure_voltage_3").MustBool(true)
	p.Voltage = make(map[models.Phase]float64)
	p.Voltage[models.PhaseA] = cfg.Section("device").Key("voltage_1").MustFloat64(230)
	p.Voltage[models.PhaseB] = cfg.Section("device").Key("voltage_2").MustFloat64(230)
	p.Voltage[models.PhaseC] = cfg.Section("device").Key("voltage_3").MustFloat64(230)

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
	p.SharedDir = cfg.Section("webserver").Key("shared_dir").MustString("/var/run")
	p.SharedFile = cfg.Section("webserver").Key("shared_file").MustString("smartpi_values")
	p.SharedEnergyFile = cfg.Section("webserver").Key("shared_energy_file").MustString("smartpi_energy_values")
	p.SharedEnergyFile = cfg.Section("webserver").Key("shared_calculated_file").MustString("smartpi_calculated_values")
	p.WebserverPort = cfg.Section("webserver").Key("port").MustInt(1080)
	p.DocRoot = cfg.Section("webserver").Key("docroot").MustString("/var/smartpi/www")
	p.AppKey = cfg.Section("webserver").Key("appkey").MustString("ew980723j35h97fqw4!234490#t33465")

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
	p.CalibrationfactorI = make(map[models.Phase]float64)
	p.CalibrationfactorI[models.PhaseA] = cfg.Section("calibration").Key("calibrationfactorI_1").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseB] = cfg.Section("calibration").Key("calibrationfactorI_2").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseC] = cfg.Section("calibration").Key("calibrationfactorI_3").MustFloat64(1)
	p.CalibrationfactorI[models.PhaseN] = cfg.Section("calibration").Key("calibrationfactorI_4").MustFloat64(1)
	p.CalibrationfactorU = make(map[models.Phase]float64)
	p.CalibrationfactorU[models.PhaseA] = cfg.Section("calibration").Key("calibrationfactorU_1").MustFloat64(1)
	p.CalibrationfactorU[models.PhaseB] = cfg.Section("calibration").Key("calibrationfactorU_2").MustFloat64(1)
	p.CalibrationfactorU[models.PhaseC] = cfg.Section("calibration").Key("calibrationfactorU_3").MustFloat64(1)

	// [GUI]
	p.GUIMaxCurrent = make(map[models.Phase]int)
	p.GUIMaxCurrent[models.PhaseA] = cfg.Section("gui").Key("gui_max_current_1").MustInt(100)
	p.GUIMaxCurrent[models.PhaseB] = cfg.Section("gui").Key("gui_max_current_2").MustInt(100)
	p.GUIMaxCurrent[models.PhaseC] = cfg.Section("gui").Key("gui_max_current_3").MustInt(100)
	p.GUIMaxCurrent[models.PhaseN] = cfg.Section("gui").Key("gui_max_current_4").MustInt(100)

	// [emeter]
	p.EmeterEnabled = cfg.Section("emeter").Key("emeter_enabled").MustBool(true)
	p.EmeterMulticastAddress = cfg.Section("emeter").Key("emeter_multicast_address").MustString("239.12.255.254")
	p.EmeterMulticastPort = cfg.Section("emeter").Key("emeter_multicast_port").MustInt(9522)
	p.EmeterSusyID = uint16(cfg.Section("emeter").Key("emeter_susy_id").MustUint(270))
	serialbytes, err := hex.DecodeString(serial[len(serial)-8:])
	if err != nil {

		log.Error(err)
		rand.Seed(time.Now().UnixNano())

		serialbytes := make([]byte, 4)
		binary.BigEndian.PutUint32(serialbytes, uint32(cfg.Section("emeter").Key("emeter_susy_id").MustUint(uint(rand.Uint32()))))
		_, err = cfg.Section("emeter").NewKey("emeter_serial", strconv.FormatUint(uint64(binary.BigEndian.Uint32(serialbytes)), 10))
		tmpFile := "/tmp/smartpi"
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
	p.EmeterSerial = serialbytes

}

func (p *Config) SaveParameterToFile() {

	var tempSendFTP [24]string

	// _, err = cfg.Section("base").NewKey("serial", p.Serial)
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
	_, err = cfg.Section("database").NewKey("sqlite_enabled", strconv.FormatBool(p.SQLLiteEnabled))
	_, err = cfg.Section("database").NewKey("sqlite_dir", p.DatabaseDir)
	_, err = cfg.Section("database").NewKey("influxversion", p.Influxversion)
	_, err = cfg.Section("database").NewKey("influxuser", p.Influxuser)
	_, err = cfg.Section("database").NewKey("influxpassword", p.Influxpassword)
	_, err = cfg.Section("database").NewKey("influxdatabase", p.Influxdatabase)
	_, err = cfg.Section("database").NewKey("influxapitoken", p.InfluxAPIToken)
	_, err = cfg.Section("database").NewKey("influxorg", p.InfluxOrg)
	_, err = cfg.Section("database").NewKey("influxbucket", p.InfluxBucket)

	// [device]
	_, err = cfg.Section("device").NewKey("i2c_device", p.I2CDevice)
	_, err = cfg.Section("device").NewKey("power_frequency", strconv.FormatInt(int64(p.PowerFrequency), 10))
	_, err = cfg.Section("device").NewKey("samplerate", strconv.FormatInt(int64(p.Samplerate), 10))
	_, err = cfg.Section("device").NewKey("integrator", strconv.FormatBool(p.Integrator))
	_, err = cfg.Section("device").NewKey("ct_type_1", p.CTType[models.PhaseA])
	_, err = cfg.Section("device").NewKey("ct_type_2", p.CTType[models.PhaseB])
	_, err = cfg.Section("device").NewKey("ct_type_3", p.CTType[models.PhaseC])
	_, err = cfg.Section("device").NewKey("ct_type_4", p.CTType[models.PhaseN])

	_, err = cfg.Section("device").NewKey("ct_type_1_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseA]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_2_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseB]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_3_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseC]), 10))
	_, err = cfg.Section("device").NewKey("ct_type_4_primary_current", strconv.FormatInt(int64(p.CTTypePrimaryCurrent[models.PhaseN]), 10))

	_, err = cfg.Section("device").NewKey("change_current_direction_1", strconv.FormatBool(p.CurrentDirection[models.PhaseA]))
	_, err = cfg.Section("device").NewKey("change_current_direction_2", strconv.FormatBool(p.CurrentDirection[models.PhaseB]))
	_, err = cfg.Section("device").NewKey("change_current_direction_3", strconv.FormatBool(p.CurrentDirection[models.PhaseC]))
	_, err = cfg.Section("device").NewKey("change_current_direction_4", strconv.FormatBool(p.CurrentDirection[models.PhaseN]))

	_, err = cfg.Section("device").NewKey("measure_current_1", strconv.FormatBool(p.MeasureCurrent[models.PhaseA]))
	_, err = cfg.Section("device").NewKey("measure_current_2", strconv.FormatBool(p.MeasureCurrent[models.PhaseB]))
	_, err = cfg.Section("device").NewKey("measure_current_3", strconv.FormatBool(p.MeasureCurrent[models.PhaseC]))
	_, err = cfg.Section("device").NewKey("measure_current_4", strconv.FormatBool(p.MeasureCurrent[models.PhaseN]))

	_, err = cfg.Section("device").NewKey("measure_voltage_1", strconv.FormatBool(p.MeasureVoltage[models.PhaseA]))
	_, err = cfg.Section("device").NewKey("measure_voltage_2", strconv.FormatBool(p.MeasureVoltage[models.PhaseB]))
	_, err = cfg.Section("device").NewKey("measure_voltage_3", strconv.FormatBool(p.MeasureVoltage[models.PhaseC]))

	_, err = cfg.Section("device").NewKey("voltage_1", strconv.FormatFloat(p.Voltage[models.PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_2", strconv.FormatFloat(p.Voltage[models.PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("voltage_3", strconv.FormatFloat(p.Voltage[models.PhaseC], 'f', -1, 64))

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
	_, err = cfg.Section("webserver").NewKey("shared_file", p.SharedFile)
	_, err = cfg.Section("webserver").NewKey("shared_energy_file", p.SharedEnergyFile)
	_, err = cfg.Section("webserver").NewKey("shared_calculated_file", p.SharedCalculatedFile)
	_, err = cfg.Section("webserver").NewKey("port", strconv.FormatInt(int64(p.WebserverPort), 10))
	_, err = cfg.Section("webserver").NewKey("docroot", p.DocRoot)
	_, err = cfg.Section("appkey").NewKey("appkey", p.AppKey)

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
	_, err = cfg.Section("device").NewKey("calibrationfactorI_1", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_2", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_3", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseC], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorI_4", strconv.FormatFloat(p.CalibrationfactorI[models.PhaseN], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_1", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseA], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_2", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseB], 'f', -1, 64))
	_, err = cfg.Section("device").NewKey("calibrationfactorU_3", strconv.FormatFloat(p.CalibrationfactorU[models.PhaseC], 'f', -1, 64))

	_, err = cfg.Section("gui").NewKey("gui_max_current_1", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseA]), 10))
	_, err = cfg.Section("gui").NewKey("gui_max_current_2", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseB]), 10))
	_, err = cfg.Section("gui").NewKey("gui_max_current_3", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseC]), 10))
	_, err = cfg.Section("gui").NewKey("gui_max_current_4", strconv.FormatInt(int64(p.GUIMaxCurrent[models.PhaseN]), 10))

	// [emeter]
	_, err = cfg.Section("emeter").NewKey("emeter_enabled", strconv.FormatBool(p.EmeterEnabled))
	_, err = cfg.Section("emeter").NewKey("emeter_multicast_address", p.EmeterMulticastAddress)
	_, err = cfg.Section("emeter").NewKey("emeter_multicast_port", strconv.FormatInt(int64(p.EmeterMulticastPort), 10))
	_, err = cfg.Section("emeter").NewKey("emeter_susy_id", strconv.FormatUint(uint64(p.EmeterSusyID), 10))
	_, err = cfg.Section("emeter").NewKey("emeter_serial", strconv.FormatUint(uint64(binary.BigEndian.Uint32(p.EmeterSerial)), 10))

	tmpFile := "/tmp/smartpi"
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

func NewConfig() *Config {

	t := new(Config)
	t.ReadParameterFromFile()
	return t
}
