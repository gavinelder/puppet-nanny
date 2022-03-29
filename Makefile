all: build

.PHONY: build

ifndef ($(GOPATH))
	GOPATH = $(HOME)/go
endif

PATH := $(GOPATH)/bin:$(PATH)
VERSION = $(shell git describe --tags --always --dirty)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
REVISION = $(shell git rev-parse HEAD)
REVSHORT = $(shell git rev-parse --short HEAD)
USER = $(shell whoami)
APP_NAME = puppet-nanny
PKGDIR_TMP = ${TMPDIR}golang

ifneq ($(OS), Windows_NT)
	CURRENT_PLATFORM = linux
	# If on macOS, set the shell to bash explicitly
	ifeq ($(shell uname), Darwin)
		SHELL := /bin/bash
		CURRENT_PLATFORM = darwin
	endif
 	# To populate version metadata, we use unix tools to get certain data
	GOVERSION = $(shell go version | awk '{print $$3}')
	NOW	= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
else
	CURRENT_PLATFORM = windows

	# To populate version metadata, we use windows tools to get the certain data
	GOVERSION_CMD = "(go version).Split()[2]"
	GOVERSION = $(shell powershell $(GOVERSION_CMD))
	NOW	= $(shell powershell Get-Date -format s)
endif

BUILD_VERSION = "\
	-X github.com/gavinelder/puppet-nanny/version.appName=${APP_NAME} \
	-X github.com/gavinelder/puppet-nanny/version.version=${VERSION} \
	-X github.com/gavinelder/puppet-nanny/version.branch=${BRANCH} \
	-X github.com/gavinelder/puppet-nanny/version.buildUser=${USER} \
	-X github.com/gavinelder/puppet-nanny/version.buildDate=${NOW} \
	-X github.com/gavinelder/puppet-nanny/version.revision=${REVISION} \
	-X github.com/gavinelder/puppet-nanny/version.goVersion=${GOVERSION}"

define HELP_TEXT

  Makefile commands

	make deps         - Install dependent programs and libraries
	make clean        - Delete all build artifacts

	make build        - Build the code

	make test         - Run the Go tests
	make lint         - Run the Go linters

endef

help:
	$(info $(HELP_TEXT))

deps:
	go get -u github.com/golang/dep/...
	go get -u github.com/golang/lint/golint
	dep ensure -vendor-only -v

clean:
	rm -rf build/
	rm -f *.zip
	rm -rf ${PKGDIR_TMP}_darwin
	rm -rf ${PKGDIR_TMP}_linux
	rm -rf ${PKGDIR_TMP}_windows

.pre-build:
	mkdir -p build/windows
	mkdir -p build/linux
	mkdir -p build/Darwin


build: .pre-build
	GOOS=windows GOARCH=amd64 go build -o build/windows/${APP_NAME}_win.exe -pkgdir ${PKGDIR_TMP}_windows  ./main.go
	GOOS=linux GOARCH=amd64 go build -o build/linux/${APP_NAME}_linux -pkgdir ${PKGDIR_TMP}_linux  ./main.go
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/${APP_NAME}_darwin_amd64 -pkgdir ${PKGDIR_TMP}_darwin ./main.go
	GOOS=darwin GOARCH=arm64 go build -o build/darwin/${APP_NAME}_darwin_arm64 -pkgdir ${PKGDIR_TMP}_darwin ./main.go
	/usr/bin/lipo -create -output build/darwin/${APP_NAME}_darwin  build/darwin/${APP_NAME}_darwin_amd64  build/darwin/${APP_NAME}_darwin_arm64 

test:
	go test -cover -race -v $(shell go list ./... | grep -v /vendor/)

lint:
	@if gofmt -l . | egrep -v ^vendor/ | grep .go; then \
	  echo "^- Repo contains improperly formatted go files; run gofmt -w *.go" && exit 1; \
	  else echo "All .go files formatted correctly"; fi
	go vet ./...
	# Bandaid until https://github.com/golang/lint/pull/325 is merged
	golint -set_exit_status `go list ./... | grep -v /vendor/`
