####################
# Configuration
####################

.DEFAULT: lint fmt vet build
all: lint fmt vet test test-long cover build cross

PKG_NAME := skyproxy
# If true, disable optimizations and does NOT strip the binary
DEBUG ?=
# If true, "build" will produce a static binary (cross compile always produce static build regardless)
STATIC ?=
# If true, turn on verbose output for build
VERBOSE ?=
# Output prefix, defaults to local directory if not specified
PREFIX ?= $(shell pwd)
# Build tags
BUILDTAGS ?=

# List of cross compilation targets
GO_OSES ?= darwin freebsd linux windows
GO_ARCHS ?= amd64 arm

####################
# Derived properties
####################

SHELL := /bin/bash
GOLINT_BIN := $(GOPATH)/bin/golint
GODEP_BIN := $(GOPATH)/bin/godep
export GO15VENDOREXPERIMENT = 1

# Get the closest annotated tag matching the pattern, number of commits away, commit sha,
# and optionally .dirty if there are uncommited files
# Example: v0.0.1-4-geaa2e87.dirty
GIT_PROP := $(shell git describe --match 'v[0-9]*' --dirty='.dirty' --always)
# Date, example: 2015-09-02.01:03:00"
DATE=$(shell date -u +%Y-%m-%d.%H:%M:%S)
# Base version
VERSION := $(GIT_PROP)-$(DATE)

# Tooling paths
GOLINT := $(shell [ -x $(GOLINT_BIN) ] && echo $(GOLINT_BIN) || echo '')
GODEP := $(shell [ -x $(GODEP_BIN) ] && echo $(GODEP_BIN) || echo '')

# Initialize the version string in place
GO_LDFLAGS := -X `go list ./version`.Version=$(VERSION)
GO_GCFLAGS :=

# Honor debug
ifeq ($(DEBUG),true)
	# Disable function inlining and variable registerization
	GO_GCFLAGS := -gcflags "-N -l"
	# Append to the version
	VERSION := "$(VERSION)-debug"
else
	# Turn of DWARF debugging information and strip the binary otherwise
	GO_LDFLAGS := $(GO_LDFLAGS) -w -s
endif

# Honor static
ifeq ($(STATIC),true)
	# Append to the version
	VERSION := "$(VERSION)-static
	GO_LDFLAGS := $(GO_LDFLAGS) -extldflags -static
endif

# Coverage
COVERDIR = $(PREFIX)/cover
COVERPROFILE = $(COVERDIR)/cover.out
COVERMODE = count

# List
PKGS := $(shell go list -tags "$(BUILDTAGS)" ./... | grep -v "/vendor/")

ifeq ($(VERBOSE),true)
	VERBOSE := -v
endif

# Colors
NC := $(shell echo -e "\033[0m")
GREEN := $(shell echo -e "\033[1;32m")
ORANGE := $(shell echo -e "\033[1;33m")
BLUE := $(shell echo -e "\033[1;34m")
RED := $(shell echo -e "\033[1;31m")

####################
# Helpers
####################

# Formatting helpers
define title
	@echo "$(GREEN) ++++++++++++++++++"
	@echo " + $(1)"
	@echo " ++++++++++++++++++ $(ORANGE)"
endef

define failif
	@echo "$(2) $(RED)"
	@test -z "$$($(1) | tee /dev/stderr)"
	@echo "$(BLUE) Success!"
endef

# Cross builder helper
define gocross
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build -o $(PREFIX)/bin/$(1)/$(2)/$(PKG_NAME) \
		-a $(VERBOSE) -tags "static_build netgo $(BUILDTAGS)" -installsuffix netgo -ldflags "$(GO_LDFLAGS) -extldflags -static" ./cmd/$(PKG_NAME);
endef

# Coverage helper
define gocover
	go test -tags "$(BUILDTAGS)" -covermode="$(COVERMODE)" -coverprofile="$(COVERDIR)/$(subst /,-,$(1)).cover" "$(1)";
endef

.PHONY: all clean build cross test test-long lint fmt vet cover
.DELETE_ON_ERROR: $(COVERDIR) $(PREFIX)/bin

clean:
	$(call title, $@)
	@rm -Rf $(PREFIX)/bin
	@rm -Rf $(COVERDIR)
	@echo "$(BLUE) Success!"

build: $(PREFIX)/bin/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME): $(shell find . -type f -name '*.go')
	$(call title, $@)
	@go build -o $@ -tags "$(BUILDTAGS)" $(VERBOSE) -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/$(PKG_NAME)
	@echo "$(BLUE) Success!"

cross:
	$(call title, $@)
	@VERSION="$(VERSION)-static"
	@$(foreach GOARCH,$(GO_ARCHS),$(foreach GOOS,$(GO_OSES),$(call gocross,$(GOOS),$(GOARCH))))
	@echo "$(BLUE) Success!"

# Quick test
test:
	$(call title, $@)
	@go test $(VERBOSE) -test.short -tags "$(BUILDTAGS)" $(PKGS)

# Runs race detection and all tests
test-long:
	$(call title, $@)
	@go test $(VERBOSE) -race -tags "$(BUILDTAGS)" $(PKGS)

lint:
	$(call title, $@)
	$(if $(GOLINT), , \
		$(error Please install golint: go get -u github.com/golang/lint/golint))
	$(call failif, $(GOLINT) ./... 2>&1 | grep -v vendor/)

fmt:
	$(call title, $@)
	$(call failif, gofmt -s -l . 2>&1 | grep -v vendor/, "Go code must be formatted with 'gofmt -s'")

# Depends on build (vet will otherwise silently fail if it can't load compiled imports)
vet: build
	$(call title, $@)
	$(call failif, go vet $(PKGS) 2>&1)

cover:
	$(call title, $@)
	@mkdir -p "$(COVERDIR)"
	@$(foreach PKG,$(PKGS),$(call gocover,$(PKG)))
	@echo "mode: $(COVERMODE)" > "$(COVERPROFILE)"
	@grep -h -v "^mode:" "$(COVERDIR)"/*.cover >> "$(COVERPROFILE)"
	@go tool cover -func="$(COVERPROFILE)"
	@go tool cover -html="$(COVERPROFILE)"

# Release management tooling and other helpers
authors: .mailmap .git/HEAD
	$(call title, $@)
	@git log --format='%aN <%aE>' | sort -fu > $@
	@echo Updated

# List deps
dependencies:
	$(call title, $@)
	@go list -f '{{ join .Deps  "\n"}}' .

# @$(GODEP) restore
# @go get -u all
update:
	$(call title, $@)
	$(if $(GODEP), , \
		$(error Please install godep: go get -u github.com/tools/godep))
	@$(GODEP) update github.com/...
	@$(GODEP) save

# Bump the static version file when cutting a new tagged release
version/version.go:
	./version/version.sh > $@

# dockerfile:
# 	@docker build --rm --force-rm -f Dockerfile -t $(PKG_NAME) .

# release: version/version.go
# 	@echo "+ $@"
