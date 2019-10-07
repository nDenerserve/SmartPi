# SmartPi Energy Monitor 

[![CircleCI](https://circleci.com/gh/nDenerserve/SmartPi.svg?style=svg)](https://circleci.com/gh/nDenerserve/SmartPi)

[SmartPi open source](http://www.emanager.eu/en/products/smartpi).

## Forum
https://forum.enerserve.eu

## Installation
Download Raspbian Jessie Lite from https://www.raspberrypi.org/downloads/raspbian/ and copy it on your SD card.
Alternatively, you may download EmonSD, a pre-built SD card image for Raspberry Pi running as an emonPi/emonBase
Download https://github.com/openenergymonitor/emonpi/wiki/emonSD-pre-built-SD-card-Download-&-Change-Log

##### Update packet list and update packages

    sudo apt-get update
    sudo apt-get upgrade

##### Install additional packages.

    $ sudo apt-get install sqlite3 git i2c-tools avahi-daemon

For building SmartPi tools, additional packages are required.

    sudo apt-get install libpam0g-dev

##### Enable i2c kernel module

`i2c-dev` is required for communicating with the SmartPi.

To check to see if the module is loaded:

    sudo lsmod | grep i2c

This should return something like this:

    i2c_dev                 5859  0
    i2c_bcm2708             4834  0

If the module is not listed, add it to the system.

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
    50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    70: -- -- -- -- -- -- -- --

##### Remove old go version

    sudo apt-get remove golang
    sudo apt-get autoremove

##### Install go
Download the archive and extract it into /usr/local, creating a Go tree in /usr/local/go.
Currently version 1.8.1 is up to date. You may need to adapt the filename according to latest version.

    cd /usr/local
    curl -s https://storage.googleapis.com/golang/go1.8.1.linux-armv6l.tar.gz | sudo tar -xvz
    echo 'PATH="/usr/local/go/bin:${PATH}"' | sudo tee -a /etc/profile

In order for the `${PATH}` to be updated, you will need to logout.

Create a directory to contain your Go workspace, for example `${HOME}/go`,
and set the GOPATH environment variable to point to that location.

    mkdir "${HOME}/go"
    export GOPATH="${HOME}/go"

##### Building source

    go get -v github.com/nDenerserve/SmartPi/src/smartpi
    cd ${GOPATH-$HOME/go}/src/github.com/nDenerserve/SmartPi
    make

NOTE: If you need to build from a fork, you will have to symlink your fork into `${GOPATH-$HOME/go}/src/github.com/nDenerserve/` to make golang dependencies work correctly.

##### emonSD Specifics
The emonSD provides a read-only file system for most areas.
You can mount the file system with write privileges by ```rpi-rw``` in order to edit files.
```rpi-ro``` reverts write privileges again. ```/home/pi/data``` is always mounted with write access.
This is the location where you need to place your SmartPi database and current values file.
In SmartPi's configif file ```/etc/smartpi``` you may set the following settings to move all files SmartPi is writing into.

    [database]
    dir="/home/pi/data/smartpi"
    
    [device]
    shared_dir="/home/pi/data/smartpi"



## Change Log

### 11/28/11/16
 * Added MQTT Client
 * producecounter and consumecounter files make use of Databasedir -> co-located to rrd database
 * fixed "}" compilation issue
 * Added this readme.md
 
## ToDo's:
   * Logging
   * Improved error handling

### 02/10/17
 * changed from rrdtool to sqlite3
 * added csv-export
 * changed from Bootstrap to Angular Material
 * change datelayout in API to RFC3339
 * fixed errors in datehandling
 * added week consumption
