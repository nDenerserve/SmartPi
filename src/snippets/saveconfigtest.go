package main

import (
	"fmt"
  "github.com/nDenerserve/SmartPi/src/smartpi"
)

func main() {

	config := smartpi.NewConfig()
  fmt.Println(config.FTPpath)
  config.FTPpath = "smartpi/"
  fmt.Println(config.FTPpath)
  config.SaveParameterToFile()
}
