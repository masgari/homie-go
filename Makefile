.PHONY: build
all: build

BUILD_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
#FALGS :=-ldflags="-s -w"
FALGS := ""

build:
	GO111MODULE=on go build $(FLAGS) -o $(GOPATH)/bin/homie-basic-example examples/basic/main.go

run:
	GO111MODULE=on go run main.go
	 	
test:
	GO111MODULE=on go test -v -timeout 10s -coverprofile=/tmp/homie-test-coverage ./...

clean:	
	rm -fr $(GOPATH)/bin/homie-basic-example
	rm -fr /tmp/homie-test-coverage