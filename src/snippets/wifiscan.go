package main

import (
	"fmt"

	"pifke.org/wpasupplicant"
)

func main() {

	// Prints the BSSID (MAC address) and SSID of each access point in range:
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		panic(err)
	}

	results, _ := w.ScanResults()
	fmt.Println(results)
	// for bss,_ := range w.ScanResults() {
	// 	fmt.Fprintf("%s\t%s\n", bss.BSSID(), bss.SSID())
	// }
}
