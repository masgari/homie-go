.PHONY: build
all: build

BUILD_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
#FALGS :=-ldflags="-s -w"
FALGS := ""

build:
	go build $(FLAGS) -o $(GOPATH)/bin/homie-sysinfo main.go

run:
	go run main.go

test:
	go test ./...

clean:	
	rm $(GOPATH)/bin/homie-sysinfo