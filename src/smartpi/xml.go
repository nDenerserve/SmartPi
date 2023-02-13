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
	"encoding/xml"
	"math"
	"reflect"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
)

const (
	// A generic XML header suitable for use with the output of Marshal.
	// This is not automatically added to any output of this package,
	// it is provided as a convenience.
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

type tXmlValue struct {
	XMLName      xml.Name `xml:"valueset"`
	Date         string   `json:"date" xml:"date"`
	Current_1    float32  `json:"current_1" xml:"current_1"`
	Current_2    float32  `json:"current_2" xml:"current_2"`
	Current_3    float32  `json:"current_3" xml:"current_3"`
	Current_4    float32  `json:"current_4" xml:"current_4"`
	Voltage_1    float32  `json:"voltage_1" xml:"voltage_1"`
	Voltage_2    float32  `json:"voltage_2" xml:"voltage_2"`
	Voltage_3    float32  `json:"voltage_3" xml:"voltage_3"`
	Power_1      float32  `json:"power_1" xml:"power_1"`
	Power_2      float32  `json:"power_2" xml:"power_2"`
	Power_3      float32  `json:"power_3" xml:"power_3"`
	Cosphi_1     float32  `json:"cosphi_1" xml:"cosphi_1"`
	Cosphi_2     float32  `json:"cosphi_2" xml:"cosphi_2"`
	Cosphi_3     float32  `json:"cosphi_3" xml:"cosphi_3"`
	Frequency_1  float32  `json:"frequency_1" xml:"frequency_1"`
	Frequency_2  float32  `json:"frequency_2" xml:"frequency_2"`
	Frequency_3  float32  `json:"frequency_3" xml:"frequency_3"`
	Energy_pos_1 float32  `json:"energyPos_1" xml:"energyPos_1"`
	Energy_pos_2 float32  `json:"energyPos_2" xml:"energyPos_2"`
	Energy_pos_3 float32  `json:"energyPos_3" xml:"energyPos_3"`
	Energy_neg_1 float32  `json:"energyNeg_1" xml:"energyNeg_1"`
	Energy_neg_2 float32  `json:"energyNeg_2" xml:"energyNeg_2"`
	Energy_neg_3 float32  `json:"energyNeg_3" xml:"energyNeg_3"`
}

func CreateXML(start time.Time, end time.Time) string {

	var valueSeries []tXmlValue

	type valuelist []tXmlValue

	type dataset struct {
		valuelist
	}

	// type serie []tChartSerie

	tempValues := make([]float32, 22)
	config := config.NewConfig()
	data := ReadChartData(config.DatabaseDir, start, end)

	for _, dataelement := range data {
		ti := dataelement.Date

		datav := reflect.ValueOf(dataelement).Elem()

		for i := 1; i < datav.NumField(); i++ {
			val := datav.Field(i).Interface().(float64)
			if math.IsNaN(val) {
				val = 0.0
			}
			tempValues[i-1] = float32(val)
		}

		valueSeries = append(valueSeries, tXmlValue{Date: ti.Format(config.CSVtimeformat), Current_1: tempValues[0], Current_2: tempValues[1], Current_3: tempValues[2], Current_4: tempValues[3], Voltage_1: tempValues[4], Voltage_2: tempValues[5], Voltage_3: tempValues[6], Power_1: tempValues[7], Power_2: tempValues[8], Power_3: tempValues[9], Cosphi_1: tempValues[10], Cosphi_2: tempValues[11], Cosphi_3: tempValues[12], Frequency_1: tempValues[13], Frequency_2: tempValues[14], Frequency_3: tempValues[15], Energy_pos_1: tempValues[16], Energy_pos_2: tempValues[17], Energy_pos_3: tempValues[18], Energy_neg_1: tempValues[19], Energy_neg_2: tempValues[20], Energy_neg_3: tempValues[21]})

	}

	if xmlstring, err := xml.MarshalIndent(dataset{valueSeries}, "", "    "); err == nil {
		xmlstring = []byte(xml.Header + string(xmlstring))
		return string(xmlstring)
	} else {
		return ""
	}

}
