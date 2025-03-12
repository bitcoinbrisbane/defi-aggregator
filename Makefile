# Makefile for DeFi Aggregator

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
BINARY_NAME=defi-aggregator
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/main

# Build flags
LDFLAGS=-ldflags "-s -w"

all: deps fmt vet test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

fmt:
	$(GOFMT) ./...

vet:
	$(GOVET) ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run: build
	./$(BINARY_NAME)

deps:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Cross-platform build
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) $(MAIN_PATH)

docker-build:
	docker build -t $(BINARY_NAME):latest .

docker-run:
	docker run --rm --env-file .env -p 8080:8080 $(BINARY_NAME):latest

# Install the application to $GOPATH/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)

# # Systemd service setup
# systemd-install: build
# 	@echo "Installing systemd service..."
# 	cp $(BINARY_NAME) /usr/local/bin/
# 	cp scripts/defi-aggregator.service /etc/systemd/system/
# 	systemctl daemon-reload
# 	systemctl enable defi-aggregator.service
# 	systemctl start defi-aggregator.service
# 	@echo "Service installed and started."

# Systemd service status
systemd-status:
	systemctl status defi-aggregator.service

.PHONY: all build test fmt vet clean run deps build-linux docker-build docker-run install systemd-install systemd-status