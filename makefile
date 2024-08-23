# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
CACHECLEAN := $(GOCMD) clean --testcache
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOLINT := golangci-lint

# Name of the executable
BINARY_NAME := midi-file-server

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

.PHONY: build
build:
	$(GOBUILD) .

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: test
test:
	$(CACHECLEAN) && $(GOTEST) -v ./...

.PHONY: get
get:
	$(GOGET) -v ./...

.PHONY: all
all: clean get build test lint
