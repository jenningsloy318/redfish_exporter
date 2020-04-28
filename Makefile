

GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
RPM          := ./scripts/build_rpm.sh
pkgs          = ./...

BIN_DIR                 ?= $(shell pwd)/build
VERSION ?= $(shell cat VERSION)
REVERSION ?=$(shell git log -1 --pretty="%H")
BRANCH ?=$(shell git rev-parse --abbrev-ref HEAD)
TIME ?=$(shell date --rfc-3339=seconds)
HOST ?=$(shell hostname)  

all:  fmt style  build rpm 

 
style:
	@echo ">> checking code style"
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

check_license:
	@echo ">> checking license header"
	@licRes=$$(for file in $$(find . -type f -iname '*.go' ! -path './vendor/*') ; do \
               awk 'NR<=3' $$file | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
       done); \
       if [ -n "$${licRes}" ]; then \
               echo "license header checking failed:"; echo "$${licRes}"; \
               exit 1; \
       fi

build: | 
	@echo ">> building binaries"
	$(GO) build -o build/redfish_exporter -ldflags  '-X "main.Version=$(VERSION)" -X  "main.BuildRevision=$(REVERSION)" -X  "main.BuildBranch=$(BRANCH)" -X "main.BuildTime=$(TIME)" -X "main.BuildHost=$(HOSTNAME)"'


rpm: | build
	@echo ">> building binaries"
	$(RPM)


fmt:
	@echo ">> format code style"
	$(GOFMT) -w $$(find . -path ./vendor -prune -o -name '*.go' -print) 



.PHONY: all style check_license format build test   fmt 