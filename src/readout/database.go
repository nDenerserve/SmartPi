// Database Exporter

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/nDenerserve/SmartPi/src/smartpi"
)

func updateSQLiteDatabase(c *smartpi.Config, data []float32) {
	t := time.Now()
	logLine := "## SQLITE Database Update ##"
	logLine += fmt.Sprintf(t.Format(" 2006-01-02 15:04:05 "))
	logLine += fmt.Sprintf("I1: %g  I2: %g  I3: %g  I4: %g  ", data[0], data[1], data[2], data[3])
	logLine += fmt.Sprintf("V1: %g  V2: %g  V3: %g  ", data[4], data[5], data[6])
	logLine += fmt.Sprintf("P1: %g  P2: %g  P3: %g  ", data[7], data[8], data[9])
	logLine += fmt.Sprintf("COS1: %g  COS2: %g  COS3: %g  ", data[10], data[11], data[12])
	logLine += fmt.Sprintf("F1: %g  F2: %g  F3: %g  ", data[13], data[14], data[15])
	logLine += fmt.Sprintf("EB1: %g  EB2: %g  EB3: %g  ", data[16], data[17], data[18])
	logLine += fmt.Sprintf("EL1: %g  EL2: %g  EL3: %g", data[19], data[20], data[21])
	log.Info(logLine)

	dbFileName := "smartpi_logdata_" + t.Format("200601") + ".db"
	if _, err := os.Stat(filepath.Join(c.DatabaseDir, dbFileName)); os.IsNotExist(err) {
		log.Debug("Creating new database file.")
		smartpi.CreateSQlDatabase(c.DatabaseDir, t)
	}
	smartpi.InsertData(c.DatabaseDir, t, data)
}
