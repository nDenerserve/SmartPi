BINARY_READOUT=smartpireadout
BINARY_SERVER=smartpiserver
BINARY_FTPUPLOAD=smartpiftpupload

#VERSION=0.3.7
VERSION := $(shell git describe --always --long --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOBUILD=$(GO) build
GOINSTALL=$(GO) install
GOCLEAN=$(GO) clean
GOGET=$(GO) get

pkgs = $(shell $(GO) list ./src/...)

all: format makedir get buildsmartpireadout buildsmartpiserver buildsmartpiftpupload
#all: makedir get buildsmartpireadout

makedir:
	@echo "start building tree..."
	@if [ ! -d $(BUILDPATH)/bin ] ; then mkdir -p $(BUILDPATH)/bin ; fi
	@if [ ! -d $(BUILDPATH)/pkg ] ; then mkdir -p $(BUILDPATH)/pkg ; fi

get:

format:
	@echo "formatting code..."
	@$(GO) fmt $(pkgs)

buildsmartpireadout:
	@echo "start building smartpireadout..."
	$(GOBUILD) -o bin/$(BINARY_READOUT) -ldflags="-X main.appVersion=${VERSION}_${BUILD_TIME}" src/readout/*.go
	@echo "building smartpireadout done"

buildsmartpiserver:
	@echo "start building smartpiserver..."
	$(GOBUILD) -o bin/$(BINARY_SERVER) -ldflags="-X main.appVersion=${VERSION}_${BUILD_TIME}" src/main/server.go
	@echo "building smartpiserver done"

buildsmartpiftpupload:
	@echo "start building smartpiftpupload..."
	$(GOBUILD) -o bin/$(BINARY_FTPUPLOAD) -ldflags="-X main.appVersion=${VERSION}_${BUILD_TIME}" src/main/ftpupload.go
	@echo "building smartpiftpupload done"

install:
	@echo install

#	if [ `pidof smartpireadout` > "0" ]
#		then killall smartpireadout
#	fi
#
#	if [ `pidof smartpiserver` > "0" ]
#		then killall smartpiserver
#	fi
#
#	if [ `pidof smartpiftpupload` > "0" ]
#		then killall smartpiftpupload
#	fi
#
#	if [ ! -d /var/tmp/smartpi ];
#		then mkdir -p /var/tmp/smartpi
#	fi
#
#	cat /etc/fstab | grep smartpi | grep -v "#"
#	if [ $? = 0 ];
#		then
#			:
#		else
#		echo "tmpfs /var/tmp/smartpi tmpfs nodev,nosuid,size=100M 0 0" >> /etc/fstab
#	fi
#
#
#	cat /etc/modules | grep i2c-bcm2708 | grep -v "#"
#	if [ $? = 0 ];
#		then
#			:
#		else
#		echo "i2c-bcm2708" >> /etc/modules
#	fi
#
#	cat /etc/modules | grep i2c-dev | grep -v "#"
#	if [ $? = 0 ];
#		then
#			:
#		else
#		echo "i2c-dev" >> /etc/modules
#	fi
#
#	cat /boot/config.txt | grep dtparam=i2c1=on | grep -v "#"
#	if [ $? = 0 ];
#		then
#			:
#		else
#			echo "dtparam=i2c1=on" >> /boot/config.txt
#	fi
#
#	cat /boot/config.txt | grep dtparam=i2c_arm=on | grep -v "#"
#	if [ $? = 0 ];
#		then
#			:
#		else
#			echo "dtparam=i2c_arm=on" >> /boot/config.txt
#	fi
#
#	if [ ! -d /var/smartpi/db ];
#		then mkdir -p /var/smartpi/db
#	fi
#
#	if [ ! -d /var/smartpi/www ];
#		then mkdir -p /var/smartpi/www
#	fi
#
#	cp $(BUILDPATH)/bin/* /usr/local/bin
#	cp $(BUILDPATH)/config/smartpi.conf /etc/smartpi
#	cp -r $(BUILDPATH)/public_html/* /var/smartpi/www
