package config

import (
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type Moduleconfig struct {
	// [base]
	I2CDevice string
	Vfs       bool
	LogLevel  log.Level

	// [digitalout]
	AllowedDigitalOutUser []string

	// [analogout420ma]
	AllowedAnalogOut420mAUser []string
	// AnalogOut420mACalibration holds per-board calibrations (offset in
	// mA), keyed by I2C address, from the calibration_<address> keys.
	AnalogOut420mACalibration map[uint16]AnalogOutCalibration

	// [analogout010v]
	AllowedAnalogOut010VUser []string
	// AnalogOut010VCalibration holds per-board output calibrations, keyed
	// by I2C address, from the calibration_<address> keys - see
	// AnalogOutCalibration. A module without an entry is used uncalibrated.
	AnalogOut010VCalibration map[uint16][4]AnalogOutCalibration

	// [analogin]
	AllowedAnalogInUser []string

	// [etemperature]
	AllowedEtemperatureUser       []string
	EtemperatureI2CAddress        uint16
	EtemperatureSamplerate        int
	EtemperatureSharedFileEnabled bool
	EtemperatureSharedDir         string
	EtemperatureSharedFile        string

	// [lorawan]
	LoRaWANEnabled             bool
	LoRaWANSharedDirs          []string
	LoRaWANSharedFilesElements []string
	LoRaWANSerialPort          string
	LoRaWANSendInterval        int
	LoRaWANApplicationEUI      string
	LoRaWANApplicationKey      string
	LoRaWANDataRate            int

	// s := strings.Split("a,b,c", ",")
}

// defaultAllowedUsers is the allowlist a module gets when its
// allowed_user key is missing or empty.
var defaultAllowedUsers = []string{"root", "smartpi"}

// splitList splits a comma-separated config value into its trimmed,
// non-empty entries, falling back to a copy of defaults when none remain.
//
// strings.Split alone is not enough: for a missing/empty key it returns
// [""], never an empty slice, so a "len == 0 -> defaults" check after it
// never fires - an allowlist then silently ends up as [""] and locks every
// user out.
func splitList(raw string, defaults []string) []string {
	var list []string
	for _, e := range strings.Split(raw, ",") {
		if e = strings.TrimSpace(e); e != "" {
			list = append(list, e)
		}
	}
	if len(list) == 0 {
		return append([]string(nil), defaults...)
	}
	return list
}

var mcfg *ini.File
var merr error

func (p *Moduleconfig) ReadParameterFromFile() {

	mcfg, merr = ini.LooseLoad("/etc/smartpiModules")
	if merr != nil {
		panic(merr)
	}

	// [base]
	p.I2CDevice = mcfg.Section("base").Key("i2c_device").MustString("/dev/i2c-1")
	p.Vfs = mcfg.Section("base").Key("vfs").MustBool(true)
	p.LogLevel, merr = log.ParseLevel(mcfg.Section("base").Key("loglevel").MustString("info"))
	if merr != nil {
		panic(merr)
	}

	// [digitalout]
	p.AllowedDigitalOutUser = splitList(mcfg.Section("digitalout").Key("allowed_user").String(), defaultAllowedUsers)

	// [analogout420ma]
	p.AllowedAnalogOut420mAUser = splitList(mcfg.Section("analogout420ma").Key("allowed_user").String(), defaultAllowedUsers)
	p.AnalogOut420mACalibration = readAnalogOut420mACalibrations(mcfg.Section("analogout420ma"))

	// [analogout010v]
	p.AllowedAnalogOut010VUser = splitList(mcfg.Section("analogout010v").Key("allowed_user").String(), defaultAllowedUsers)
	p.AnalogOut010VCalibration = readAnalogOut010VCalibrations(mcfg.Section("analogout010v"))

	// [analogin]
	p.AllowedAnalogInUser = splitList(mcfg.Section("analogin").Key("allowed_user").String(), defaultAllowedUsers)

	// [etemperature]
	p.AllowedEtemperatureUser = splitList(mcfg.Section("etemperature").Key("allowed_user").String(), defaultAllowedUsers)
	if p.EtemperatureI2CAddress, err = utils.DecodeUint16(mcfg.Section("etemperature").Key("i2c_address").MustString("0x52")); err != nil {
		log.Fatal(err)
	}
	p.EtemperatureSamplerate = mcfg.Section("etemperature").Key("samplerate").MustInt(6)
	p.EtemperatureSharedFileEnabled = mcfg.Section("etemperature").Key("shared_file_enabled").MustBool(true)
	p.EtemperatureSharedDir = mcfg.Section("etemperature").Key("shared_dir").MustString("/var/run/smartpi")
	p.EtemperatureSharedFile = mcfg.Section("etemperature").Key("shared_file").MustString("smartpi_etemperature_values")

	// [lorawan]
	p.LoRaWANEnabled = mcfg.Section("lorawan").Key("lorawan_enabled").MustBool(true)
	p.LoRaWANSharedDirs = splitList(mcfg.Section("lorawan").Key("shared_files_path").String(),
		[]string{"/var/run/smartpi/smartpi_values"})
	p.LoRaWANSharedFilesElements = splitList(mcfg.Section("lorawan").Key("shared_files_elements").String(),
		[]string{"1:2f:1.0|2:2f:1.0|3:2f:1.0|4:2f:1.0|5:2f:1.0|6:2f:1.0|7:2f:1.0", "1:2f:1.0|2:2f:1.0|3:2f:1.0|4:2f:1.0|5:2f:1.0|6:2f:1.0|7:2f:1.0"})
	p.LoRaWANSendInterval = mcfg.Section("lorawan").Key("interval").MustInt(60)
	p.LoRaWANSerialPort = mcfg.Section("lorawan").Key("serial_port").MustString("/dev/ttyS0")
	p.LoRaWANApplicationEUI = mcfg.Section("lorawan").Key("applicationEUI").MustString("")
	p.LoRaWANApplicationKey = mcfg.Section("lorawan").Key("applicationKey").MustString("")
	p.LoRaWANDataRate = mcfg.Section("lorawan").Key("datarate").MustInt(5)

}

func (p *Moduleconfig) SaveParameterToFile() {

	// [base]
	_, merr = mcfg.Section("base").NewKey("i2c_device", p.I2CDevice)
	_, merr = mcfg.Section("base").NewKey("vfs", strconv.FormatBool(p.Vfs))
	_, merr = mcfg.Section("base").NewKey("loglevel", p.LogLevel.String())

	// [digitalout]
	_, merr = mcfg.Section("digitalout").NewKey("allowed_user", strings.Join(p.AllowedDigitalOutUser, ","))

	// [analogout420ma]
	_, merr = mcfg.Section("analogout420ma").NewKey("allowed_user", strings.Join(p.AllowedAnalogOut420mAUser, ","))
	for addr, cal := range p.AnalogOut420mACalibration {
		_, merr = mcfg.Section("analogout420ma").NewKey(analogOut420mACalibrationKey(addr), formatCalibration(cal))
	}

	// [analogout010v]
	_, merr = mcfg.Section("analogout010v").NewKey("allowed_user", strings.Join(p.AllowedAnalogOut010VUser, ","))
	for addr, cal := range p.AnalogOut010VCalibration {
		_, merr = mcfg.Section("analogout010v").NewKey(analogOut010VCalibrationKey(addr), FormatAnalogOutCalibration(cal))
	}

	// [analogin]
	_, merr = mcfg.Section("analogin").NewKey("allowed_user", strings.Join(p.AllowedAnalogInUser, ","))

	//[etemperature]
	_, merr = mcfg.Section("etemperature").NewKey("allowed_user", strings.Join(p.AllowedEtemperatureUser, ","))
	_, merr = mcfg.Section("etemperature").NewKey("i2c_address", utils.EncodeUint64(uint64(p.EtemperatureI2CAddress)))
	_, merr = mcfg.Section("etemperature").NewKey("samplerate", strconv.FormatInt(int64(p.EtemperatureSamplerate), 10))
	_, merr = mcfg.Section("etemperature").NewKey("shared_file_enabled", strconv.FormatBool(p.EtemperatureSharedFileEnabled))
	_, merr = mcfg.Section("etemperature").NewKey("shared_dir", p.EtemperatureSharedDir)
	_, merr = mcfg.Section("etemperature").NewKey("shared_file", p.EtemperatureSharedFile)

	//[lorawan]
	_, merr = mcfg.Section("lorawan").NewKey("lorawan_enabled", strconv.FormatBool(p.LoRaWANEnabled))
	_, merr = mcfg.Section("lorawan").NewKey("shared_files_path", strings.Join(p.LoRaWANSharedDirs[:], ","))
	_, merr = mcfg.Section("lorawan").NewKey("shared_files_elements", strings.Join(p.LoRaWANSharedFilesElements[:], ","))
	_, merr = mcfg.Section("lorawan").NewKey("interval", strconv.FormatInt(int64(p.LoRaWANSendInterval), 10))
	_, merr = mcfg.Section("lorawan").NewKey("serial_port", p.LoRaWANSerialPort)
	_, merr = mcfg.Section("lorawan").NewKey("applicationEUI", p.LoRaWANApplicationEUI)
	_, merr = mcfg.Section("lorawan").NewKey("applicationKey", p.LoRaWANApplicationKey)
	_, merr = mcfg.Section("lorawan").NewKey("datarate", strconv.FormatInt(int64(p.LoRaWANDataRate), 10))

	if merr := writeConfigFile(mcfg, "/etc/smartpiModules"); merr != nil {
		log.Errorf("saving /etc/smartpiModules: %v", merr)
	}
}

func NewModuleconfig() *Moduleconfig {

	t := new(Moduleconfig)
	t.ReadParameterFromFile()
	return t
}
