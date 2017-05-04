package main

import (
	"fmt"
	"os/exec"
)

func main() {

	// var param smartpi.Udhcpdparams
	//
	// param.Ipstart = "192.168.7.2"
	// param.Ipend = "192.168.7.254"
	// param.Dns = "8.8.8.8"
	// param.Subnet = "255.255.255.0"
	// param.Router = "192.168.7.1"
	// param.Leasetime = 604800
	//
	// fmt.Println(smartpi.CreateUDHCPD(param))

	var v = "192.168.0.1"
	out, err := exec.Command("/bin/sh", "-c", "echo \"Hallo"+v+"\" > test.txt").Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
}
