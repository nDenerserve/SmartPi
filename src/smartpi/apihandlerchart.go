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
/*
File: apihandlerchart.go
Description: Handels API requests for charts
*/



package smartpi

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "time"
    "github.com/ziutek/rrd"
    "log"
    "strconv"
    "strings"
    "fmt"
    "math"
)




func ServeChartValues(w http.ResponseWriter, r *http.Request) {

  type tChartValue struct {
    Time string `json:"time"`
    Value float32 `json:"value"`
  }

  type tChartSerie struct {
    Key string `json:"key"`
    Values []tChartValue `json:"values"`
    }

  // type tChartSeries []tChartSerie

	var timeSeries []tChartSerie

  vars := mux.Vars(r)
  from := vars["fromDate"]
  to := vars["toDate"]
  phaseId := vars["phaseId"]
  valueId := vars["valueId"]


  config := NewConfig()
  dbfile := config.Databasedir+"/"+config.Databasefile

  location := time.Now().Location()

  end, err := time.ParseInLocation("2006-01-02 15:04:05",to,location)
  if err != nil {
    log.Fatal(err)
  }
  end = end.UTC()
	start, err := time.ParseInLocation("2006-01-02 15:04:05",from,location)
  if err != nil {
    log.Fatal(err)
  }
  start = start.UTC()

  if end.Before(start) {
    start = start.AddDate(0,0,-1)
  }

      e := rrd.NewExporter()

      for i:=1; i<=3; i++ {

        if strings.Contains(phaseId,strconv.Itoa(i)) {
          e.Def("def"+strconv.Itoa(i), dbfile, valueId+"_"+strconv.Itoa(i), "AVERAGE")
          e.XportDef("def"+strconv.Itoa(i), valueId+"_"+strconv.Itoa(i))
        }

      }

      xportRes, err := e.Xport(start, end, STEP*time.Second)
    	if err != nil {
        if err := json.NewEncoder(w).Encode("error"); err != nil {
          log.Fatal(err)
        }
    	}
      defer xportRes.FreeValues()



      for i := 0; i < len(xportRes.Legends); i++ {
      row := 0
        var values []tChartValue
        for ti := xportRes.Start.Add(xportRes.Step); ti.Before(end) || ti.Equal(end); ti = ti.Add(xportRes.Step) {

          val := xportRes.ValueAt(i, row)
          if math.IsNaN(val) {
            val = 0.0
          }
          values = append(values, tChartValue{Time: fmt.Sprintf("%d",ti.Unix()), Value: float32( val )})
          row++
        }


        timeSeries = append(timeSeries, tChartSerie{Key: xportRes.Legends[i], Values: values})
      }


/*
      row := 0
    	for ti := xportRes.Start.Add(xportRes.Step); ti.Before(end) || ti.Equal(end); ti = ti.Add(xportRes.Step) {
    		fmt.Printf("%s / %d", ti, ti.Unix())
    		for i := 0; i < len(xportRes.Legends); i++ {
    			v := xportRes.ValueAt(i, row)
    			fmt.Printf("\t%e", v)
    		}
    		fmt.Printf("\n")
    		row++
    	}


  for j := 1; j < 3; j++ {
		var values []tChartValue

		for i := 1; i < 4; i++ {
			values = append(values, tChartValue{Time: "2016090815" + strconv.Itoa(15+i), Value: float32(i * j)})
		}

		timeSeries = append(timeSeries, values)
	}



*/

  // JSON output of request
  if err := json.NewEncoder(w).Encode(timeSeries); err != nil {
     panic(err)
  }

}
