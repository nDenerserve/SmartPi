package network

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type NetworkInfo struct {
	Name  string   `json:"name"`
	Flags string   `json:"flags"`
	Addrs []string `json:"addrs"`
}

func LocalAddresses() ([]NetworkInfo, error) {

	var networklist []NetworkInfo
	var addrlist []string

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
			return nil, err
		}
		addrlist = nil
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				addrlist = append(addrlist, ipnet.IP.String())
			}
		}
		networklist = append(networklist, NetworkInfo{Name: i.Name, Flags: i.Flags.String(), Addrs: addrlist})
	}
	return networklist, nil
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// getMacAddr gets the MAC hardware
// address of the host machine
func GetMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}
