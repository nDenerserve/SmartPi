BINARY_READOUT=smartpireadout
BINARY_SMARTPIDC=smartpidcreadout
BINARY_SERVER=smartpiserver
BINARY_FTPUPLOAD=smartpiftpupload
BINARY_MODBUSSERVER=smartpimodbusserver
BINARY_EMETERSERVER=smartpiemeterserver
BINARY_ETEMPERATUREREADOUT=smartpietemperaturereadout
BINARY_LORAWAN=smartpilorawan

#VERSION=0.3.7
VERS = $(shell git describe --always --long --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
LBITS := $(shell getconf LONG_BIT)

BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOFMT?=$(GO)fmt
GOBUILD=$(GO) build
GOINSTALL=$(GO) install
GOCLEAN=$(GO) clean
GOGET=$(GO) get

pkgs = $(shell $(GO) list ./...)


ifdef VERSION
		VERSION := $(VERSION)
else
		VERSION := $(VERS)_${BUILD_TIME}
endif

$(info    VERSION is $(VERSION))


all: format makedir get buildsmartpireadout buildsmartpidcreadout buildsmartpiserver buildsmartpiftpupload buildsmartpimodbusserver buildsmartpiemeterserver buildsmartpietemperaturereadout buildsmartpielorawan


makedir:
	@echo "start building tree..."
	@if [ ! -d $(BUILDPATH)/bin ] ; then mkdir -p $(BUILDPATH)/bin ; fi
	@if [ ! -d $(BUILDPATH)/pkg ] ; then mkdir -p $(BUILDPATH)/pkg ; fi

get:

style:
	@echo "checking code style"
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'


format:
	@echo "formatting code..."
	@$(GO) fmt $(pkgs)

buildsmartpireadout:
	@echo "start building smartpireadout..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_READOUT) -ldflags="-X main.appVersion=${VERSION}" readout/*.go
	@echo "building smartpireadout done"

buildsmartpidcreadout:
	@echo "start building smartpidcreadout..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_SMARTPIDC) -ldflags="-X main.appVersion=${VERSION}" dcreadout/readout.go
	@echo "building smartpidcreadout done"

buildsmartpiserver:
	@echo "start building smartpiserver..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_SERVER) -ldflags="-X main.appVersion=${VERSION}" server/server.go
	@echo "building smartpiserver done"

buildsmartpiftpupload:
	@echo "start building smartpiftpupload..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_FTPUPLOAD) -ldflags="-X main.appVersion=${VERSION}" ftpupload/ftpupload.go
	@echo "building smartpiftpupload done"

buildsmartpimodbusserver:
	@echo "start building smartpimodbusserver..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_MODBUSSERVER) -ldflags="-X main.appVersion=${VERSION}" services/modbus/modbusserver.go
	@echo "building smartpimodbusserver done"

buildsmartpiemeterserver:
	@echo "start building smartpiemeterserver..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_EMETERSERVER) -ldflags="-X main.appVersion=${VERSION}" services/emeter/emeter.go
	@echo "building smartpiemeterserver done"

buildsmartpietemperaturereadout:
	@echo "start building smartpietemperaturereadout..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_ETEMPERATUREREADOUT) -ldflags="-X main.appVersion=${VERSION}" modules/etemperature/etemperature.go
	@echo "building smartpietemperaturereadout done"

buildsmartpielorawan:
	@echo "start building smartpilorawan..."
	$(GOBUILD) -o ../bin$(LBITS)/$(BINARY_LORAWAN) -ldflags="-X main.appVersion=${VERSION}" modules/lorawan/lorawan.go
	@echo "building smartpilorawan done"

install:
	@echo install

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@for m in $$($(GO) list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		$(GO) get $$m; \
	done
	$(GO) mod tidy

