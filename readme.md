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

##### Install InfluxDB 2

    wget -q https://repos.influxdata.com/influxdata-archive_compat.key

    echo '393e8779c89ac8d958f81f942f9ad7fb82a25e133faddaf92e15b16e6ac9ce4c influxdata-archive_compat.key' | sha256sum -c && cat influxdata-archive_compat.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg > /dev/null

    echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list

    sudo apt update && sudo apt install influxdb2
    sudo systemctl start influxdb

Check if influxdb is running

    sudo systemctl status influxdb

###### Create InfluxDB user and tables
Go to http://<<ip-address of smartpi>>:8086
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx01.jpg?raw=true)

Create user with name smartpi and password smart4pi. Use **smartpi** for Initial Organization Name and **meteringdata** for Initial Bucket Name
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx02.jpg?raw=true)

Login and add an addition Bucket called **fastmeasurement**
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx03.jpg?raw=true)

Create an API-Key and save it for later use. You have to adde the API-Key later in the SmartPiconfiguration
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx04.jpg?raw=true)

##### Update packet list and update packages
    
    wget -q -O - https://packages.grafana.com/gpg.key | gpg --dearmor | sudo tee -a /etc/apt/trusted.gpg.d/grafana-archive_compat.gpg && clear
    echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list

    sudo apt update
    sudo apt upgrade


##### Install additional packages

    sudo apt-get install -y grafana sqlite3 libpam0g i2c-tools watchdog
    sudo systemctl start grafana-server

Check if grafana is running

    sudo systemctl status grafana-server


##### Install nodered

    bash <(curl -sL https://raw.githubusercontent.com/node-red/linux-installers/master/deb/update-nodejs-and-nodered)

    sudo systemctl enable nodered
    sudo systemctl start nodered


##### Create tmpfs in /etc/fstab

    echo "tmpfs /var/tmp/smartpi tmpfs nodev,nosuid,size=20M 0 0" | sudo tee -a /etc/fstab
    
For secure 24/7 operation, we recommend that you also create a tmpf for the log and tmp files.

    echo "tmpfs /var/log tmpfs defaults,noatime,mode=1777,size=50M 0 0" | sudo tee -a /etc/fstab
    echo "tmpfs /var/tmp tmpfs defaults,noatime,mode=1777,size=30M 0 0" | sudo tee -a /etc/fstab
    echo "tmpfs /tmp tmpfs defaults,noatime,mode=1777,size=20M 0 0" | sudo tee -a /etc/fstab
    
If you want to compile yourself, increase the value for tmp to 200M.
    

##### Optimize the logfile (for bullseye):

    sudo nano /etc/logrotate.d/rsyslog

And add size 10M and change from weekly to daily.
The file looks like:
                                                                                                                                                                                                                                                
    /var/log/syslog
    /var/log/mail.info
    /var/log/mail.warn
    /var/log/mail.err
    /var/log/mail.log
    /var/log/daemon.log
    /var/log/kern.log
    /var/log/auth.log
    /var/log/user.log
    /var/log/lpr.log
    /var/log/cron.log
    /var/log/debug
    /var/log/messages
    {
            rotate 4
            size 10M
            daily
            missingok
            notifempty
            compress
            delaycompress
            sharedscripts
            postrotate
                    /usr/lib/rsyslog/rsyslog-rotate
            endscript
    }

##### Optimize the logfile (for bookworm):

    sudo nano /etc/logrotate.conf
    
The file should be looks like this:
   
    daily
    rotate 7
    create
    #dateext
    compress
    include /etc/logrotate.d


##### Add rescue IP

    sudo nano /etc/dhcpcd.enter-hook
    
Insert the following content:
                                                                                                      
    # File /etc/dhcpcd.enter-hook
    # Assign an IP alias to the eth0 interface.
    if [ "$interface" = "eth0" ]; then
        case $reason in
            PREINIT)
                # Other reasons are: NOCARRIER|CARRIER|BOUND
                /usr/sbin/ip addr add 169.254.3.10/16 dev eth0 label eth0:0 || true
                ;;
        esac
    fi


##### Reboot
    
    sudo reboot
    

    
##### Enable i2c kernel module

By default, the I2C bus is disabled on the raspberry Pi. It must be activated for the operation of the SmartPi.
There are 2 ways to activate tccce I2C bus.

1. With the help of raspi-config:

    sudo raspi-config

Select 3 Interface Options --> I5 I2C --> Yes

2. Manually:

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
    
    
##### Activate Hardware RTC

You can add support for the RTC by adding a device tree overlay. Run

    sudo nano /boot/config.txt
    
to edit the pi configuration and add the RTC chip of the SmartPi

    dtoverlay=i2c-rtc,pcf8523
    
to the end of the file.

    sudo reboot
    
Disable the "fake hwclock" which interferes with the 'real' hwclock

    sudo apt-get -y remove fake-hwclock
    sudo update-rc.d -f fake-hwclock remove
    sudo systemctl disable fake-hwclock
Run

    sudo nano /lib/udev/hwclock-set 
    
and comment out these three lines:

    #if [ -e /run/systemd/system ] ; then
    # exit 0
    #fi

And also comment out:

    #/sbin/hwclock --rtc=$dev --systz
    #/sbin/hwclock --rtc=$dev --hctosys

Use
    
    sudo hwclock -w
    
to write the time to the RTC and

    sudo hwclock -r

to read the time from the RTC.

##### If you would like to use our RS485- or Lora-Module:
Open

    sudo nano /boot/config.txt
and add:

    dtoverlay=sc16is752-i2c,int_pin=24,addr=0x4D,xtal=14745600

    

##### Remove old go version

    sudo apt-get remove golang
    sudo apt-get autoremove


##### Install go
Download the archive and extract it into /usr/local, creating a Go tree in /usr/local/go.
Currently version 1.20.5 is up to date. You may need to adapt the filename according to latest version.

    cd /usr/local

    sudo wget https://go.dev/dl/go1.21.4.linux-arm64.tar.gz
    sudo tar -xvzf go1.21.4.linux-arm64.tar.gz
    sudo rm go1.21.4.linux-arm64.tar.gz
    echo 'PATH="/usr/local/go/bin:${PATH}"' | sudo tee -a /etc/profile


In order for the `${PATH}` to be updated, you will need to logout.

##### Building source

    cd ~
    git clone github.com:nDenerserve/SmartPi.git
    cd ~/SmartPi/src
    make

NOTE: Executables files are located in the bin directory


##### Add Api-Key to config-file

Add the saved API-Key into the /etc/smartpi
    influxapitoken      = <API-Key>
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx06.jpg?raw=true)


Or you can add it later via webgui:
![Login Screen](https://github.com/nDenerserve/SmartPi/blob/master/img/influx05.jpg?raw=true)

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
 
 ### 02/14/23
 * changed to InfluxDB2
 * added support for nD-enerserve Digital-Out-Module (https://shop.enerserve.eu/e.digital-OUT/100054)
 * removed support for UMTS-Sticks

 ### 06/19/23
 * added support for SMA energy meter (test)
 * SmartPi can replace the SMA energy meter
