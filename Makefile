# Binary name
BINARY_NAME=gitid
VERSION ?= $(shell git describe --tags --always --dirty)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -s -w"
CGO_ENABLED=0

# Supported platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

# Output directories
RELEASE_DIR=release
BUILD_DIR=build

.PHONY: all build clean test release

all: clean build

build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS)

build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	$(GOBUILD) -tags netgo -ldflags '-extldflags "-static" $(LDFLAGS)' \
	-o $(BUILD_DIR)/$(BINARY_NAME)-static

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(RELEASE_DIR)

test:
	$(GOTEST) -v ./...

release: clean
	mkdir -p $(RELEASE_DIR) $(BUILD_DIR)
	# Build for each platform/architecture
	$(foreach GOOS, $(PLATFORMS), \
		$(foreach GOARCH, $(ARCHITECTURES), \
			$(eval EXTENSION := $(if $(filter windows,$(GOOS)),.exe,)) \
			GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) \
			$(GOBUILD) $(LDFLAGS) \
			-o $(BUILD_DIR)/$(BINARY_NAME)_$(GOOS)_$(GOARCH)$(EXTENSION); \
		) \
	)

	# Build musl version for linux_amd64
	CC=musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	$(GOBUILD) -tags musl \
	-ldflags "-linkmode external -extldflags '-static' -X main.version=${VERSION} -s -w" \
	-o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64_musl

	# Create Linux packages for each architecture
	$(foreach ARCH, $(ARCHITECTURES), \
		ARCH=$(ARCH) VERSION=$(VERSION) nfpm package \
			-f nfpm.yaml \
			-p deb \
			-t $(RELEASE_DIR) && \
		ARCH=$(ARCH) VERSION=$(VERSION) nfpm package \
			-f nfpm.yaml \
			-p rpm \
			-t $(RELEASE_DIR); \
	)

.DEFAULT_GOAL := all
