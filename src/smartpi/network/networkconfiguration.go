/*
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
package network

import (
	"errors"
	// log "github.com/Sirupsen/logrus"
	"bufio"
	"fmt"
	"log"
	// "os"
	"os/exec"
	// "regexp"
	"strconv"
	"strings"
)


// WifiInfo represents meta data about a WIFI network
type WifiInfo struct {
	SSID     string `json:"ssid"`
	RSSI     int    `json:"rssi"`
	Channel  int    `json:"channel"`
	Security string `json:"security"`
	Mode	 string `json:"mode"`
	Bars	 string `json:"bars"`
	Active   bool	`json:"active"`
}

type NetworkInfo struct {
	Name string`json:"name"`
	Wireless bool `json:"wireless"`
	Active bool `json:"active"`
}

func ScanWifi() ([]WifiInfo, error) {
	var wifilist []WifiInfo
	var wifissid = ""
	var wifichannel = 0
	var wifisecurity = ""
	var wifisignal = 0
	var wifibars = ""
	var wifimode = ""
	var active = false

	out, err := exec.Command("/bin/sh", "-c", `nmcli -t -f in-use,ssid,mode,chan,signal,bars,security dev wifi`).Output()
	if err != nil {
		return wifilist, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		for i := range parts {
			switch i {
				case 0:
					if parts[i]=="*" {
						active = true
					} else {
						active = false
					}
				case 1:
					wifissid = parts[i]
				case 2:
					wifimode = parts[i]
				case 3:
					wifichannel, _ = strconv.Atoi(parts[i])
				case 4:
					wifisignal, _ = strconv.Atoi(parts[i])
				case 5:
					wifibars = parts[i]
				case 6:
					wifisecurity = parts[i]
			}
		}
		wifilist = append(wifilist, WifiInfo{SSID: wifissid, Mode: wifimode, Channel: wifichannel, RSSI: wifisignal, Bars: wifibars, Security: wifisecurity, Active: active})
       
	}
	return wifilist, nil
}



func ListNetworkConnections() ([]NetworkInfo, error) {

	var networklist []NetworkInfo
	var networkname = ""
	var networkwireless = false
	var networkactive = false

	out, err := exec.Command("/bin/sh", "-c", `sudo nmcli -t -f name,type,device connection show`).Output()	
	if err != nil {
		log.Println(err)
		return networklist, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		for i := range parts {
			switch i {
				case 0:
					networkname = parts[i]
				case 1:
					if strings.Contains(parts[i], "wireless") {
						networkwireless = true
					} else {
						networkwireless = false
					}
				case 2:
					if strings.Contains(parts[i], "eth") || strings.Contains(parts[i], "wlan") {
						networkactive = true
					} else {
						networkactive = false
					}
			}
		}
		networklist = append(networklist, NetworkInfo{Name: networkname, Wireless: networkwireless, Active: networkactive})
       
	}
	return networklist, nil
}

func AddWifi(ssid string, name string, key string) error {

	_, err := exec.Command("/bin/sh", "-c", `sudo nmcli con add con-name `+name+` ifname wlan0 type wifi ssid `+ssid).Output()	
	if err != nil {
		log.Println(err)
		return err
	}
	err = ChangeWifiSecurity(ssid,name,key,"wpa-psk")
	if err != nil {
		log.Println(err)
		return err
	}
	err = ActivateWifi(name)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func ChangeWifiSecurity(ssid string, name string, newkey string, keymgmt string) error {
	_, err := exec.Command("/bin/sh", "-c", `sudo nmcli con modify `+name+` 802-11-wireless-security.key-mgmt `+keymgmt).Output()	
	if err != nil {
		log.Println(err)
		return err
	}
	
	_, err = exec.Command("/bin/sh", "-c", `sudo nmcli con modify `+name+` 802-11-wireless-security.psk `+newkey).Output()	
	if err != nil {
		log.Println(err)
	}
	// sudo nmcli con modify dev_zg_wlan 802-11-wireless-security.key-mgmt wpa-psk
	// sudo nmcli con modify dev_zg_wlan 802-11-wireless-security.psk pskkey
	return nil
}

func ActivateWifi(name string) error {
	_, err := exec.Command("/bin/sh", "-c", `sudo nmcli -p con up '`+name+`' ifname wlan0`).Output()	
	if err != nil {
		log.Println(err)
		return errors.New("activation faild")
	}
	return nil
}

func DeactivateWifi(name string) error {
	_, err := exec.Command("/bin/sh", "-c", `sudo nmcli -p con down '`+name+`'`).Output()	
	if err != nil {
		log.Println(err)
		return errors.New("activation faild")
	}
	return nil
}

func RemoveWifi(name string) error {
	_, err := exec.Command("/bin/sh", "-c", `sudo nmcli connection delete id `+name).Output()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
} 




func parseBool(str string) (value bool, err error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "y", "ON", "on", "On":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "n", "OFF", "off", "Off":
		return false, nil
	}
	return false, fmt.Errorf("parsing \"%s\": invalid syntax", str)
}

func bool2string(bl bool) (str string) {
	switch bl {
	case true:
		return "yes"
	case false:
		return "no"
	}
	return
}

func parseOption(option string) (opt, value string) {
	split := func(i int, delim string) (opt, value string) {
		// strings.Split cannot handle wsrep_provider_options settings
		opt = strings.Trim(option[:i], " ")
		value = strings.Trim(option[i+1:], " ")
		return
	}

	if i := strings.Index(option, "="); i != -1 {
		opt, value = split(i, "=")
	} else if i := strings.Index(option, ":"); i != -1 {
		opt, value = split(i, ":")
	} else {
		opt = option
	}
	return
}
