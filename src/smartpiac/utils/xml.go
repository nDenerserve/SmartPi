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

package smartpiacUtils

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
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
	Current_1    float64  `json:"current_1" xml:"current_1"`
	Current_2    float64  `json:"current_2" xml:"current_2"`
	Current_3    float64  `json:"current_3" xml:"current_3"`
	Current_4    float64  `json:"current_4" xml:"current_4"`
	Voltage_1    float64  `json:"voltage_1" xml:"voltage_1"`
	Voltage_2    float64  `json:"voltage_2" xml:"voltage_2"`
	Voltage_3    float64  `json:"voltage_3" xml:"voltage_3"`
	Power_1      float64  `json:"power_1" xml:"power_1"`
	Power_2      float64  `json:"power_2" xml:"power_2"`
	Power_3      float64  `json:"power_3" xml:"power_3"`
	Cosphi_1     float64  `json:"cosphi_1" xml:"cosphi_1"`
	Cosphi_2     float64  `json:"cosphi_2" xml:"cosphi_2"`
	Cosphi_3     float64  `json:"cosphi_3" xml:"cosphi_3"`
	Frequency_1  float64  `json:"frequency_1" xml:"frequency_1"`
	Frequency_2  float64  `json:"frequency_2" xml:"frequency_2"`
	Frequency_3  float64  `json:"frequency_3" xml:"frequency_3"`
	Energy_pos_1 float64  `json:"energyPos_1" xml:"energyPos_1"`
	Energy_pos_2 float64  `json:"energyPos_2" xml:"energyPos_2"`
	Energy_pos_3 float64  `json:"energyPos_3" xml:"energyPos_3"`
	Energy_neg_1 float64  `json:"energyNeg_1" xml:"energyNeg_1"`
	Energy_neg_2 float64  `json:"energyNeg_2" xml:"energyNeg_2"`
	Energy_neg_3 float64  `json:"energyNeg_3" xml:"energyNeg_3"`
	Energy_pos_b float64  `json:"energyPos_balanced" xml:"energyPos_balanced"`
	Energy_neg_b float64  `json:"energyNeg_balanced" xml:"energyNeg_balanced"`
}

func CreateLegacyXML(result *api.QueryTableResult) (string, error) {

	var valueSeries []tXmlValue
	var i4, epb, ecb float64

	type valuelist []tXmlValue

	type dataset struct {
		valuelist
	}

	for result.Next() {

		fmt.Println("Next:")

		line := result.Record().Values()

		if line["I4"] != nil {
			i4 = line["I4"].(float64)
		} else {
			i4 = 0.0
		}
		if line["Epb"] != nil {
			epb = line["Epb"].(float64)
		} else {
			epb = 0.0
		}
		if line["Ecb"] != nil {
			ecb = line["Ecb"].(float64)
		} else {
			ecb = 0.0
		}

		valueSeries = append(valueSeries, tXmlValue{Date: result.Record().Time().Local().Format(time.DateTime), Current_1: line["I1"].(float64), Current_2: line["I2"].(float64),
			Current_3: line["I3"].(float64), Current_4: i4, Voltage_1: line["U1"].(float64), Voltage_2: line["U2"].(float64),
			Voltage_3: line["U3"].(float64), Power_1: line["P1"].(float64), Power_2: line["P2"].(float64),
			Power_3: line["P3"].(float64), Cosphi_1: line["CosPhi1"].(float64), Cosphi_2: line["CosPhi2"].(float64),
			Cosphi_3: line["CosPhi3"].(float64), Frequency_1: line["F1"].(float64), Frequency_2: line["F2"].(float64),
			Frequency_3: line["F3"].(float64), Energy_pos_1: line["Ep1"].(float64), Energy_pos_2: line["Ep2"].(float64),
			Energy_pos_3: line["Ep3"].(float64), Energy_neg_1: line["Ec1"].(float64), Energy_neg_2: line["Ec2"].(float64),
			Energy_neg_3: line["Ec3"].(float64), Energy_pos_b: epb, Energy_neg_b: ecb})
		fmt.Println(valueSeries)
	}

	if xmlstring, err := xml.MarshalIndent(dataset{valueSeries}, "", "    "); err == nil {
		xmlstring = []byte(xml.Header + string(xmlstring))
		return string(xmlstring), nil
	} else {
		return "", err
	}

}
