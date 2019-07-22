.PHONY: build

CWD = $(shell pwd)
OUT := ${PWD}/bin/logbay
VERSION := $(shell git describe --always --long)
LD_FLAGS := -X main.BUILD_VERSION=${VERSION}

default: build

build: clean
	@echo Building...
	go build -o ${OUT} -ldflags '${LD_FLAGS}' .

clean:
	 @echo Cleaning up...
	 go clean -cache

run: build
	./bin/logbay
