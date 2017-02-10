package main


import (
	  "smartpi"
    "time"
		"fmt"
)

func main() {


	location := time.Now().Location()

  start,_ := time.ParseInLocation("2006-01-02 15:04:05","2017-02-0 00:00:00",location)
	end,_ := time.ParseInLocation("2006-01-02 15:04:05","2017-02-08 14:00:00",location)

	csvfile := smartpi.CreateCSV(start,end)

	fmt.Println(csvfile)

}
