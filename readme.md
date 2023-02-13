# SmartPi Energy Monitor 

[![CircleCI](https://circleci.com/gh/nDenerserve/SmartPi.svg?style=svg)](https://circleci.com/gh/nDenerserve/SmartPi)

[SmartPi open source](http://www.enerserve.eu.eu/products/smartpi).

## Forum
https://forum.enerserve.eu

## Installation
The easiest way is to download a prebuild image.
Further information under: https://www.enerserve.eu/service.html

Download Raspbian Bullseye Lite (64bit) (the big bullsyeye should also work) from https://www.raspberrypi.com/software/ and copy it on your SD card. The easiest way is to use the Raspbbery Pi Imager.

Create a user with the name smartpi and a password of your choice. We use the password smart4pi here. During installation, please use the password you have chosen and replace smart4pi with the one you have chosen.

##### Update packet list and update packages
    
    wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
    echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list

    sudo apt update
    sudo apt upgrade

##### Install additional packages.

    sudo apt-get install -y grafana, sqlite3, ppp, wvdial, libpam0g, npm, influxdb, grafana

##### Start InfluxDB at startup and create all needed tables

    sudo systemctl unmask influxdb
    sudo systemctl enable influxdb
    sudo systemctl start influxdb

Create needed tables. You can use your own password, but keep in mind that you have to change the password in /etc/smartpi too.

    influx -execute "CREATE USER smartpi WITH PASSWORD 'smart4pi' WITH ALL PRIVILEGES"
    influx -username admin -password smart4pi -execute "CREATE DATABASE MeteringData"
    influx -username admin -password smart4pi -execute "CREATE CONTINUOUS QUERY minmax_1h ON MeteringData BEGIN SELECT max(CosPhi1) AS MAX_CosPhi1,max(CosPhi2) AS MAX_CosPhi2,max(CosPhi3) AS MAX_CosPhi3,max(F1) AS MAX_F1,max(F2) AS MAX_F2,max(F3) AS MAX_F3,max(U1) AS MAX_U1,max(U2) AS MAX_U2,max(U3) AS MAX_U3,max(I1) AS MAX_I1,max(I2) AS MAX_I2,max(I3) AS MAX_I3,max(I4) AS MAX_I4,max(P1) AS MAX_P1,max(P2) AS MAX_P2,max(P3) AS MAX_P3,max(P3) AS MAX_P3,min(CosPhi1) AS MIN_CosPhi1,min(CosPhi2) AS MIN_CosPhi2,min(CosPhi3) AS MIN_CosPhi3,min(F1) AS MIN_F1,min(F2) AS MIN_F2,min(F3) AS MIN_F3,min(U1) AS MIN_U1,min(U2) AS MIN_U2,min(U3) AS MIN_U3,min(I1) AS MIN_I1,min(I2) AS MIN_I2,min(I3) AS MIN_I3,min(I4) AS MIN_I4,min(P1) AS MIN_P1,min(P2) AS MIN_P2,min(P3) AS MIN_P3,min(P3) AS MIN_P3 INTO hour FROM data GROUP BY time(1h),serial,type END"
    influx -username admin -password smart4pi -execute "CREATE DATABASE FastMeasurement"

##### Create tmpfs in /etc/fstab

    tmpfs /var/tmp/smartpi tmpfs nodev,nosuid,size=20M 0 0
    

For secure 24/7 operation, we recommend that you also create a tmpf for the log and tmp files.

    tmpfs /var/log tmpfs defaults,noatime,mode=1777,size=10M 0 0
    tmpfs /var/tmp tmpfs defaults,noatime,mode=1777,size=30M 0 0
    tmpfs /tmp tmpfs defaults,noatime,mode=1777,size=20M 0 0

##### Enable i2c kernel module

`i2c-dev` is required for communicating with the SmartPi.

To check to see if the module is loaded:

    sudo lsmod | grep i2c

This should return something like this:

    i2c_dev                20480  0
    i2c_bcm2835            16384  0


If the module is not listed, activate i2c via 
sudo raspi-config or add it in following way to the system: 

    echo 'i2c-dev' | sudo tee -a /etc/modules
    sudo modprobe 'i2c-dev'

##### Test if i2c is correctly enabled:

    i2cdetect -l

The output should list your I2C channel. "i2c-1" in my case.

    i2c-1   i2c             20804000.i2c                            I2C adapter

##### Scan I2C bus for connected devices
Channel 1 in my case

    i2cdetect -y 1

In case of an SmartPi connected to an RPI3, the output should look like this:

         0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
    00:          -- -- -- -- -- -- -- -- -- -- -- -- --
    10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    30: -- -- -- -- -- -- -- -- 38 -- -- -- -- -- -- --
    40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    50: -- 51 -- -- -- -- -- -- -- -- -- -- -- -- -- --
    60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    70: -- -- -- -- -- -- -- --

##### Remove old go version

    sudo apt-get remove golang
    sudo apt-get autoremove

##### Install go
Download the archive and extract it into /usr/local, creating a Go tree in /usr/local/go.
Currently version 1.18.2 is up to date. You may need to adapt the filename according to latest version.

    cd /usr/local

    sudo wget https://go.dev/dl/go1.19.5.linux-arm64.tar.gz
    sudo tar -xvzf go1.19.5.linux-arm64.tar.gz
    echo 'PATH="/usr/local/go/bin:${PATH}"' | sudo tee -a /etc/profile




In order for the `${PATH}` to be updated, you will need to logout.

Create a directory to contain your Go workspace, for example `${HOME}/go`,
and set the GOPATH environment variable to point to that location.

    mkdir "${HOME}/go"
    export GOPATH="${HOME}/go"

##### Building source

    cd ~
    git clone github.com:nDenerserve/SmartPi.git
    cd ~/SmartPi
    make

NOTE: Executables files are located in the bin directory



## Change Log

### 11/28/11/16
 * Added MQTT Client
 * producecounter and consumecounter files make use of Databasedir -> co-located to rrd database
 * fixed "}" compilation issue
 * Added this readme.md

### 02/10/17
 * changed from rrdtool to sqlite3
 * added csv-export
 * changed from Bootstrap to Angular Material
 * change datelayout in API to RFC3339
 * fixed errors in datehandling
 * added week consumption

 ### 05/24/22
 * added Influxdb-support
 * added calibration possibilities
 * added modbus-server
 
