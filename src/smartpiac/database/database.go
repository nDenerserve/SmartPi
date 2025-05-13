// Database Exporter

package smartpiacDatabase

import (
	"context"
	"fmt"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "github.com/sirupsen/logrus"
)

func UpdateInfluxDatabase(c *config.SmartPiConfig, data models.ReadoutAccumulator, consumedWattHourBalanced float64, producedWattHourBalanced float64) {
	t := time.Now()

	logLine := "## InfluxDB database update minute values ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	// logLine += dbFileName
	log.Info(logLine)

	InsertInfluxData(c, t, data, consumedWattHourBalanced, producedWattHourBalanced)
}

func UpdateSampleInfluxDatabase(c *config.SmartPiConfig, data *models.ADE7878Readout, wattHourBalanced float64) {
	t := time.Now()

	logLine := "## InfluxDB database update sample values ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	// logLine += dbFileName
	log.Info(logLine)

	InsertInfluxSampleData(c, t, data, wattHourBalanced)
}

func UpdateCalculatedInfluxDatabase(c *config.SmartPiConfig, consumedWattHourBalanced float64, producedWattHourBalanced float64) {
	t := time.Now()

	logLine := "## InfluxDB database update calculated values ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	// logLine += dbFileName
	log.Info(logLine)

	InsertCalculatedInfluxData(c, t, consumedWattHourBalanced, producedWattHourBalanced)
}

func ReadData(conf *config.SmartPiConfig, starttime time.Time, stoptime time.Time, params ...string) (*api.QueryTableResult, error) {

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
	fmt.Println(query)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Error(err)
		return nil, result.Err()
	}

	return result, nil
}
