SHELL = /usr/bin/env bash
WORK_DIR = $(shell pwd)
BUILD_DIR = $(WORK_DIR)/bin
BINARY_NAME = spec
OUTPUT = $(BUILD_DIR)/$(BINARY_NAME)
GO = $(shell which go)
GOARCH ?= amd64
GOOS ?= linux

.PHONY: help
## help: prints this help message
help:
	@echo -e "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: build
## build: builds Go binary
build:
	$(GO) build -o $(OUTPUT) cmd/main.go

