package smartpi

import (
	// "database/sql"
	// "fmt"
	// "os"

	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/nDenerserve/SmartPi/src/smartpi/network"
	log "github.com/sirupsen/logrus"
)

// type MinuteValues struct {
// 	Date                                                                                                                                                                                                                                                            time.Time
// 	Current_1, Current_2, Current_3, Current_4, Voltage_1, Voltage_2, Voltage_3, Power_1, Power_2, Power_3, Cosphi_1, Cosphi_2, Cosphi_3, Frequency_1, Frequency_2, Frequency_3, Energy_pos_1, Energy_pos_2, Energy_pos_3, Energy_neg_1, Energy_neg_2, Energy_neg_3 float64
// }

func InsertInfluxData(c *Config, t time.Time, v ReadoutAccumulator, consumedWattHourBalanced float64, producedWattHourBalanced float64) {

	dbc, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     c.Influxdatabase,
		Username: c.Influxuser,
		Password: c.Influxpassword,
	})
	if err != nil {
		log.Printf("Error creating InfluxDB Client: ", err.Error())
	}
	defer dbc.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "MeteringData",
		Precision: "s",
	})

	// Create a point and add to batch
	macaddress := network.GetMacAddr()
	tags := map[string]string{"serial": macaddress, "type": "electric"}
	fields := map[string]interface{}{
		"I1":      float64(v.Current[PhaseA]),
		"I2":      float64(v.Current[PhaseB]),
		"I3":      float64(v.Current[PhaseC]),
		"I4":      float64(v.Current[PhaseN]),
		"U1":      float64(v.Voltage[PhaseA]),
		"U2":      float64(v.Voltage[PhaseB]),
		"U3":      float64(v.Voltage[PhaseC]),
		"P1":      float64(v.ActiveWatts[PhaseA]),
		"P2":      float64(v.ActiveWatts[PhaseB]),
		"P3":      float64(v.ActiveWatts[PhaseC]),
		"CosPhi1": float64(v.CosPhi[PhaseA]),
		"CosPhi2": float64(v.CosPhi[PhaseB]),
		"CosPhi3": float64(v.CosPhi[PhaseC]),
		"F1":      float64(v.Frequency[PhaseA]),
		"F2":      float64(v.Frequency[PhaseB]),
		"F3":      float64(v.Frequency[PhaseC]),
		"Ec1":     float64(v.WattHoursConsumed[PhaseA]),
		"Ec2":     float64(v.WattHoursConsumed[PhaseB]),
		"Ec3":     float64(v.WattHoursConsumed[PhaseC]),
		"Ep1":     float64(v.WattHoursProduced[PhaseA]),
		"Ep2":     float64(v.WattHoursProduced[PhaseB]),
		"Ep3":     float64(v.WattHoursProduced[PhaseC]),
		"bEc":     float64(consumedWattHourBalanced),
		"bEp":     float64(producedWattHourBalanced),
	}
	pt, err := client.NewPoint("data", tags, fields, time.Now())
	if err != nil {
		log.Printf("Error: ", err.Error())
	}
	bp.AddPoint(pt)

	// Write the batch
	err = dbc.Write(bp)
	if err != nil {
		log.Printf("Error: ", err.Error())
	}
}
