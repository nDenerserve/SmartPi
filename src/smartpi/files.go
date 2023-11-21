// File Exporter

package smartpi

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	log "github.com/sirupsen/logrus"
)

func WriteSharedFile(c *config.Config, values *ADE7878Readout, balancedvalue float64) {
	var f *os.File
	var err error
	var p models.Phase
	s := make([]string, 26)
	i := 0
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.Current[p])
		i++
	}
	s[i] = fmt.Sprint(values.Current[models.PhaseN])
	i++
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.Voltage[p])
		i++
	}
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.ActiveWatts[p])
		i++
	}
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.CosPhi[p])
		i++
	}
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.Frequency[p])
		i++
	}
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.Energyconsumption[p])
		i++
	}
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.Energyproduction[p])
		i++
	}

	// sald Values
	s[i] = fmt.Sprint(balancedvalue)
	i++

	//PowerFactor
	for _, p = range MainPhases {
		s[i] = fmt.Sprint(values.PowerFactor[p])
		i++
	}

	t := time.Now()
	timeStamp := t.Format("2006-01-02 15:04:05")
	logLine := "## Shared File Update ## "
	logLine += fmt.Sprintf(timeStamp)
	logLine += fmt.Sprintf(" I1: %s  I2: %s  I3: %s  I4: %s  ", s[0], s[1], s[2], s[3])
	logLine += fmt.Sprintf("V1: %s  V2: %s  V3: %s  ", s[4], s[5], s[6])
	logLine += fmt.Sprintf("P1: %s  P2: %s  P3: %s  ", s[7], s[8], s[9])
	logLine += fmt.Sprintf("COS1: %s  COS2: %s  COS3: %s  ", s[10], s[11], s[12])
	logLine += fmt.Sprintf("F1: %s  F2: %s  F3: %s  ", s[13], s[14], s[15])
	logLine += fmt.Sprintf("Ec1: %s  Ec2: %s  Ec3: %s  ", s[16], s[17], s[18])
	logLine += fmt.Sprintf("Ep1: %s  Ep2: %s  Ep3: %s  ", s[19], s[20], s[21])
	logLine += fmt.Sprintf("Balanced: %s  ", s[22])
	logLine += fmt.Sprintf("PF1: %s  PF2: %s  PF3: %s  ", s[23], s[24], s[25])
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

func UpdateCounterFile(c *config.Config, f string, v float64) float64 {
	t := time.Now()
	var counter float64
	counterFile, err := os.ReadFile(f)
	if err == nil {
		counter, err = strconv.ParseFloat(string(counterFile), 64)
		if err != nil {
			log.Errorf("unable to read counter file %q, %q", f, err)
			log.Errorf("try to create new counterfile")
			counter = 0.0
			err = os.WriteFile(f, []byte(strconv.FormatFloat(counter, 'f', 8, 64)), 0644)
			if err != nil {
				log.Errorf("unable to create counter file %q, %q", f, err)
			}
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
	return counter + v
}

func ReadCounterFile(c *config.Config, f string) float64 {
	t := time.Now()
	var counter float64
	counterFile, err := ioutil.ReadFile(f)
	if err == nil {
		counter, err = strconv.ParseFloat(string(counterFile), 64)
		if err != nil {
			log.Errorf("unable to read counter file %q, %q", f, err)
			counter = 0.0
		}
	} else {
		counter = 0.0
	}

	logLine := "## Read Persistent counter file ##"
	logLine += t.Format(" 2006-01-02 15:04:05 ")
	logLine += fmt.Sprintf("File: %q  Current: %g  ", f, counter)
	log.Info(logLine)

	return counter
}
