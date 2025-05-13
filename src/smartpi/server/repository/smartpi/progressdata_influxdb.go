package smartpiRepository

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	log "github.com/sirupsen/logrus"
)

func (s SmartPiRepository) Progressdata(starttime time.Time, stoptime time.Time, aggregatewindow string, valuelist []string, conf *config.SmartPiConfig) (models.Progressdatalist, error) {

	// Create a client
	// You can generate an API Token from the "API Tokens Tab" in the UI

	client := influxdb2.NewClientWithOptions(conf.Influxdatabase, conf.InfluxAPIToken,
		influxdb2.DefaultOptions().
			SetPrecision(time.Second).
			SetHTTPRequestTimeout(900))
	// always close client at the end
	defer client.Close()

	// Get query client
	queryAPI := client.QueryAPI("smartpi")

	devicelist := models.Devices{}
	devicelist.AddItem(models.Device{DeviceId: conf.Serial})

	progdata := models.Progressdatalist{}
	progdata.AddDevicelist(devicelist)

	fmt.Println(starttime.Format(time.RFC3339))
	fmt.Println(stoptime.Format(time.RFC3339))

	query := `
	import "timezone"

	option location = timezone.location(name: "Europe/Berlin")
	
	from(bucket: "meteringdata")
	|> range(start: ` + starttime.Format(time.RFC3339) + `, stop: ` + stoptime.Format(time.RFC3339) + `)
	|> filter(fn: (r) => r["_measurement"] == "data")`
	if !slices.Contains(valuelist, "all") {
		query = query + `|> filter(fn: (r) => `
		for i, value := range valuelist {

			if i > 0 {
				query = query + ` or `
			}
			query = query + `r["_field"] == "` + value + `"`
		}
		query = query + `)`
	}
	query = query + `|> aggregateWindow(every: ` + aggregatewindow + `, fn: mean, timeSrc: "_start")
	|> map(fn: (r) => ({
            r with
            _value: if exists r._value then float(v: r._value) * 1.0  else 0.0
            })
        )
  	|> yield(name: "mean") `

	log.Debug(query)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Error(err)
		return progdata, err
	}

	for result.Next() {
		log.Debug(result.Record())
		for i := 0; i < len(progdata.Progressdatalist); i++ {

			// if result.Record().ValueByKey("identifier").(string) == progdata.Progressdatalist[i].Device.DeviceId {
			if len(progdata.Progressdatalist[i].Data) == 0 {
				progdata.Progressdatalist[i].AddTimeseries(models.Timeseriesdata{Field: result.Record().ValueByKey("_field").(string), Datapoint: make([]models.Datapoint, 0)})
				datap := models.Datapoint{Time: result.Record().ValueByKey("_time").(time.Time), Value: result.Record().Value().(float64)}
				progdata.Progressdatalist[i].Data[0].AddDatapoint(datap)
			} else {
				for j := 0; j < len(progdata.Progressdatalist[i].Data); j++ {
				restartloop:
					if progdata.Progressdatalist[i].Data[j].Field == result.Record().ValueByKey("_field").(string) {
						if result.Record().Value() != nil {
							progdata.Progressdatalist[i].Data[j].AddDatapoint(models.Datapoint{Time: result.Record().ValueByKey("_time").(time.Time), Value: result.Record().Value().(float64)})
						} else {
							progdata.Progressdatalist[i].Data[j].AddDatapoint(models.Datapoint{Time: result.Record().ValueByKey("_time").(time.Time), Value: 0.0})
						}
					} else {
						if j < len(progdata.Progressdatalist[i].Data)-1 {
							j++
							goto restartloop
						} else {
							progdata.Progressdatalist[i].AddTimeseries(models.Timeseriesdata{Field: result.Record().ValueByKey("_field").(string), Datapoint: make([]models.Datapoint, 0)})
						}
					}
				}
			}
			// }
		}

	}
	// check for an error
	if result.Err() != nil {
		log.Error("query parsing error: %\n", result.Err().Error())
		return progdata, result.Err()
	}

	return progdata, nil

}
