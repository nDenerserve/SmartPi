package main

import (
	"fmt"

	"github.com/Nitroman605/SmartPi/src/smartpi"
)

func main() {

	config := smartpi.NewConfig()
	fmt.Println(config.FTPpath)
	config.FTPpath = "smartpi/"
	fmt.Println(config.FTPpath)
	config.SaveParameterToFile()
}
