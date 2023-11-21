package config

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type Moduleconfig struct {
	// [base]
	I2CDevice string
	Webserver bool
	Vfs       bool
	LogLevel  log.Level

	// [webserver]
	WebserverPort int
	AppKey        string

	// [digitalout]
	AllowedDigitalOutUser []string

	// [etemperature]
	AllowedEtemperatureUser       []string
	EtemperatureI2CAddress        uint16
	EtemperatureSamplerate        int
	EtemperatureSharedFileEnabled bool
	EtemperatureSharedDir         string
	EtemperatureSharedFile        string

	// s := strings.Split("a,b,c", ",")
}

var mcfg *ini.File
var merr error

func (p *Moduleconfig) ReadParameterFromFile() {

	mcfg, merr = ini.Load("/etc/smartpi_modules")
	if merr != nil {
		panic(merr)
	}

	// [base]
	p.I2CDevice = mcfg.Section("base").Key("i2c_device").MustString("/dev/i2c-1")
	p.Webserver = mcfg.Section("base").Key("webserver").MustBool(true)
	p.Vfs = mcfg.Section("base").Key("vfs").MustBool(true)
	p.LogLevel, merr = log.ParseLevel(mcfg.Section("base").Key("loglevel").MustString("info"))
	if merr != nil {
		panic(merr)
	}

	// [webserver]
	p.WebserverPort = mcfg.Section("webserver").Key("port").MustInt(1080)
	p.AppKey = mcfg.Section("webserver").Key("appkey").MustString("ew980723j35h546ergr97fqw4!234490#t33465")

	// [digitalout]
	p.AllowedDigitalOutUser = strings.Split(mcfg.Section("digitalout").Key("allowed_user").String(), ",")

	//[etemperature]
	p.AllowedEtemperatureUser = strings.Split(mcfg.Section("etemperature").Key("allowed_user").String(), ",")
	if p.EtemperatureI2CAddress, err = utils.DecodeUint16(mcfg.Section("etemperature").Key("i2c_address").MustString("0x52")); err != nil {
		log.Fatal(err)
	}
	p.EtemperatureSamplerate = mcfg.Section("etemperature").Key("samplerate").MustInt(6)
	p.EtemperatureSharedFileEnabled = mcfg.Section("etemperature").Key("shared_file_enabled").MustBool(true)
	p.EtemperatureSharedDir = mcfg.Section("etemperature").Key("shared_dir").MustString("/var/run")
	p.EtemperatureSharedFile = mcfg.Section("etemperature").Key("shared_file").MustString("smartpi_etemperature_values")

}

func (p *Moduleconfig) SaveParameterToFile() {

	// [base]
	_, merr = mcfg.Section("base").NewKey("i2c_device", p.I2CDevice)
	_, merr = mcfg.Section("base").NewKey("webserver", strconv.FormatBool(p.Webserver))
	_, merr = mcfg.Section("base").NewKey("vfs", strconv.FormatBool(p.Vfs))
	_, merr = mcfg.Section("base").NewKey("loglevel", p.LogLevel.String())

	// [webserver]
	_, merr = mcfg.Section("webserver").NewKey("port", strconv.FormatInt(int64(p.WebserverPort), 10))
	_, merr = mcfg.Section("appkey").NewKey("appkey", p.AppKey)

	// [digitalout]
	_, merr = mcfg.Section("digitalout").NewKey("allowed_user", strings.Join(p.AllowedDigitalOutUser, ","))

	//[etemperature]
	_, merr = mcfg.Section("etemperature").NewKey("allowed_user", strings.Join(p.AllowedEtemperatureUser, ","))
	_, merr = mcfg.Section("etemperature").NewKey("i2c_address", utils.EncodeUint64(uint64(p.EtemperatureI2CAddress)))
	_, merr = mcfg.Section("etemperature").NewKey("samplerate", strconv.FormatInt(int64(p.EtemperatureSamplerate), 10))
	_, merr = mcfg.Section("etemperature").NewKey("shared_file_enabled", strconv.FormatBool(p.EtemperatureSharedFileEnabled))
	_, merr = mcfg.Section("etemperature").NewKey("shared_dir", p.EtemperatureSharedDir)
	_, merr = mcfg.Section("etemperature").NewKey("shared_file", p.EtemperatureSharedFile)

	tmpFile := "/tmp/smartpi_modules"
	merr := mcfg.SaveTo(tmpFile)
	if merr != nil {
		panic(merr)
	}

	srcFile, merr := os.Open(tmpFile)
	utils.Checklog(merr)
	defer srcFile.Close()

	destFile, merr := os.Create("/etc/smartpi_modules") // creates if file doesn't exist
	utils.Checklog(merr)
	defer destFile.Close()

	_, merr = io.Copy(destFile, srcFile)
	utils.Checklog(merr)

	merr = destFile.Sync()
	utils.Checklog(merr)

	defer os.Remove(tmpFile)
}

func NewModuleconfig() *Moduleconfig {

	t := new(Moduleconfig)
	t.ReadParameterFromFile()
	return t
}
