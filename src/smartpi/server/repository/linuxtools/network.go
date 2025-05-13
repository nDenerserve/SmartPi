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
package linuxtoolsRepository

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

// addConnection adds a new network connection (Ethernet or WiFi)
func (l LinuxToolsRepository) AddConnection(interfaceName, connectionName, connectionType, ssid, password string) error {
	log.Debugf("AddConnection: %s, %s, %s, %s, %s", interfaceName, connectionName, connectionType, ssid, password)
	var cmdArgs []string

	if connectionType == "ethernet" {
		cmdArgs = []string{"connection", "add", "type", "ethernet", "ifname", interfaceName, "con-name", connectionName}
	} else if connectionType == "wifi" {
		cmdArgs = []string{"connection", "add", "type", "wifi", "ifname", interfaceName, "con-name", connectionName, "ssid", ssid}
		if password != "" {
			cmdArgs = append(cmdArgs, "wifi-sec.key-mgmt", "wpa-psk", "wifi-sec.psk", password)
		}
	} else {
		return fmt.Errorf("unsupported connection type: %s", connectionType)
	}

	_, err := utils.RunCommand("nmcli", cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to add connection %s: %v", connectionName, err)
	}
	return nil
}

// configureNetwork sets up the network interface with the specified IP address and netmask
func (l LinuxToolsRepository) ConfigureNetwork(interfaceName, ipAddress, netmask string) error {
	log.Debugf("ConfigureNetwork: %s, %s, %s", interfaceName, ipAddress, netmask)
	// Bring the interface up
	if err := exec.Command("nmcli", "connection", "up", interfaceName).Run(); err != nil {
		return fmt.Errorf("failed to bring up interface %s: %v", interfaceName, err)
	}

	// Set the IP address and netmask
	if err := exec.Command("nmcli", "connection", "modify", interfaceName, "ipv4.addresses", fmt.Sprintf("%s/%s", ipAddress, netmask)).Run(); err != nil {
		return fmt.Errorf("failed to set IP address on interface %s: %v", interfaceName, err)
	}

	if err := exec.Command("nmcli", "connection", "modify", interfaceName, "ipv4.method", "manual").Run(); err != nil {
		return fmt.Errorf("failed to set IP address method on interface %s: %v", interfaceName, err)
	}

	return nil
}

// modifyConnection modifies the network connection with the given parameters
func (l LinuxToolsRepository) ModifyConnection(connectionName, ipAddress, netmask, gateway, dns string, useDHCP bool) error {
	log.Debugf("ListConnections: %s, IpAddress: %s, Netmask: %s, Gateway: %s, DNS; %s, UseDHCP: %s", connectionName, ipAddress, netmask, gateway, dns, useDHCP)
	if useDHCP {
		if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.method", "auto").Run(); err != nil {
			return fmt.Errorf("failed to set DHCP on connection %s: %v", connectionName, err)
		}
	} else {
		if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.addresses", fmt.Sprintf("%s/%s", ipAddress, netmask)).Run(); err != nil {
			return fmt.Errorf("failed to set IP address on connection %s: %v", connectionName, err)
		}

		if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.gateway", gateway).Run(); err != nil {
			return fmt.Errorf("failed to set gateway on connection %s: %v", connectionName, err)
		}

		if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.dns", dns).Run(); err != nil {
			return fmt.Errorf("failed to set DNS on connection %s: %v", connectionName, err)
		}

		if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.method", "manual").Run(); err != nil {
			return fmt.Errorf("failed to set IP address method on connection %s: %v", connectionName, err)
		}
	}

	return nil
}

// listConnections lists all network connections managed by NetworkManager
func (l LinuxToolsRepository) ListConnections() ([]models.NetworkConnection, error) {
	log.Debugf("ListConnections: ")
	out, err := utils.RunCommand("nmcli", "-t", "-f", "NAME,UUID,TYPE,DEVICE", "connection", "show")
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %v", err)
	}
	lines := strings.Split(out, "\n")
	var connections []models.NetworkConnection
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		connectionAddresses, state, err := l.getConnectionDetails(fields[0])
		if err != nil {
			return nil, err
		}

		connections = append(connections, models.NetworkConnection{
			Name:                fields[0],
			Uuid:                fields[1],
			Type:                fields[2],
			Device:              fields[3],
			ConnectionAddresses: connectionAddresses,
			State:               state,
		})
	}
	return connections, nil
}

// TODO: Subnetmask
// getConnectionDetails gets the IP address and state of a specific network connection
func (l LinuxToolsRepository) getConnectionDetails(connectionName string) ([]models.NetworkConnectionAddress, string, error) {
	log.Debugf("getConnectionDetails: %s", connectionName)
	ipMethodOutput, err := utils.RunCommand("nmcli", "-t", "-f", "ipv4.method", "connection", "show", connectionName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get IP method for connection %s: %v", connectionName, err)
	}
	ipMethod := strings.TrimPrefix(strings.TrimSpace(ipMethodOutput), "ipv4.method:")

	ipAddressOutput, err := utils.RunCommand("nmcli", "-t", "-f", "IP4.ADDRESS", "connection", "show", connectionName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get IP addresses for connection %s: %v", connectionName, err)
	}
	ipAddressOutput = strings.TrimSpace(ipAddressOutput)
	connectionAddresses := []models.NetworkConnectionAddress{}

	for _, line := range strings.Split(ipAddressOutput, "\n") {
		if line != "" {
			ip := strings.Split(line, ":")[1]
			ipWithoutSuffix := strings.Split(ip, "/")[0]
			cidrSuffix, err := strconv.Atoi(strings.Split(ip, "/")[1])
			if err != nil {
				return nil, "", fmt.Errorf("getConnectionDetails: Failed to convert CIDR-suffix %v", err)
			}
			connectionAddresses = append(connectionAddresses, models.NetworkConnectionAddress{Ipv4Address: ipWithoutSuffix, IpMethod: ipMethod, CidrSuffix: uint8(cidrSuffix)})
		}
	}

	stateOutput, err := utils.RunCommand("nmcli", "-t", "-f", "GENERAL.STATE", "connection", "show", connectionName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get state for connection %s: %v", connectionName, err)
	}
	stateOutput = strings.TrimSpace(stateOutput)
	state := strings.Split(stateOutput, ":")[1] // Remove the GENERAL.STATE prefix

	staticipv4Output, err := utils.RunCommand("nmcli", "-t", "-f", "ipv4.addresses", "connection", "show", connectionName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get static IPv4 for connection %s: %v", connectionName, err)
	}

	staticipv4Output = strings.TrimSpace(staticipv4Output)
	staticipv4Output = strings.Split(staticipv4Output, ":")[1]
	log.Debugf("getConnectionDetails: Length staticipv4Output: %s", len(staticipv4Output))
	if len(staticipv4Output) == 0 {
		return connectionAddresses, state, nil
	}
	staticIpv4Addresses := strings.Split(staticipv4Output, ",")
	staticIpv4Cidr := make([]string, len(staticIpv4Addresses))
	log.Debugf("getConnectionDetails: Length staticIpv4Addresses: %s, staticIpv4Addresses: %s, staticIpv4Cidr: %s", len(staticIpv4Addresses), staticIpv4Addresses, staticIpv4Cidr)

	for i := range staticIpv4Addresses {
		staticIpv4Addresses[i] = strings.TrimSpace(staticIpv4Addresses[i])
		staticIpv4Cidr[i] = strings.Split(staticIpv4Addresses[i], "/")[1]      // Get the cidr-suffix
		staticIpv4Addresses[i] = strings.Split(staticIpv4Addresses[i], "/")[0] // Remove the prefix

	}

	log.Debugf("getConnectionDetails: connectionAddresses: %s", connectionAddresses)
	log.Debugf("getConnectionDetails: trimmed staticIpv4Addresses: %s", staticIpv4Addresses)

	for i := range staticIpv4Addresses {
		for j := range connectionAddresses {
			if connectionAddresses[j].Ipv4Address == staticIpv4Addresses[i] {
				connectionAddresses[j].IpMethod = "manual"
				cidrsuffix, err := strconv.Atoi(staticIpv4Cidr[i])
				if err != nil {
					return nil, "", fmt.Errorf("getConnectionDetails: Failed to convert CIDR-suffix %v", err)
				}
				connectionAddresses[j].CidrSuffix = uint8(cidrsuffix)
			}
		}
	}

	log.Debugf("getConnectionDetails: ConnectionAddresses: %s", connectionAddresses)
	log.Debugf("getConnectionDetails: State: %s", state)
	return connectionAddresses, state, nil
}

// scanWifiNetworks scans for available WiFi networks and returns a list of WifiNetwork structs
func (l LinuxToolsRepository) ScanWifiNetworks() ([]models.WifiNetwork, error) {
	log.Debugf("ScanWifiNetworks..")
	out, err := utils.RunCommand("nmcli", "-t", "-f", "ssid,mode,chan,rate,signal,bars,security", "device", "wifi", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to scan WiFi networks: %v", err)
	}
	lines := strings.Split(out, "\n")
	var networks []models.WifiNetwork
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}
		networks = append(networks, models.WifiNetwork{
			SSID:     fields[0],
			Mode:     fields[1],
			Channel:  fields[2],
			Rate:     fields[3],
			Signal:   fields[4],
			Bars:     fields[5],
			Security: fields[6],
		})
	}
	log.Debugf("ScanWifiNetworks: %s", networks)
	return networks, nil
}

// changeToDHCP changes the network connection from static IP to DHCP
func (l LinuxToolsRepository) ChangeToDHCP(connectionName string) error {
	log.Debugf("ChangeToDHCP: %s %s", connectionName)
	if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.method", "auto").Run(); err != nil {
		return fmt.Errorf("failed to change connection %s to DHCP: %v", connectionName, err)
	}
	return nil
}

// removeConnection removes a network connection by name
func (l LinuxToolsRepository) RemoveConnection(connectionName string) error {
	log.Debugf("RemoveConnection: %s %s", connectionName)
	_, err := utils.RunCommand("nmcli", "connection", "delete", connectionName)
	if err != nil {
		return fmt.Errorf("failed to remove connection %s: %v", connectionName, err)
	}
	return nil
}

// restartConnection restarts a network connection by bringing it down and then up
func (l LinuxToolsRepository) RestartConnection(connectionName string) error {
	log.Debugf("RestartConnection: %s %s", connectionName)
	if err := exec.Command("nmcli", "connection", "down", connectionName).Run(); err != nil {
		return fmt.Errorf("failed to bring down connection %s: %v", connectionName, err)
	}
	if err := exec.Command("nmcli", "connection", "up", connectionName).Run(); err != nil {
		return fmt.Errorf("failed to bring up connection %s: %v", connectionName, err)
	}
	return nil
}

// changeToStatic changes the network connection from DHCP to static IP
func (l LinuxToolsRepository) ChangeToStatic(connectionName, ipAddress, netmask, gateway, dns string) error {
	if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.addresses", fmt.Sprintf("%s/%s", ipAddress, netmask)).Run(); err != nil {
		return fmt.Errorf("failed to set IP address on connection %s: %v", connectionName, err)
	}

	if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.gateway", gateway).Run(); err != nil {
		return fmt.Errorf("failed to set gateway on connection %s: %v", connectionName, err)
	}

	if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.dns", dns).Run(); err != nil {
		return fmt.Errorf("failed to set DNS on connection %s: %v", connectionName, err)
	}

	if err := exec.Command("nmcli", "connection", "modify", connectionName, "ipv4.method", "manual").Run(); err != nil {
		return fmt.Errorf("failed to set IP address method on connection %s: %v", connectionName, err)
	}

	return nil
}

// addIpAddressToConnection adds an IP address to an existing network connection
func (l LinuxToolsRepository) AddIpAddressToConnection(connectionName, ipAddress string, cidrsuffix uint8) error {
	log.Debugf("AddIpAddressToConnection: %s %s", connectionName, ipAddress+"/"+fmt.Sprint(cidrsuffix))
	_, err := utils.RunCommand("nmcli", "connection", "modify", connectionName, "+ipv4.addresses", ipAddress+"/"+fmt.Sprint(cidrsuffix))
	if err != nil {
		return fmt.Errorf("failed to add IP address %s to connection %s: %v", ipAddress, connectionName, err)
	}
	return nil
}

// removeIpAddressFromConnection removes an IP address from an existing network connection
func (l LinuxToolsRepository) RemoveIpAddressFromConnection(connectionName, ipAddress string, cidrsuffix uint8) error {
	log.Debugf("RemoveIpAddressFromConnection: %s %s", connectionName, ipAddress)
	_, err := utils.RunCommand("nmcli", "connection", "modify", connectionName, "-ipv4.addresses", ipAddress+"/"+fmt.Sprint(cidrsuffix))
	if err != nil {
		return fmt.Errorf("failed to remove IP address %s from connection %s: %v", ipAddress, connectionName, err)
	}
	return nil
}
