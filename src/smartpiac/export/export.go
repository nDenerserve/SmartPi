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
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	smartpiacDatabase "github.com/nDenerserve/SmartPi/smartpiac/database"
	smartpiacUtils "github.com/nDenerserve/SmartPi/smartpiac/utils"
	log "github.com/sirupsen/logrus"
)

func parseTime(formats []string, dt string) (time.Time, error) {
	loc := time.Time.Location(time.Now())
	for _, format := range formats {
		parsedTime, err := time.ParseInLocation(format, dt, loc)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", dt)
}

func writeCSV(csv string, path string) {

	fmt.Println(path)

	f, err := os.Create(path)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(csv)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()

}

func main() {

	var csv string
	var filepath = "./smartpi_csv.csv"
	var result *api.QueryTableResult

	smartpiConfig := config.NewSmartPiConfig()

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC850,
		time.RFC822Z,
		time.RFC822,
		time.Layout,
		time.RubyDate,
		time.UnixDate,
		time.ANSIC,
		time.StampNano,
		time.StampMicro,
		time.StampMilli,
		time.Stamp,
		time.Kitchen,
		time.DateTime,
	}

	starttimePtr := flag.String("start", "", "starttime")
	stoptimePtr := flag.String("stop", "", "stoptime")
	rangePtr := flag.Int("range", 1, "range")
	decimalpointPtr := flag.String("decimalpoint", ".", "decimalpoint")
	aggregatePtr := flag.String("aggregate", "", "aggregate")

	flag.Parse()

	stop := time.Now()
	start := stop.Add(time.Duration(*rangePtr*24*(-1)) * time.Hour)

	if *starttimePtr != "" {
		fmt.Println(*starttimePtr)
		// start, _ = parseTime(formats, *starttimePtr)
		start, _ = parseTime(formats, *starttimePtr)
	}
	if *stoptimePtr != "" {
		fmt.Println(*starttimePtr)
		stop, _ = parseTime(formats, *starttimePtr)
	}
	if *rangePtr != 24 {
		year, month, day := start.Date()
		start = time.Date(year, month, day, 0, 0, 0, 0, start.Location())
	}
	if *aggregatePtr != "" {
		// csv, _ = exportCSV(start, stop, smartpiConfig, *decimalpointPtr, *aggregatePtr)
		result, _ = smartpiacDatabase.ReadData(smartpiConfig, start, stop, *aggregatePtr)
	} else {
		// csv, _ = exportCSV(start, stop, smartpiConfig, *decimalpointPtr)
		result, _ = smartpiacDatabase.ReadData(smartpiConfig, start, stop)
	}

	csv, _ = smartpiacUtils.CreateLegacyCSV(result, *decimalpointPtr)

	writeCSV(csv, filepath)
}
