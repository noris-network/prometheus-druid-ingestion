SHELL = /usr/bin/env bash
WORK_DIR = $(shell pwd)
BUILD_DIR = $(WORK_DIR)/bin
BINARY_NAME = generate-ingestion
OUTPUT = $(BUILD_DIR)/$(BINARY_NAME)
GO = $(shell which go)
GOARCH ?= amd64
GOOS ?= linux
DOCKER_IMAGE_NAME = noris-network/prometheus-kafka-druid-ingestion
DOCKER_IMAGE_TAG = v0.1.0

.PHONY: help
## help: prints this help message
help:
	@echo -e "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: lint
## lint: runs golint
lint:
	golint

.PHONY: test
## test: runs unit tests
test:
	$(GO) test -v ./...

.PHONY: build
## build: builds Go binary
build:
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO) build -o $(OUTPUT) ./cmd

.PHONY: build/docker
## build/docker: builds Docker iamge
build/docker:
	 docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
