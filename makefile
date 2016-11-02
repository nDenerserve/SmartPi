BINARY_READOUT=smartpireadout
BINARY_SERVER=smartpiserver
BINARY_FTPUPLOAD=smartpiftpupload

VERSION=0.3.3
BUILD_TIME=`date +%FT%T%z`

BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOINSTALL=$(GO) install
GOCLEAN=$(GO) clean
GOGET=$(GO) get

export GOPATH=$(CURDIR)

all: makedir get buildsmartpireadout buildsmartpiserver buildsmartpiftpupload

makedir:
	@echo "start building tree..."
	@if [ ! -d $(BUILDPATH)/bin ] ; then mkdir -p $(BUILDPATH)/bin ; fi
	@if [ ! -d $(BUILDPATH)/pkg ] ; then mkdir -p $(BUILDPATH)/pkg ; fi

get:

buildsmartpireadout:
	@echo "start building smartpireadout..."
	#$(GOINSTALL) $(BINARY_READOUT)
	env CGO_ENABLED=1 GOOS=linux GOARCH=arm CC=arm-linux-gnueabi-gcc go build  -o ../smartpireadout main/readout.go
	@echo "building smartpireadout done"

buildsmartpiserver:
	@echo "start building smartpiserver..."
	$(GOINSTALL) $(BINARY_SERVER)
	@echo "building smartpiserver done"

buildsmartpiftpupload:
	@echo "start building smartpiftpupload..."
	$(GOINSTALL) $(BINARY_FTPUPLOAD)
	@echo "building smartpiftpupload done"
