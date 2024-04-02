/*
    Copyright (C) Jens Ramhorst
	  This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var appVersion = "No Version Provided"

func main() {

	config := config.NewConfig()

	version := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if !config.FTPupload {
		os.Exit(0)
	}

	startDate := time.Now()

	location := time.Now().Location()

	lastdate, err := os.ReadFile("/var/smartpi/csvftp")
	if err == nil {
		startDate, err = time.ParseInLocation("2006-01-02 15:04:05", string(lastdate), location)
		if err != nil {
			log.Println(err)
		}

	} else {
		startDate = time.Now().AddDate(0, 0, -1)
	}

	tlsConfig := &tls.Config{
		// Enable TLS 1.2.
		InsecureSkipVerify: true,
		MaxVersion:         tls.VersionTLS12,
	}
	ftpdialoption := ftp.DialWithExplicitTLS(tlsConfig)

	// startDate = startDate.UTC()
	endDate := time.Now()

	log.Debugf("Startdate: " + startDate.Format("2006-01-02 15:04:05"))
	log.Debugf("Enddate: " + endDate.Format("2006-01-02 15:04:05"))

	if config.FTPcsv {

		file := bytes.NewBufferString(smartpi.CreateCSV(startDate, endDate))

		ftpserver := config.FTPserver
		if !strings.Contains(ftpserver, ":") {
			ftpserver = ftpserver + ":21"
		}

		client, err := ftp.Dial(ftpserver, ftpdialoption)
		if err != nil {
			log.Fatal(err)
		}

		err = client.Login(config.FTPuser, config.FTPpass)
		if err != nil {
			log.Fatal(err)
		}

		ftp_path := config.FTPpath
		pathlist := strings.Split(ftp_path, "/")
		for i := 0; i < len(pathlist); i++ {
			if len(pathlist[i]) == 0 {
				pathlist = append(pathlist[:i], pathlist[i+1:]...)
			}
		}

		workingpath := "/"
		createpath := ""

		for j := 0; j < len(pathlist); j++ {
			if j > 0 {
				workingpath = workingpath + pathlist[j-1] + "/"
			}
			createpath = createpath + "/" + pathlist[j]
		}

		filename := time.Now().Format("20060102150405") + "_" + config.Serial + ".csv"
		err = client.Stor(createpath+"/"+filename, file)
		if err != nil {
			panic(err)
		} else {
			err = os.WriteFile("/var/smartpi/csvftp", []byte(endDate.Local().Format("2006-01-02 15:04:05")), 0644)
			if err != nil {
				panic(err)
			}
			if err := client.Quit(); err != nil {
				log.Fatal(err)
			}

		}

	}

	if config.FTPxml {

		file := bytes.NewBufferString(smartpi.CreateXML(startDate, endDate))

		ftpserver := config.FTPserver
		if !strings.Contains(ftpserver, ":") {
			ftpserver = ftpserver + ":21"
		}

		client, err := ftp.Dial(ftpserver, ftpdialoption)
		if err != nil {
			log.Fatal(err)
		}

		err = client.Login(config.FTPuser, config.FTPpass)
		if err != nil {
			log.Fatal(err)
		}

		ftp_path := config.FTPpath
		pathlist := strings.Split(ftp_path, "/")
		for i := 0; i < len(pathlist); i++ {
			if len(pathlist[i]) == 0 {
				pathlist = append(pathlist[:i], pathlist[i+1:]...)
			}
		}

		workingpath := "/"
		createpath := ""

		for j := 0; j < len(pathlist); j++ {
			if j > 0 {
				workingpath = workingpath + pathlist[j-1] + "/"
			}
			createpath = createpath + "/" + pathlist[j]
		}

		filename := time.Now().Format("20060102150405") + "_" + config.Serial + ".xml"
		err = client.Stor(createpath+"/"+filename, file)
		if err != nil {
			panic(err)
		} else {
			err = os.WriteFile("/var/smartpi/csvftp", []byte(endDate.Local().Format("2006-01-02 15:04:05")), 0644)
			if err != nil {
				panic(err)
			}
			if err := client.Quit(); err != nil {
				log.Fatal(err)
			}

		}

	}

}
