package main

import (
	"fmt"
	// "os/exec"
	"github.com/nDenerserve/SmartPi/src/smartpi/network"
)

func main() {

	// wifiNetworks := []smartpi.Wpa{}

	// n := smartpi.Wpa{"dev_zg_wlan", "Thorsten04041975Krakau"}

	// wifiNetworks = append (wifiNetworks,n)
	// smartpi.AddWPASupplicant(wifiNetworks)
//  smartpi.DeleteWPASupplicant("dev_zg_wlan")
	// fmt.Println(smartpi.ScanWIFI())
	// smartpi.RemoveWifi("dev_zg_wlan")
	network.AddWifi("dev_zg_wlan","dev_zg_wlan","Thorsten04041975Krakau")
	fmt.Println(network.ListNetworkConnections())
	// fmt.Println(smartpi.ReadWPASupplicant())
}
