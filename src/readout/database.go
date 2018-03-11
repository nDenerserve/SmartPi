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

func updateSQLiteDatabase(c *smartpi.Config, data smartpi.ReadoutAccumulator) {
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
	smartpi.InsertData(c.DatabaseDir, t, data)
}
