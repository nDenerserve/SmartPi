// File Exporter

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Nitroman605/SmartPi/src/smartpi"

	log "github.com/Sirupsen/logrus"
)

func writeSharedFile(c *smartpi.Config, values *smartpi.ADE7878Readout, balancedvalue float64) {
	var f *os.File
	var err error
	var p smartpi.Phase
	s := make([]string, 17)
	i := 0
	for _, p = range smartpi.MainPhases {
		s[i] = fmt.Sprint(values.Current[p])
		i++
	}
	s[i] = fmt.Sprint(values.Current[smartpi.PhaseN])
	i++
	for _, p = range smartpi.MainPhases {
		s[i] = fmt.Sprint(values.Voltage[p])
		i++
	}
	for _, p = range smartpi.MainPhases {
		s[i] = fmt.Sprint(values.ActiveWatts[p])
		i++
	}
	for _, p = range smartpi.MainPhases {
		s[i] = fmt.Sprint(values.CosPhi[p])
		i++
	}
	for _, p = range smartpi.MainPhases {
		s[i] = fmt.Sprint(values.Frequency[p])
		i++
	}
	// sald Values
	s[i] = fmt.Sprint(balancedvalue)

	t := time.Now()
	timeStamp := t.Format("2006-01-02 15:04:05")
	logLine := "## Shared File Update ## "
	logLine += fmt.Sprintf(timeStamp)
	logLine += fmt.Sprintf(" I1: %s  I2: %s  I3: %s  I4: %s  ", s[0], s[1], s[2], s[3])
	logLine += fmt.Sprintf("V1: %s  V2: %s  V3: %s  ", s[4], s[5], s[6])
	logLine += fmt.Sprintf("P1: %s  P2: %s  P3: %s  ", s[7], s[8], s[9])
	logLine += fmt.Sprintf("COS1: %s  COS2: %s  COS3: %s  ", s[10], s[11], s[12])
	logLine += fmt.Sprintf("F1: %s  F2: %s  F3: %s  ", s[13], s[14], s[15])
	logLine += fmt.Sprintf("Balanced: %s  ", s[16])
	log.Info(logLine)
	sharedFile := filepath.Join(c.SharedDir, c.SharedFile)
	if _, err = os.Stat(sharedFile); os.IsNotExist(err) {
		os.MkdirAll(c.SharedDir, 0777)
		f, err = os.Create(sharedFile)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(sharedFile, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()
	_, err = f.WriteString(timeStamp + ";" + strings.Join(s, ";") + ";")
	if err != nil {
		panic(err)
	}
	f.Close()
}

func updateCounterFile(c *smartpi.Config, f string, v float64) {
	t := time.Now()
	var counter float64
	counterFile, err := ioutil.ReadFile(f)
	if err == nil {
		counter, err = strconv.ParseFloat(string(counterFile), 64)
		if err != nil {
			counter = 0.0
			log.Fatal(err)
		}
	} else {
		counter = 0.0
	}

	logLine := "## Persistent counter file update ##"
	logLine += t.Format(" 2006-01-02 15:04:05 ")
	logLine += fmt.Sprintf("File: %q  Current: %g  Increment: %g", f, counter, v)
	log.Info(logLine)

	err = ioutil.WriteFile(f, []byte(strconv.FormatFloat(counter+v, 'f', 8, 64)), 0644)
	if err != nil {
		panic(err)
	}
}
