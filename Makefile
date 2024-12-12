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
SERVER_NAME ?= modelx

IMAGE ?= ghcr.io/kubeservice-stack/$(SERVER_NAME)
DOCKERFILE ?= ./hack/build/Dockerfile
TAG?=$(shell git rev-parse --short HEAD)
VERSION?=$(shell cat VERSION | grep -Eo "v[0-9]+\.[0-9]+.*")

BUILD_DATE=$(shell date +"%Y%m%d-%T")
ifndef CI
	BUILD_USER?=$(USER)
	BUILD_BRANCH?=$(shell git branch --show-current)
	BUILD_REVISION?=$(shell git rev-parse --short HEAD)
else
	BUILD_USER=Action-Run-ID-$(GITHUB_RUN_ID)
	BUILD_BRANCH=$(GITHUB_REF:refs/heads/%=%)
	BUILD_REVISION=$(GITHUB_SHA)
endif

TOOLS_BIN_DIR ?= $(shell pwd)/tmp/bin
export PATH := $(TOOLS_BIN_DIR):$(PATH)

# tools
GOLANGCILINTER_BINARY=$(TOOLS_BIN_DIR)/golangci-lint
GOSEC_BINARY=$(TOOLS_BIN_DIR)/gosec
GOSYCLO_BINARY=$(TOOLS_BIN_DIR)/gocyclo
SWAGGO_BINARY=$(TOOLS_BIN_DIR)/swag
TOOLING=$(GOLANGCILINTER_BINARY) $(GOSEC_BINARY) $(GOSYCLO_BINARY) $(SWAGGO_BINARY)

GO_PKG=kubegems.io/modelx/$(SERVER_NAME)
COMMON_PKG ?= $(GO_PKG)/pkg
COMMON_CMD ?= $(GO_PKG)/cmd
COMMON_INTERNAL ?= $(GO_PKG)/internal
# The ldflags for the go build process to set the version related data.
GO_BUILD_LDFLAGS=\
	-s \
	-w \
	-X $(COMMON_PKG)/version.Revision=$(BUILD_REVISION)  \
	-X $(COMMON_PKG)/version.BuildUser=$(BUILD_USER) \
	-X $(COMMON_PKG)/version.BuildDate=$(BUILD_DATE) \
	-X $(COMMON_PKG)/version.Branch=$(BUILD_BRANCH) \
	-X $(COMMON_PKG)/version.Version=$(VERSION)

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
	$(SWAGGO_BINARY) init -g ./cmd/main.go

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
build: build-binary build-image # go build both binary and docker image

.PHONY: build-image # go build output docker image
build-image: GOOS := linux # Overriding GOOS value for docker image build
build-image: 
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -f $(DOCKERFILE) -t $(IMAGE):$(TAG) .

.PHONY: build-binary
build-binary: # go build output binary
	$(GO_BUILD_RECIPE) -o ${BIN_DIR}/$(SERVER_NAME) cmd/main.go

define asserts
	@echo "Building ${SERVER_NAME}-${1}-${2}"
	@GOOS=${1} GOARCH=${2} CGO_ENABLED=0 go build -ldflags="$(GO_BUILD_LDFLAGS)" -o ${BIN_DIR}/$(SERVER_NAME)-$(1)-$(2) cmd/main.go
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

