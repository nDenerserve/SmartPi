package smartpidc

import (
	// "database/sql"
	// "fmt"
	// "os"

	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi/network"

	log "github.com/sirupsen/logrus"
)

func InsertInfluxData(c *config.DCconfig, inputconfig []int, values []float64, power []float64, energyConsumed []float64, energyProduced []float64) {

	client := influxdb2.NewClient(c.Influxdatabase, c.InfluxAPIToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(c.InfluxOrg, c.InfluxBucket)

	log.Debug("InfluxDB: " + c.Influxdatabase + "  User: " + c.Influxuser + "  Password: " + c.Influxpassword)

	// Create a point and add to batch
	macaddress := network.GetMacAddr()
	tags := map[string]string{"mac": macaddress, "type": "electric"}
	fields := map[string]interface{}{

		// "I1": float64(13.5),
	}

	for i := 0; i < len(values); i++ {
		fields[c.InputName[i]] = float64(values[i])
	}

	for i := 0; i < len(power); i++ {
		fields[c.PowerName[i]] = float64(power[i])
	}

	for i := 0; i < len(energyConsumed); i++ {
		fields[c.EnergyConsumptionName[i]] = float64(energyConsumed[i])
	}

	for i := 0; i < len(energyProduced); i++ {
		fields[c.EnergyProductionName[i]] = float64(energyProduced[i])
	}

	fmt.Println(fields)

	pt := influxdb2.NewPoint("data", tags, fields, time.Now())

	writeAPI.WritePoint(context.Background(), pt)
}

func ReadCSVData(c *config.DCconfig, starttime time.Time, endtime time.Time) string {

	loc, _ := time.LoadLocation("UTC")
	// endtime := time.Now()
	// starttime := endtime.Add(time.Minute * -10)

	querystring := url.QueryEscape("SELECT mean(\"P1\") AS \"P1\", mean(\"P2\") AS \"P2\", mean(\"P3\") AS \"P3\", mean(\"bEc\") AS \"bEc\", mean(\"bEp\") AS \"bEp\", mean(\"I1\") AS \"i1\", mean(\"I2\") AS \"i2\", mean(\"I3\") AS \"i3\", mean(\"I4\") AS \"i4\", mean(\"U1\") AS \"u1\", mean(\"U2\") AS \"u2\", mean(\"U3\") AS \"u3\", mean(\"Ec1\") AS \"Ec1\", mean(\"Ec2\") AS \"Ec2\", mean(\"Ec3\") AS \"Ec3\", mean(\"Ep1\") AS \"Ep1\", mean(\"Ep2\") AS \"Ep2\", mean(\"Ep3\") AS \"Ep3\", mean(\"CosPhi1\") AS \"CosPhi1\", mean(\"CosPhi2\") AS \"CosPhi2\", mean(\"CosPhi3\") AS \"CosPhi3\", mean(\"F1\") AS \"f1\", mean(\"F2\") AS \"f2\", mean(\"F3\") AS \"f3\" FROM \"data\" WHERE time >= '" + starttime.In(loc).Format("2006-01-02 15:04:05") + "' AND time <= '" + endtime.In(loc).Format("2006-01-02 15:04:05") + "' GROUP BY time(1m) fill(null)")
	req, err := http.NewRequest("POST", c.Influxdatabase+"/query?db=MeteringData&u="+c.Influxuser+"&p="+c.Influxpassword+"&q="+querystring, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Accept", "application/csv")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	stringbody := strings.Replace(string(body), ",", ";", -1)
	stringbody = strings.Replace(string(stringbody), ".", c.CSVdecimalpoint, -1)

	return stringbody
}

func ExampleClient_query(c *config.DCconfig) {

	dbc, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     c.Influxdatabase,
		Username: c.Influxuser,
		Password: c.Influxpassword,
	})
	if err != nil {
		log.Printf("Error creating InfluxDB Client: ", err.Error())
	}
	defer dbc.Close()

	q := client.NewQuery("SELECT * FROM meteringdata", "meteringdata", "ns")
	if response, err := dbc.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}
}
