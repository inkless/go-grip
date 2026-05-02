BIN_DIR  := bin
BIN_NAME := grip
GOPATH   ?= $(shell go env GOPATH)

.PHONY: all build install run test lint format vendor compile clean help

all: format lint test build

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) main.go

install:
	go build -o $(GOPATH)/bin/$(BIN_NAME) main.go

run:
	go run main.go $(ARGS)

test:
	go test ./...

lint:
	golangci-lint run

format:
	go fmt ./...

vendor:
	go mod vendor

compile:
	GOOS=darwin  GOARCH=amd64 go build -o $(BIN_DIR)/$(BIN_NAME)-darwin-amd64      main.go
	GOOS=darwin  GOARCH=arm64 go build -o $(BIN_DIR)/$(BIN_NAME)-darwin-arm64      main.go
	GOOS=linux   GOARCH=amd64 go build -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64       main.go
	GOOS=linux   GOARCH=arm64 go build -o $(BIN_DIR)/$(BIN_NAME)-linux-arm64       main.go
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(BIN_NAME)-windows-arm64.exe main.go

clean:
	rm -rf $(BIN_DIR)

help:
	@echo "Targets: all build install run test lint format vendor compile clean"
	@echo "Run with args: make run ARGS=\"--help\""
