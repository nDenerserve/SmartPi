# Installation
Download Raspbian Jessie Lite from https://www.raspberrypi.org/downloads/raspbian/ and copy it on your SD card.
Alternatively, you may download EmonSD, a pre-built SD card image for Raspberry Pi running as an emonPi/emonBase
Download https://github.com/openenergymonitor/emonpi/wiki/emonSD-pre-built-SD-card-Download-&-Change-Log

* Expand Filesystem
* Change user (pi) password
* Advanced Options > Enable I2C
	sudo raspi-config 

##### Update packet list and update packets

    sudo apt-get update
    sudo apt-get upgrade

##### Install additional packets. Especially, "librrd-dev" is required to build SmartPi tools.

    sudo apt-get install librrd-dev rrdtool git i2c-tools avahi-daemon

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
	
##### Install go
Download the archive and extract it into /usr/local, creating a Go tree in /usr/local/go.
Currently version 1.7.3 is up to date. You may need to adapt the filenamy according to latest version.
	
    cd /usr/local
    sudo wget https://storage.googleapis.com/golang/go1.8rc2.linux-armv6l.tar.gz
    sudo tar -xvzf go1.8rc2.linux-armv6l.tar.gz
    sudo rm go1.8rc2.linux-armv6l.tar.gz
    nano /etc/profile

Paste export PATH=$PATH:/usr/local/go/bin at the end of file.
Add "/usr/local/go/bin" to the PATH environment variable.
You can do this by adding this line to your /etc/profile (for a system-wide installation) before the path is being exported.

    PATH=$PATH:/usr/local/go/bin

Create a directory to contain your workspace and SmartPi git-repo, $HOME/SmartPi in this case,
and set the GOPATH environment variable to point to that location.

    $ export GOPATH=$HOME/SmartPi

Download of SmartPi sources

    cd ~
    git clone https://github.com/nDenerserve/SmartPi.git

##### build SmartPi tools
	cd  ~/SmartPi
	make

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
