SHELL=/usr/bin/env bash -o pipefail

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif

CONTAINER_CLI ?= docker

BIN_DIR ?= $(shell pwd)/bin
# auto add modelxd and modelxdl
SERVER_NAME ?= modelx

IMAGE ?= ghcr.io/kubeservice-stack/$(SERVER_NAME)
DOCKERFILE ?= ./hack/build/Dockerfile
TAG ?= $(shell git rev-parse --short HEAD)

BUILD_DATE=$(shell date +"%Y%m%d-%T")
GIT_VERSION=$(shell git describe --tags --dirty --abbrev=0 2>/dev/null || git symbolic-ref --short HEAD)
ifndef CI
	BUILD_BRANCH?=$(shell git branch --show-current)
	BUILD_REVISION?=$(shell git rev-parse --short HEAD)
else
	BUILD_BRANCH=$(GITHUB_REF:refs/heads/%=%)
	BUILD_REVISION=$(GITHUB_SHA)
endif

VERSION?=$(shell echo "${GIT_VERSION}" | sed -e 's/^v//')

TOOLS_BIN_DIR ?= $(shell pwd)/tmp/bin
export PATH := $(TOOLS_BIN_DIR):$(PATH)

# tools
GOLANGCILINTER_BINARY=$(TOOLS_BIN_DIR)/golangci-lint
GOSEC_BINARY=$(TOOLS_BIN_DIR)/gosec
GOSYCLO_BINARY=$(TOOLS_BIN_DIR)/gocyclo
SWAGGO_BINARY=$(TOOLS_BIN_DIR)/swag
TOOLING=$(GOLANGCILINTER_BINARY) $(GOSEC_BINARY) $(GOSYCLO_BINARY) $(SWAGGO_BINARY)

GO_PKG=$(shell go list -m)
COMMON_PKG ?= $(GO_PKG)/pkg
COMMON_CMD ?= $(GO_PKG)/cmd
COMMON_INTERNAL ?= $(GO_PKG)/internal
# The ldflags for the go build process to set the version related data.
GO_BUILD_LDFLAGS=\
	-s \
	-w \
	-X '$(COMMON_PKG)/version.gitVersion=$(GIT_VERSION)'  \
	-X '$(COMMON_PKG)/version.buildDate=$(BUILD_DATE)' \
	-X '$(COMMON_PKG)/version.gitCommit=$(BUILD_REVISION)' \
	-X '$(COMMON_PKG)/version.defaultVersion=$(VERSION)'

GO_BUILD_RECIPE=\
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=0 \
	go build -ldflags="$(GO_BUILD_LDFLAGS)"

pkgs = $(shell go list ./... | grep -v /test/ | grep -v /vendor/)

.PHONY: all
all: format test build e2e # format test build e2e type.

.PHONY: clean
clean: # Remove all files and directories ignored by git.
	git clean -Xfd .
	rm -rf ${TOOLS_BIN_DIR}
	rm -rf ${BIN_DIR}

##############
# Formatting #
##############

.PHONY: format
format: go-fmt go-vet gocyclo golangci-lint # make all go-fmt go-vet gocyclo golangci-lint format type.

.PHONY: go-fmt # gofmt rewrite to go file
go-fmt:
	gofmt -s -w ./pkg ./cmd ./internal

.PHONY: go-vet 
go-vet: # go vet
	go vet -stdmethods=false $(pkgs)

.PHONY: gocyclo
gocyclo: $(GOSYCLO_BINARY) # gocyclo
	$(GOSYCLO_BINARY) -top 20 -avg -ignore "_test|test/|vendor/" .

.PHONY: golangci-lint
golangci-lint: $(GOLANGCILINTER_BINARY)  # golangci-lint  
	$(GOLANGCILINTER_BINARY) run -v

.PHONY: swag
swag: $(SWAGGO_BINARY) # swag
	$(SWAGGO_BINARY) init -g ./cmd/$(SERVER_NAME)d/$(SERVER_NAME)d.go

###########
# Testing #
###########

.PHONY: test
test: test-unit test-coverage # go test for test-unit test-coverage

.PHONY: test-unit
test-unit: # go test unittest cases
	go test -short $(pkgs) -count=1 -v

.PHONY: test-coverage
test-coverage: # go test unittest coverage
	go test -coverprofile=coverage.out -covermode count $(pkgs)
	go tool cover -func coverage.out

###########
# E2e #
###########

.PHONY: e2e
e2e: KUBECONFIG?=$(HOME)/.kube/config
e2e: # go test e2e cases
	go test -mod=readonly -timeout 120m -v ./test/e2e/ --kubeconfig=$(KUBECONFIG) --test-image=$(IMAGE):$(TAG) -count=1

############
# Security #
############

.PHONY: go-sec
go-sec: $(GOSEC_BINARY) # go security
	$(GOSEC_BINARY) $(pkgs)

############
# Building #
############

.PHONY: build
build: build-modelx build-modelxd build-modelxdl modelxd-build-image modelxdl-build-image # go build both binary and docker image

.PHONY: modelxd-build-image # go build output docker image
modelxd-build-image: GOOS := linux # Overriding GOOS value for docker image build
modelxd-build-image: 
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -f $(DOCKERFILE) -t $(IMAGE)d:$(TAG) .

.PHONY: modelxdl-build-image # go build output docker image
modelxdl-build-image: GOOS := linux # Overriding GOOS value for docker image build
modelxdl-build-image: 
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -f $(DOCKERFILE).dl -t $(IMAGE)dl:$(TAG) .

.PHONY: build-modelx
build-modelx: # go build output binary
	$(GO_BUILD_RECIPE) -o ${BIN_DIR}/$(SERVER_NAME) cmd/$(SERVER_NAME)/$(SERVER_NAME).go

.PHONY: build-modelxd
build-modelxd: # go build output binary
	$(GO_BUILD_RECIPE) -o ${BIN_DIR}/$(SERVER_NAME)d cmd/$(SERVER_NAME)d/$(SERVER_NAME)d.go

.PHONY: build-modelxdl
build-modelxdl: # go build output binary
	$(GO_BUILD_RECIPE) -o ${BIN_DIR}/$(SERVER_NAME)dl cmd/$(SERVER_NAME)dl/$(SERVER_NAME)dl.go

define asserts
	@echo "Building ${SERVER_NAME}-${1}-${2}"
	@GOOS=${1} GOARCH=${2} CGO_ENABLED=0 go build -ldflags="$(GO_BUILD_LDFLAGS)" -o ${BIN_DIR}/$(SERVER_NAME)-$(1)-$(2) cmd/$(SERVER_NAME)/$(SERVER_NAME).go
endef

.PHONY: build-assets
build-assets: # go build muti arch for github assets
	mkdir -p $(BIN_DIR)
	$(call asserts,linux,amd64)
	$(call asserts,linux,arm64)
	$(call asserts,darwin,amd64)
	$(call asserts,darwin,arm64)
	$(call asserts,windows,amd64)

$(TOOLS_BIN_DIR): 
	mkdir -p $(TOOLS_BIN_DIR)

$(TOOLING): $(TOOLS_BIN_DIR) # Install tools
	@echo Installing tools from scripts/tools.go
	@cat scripts/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(TOOLS_BIN_DIR) xargs -tI % go install -mod=readonly -modfile=scripts/go.mod %

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#' $(MAKEFILE_LIST) | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

