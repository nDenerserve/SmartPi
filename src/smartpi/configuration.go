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
	"bytes"
	"errors"
	"net"
	"os/exec"
	"text/template"
)

type tWIFI struct {
	SSID    string `json:"ssid"`
	Quality string `json:"quality"`
	Level   string `json:"level"`
}

type tWIFIList struct {
	Values []tWIFI `json:"wifi"`
}

type Udhcpdparams struct {
	Ipstart   string
	Ipend     string
	Dns       string
	Subnet    string
	Router    string
	Leasetime int
}

type Hostapdparams struct {
	Ssid          string
	WpaPassphrase string
}

func ScanWIFI() (string, error) {
	out, err := exec.Command("/bin/sh", "-c", "sudo iwlist wlan0 scan | egrep \"(ESSID|IEEE|Quality)\"").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

/*
Create /etc/udhcpd.conf
*/
func CreateUDHCPD(params Udhcpdparams) string {

	var buf bytes.Buffer

	const templateUDHCPD = `
		# Adressbereich
		start {{.Ipstart}}
		end {{.Ipend}}
		# Interface
		interface wlan0
		remaining yes
		#DNS und subnet
		opt dns {{.Dns}}
		opt subnet {{.Subnet}}
		# Adresse Router
		opt router {{.Router}}
		# Leasetime in Sekunden (7 Tage)
		opt lease {{.Leasetime}}
    `
	t, _ := template.New("template_name").Parse(templateUDHCPD)
	t.Execute(&buf, params)

	return buf.String()
}

/*
Create /etc/hostapd/hostapd.conf
*/
func CreateHOSTAPD(params Hostapdparams) string {

	var buf bytes.Buffer

	const templateHOSTAPD = `
		# This is the name of the WiFi interface we configured above
		interface=wlan0
		# Use the nl80211 driver with the brcmfmac driver
		driver=nl80211
		# This is the name of the network
		ssid={{.Ssid}}
		# Use the 2.4GHz band
		hw_mode=g
		# Use channel 6
		channel=6
		# Enable 802.11n
		ieee80211n=1
		# Enable WMM
		wmm_enabled=1
		# Enable 40MHz channels with 20ns guard interval
		ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]
		# Accept all MAC addresses
		macaddr_acl=0
		# Use WPA authentication
		auth_algs=1
		# Require clients to know the network name
		ignore_broadcast_ssid=0
		# Use WPA2
		wpa=2
		# Use a pre-shared key
		wpa_key_mgmt=WPA-PSK
		# The network passphrase
		wpa_passphrase={{.WpaPassphrase}}
		# Use AES, instead of TKIP
		rsn_pairwise=CCMP
    `
	t, _ := template.New("template_name").Parse(templateHOSTAPD)
	t.Execute(&buf, params)

	return buf.String()
}

func EnableDHCPDinUDHCPD() error {
	_, err := exec.Command("/bin/sh", "-c", "sudo sed -i '/#DHCPD_ENABLED/!s/DHCPD_ENABLED/#DHCPD_ENABLED/' /etc/default/udhcpd").Output()
	if err != nil {
		return err
	}
	return nil
}

func SetWlanAdapterIp(ipaddress string) error {
	_, err := exec.Command("/bin/sh", "-c", "sudo ifconfig wlan0 "+ipaddress).Output()
	if err != nil {
		return err
	}
	return nil
}
