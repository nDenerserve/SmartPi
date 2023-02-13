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
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
)

func CreateCSV(start time.Time, end time.Time) string {

	config := config.NewConfig()

	data := ReadChartData(config.DatabaseDir, start, end)

	csv := "date;current_1;current_2;current_3;current_4;voltage_1;voltage_2;voltage_3;power_1;power_2;power_3;cosphi_1;cosphi_2;cosphi_3;frequency_1;frequency_2;frequency_3;energy_pos_1;energy_pos_2;energy_pos_3;energy_neg_1;energy_neg_2;energy_neg_3"
	csv = csv + "\n"

	for _, dataelement := range data {
		ti := dataelement.Date
		csv = csv + ti.Format(config.CSVtimeformat)

		datav := reflect.ValueOf(dataelement).Elem()

		for i := 1; i < datav.NumField(); i++ {
			val := datav.Field(i).Interface().(float64)
			if math.IsNaN(val) {
				val = 0.0
			}
			csv = csv + ";" + strings.Replace(strconv.FormatFloat(val, 'f', 5, 64), ".", config.CSVdecimalpoint, -1)
		}

		csv = csv + "\n"
	}
	return csv
}
