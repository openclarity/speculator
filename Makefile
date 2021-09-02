# Go parameters
	GOCMD=go
	GOBUILD=$(GOCMD) build
	GOCLEAN=$(GOCMD) clean
	GOTEST=$(GOCMD) test
	GOGET=$(GOCMD) get
	CLI_BINARY_NAME=cli
	MAKE_BUILD_COMMAND ?= build

all: test build

build: build_cli

.PHONY: build_cli
build_cli:
	$(GOBUILD) -v -o ./bin/$(CLI_BINARY_NAME) ./cmd/$(CLI_BINARY_NAME)

.PHONY: test
test:
	$(GOTEST) -v `go list ./pkg/...` -coverprofile=coverage.out
	# go tool cover -html=coverage.out

clean:
	$(GOCLEAN)
	rm -f ./bin/$(CLI_BINARY_NAME)
