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
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/network"
	log "github.com/sirupsen/logrus"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type MinuteValues struct {
	Date                                                                                                                                                                                                                                                                             time.Time
	Current_1, Current_2, Current_3, Current_4, Voltage_1, Voltage_2, Voltage_3, Power_1, Power_2, Power_3, Cosphi_1, Cosphi_2, Cosphi_3, Frequency_1, Frequency_2, Frequency_3, Energy_pos_1, Energy_pos_2, Energy_pos_3, Energy_neg_1, Energy_neg_2, Energy_neg_3, Energy_balanced float64
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./sqlite2influxdb <sqlitefile>")
		os.Exit(1)
	}
	// Parse command line arguments
	fmt.Println(os.Args[1])

	smartpiConfig := config.NewSmartPiConfig()

	values := ReadChartData(os.Args[1])

	for _, element := range values {
		fmt.Println(element.Date, element.Current_1, element.Current_2, element.Current_3, element.Current_4, element.Voltage_1, element.Voltage_2, element.Voltage_3, element.Power_1, element.Power_2, element.Power_3, element.Cosphi_1, element.Cosphi_2, element.Cosphi_3, element.Frequency_1, element.Frequency_2, element.Frequency_3, element.Energy_pos_1, element.Energy_pos_2, element.Energy_pos_3, element.Energy_neg_1, element.Energy_neg_2, element.Energy_neg_3, element.Energy_balanced)
		InsertInfluxSampleData(smartpiConfig, element.Date, element)
	}
}

func ReadChartData(databasefile string) []*MinuteValues {

	values := []*MinuteValues{}

	db, err := sql.Open("sqlite3", databasefile)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT date, current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3, energy_pos_balanced, energy_neg_balanced FROM " + RemoveFileExtensionAndPath(databasefile) + " ORDER BY date")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	var rowcounter = 0

	for rows.Next() {
		var dateentry string
		var current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3, energy_pos_balanced, energy_neg_balanced float64
		err = rows.Scan(&dateentry, &current_1, &current_2, &current_3, &current_4, &voltage_1, &voltage_2, &voltage_3, &power_1, &power_2, &power_3, &cosphi_1, &cosphi_2, &cosphi_3, &frequency_1, &frequency_2, &frequency_3, &energy_pos_1, &energy_pos_2, &energy_pos_3, &energy_neg_1, &energy_neg_2, &energy_neg_3, &energy_pos_balanced, &energy_neg_balanced)
		if err != nil {
			log.Println(err)
		}

		val := new(MinuteValues)

		val.Date, err = time.ParseInLocation("2006-01-02T15:04:05Z", dateentry, time.Now().Location())
		// val.Date, err = time.Parse("2006-01-02T15:04:05Z",dateentry)
		val.Current_1 = current_1
		val.Current_2 = current_2
		val.Current_3 = current_3
		val.Current_4 = current_4
		val.Voltage_1 = voltage_1
		val.Voltage_2 = voltage_2
		val.Voltage_3 = voltage_3
		val.Power_1 = power_1
		val.Power_2 = power_2
		val.Power_3 = power_3
		val.Cosphi_1 = cosphi_1
		val.Cosphi_2 = cosphi_2
		val.Cosphi_3 = cosphi_3
		val.Frequency_1 = frequency_1
		val.Frequency_2 = frequency_2
		val.Frequency_3 = frequency_3
		val.Energy_pos_1 = energy_pos_1
		val.Energy_pos_2 = energy_pos_2
		val.Energy_pos_3 = energy_pos_3
		val.Energy_neg_1 = energy_neg_1
		val.Energy_neg_2 = energy_neg_2
		val.Energy_neg_3 = energy_neg_3
		val.Energy_balanced = energy_pos_balanced - energy_neg_balanced

		values = append(values, val)

		if err != nil {
			log.Println(err)
		}
		rowcounter++
	}

	return values

}

func InsertInfluxSampleData(c *config.SmartPiConfig, t time.Time, values *MinuteValues) {

	client := influxdb2.NewClient(c.Influxdatabase, c.InfluxAPIToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(c.InfluxOrg, c.InfluxBucket)

	// Create a point and add to batch
	macaddress := network.GetMacAddr()
	tags := map[string]string{"mac": macaddress, "type": "electric"}
	fields := map[string]interface{}{
		"I1":      float64(values.Current_1),
		"I2":      float64(values.Current_2),
		"I3":      float64(values.Current_3),
		"I4":      float64(values.Current_4),
		"U1":      float64(values.Voltage_1),
		"U2":      float64(values.Voltage_1),
		"U3":      float64(values.Voltage_1),
		"P1":      float64(values.Power_1),
		"P2":      float64(values.Power_2),
		"P3":      float64(values.Power_3),
		"CosPhi1": float64(values.Cosphi_1),
		"CosPhi2": float64(values.Cosphi_2),
		"CosPhi3": float64(values.Cosphi_3),
		"F1":      float64(values.Frequency_1),
		"F2":      float64(values.Frequency_2),
		"F3":      float64(values.Frequency_3),
		"Ec1":     float64(values.Energy_pos_1),
		"Ec2":     float64(values.Energy_pos_2),
		"Ec3":     float64(values.Energy_pos_3),
		"Ep1":     float64(values.Energy_neg_1),
		"Ep2":     float64(values.Energy_neg_2),
		"Ep3":     float64(values.Energy_neg_3),
		"Ebal":    float64(values.Energy_balanced),
	}
	pt := influxdb2.NewPoint("data", tags, fields, t)
	writeAPI.WritePoint(context.Background(), pt)
}

func RemoveFileExtensionAndPath(filename string) string {
	base := filepath.Base(filename) // Removes the directory path
	if dotIndex := strings.LastIndex(base, "."); dotIndex != -1 {
		return base[:dotIndex] // Removes the file extension
	}
	return base
}
