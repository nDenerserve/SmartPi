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
File: csv.go
Description: create csv-file
*/



package smartpi

import (
    "time"
    "github.com/ziutek/rrd"
    "log"
    "strconv"
    "strings"
    // "fmt"
    "math"
)


func CreateCSV(start time.Time, end time.Time) (string) {


  config := NewConfig()
  dbfile := config.Databasedir+"/"+config.Databasefile


  e := rrd.NewExporter()
  csv := "date;current_1;current_2;current_3;current_4;voltage_1;voltage_2;voltage_3;power_1;power_2;power_3;cosphi_1;cosphi_2;cosphi_3;energy_pos_1;energy_pos_2;energy_pos_3;energy_neg_1;energy_neg_2;energy_neg_3"
  csv = csv + "\n"


  e.Def("def1", dbfile, "current_1", "AVERAGE")
  e.Def("def2", dbfile, "current_2", "AVERAGE")
  e.Def("def3", dbfile, "current_3", "AVERAGE")
  e.Def("def4", dbfile, "current_4", "AVERAGE")
  e.Def("def5", dbfile, "voltage_1", "AVERAGE")
  e.Def("def6", dbfile, "voltage_2", "AVERAGE")
  e.Def("def7", dbfile, "voltage_3", "AVERAGE")
  e.Def("def8", dbfile, "power_1", "AVERAGE")
  e.Def("def9", dbfile, "power_2", "AVERAGE")
  e.Def("def10", dbfile, "power_3", "AVERAGE")
  e.Def("def11", dbfile, "cosphi_1", "AVERAGE")
  e.Def("def12", dbfile, "cosphi_2", "AVERAGE")
  e.Def("def13", dbfile, "cosphi_3", "AVERAGE")
  e.Def("def14", dbfile, "energy_pos_1", "AVERAGE")
  e.Def("def15", dbfile, "energy_pos_2", "AVERAGE")
  e.Def("def16", dbfile, "energy_pos_3", "AVERAGE")
  e.Def("def17", dbfile, "energy_neg_1", "AVERAGE")
  e.Def("def18", dbfile, "energy_neg_2", "AVERAGE")
  e.Def("def19", dbfile, "energy_neg_3", "AVERAGE")

  e.XportDef("def1", "current_1")
  e.XportDef("def2", "current_2")
  e.XportDef("def3", "current_3")
  e.XportDef("def4", "current_4")
  e.XportDef("def5", "voltage_1")
  e.XportDef("def6", "voltage_2")
  e.XportDef("def7", "voltage_3")
  e.XportDef("def8", "power_1")
  e.XportDef("def9", "power_2")
  e.XportDef("def10", "power_3")
  e.XportDef("def11", "cosphi_1")
  e.XportDef("def12", "cosphi_2")
  e.XportDef("def13", "cosphi_3")
  e.XportDef("def14", "energy_pos_1")
  e.XportDef("def15", "energy_pos_2")
  e.XportDef("def16", "energy_pos_3")
  e.XportDef("def17", "energy_neg_1")
  e.XportDef("def17", "energy_neg_2")
  e.XportDef("def17", "energy_neg_3")

  xportRes, err := e.Xport(start, end, STEP*time.Second)
	if err != nil {
		log.Fatal(err)
	}
  defer xportRes.FreeValues()

  row := 0
	for ti := xportRes.Start.Add(xportRes.Step); ti.Before(end) || ti.Equal(end); ti = ti.Add(xportRes.Step) {
		// fmt.Printf("%s / %d", ti, ti.Unix())
    csv = csv + ti.Format(config.CSVtimeformat)
		for i := 0; i < len(xportRes.Legends); i++ {
			val := xportRes.ValueAt(i, row)
      if math.IsNaN(val) {
        val = 0.0
      }
      csv = csv + ";"+strings.Replace(strconv.FormatFloat(val,'f',5, 64),".",config.CSVdecimalpoint,-1)
			// fmt.Printf(";%f", strings.Replace(strconv.FormatFloat(val,'f',5, 64),".",config.Decimalpoint,-1))
		}
		// fmt.Printf("\n")
    csv = csv + "\n"

		row++
	}
  // fmt.Println(csv)
  return csv

}
