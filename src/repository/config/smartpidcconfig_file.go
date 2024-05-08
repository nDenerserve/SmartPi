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
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/utils"

	log "github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
)

type DCconfig struct {
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
	CounterEnabled         bool
	CounterDir             string
	DatabaseEnabled        bool
	DatabaseStoreIntervall string
	Influxversion          string
	Influxuser             string
	Influxpassword         string
	Influxdatabase         string
	InfluxAPIToken         string
	InfluxOrg              string
	InfluxBucket           string

	// [device]
	I2CDevice              string
	ADCAddress             []byte
	InputType              map[int]string
	InputName              map[int]string
	InputCalibrationOffset map[int]float64
	Samplerate             int

	// [calculations]
	Power                 map[int][]int
	PowerName             map[int]string
	EnergyProductionName  map[int]string
	EnergyConsumptionName map[int]string
	EnergyBalancedName    map[int]string

	// [ftp]
	FTPupload    bool
	FTPserver    string
	FTPuser      string
	FTPpass      string
	FTPpath      string
	FTPcsv       bool
	FTPxml       bool
	FTPsendtimes [24]bool

	// [shared]
	SharedFileEnabled    bool
	SharedDir            string
	SharedFile           string
	SharedEnergyFile     string
	SharedCalculatedFile string

	// [csv]
	CSVdecimalpoint string
	CSVtimeformat   string

	// [mqtt]
	MQTTenabled          bool
	MQTTbrokerscheme     string
	MQTTbroker           string
	MQTTbrokerport       string
	MQTTuser             string
	MQTTpass             string
	MQTTtopic            string
	MQTTpublishintervall string

	// [modbus slave]
	ModbusRTUenabled bool
	ModbusTCPenabled bool
	ModbusRTUAddress uint8
	ModbusRTUDevice  string
	ModbusTCPAddress string
}

var dccfg *ini.File
var dcerr error

func (p *DCconfig) ReadDCParameterFromFile() {

	dccfg, dcerr = ini.LooseLoad("/etc/smartpidc")
	if dcerr != nil {
		panic(dcerr)
	}

	// [base]
	p.Serial = dccfg.Section("base").Key("serial").String()
	p.Name = dccfg.Section("base").Key("name").MustString("House")
	// Handle logging levels
	p.LogLevel, dcerr = log.ParseLevel(dccfg.Section("base").Key("loglevel").MustString("info"))
	if dcerr != nil {
		panic(dcerr)
	}
	// Handle old debuglevel config key as log.Debug.
	p.DebugLevel, dcerr = dccfg.Section("base").Key("debuglevel").Int()
	if dcerr == nil && p.DebugLevel > 0 {
		p.LogLevel = log.DebugLevel
		log.Debug("Config option debuglevel is deprecated, use loglevel=debug.")
	} else {
		p.DebugLevel = 0
	}
	p.MetricsListenAddress = dccfg.Section("base").Key("metrics_listen_address").MustString(":9246")

	// [location]
	p.Lat = dccfg.Section("location").Key("lat").MustFloat64(52.3667)
	p.Lng = dccfg.Section("location").Key("lng").MustFloat64(9.7167)

	// [database]
	p.CounterEnabled = dccfg.Section("database").Key("counter_enabled").MustBool(true)
	p.CounterDir = dccfg.Section("database").Key("counterdir").MustString("/var/smartpi")
	p.DatabaseEnabled = dccfg.Section("database").Key("database_enabled").MustBool(true)
	p.DatabaseStoreIntervall = dccfg.Section("database").Key("database_storeintervall").MustString("minute")
	p.Influxversion = dccfg.Section("database").Key("influxversion").MustString("2")
	p.Influxuser = dccfg.Section("database").Key("influxuser").MustString("smartpi")
	p.Influxpassword = dccfg.Section("database").Key("influxpassword").MustString("smart4pi")
	p.Influxdatabase = dccfg.Section("database").Key("influxdatabase").MustString("http://localhost:8086")
	p.InfluxAPIToken = dccfg.Section("database").Key("influxapitoken").MustString("847583öjkhldkjfg9er)/(&jljh)")
	p.InfluxOrg = dccfg.Section("database").Key("influxorg").MustString("smartpi")
	p.InfluxBucket = dccfg.Section("database").Key("influxbucket").MustString("meteringdata")

	// [device]
	p.I2CDevice = dccfg.Section("device").Key("i2c_device").MustString("/dev/i2c-1")
	adcaddresses := strings.Split(dccfg.Section("device").Key("adc_address").MustString("0x6e"), ",")

	if len(adcaddresses) == 0 {
		adcaddresses = append(adcaddresses, "0x6e")
	}

	for _, element := range adcaddresses {
		hexval, _ := hex.DecodeString(strings.Trim(element, "0x"))
		p.ADCAddress = append(p.ADCAddress, hexval[0])
	}

	p.Samplerate = dccfg.Section("device").Key("samplerate").MustInt(1)

	// [input]
	p.InputType = make(map[int]string)
	p.InputType[models.Input1] = dccfg.Section("input").Key("input_type_1").MustString("Voltage 0-5V")
	p.InputType[models.Input2] = dccfg.Section("input").Key("input_type_2").MustString("Voltage 0-5V")
	p.InputType[models.Input3] = dccfg.Section("input").Key("input_type_3").MustString("HSTS016L 10A")
	p.InputType[models.Input4] = dccfg.Section("input").Key("input_type_4").MustString("HSTS016L 10A")
	p.InputCalibrationOffset = make(map[int]float64)
	p.InputCalibrationOffset[models.Input1] = dccfg.Section("input").Key("input_calibration_offset_1").MustFloat64(0.0)
	p.InputCalibrationOffset[models.Input2] = dccfg.Section("input").Key("input_calibration_offset_2").MustFloat64(0.0)
	p.InputCalibrationOffset[models.Input3] = dccfg.Section("input").Key("input_calibration_offset_3").MustFloat64(0.0)
	p.InputCalibrationOffset[models.Input4] = dccfg.Section("input").Key("input_calibration_offset_4").MustFloat64(0.0)
	p.InputName = make(map[int]string)
	p.InputName[models.Input1] = dccfg.Section("input").Key("input_name_1").MustString("U1")
	p.InputName[models.Input2] = dccfg.Section("input").Key("input_name_2").MustString("U2")
	p.InputName[models.Input3] = dccfg.Section("input").Key("input_name_3").MustString("I1")
	p.InputName[models.Input4] = dccfg.Section("input").Key("input_name_4").MustString("I2")

	// [calculations]
	p.Power = make(map[int][]int)
	p.Power[0] = dccfg.Section("calculations").Key("power_1").Ints("|")
	p.Power[1] = dccfg.Section("calculations").Key("power_2").Ints("|")
	p.Power[2] = dccfg.Section("calculations").Key("power_3").Ints("|")
	p.PowerName = make(map[int]string)
	p.PowerName[0] = dccfg.Section("calculations").Key("power_name_1").MustString("P1")
	p.PowerName[1] = dccfg.Section("calculations").Key("power_name_2").MustString("P2")
	p.PowerName[2] = dccfg.Section("calculations").Key("power_name_3").MustString("P3")
	p.EnergyProductionName = make(map[int]string)
	p.EnergyProductionName[0] = dccfg.Section("calculations").Key("energy_production_name_1").MustString("Ep1")
	p.EnergyProductionName[1] = dccfg.Section("calculations").Key("energy_production_name_2").MustString("Ep2")
	p.EnergyProductionName[2] = dccfg.Section("calculations").Key("energy_production_name_3").MustString("Ep3")
	p.EnergyConsumptionName = make(map[int]string)
	p.EnergyConsumptionName[0] = dccfg.Section("calculations").Key("energy_consumption_name_1").MustString("Ec1")
	p.EnergyConsumptionName[1] = dccfg.Section("calculations").Key("energy_consumption_name_2").MustString("Ec2")
	p.EnergyConsumptionName[2] = dccfg.Section("calculations").Key("energy_consumption_name_3").MustString("Ec3")
	p.EnergyBalancedName = make(map[int]string)
	p.EnergyBalancedName[0] = dccfg.Section("calculations").Key("energy_balanced_name_1").MustString("Eb1")
	p.EnergyBalancedName[1] = dccfg.Section("calculations").Key("energy_balanced_name_2").MustString("Eb2")
	p.EnergyBalancedName[2] = dccfg.Section("calculations").Key("energy_balanced_name_3").MustString("Eb3")

	// [ftp]
	p.FTPupload = dccfg.Section("ftp").Key("ftp_upload").MustBool(false)
	p.FTPserver = dccfg.Section("ftp").Key("ftp_server").String()
	p.FTPuser = dccfg.Section("ftp").Key("ftp_user").String()
	p.FTPpass = dccfg.Section("ftp").Key("ftp_pass").String()
	p.FTPpath = dccfg.Section("ftp").Key("ftp_path").String()
	p.FTPcsv = dccfg.Section("ftp").Key("ftp_csv").MustBool(true)
	p.FTPxml = dccfg.Section("ftp").Key("ftp_xml").MustBool(true)
	sendtimes := strings.Split(dccfg.Section("ftp").Key("ftp_sendtimes").MustString("1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"), ",")
	for i, r := range sendtimes {
		p.FTPsendtimes[i], _ = strconv.ParseBool(r)
	}

	// [webserver]
	p.SharedFileEnabled = dccfg.Section("sharedfile").Key("shared_file_enabled").MustBool(true)
	p.SharedDir = dccfg.Section("sharedfile").Key("shared_dir").MustString("/var/run")
	p.SharedFile = dccfg.Section("sharedfile").Key("shared_file").MustString("smartpi_values")
	p.SharedEnergyFile = dccfg.Section("sharedfile").Key("shared_energy_file").MustString("smartpi_energy_values")
	p.SharedEnergyFile = dccfg.Section("sharedfile").Key("shared_calculated_file").MustString("smartpi_calculated_values")

	// [csv]
	p.CSVdecimalpoint = dccfg.Section("csv").Key("decimalpoint").String()
	p.CSVtimeformat = dccfg.Section("csv").Key("timeformat").String()

	// [mqtt]
	p.MQTTenabled = dccfg.Section("mqtt").Key("mqtt_enabled").MustBool(false)
	p.MQTTbrokerscheme = dccfg.Section("mqtt").Key("mqtt_broker_scheme").MustString("tcp://")
	p.MQTTbroker = dccfg.Section("mqtt").Key("mqtt_broker_url").String()
	p.MQTTbrokerport = dccfg.Section("mqtt").Key("mqtt_broker_port").String()
	p.MQTTuser = dccfg.Section("mqtt").Key("mqtt_username").String()
	p.MQTTpass = dccfg.Section("mqtt").Key("mqtt_password").String()
	p.MQTTtopic = dccfg.Section("mqtt").Key("mqtt_topic").String()
	p.MQTTpublishintervall = dccfg.Section("mqtt").Key("mqtt_publishintervall").MustString("sample")

	// [modbus slave]
	p.ModbusRTUenabled = dccfg.Section("modbus").Key("modbus_rtu_enabled").MustBool(false)
	p.ModbusTCPenabled = dccfg.Section("modbus").Key("modbus_tcp_enabled").MustBool(false)
	p.ModbusRTUAddress = uint8(dccfg.Section("modbus").Key("modbus_rtu_address").MustInt(1))
	p.ModbusRTUDevice = dccfg.Section("modbus").Key("modbus_rtu_device_id").MustString("/dev/serial0")
	p.ModbusTCPAddress = dccfg.Section("modbus").Key("modbus_tcp_address").MustString(":502")

}

func (p *DCconfig) SaveDCParameterToFile() {

	var tempSendFTP [24]string

	_, dcerr = dccfg.Section("base").NewKey("serial", p.Serial)
	_, dcerr = dccfg.Section("base").NewKey("name", p.Name)
	_, dcerr = dccfg.Section("base").NewKey("loglevel", p.LogLevel.String())
	_, dcerr = dccfg.Section("base").NewKey("metrics_listen_address", p.MetricsListenAddress)

	// [location]
	_, dcerr = dccfg.Section("location").NewKey("lat", strconv.FormatFloat(p.Lat, 'f', -1, 64))
	_, dcerr = dccfg.Section("location").NewKey("lng", strconv.FormatFloat(p.Lng, 'f', -1, 64))

	// [database]
	_, dcerr = dccfg.Section("database").NewKey("counter_enabled", strconv.FormatBool(p.CounterEnabled))
	_, dcerr = dccfg.Section("database").NewKey("counterdir", p.CounterDir)
	_, dcerr = dccfg.Section("database").NewKey("database_enabled", strconv.FormatBool(p.DatabaseEnabled))
	_, dcerr = dccfg.Section("database").NewKey("database_storeintervall", p.DatabaseStoreIntervall)
	_, dcerr = dccfg.Section("database").NewKey("influxversion", p.Influxversion)
	_, dcerr = dccfg.Section("database").NewKey("influxuser", p.Influxuser)
	_, dcerr = dccfg.Section("database").NewKey("influxpassword", p.Influxpassword)
	_, dcerr = dccfg.Section("database").NewKey("influxdatabase", p.Influxdatabase)
	_, dcerr = dccfg.Section("database").NewKey("influxapitoken", p.InfluxAPIToken)
	_, dcerr = dccfg.Section("database").NewKey("influxorg", p.InfluxOrg)
	_, dcerr = dccfg.Section("database").NewKey("influxbucket", p.InfluxBucket)

	// [device]
	_, dcerr = dccfg.Section("device").NewKey("i2c_device", p.I2CDevice)

	var slice []byte
	var hexstring []string
	slice = nil
	hexstring = nil
	for _, element := range p.ADCAddress {
		slice = append(slice, element)
		hexstring = append(hexstring, "0x"+hex.EncodeToString(slice))
	}
	_, dcerr = dccfg.Section("device").NewKey("adc_address", strings.Join(hexstring, ","))
	_, dcerr = dccfg.Section("device").NewKey("samplerate", strconv.FormatInt(int64(p.Samplerate), 10))

	// [input]
	_, dcerr = dccfg.Section("input").NewKey("input_type_1", p.InputType[models.Input1])
	_, dcerr = dccfg.Section("input").NewKey("input_type_2", p.InputType[models.Input2])
	_, dcerr = dccfg.Section("input").NewKey("input_type_3", p.InputType[models.Input3])
	_, dcerr = dccfg.Section("input").NewKey("input_type_4", p.InputType[models.Input4])
	_, dcerr = dccfg.Section("input").NewKey("input_calibration_offset_1", strconv.FormatFloat(p.InputCalibrationOffset[models.Input1], 'f', -1, 64))
	_, dcerr = dccfg.Section("input").NewKey("input_calibration_offset_2", strconv.FormatFloat(p.InputCalibrationOffset[models.Input2], 'f', -1, 64))
	_, dcerr = dccfg.Section("input").NewKey("input_calibration_offset_3", strconv.FormatFloat(p.InputCalibrationOffset[models.Input3], 'f', -1, 64))
	_, dcerr = dccfg.Section("input").NewKey("input_calibration_offset_4", strconv.FormatFloat(p.InputCalibrationOffset[models.Input4], 'f', -1, 64))
	_, dcerr = dccfg.Section("input").NewKey("input_name_1", p.InputName[models.Input1])
	_, dcerr = dccfg.Section("input").NewKey("input_name_2", p.InputName[models.Input2])
	_, dcerr = dccfg.Section("input").NewKey("input_name_3", p.InputName[models.Input3])
	_, dcerr = dccfg.Section("input").NewKey("input_name_4", p.InputName[models.Input4])

	// [calculations]
	_, dcerr = dccfg.Section("calculations").NewKey("power_1", strings.Join(utils.Int2StringSlice(p.Power[0][:]), "|"))
	_, dcerr = dccfg.Section("calculations").NewKey("power_2", strings.Join(utils.Int2StringSlice(p.Power[1][:]), "|"))
	_, dcerr = dccfg.Section("calculations").NewKey("power_3", strings.Join(utils.Int2StringSlice(p.Power[2][:]), "|"))
	_, dcerr = dccfg.Section("calculations").NewKey("power_name_1", p.PowerName[0])
	_, dcerr = dccfg.Section("calculations").NewKey("power_name_3", p.PowerName[2])
	_, dcerr = dccfg.Section("calculations").NewKey("power_name_2", p.PowerName[1])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_production_name_1", p.EnergyProductionName[0])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_production_name_2", p.EnergyProductionName[1])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_production_name_3", p.EnergyProductionName[2])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_consumption_name_1", p.EnergyConsumptionName[0])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_consumption_name_2", p.EnergyConsumptionName[1])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_consumption_name_3", p.EnergyConsumptionName[2])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_balanced_name_1", p.EnergyBalancedName[0])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_balanced_name_2", p.EnergyBalancedName[1])
	_, dcerr = dccfg.Section("calculations").NewKey("energy_balanced_name_3", p.EnergyBalancedName[2])

	// [ftp]
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_upload", strconv.FormatBool(p.FTPupload))
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_server", p.FTPserver)
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_user", p.FTPuser)
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_pass", p.FTPpass)
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_path", p.FTPpath)
	for i, r := range p.FTPsendtimes {
		tempSendFTP[i] = strconv.FormatBool(r)
	}
	_, dcerr = dccfg.Section("ftp").NewKey("ftp_sendtimes", strings.Join(tempSendFTP[:], ","))

	// [sharedfile]
	_, dcerr = dccfg.Section("sharedfile").NewKey("shared_file_enabled", strconv.FormatBool(p.SharedFileEnabled))
	_, dcerr = dccfg.Section("sharedfile").NewKey("shared_dir", p.SharedDir)
	_, dcerr = dccfg.Section("sharedfile").NewKey("shared_file", p.SharedFile)
	_, dcerr = dccfg.Section("sharedfile").NewKey("shared_energy_file", p.SharedEnergyFile)
	_, dcerr = dccfg.Section("sharedfile").NewKey("shared_calculated_file", p.SharedCalculatedFile)

	// [csv]
	_, dcerr = dccfg.Section("csv").NewKey("decimalpoint", p.CSVdecimalpoint)
	_, dcerr = dccfg.Section("csv").NewKey("timeformat", p.CSVtimeformat)

	// [mqtt]
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_enabled", strconv.FormatBool(p.MQTTenabled))
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_broker_scheme", p.MQTTbrokerscheme)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_broker_url", p.MQTTbroker)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_broker_port", p.MQTTbrokerport)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_username", p.MQTTuser)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_password", p.MQTTpass)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_topic", p.MQTTtopic)
	_, dcerr = dccfg.Section("mqtt").NewKey("mqtt_publishintervall", p.MQTTpublishintervall)

	// [modbus slave]
	_, dcerr = dccfg.Section("modbus").NewKey("modbus_rtu_enabled", strconv.FormatBool(p.ModbusRTUenabled))
	_, dcerr = dccfg.Section("modbus").NewKey("modbus_tcp_enabled", strconv.FormatBool(p.ModbusTCPenabled))
	_, dcerr = dccfg.Section("modbus").NewKey("modbus_rtu_address", strconv.FormatUint(uint64(p.ModbusRTUAddress), 10))
	_, dcerr = dccfg.Section("modbus").NewKey("modbus_rtu_device_id", p.ModbusRTUDevice)
	_, dcerr = dccfg.Section("modbus").NewKey("modbus_tcp_address", p.ModbusTCPAddress)

	tmpFile := "/tmp/smartpidc"
	dcerr := dccfg.SaveTo(tmpFile)
	if dcerr != nil {
		panic(dcerr)
	}

	srcFile, dcerr := os.Open(tmpFile)
	utils.Checklog(dcerr)
	defer srcFile.Close()

	destFile, dcerr := os.Create("/etc/smartpidc") // creates if file doesn't exist
	utils.Checklog(dcerr)
	defer destFile.Close()

	_, dcerr = io.Copy(destFile, srcFile)
	utils.Checklog(dcerr)

	dcerr = destFile.Sync()
	utils.Checklog(dcerr)

	defer os.Remove(tmpFile)
}

func NewDCconfig() *DCconfig {

	t := new(DCconfig)
	t.ReadDCParameterFromFile()
	return t
}
