// Database Exporter

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi"
	log "github.com/sirupsen/logrus"
)

func updateInfluxDatabase(c *config.Config, data smartpi.ReadoutAccumulator, consumedWattHourBalanced float64, producedWattHourBalanced float64) {
	t := time.Now()

	logLine := "## InfluxDB (meteringdata) Database Update ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	// logLine += dbFileName
	log.Info(logLine)

	smartpi.InsertInfluxData(c, t, data, consumedWattHourBalanced, producedWattHourBalanced)
}

func updateInfluxFastdata(c *config.Config, data *smartpi.ADE7878Readout) {
	t := time.Now()

	logLine := "## InfluxDB (fastmeasurement) Database Update ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	// logLine += dbFileName
	log.Info(logLine)

	smartpi.InsertFastData(c, t, data)
}

func updateSQLiteDatabase(c *config.Config, data smartpi.ReadoutAccumulator, consumedWattHourBalanced float64, producedWattHourBalanced float64) {
	t := time.Now()
	dbFileName := "smartpi_logdata_" + t.Format("200601") + ".db"

	logLine := "## SQLITE Database Update ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	logLine += dbFileName
	log.Info(logLine)

	if _, err := os.Stat(filepath.Join(c.DatabaseDir, dbFileName)); os.IsNotExist(err) {
		log.Debug("Creating new database file.")
		smartpi.CreateSQlDatabase(c.DatabaseDir, t)
	}
	smartpi.InsertSQLData(c.DatabaseDir, t, data, consumedWattHourBalanced, producedWattHourBalanced)
}
