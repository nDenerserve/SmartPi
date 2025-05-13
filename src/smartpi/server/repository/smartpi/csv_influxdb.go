package smartpiRepository

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	log "github.com/sirupsen/logrus"
)

func (s SmartPiRepository) ExportCSV(conf *config.SmartPiConfig, starttime time.Time, stoptime time.Time, decimalpoint string, params ...string) (string, error) {

	var aggregate string
	var influxFunc string

	if len(params) == 1 {
		aggregate = params[0]
	} else if len(params) == 2 {
		aggregate = params[0]
		influxFunc = params[1]
	}

	client := influxdb2.NewClientWithOptions(conf.Influxdatabase, conf.InfluxAPIToken,
		influxdb2.DefaultOptions().
			SetPrecision(time.Second).
			SetHTTPRequestTimeout(900))

	// always close client at the end
	defer client.Close()

	// Get query client
	queryAPI := client.QueryAPI("smartpi")

	query := `
	import "timezone"

	option location = timezone.location(name: "Europe/Berlin")
	
	from(bucket: "meteringdata")
		|> range(start: ` + starttime.Format(time.RFC3339) + `, stop: ` + stoptime.Format(time.RFC3339) + `)
		|> filter(fn: (r) => r["_measurement"] == "data")`
	if (aggregate != "") && (influxFunc != "") {
		query = query + `
		|> aggregateWindow(every: ` + aggregate + `, fn: ` + influxFunc + `, createEmpty: false)`
	} else if aggregate != "" {
		query = query + `
		|> aggregateWindow(every: ` + aggregate + `, fn: mean, createEmpty: false)`
	}

	query = query + `
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> yield(name: "mean")
	`
	log.Debug(query)
	// fmt.Println(query)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Error(err)
		return "", result.Err()
	}

	csv, err := s.createLegacyCSV(result, decimalpoint)

	// check for an error
	if err != nil {
		log.Error("query parsing error: %\n", result.Err().Error())
		return "", result.Err()
	}

	return csv, nil

}

func (s SmartPiRepository) createLegacyCSV(result *api.QueryTableResult, csvDecimalpoint string) (string, error) {

	csvLine := []string{"date", "current_1", "current_2", "current_3", "current_4", "voltage_1", "voltage_2", "voltage_3", "power_1", "power_2", "power_3", "cosphi_1", "cosphi_2", "cosphi_3", "frequency_1", "frequency_2", "frequency_3", "energy_pos_1", "energy_pos_2", "energy_pos_3", "energy_neg_1", "energy_neg_2", "energy_neg_3", "energy_pos_balanced", "energy_neg_balanced"}
	var buff bytes.Buffer
	w := csv.NewWriter(io.Writer(&buff))
	w.Comma = ';'
	w.UseCRLF = true

	if err := w.Write(csvLine); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}
	// strings.Replace(strings.Replace(strconv.FormatFloat(val, 'f', 5, 64), ".", config.CSVdecimalpoint
	for result.Next() {
		csvLine = []string{"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0"}
		csvLine[0] = result.Record().Time().Local().Format(time.DateTime)
		line := result.Record().Values()
		csvLine[1] = strings.Replace(strconv.FormatFloat(line["I1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[2] = strings.Replace(strconv.FormatFloat(line["I2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[3] = strings.Replace(strconv.FormatFloat(line["I3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		if line["I4"] != nil {
			csvLine[4] = strings.Replace(strconv.FormatFloat(line["I4"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		}
		csvLine[5] = strings.Replace(strconv.FormatFloat(line["U1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[6] = strings.Replace(strconv.FormatFloat(line["U2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[7] = strings.Replace(strconv.FormatFloat(line["U3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[8] = strings.Replace(strconv.FormatFloat(line["P1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[9] = strings.Replace(strconv.FormatFloat(line["P2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[10] = strings.Replace(strconv.FormatFloat(line["P3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[11] = strings.Replace(strconv.FormatFloat(line["CosPhi1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[12] = strings.Replace(strconv.FormatFloat(line["CosPhi2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[13] = strings.Replace(strconv.FormatFloat(line["CosPhi3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[14] = strings.Replace(strconv.FormatFloat(line["F1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[15] = strings.Replace(strconv.FormatFloat(line["F2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[16] = strings.Replace(strconv.FormatFloat(line["F3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[17] = strings.Replace(strconv.FormatFloat(line["Ep1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[18] = strings.Replace(strconv.FormatFloat(line["Ep2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[19] = strings.Replace(strconv.FormatFloat(line["Ep3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[20] = strings.Replace(strconv.FormatFloat(line["Ec1"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[21] = strings.Replace(strconv.FormatFloat(line["Ec2"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		csvLine[22] = strings.Replace(strconv.FormatFloat(line["Ec3"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		if line["Epb"] != nil {
			csvLine[22] = strings.Replace(strconv.FormatFloat(line["Epb"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		}
		if line["Ecb"] != nil {
			csvLine[23] = strings.Replace(strconv.FormatFloat(line["Ecb"].(float64), 'f', -1, 64), ".", csvDecimalpoint, -1)
		}

		if err := w.Write(csvLine); err != nil {
			log.Error("error writing record to csv:", err)
			return "", err
		}
	}

	w.Flush()

	return buff.String(), nil

}
