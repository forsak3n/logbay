export GOPATH = ${PWD}/vendor

CWD = $(shell pwd)
OUT := ${PWD}/bin/logthing
GO_FILES := $(shell find ./src -name '*.go' | grep -v /vendor/)
VERSION := $(shell git describe --always --long)
LD_FLAGS := -X main.BUILD_VERSION=${VERSION}

clean-deps:
	rm -rf vendor/*

deps: clean-deps
	go get -u -v \
	github.com/sirupsen/logrus \
	github.com/go-redis/redis \
	github.com/x-cray/logrus-prefixed-formatter \
	github.com/BurntSushi/toml \
	github.com/sevlyar/go-daemon

build: clean
	@echo ${GOPATH}
	@echo Building...
	go build -i -v -o ${OUT} -ldflags '${LD_FLAGS}' ./src/*.go

clean:
	@echo Cleaning up...
	-@rm -rf ${CWD}/dist/*
	-@rm -rf ${CWD}/bin/proton

run: build
	./bin/logthing
