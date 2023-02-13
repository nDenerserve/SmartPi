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
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// WifiInfo represents meta data about a WIFI network
type WifiInfo struct {
	SSID     string `json:"ssid"`
	RSSI     int    `json:"rssi"`
	BSSID    string `json:"bssid"`
	Channel  int    `json:"channel"`
	Security bool   `json:"security"`
	Active   bool   `json:"active"`
}

func ScanWifi() ([]WifiInfo, error) {
	var wifilist []WifiInfo
	var wifissid = ""
	var wifibssid = ""
	var wifichannel = 0
	var wifisecurity = false
	var wifisignal = 0
	var wifiactive = false

	var activessid = ""

	out, err := exec.Command("/bin/sh", "-c", `sudo iwgetid wlan0 | sed -e "s#^.*ESSID:##" | tr -d '"'`).Output()
	if err != nil {
		return wifilist, err
	}
	activessid = string(out)

	out, err = exec.Command("/bin/sh", "-c", `sudo iwlist wlan0 scan | egrep "ESSID:|Address:|Quality=|Encryption key:|Channel:" | sed -e  "s#^.*Channel:##" -e "s#^.*ESSID:##" -e "s#^.*Encryption key:##" -e "s#^.*Address: ##" -e "s#^.*Signal level=##" -e "s/\"//" -e "s/\"//"`).Output()
	if err != nil {
		return wifilist, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	linenumber := 0
	for scanner.Scan() {
		linenumber++
		line := scanner.Text()
		switch linenumber {
		case 1:
			wifibssid = line
		case 2:
			re := regexp.MustCompile("-?[0-9]+")
			wifichannel, _ = strconv.Atoi(re.FindString(line))
		case 3:
			re := regexp.MustCompile("-?[0-9]+")
			wifisignal, _ = strconv.Atoi(re.FindString(line))
		case 4:
			wifisecurity, _ = parseBool(line)
		case 5:
			wifiactive = false
			wifissid = line
			if strings.Contains(activessid, wifissid) {
				wifiactive = true
			}
			wifilist = append(wifilist, WifiInfo{SSID: wifissid, BSSID: wifibssid, RSSI: wifisignal, Channel: wifichannel, Security: wifisecurity, Active: wifiactive})
			linenumber = 0
		}
	}
	return wifilist, nil
}

func ListNetworkConnections() ([]NetworkInfo, error) {

	var listOfActiveDevices []NetworkInfo

	networklist, err := LocalAddresses()
	if err != nil {
		log.Println(err)
		return networklist, err
	}

	for _, i := range networklist {
		if len(i.Addrs) > 1 {
			listOfActiveDevices = append(listOfActiveDevices, i)
		}
	}

	return listOfActiveDevices, nil
}

func AddWifi(ssid string, key string) error {

	text :=
		`
	network={
			ssid=\"` + ssid + `\"
			psk=\"` + key + `\"
			key_mgmt=WPA-PSK
	}
	`

	_, err := exec.Command("/bin/sh", "-c", `echo "`+text+`" | sudo tee --append /etc/wpa_supplicant/wpa_supplicant.conf > /dev/null`).Output()
	if err != nil {
		log.Println(err)
		return errors.New("activation faild")
	}

	if err = ReconfigureWifi(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func ReconfigureWifi() error {
	_, err := exec.Command("/bin/sh", "-c", `sudo wpa_cli -i wlan0 reconfigure`).Output()
	if err != nil {
		log.Println(err)
		return errors.New("activation faild")
	}
	return nil
}

func RemoveWifi(ssid string) error {
	out, err := exec.Command("/bin/sh", "-c", `sudo sed -n '1 !H;1 h;$ {x;s/[[:space:]]*network={\n[[:space:]]*ssid="`+ssid+`"[^}]*}//g;p;}' /etc/wpa_supplicant/wpa_supplicant.conf`).Output()
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = exec.Command("/bin/sh", "-c", `echo "`+string(out)+`" | sudo tee /etc/wpa_supplicant/wpa_supplicant.conf > /dev/null`).Output()
	if err != nil {
		log.Println(err)
		return errors.New("Remove faild")
	}

	if err = ReconfigureWifi(); err != nil {
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
