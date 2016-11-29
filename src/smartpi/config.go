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
  Serial          string
  Name            string
  Debuglevel      int
  Lat             float64
  Lng             float64
  Databasedir     string
  Databasefile    string
  I2cdevice       string
  Shareddir      string
  Sharedfile      string
  Powerfrequency  int
  Measurevoltage1  int
  Measurevoltage2  int
  Measurevoltage3  int
  Voltage1 float64
  Voltage2 float64
  Voltage3 float64
  FTPupload int
  FTPserver string
  FTPuser string
  FTPpass string
  FTPpath string
  Webserverport int
  Docroot string
  Currentdirection1 int
  Currentdirection2 int
  Currentdirection3 int
  CSVdecimalpoint string
  CSVtimeformat string
  
  // MQTT Settings
  MQTTenabled int
  MQTTbroker string
  MQTTbrokerport string
  MQTTuser string
  MQTTpass string
  MQTTtopic string
}



func (p *Config) ReadParameterFromFile() {

  cfg, err := ini.Load("/etc/smartpi")
  if err != nil {
      panic(err)
  }

  p.Serial = cfg.Section("base").Key("serial").String()
  p.Name = cfg.Section("base").Key("name").String()
  p.Debuglevel, _ = cfg.Section("base").Key("debuglevel").Int()
  p.Lat, _ = cfg.Section("location").Key("lat").Float64()
  p.Lng, _ = cfg.Section("location").Key("lng").Float64()
  p.Databasedir = cfg.Section("database").Key("dir").String()
  p.Databasefile = cfg.Section("database").Key("file").String()
  p.I2cdevice = cfg.Section("device").Key("i2c_device").String()
  p.Shareddir = cfg.Section("device").Key("shared_dir").String()
  p.Sharedfile = cfg.Section("device").Key("shared_file").String()
  p.Powerfrequency, _ = cfg.Section("device").Key("power_frequency").Int()
  p.Measurevoltage1, _ = cfg.Section("device").Key("measure_voltage_1").Int()
  p.Measurevoltage2, _ = cfg.Section("device").Key("measure_voltage_2").Int()
  p.Measurevoltage3, _ = cfg.Section("device").Key("measure_voltage_3").Int()
  p.Voltage1, _ = cfg.Section("device").Key("voltage_1").Float64()
  p.Voltage2, _ = cfg.Section("device").Key("voltage_2").Float64()
  p.Voltage3, _ = cfg.Section("device").Key("voltage_3").Float64()
  p.FTPupload, _ = cfg.Section("ftp").Key("ftp_upload").Int()
  p.FTPserver = cfg.Section("ftp").Key("ftp_server").String()
  p.FTPuser = cfg.Section("ftp").Key("ftp_user").String()
  p.FTPpass = cfg.Section("ftp").Key("ftp_pass").String()
  p.FTPpath = cfg.Section("ftp").Key("ftp_path").String()
  p.Webserverport, _ = cfg.Section("webserver").Key("port").Int()
  p.Docroot = cfg.Section("webserver").Key("docroot").String()
  p.Currentdirection1, _ = cfg.Section("device").Key("change_current_direction_1").Int()
  p.Currentdirection2, _ = cfg.Section("device").Key("change_current_direction_2").Int()
  p.Currentdirection3, _ = cfg.Section("device").Key("change_current_direction_3").Int()
  p.CSVdecimalpoint = cfg.Section("csv").Key("decimalpoint").String()
  p.CSVtimeformat = cfg.Section("csv").Key("timeformat").String()
  
  //MQTT
  p.MQTTenabled, _ 	= cfg.Section("mqtt").Key("mqtt_enabled").Int()
  p.MQTTbroker  	= cfg.Section("mqtt").Key("mqtt_broker_url").String()
  p.MQTTbrokerport	= cfg.Section("mqtt").Key("mqtt_broker_port").String()
  p.MQTTuser	 	= cfg.Section("mqtt").Key("mqtt_username").String()
  p.MQTTpass		= cfg.Section("mqtt").Key("mqtt_password").String()
  p.MQTTtopic		= cfg.Section("mqtt").Key("mqtt_topic").String()
}

func NewConfig() *Config {

  t := new(Config)
  t.ReadParameterFromFile()
  return t
}
