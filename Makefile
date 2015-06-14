.PHONY: all validate test build

all: validate test build

GOLINT_BIN := $(GOPATH)/bin/golint
GOLINT := $(shell [ -x $(GOLINT_BIN) ] && echo $(GOLINT_BIN) || echo '')

GODEP_BIN := $(GOPATH)/bin/godep
GODEP := $(shell [ -x $(GODEP_BIN) ] && echo $(GODEP_BIN) || echo '')

validate:
	find . -type d -not -path '*/.*' -not -path './Godeps*' -exec go fmt {} \;
	$(if $(GOLINT), , \
		$(error Please install golint: go get -u github.com/golang/lint/golint))
	find . -type d -not -path '*/.*' -not -path './Godeps*' -exec $(GOLINT) {} \;
	find . -type d -not -path '*/.*' -not -path './Godeps*' -exec go vet {} \;

test:
	go test

build:
	$(if $(GODEP), , \
		$(error Please install godep: go get -u github.com/tools/godep))
	$(GODEP) go build
