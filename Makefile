

GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
RPM          := ./scripts/build_rpm.sh
pkgs          = ./...

BIN_DIR ?= $(shell pwd)/build
VERSION ?= $(shell cat VERSION)
REVERSION ?=$(shell git log -1 --pretty="%H")
BRANCH ?=$(shell git rev-parse --abbrev-ref HEAD)
TIME ?=$(shell date --rfc-3339=seconds)
HOST ?=$(shell hostname)  
DOCKER := $(shell { command -v podman || command -v docker; } 2>/dev/null)


all:  fmt style build docker-build docker-rpm
 
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

docker-build-centos7:
	$(DOCKER) run -v `pwd`:/go/src/github.com/jenningsloy318/redfish_exporter  -w /go/src/github.com/jenningsloy318/redfish_exporter docker.io/jenningsloy318/prom-builder:centos7  make build


docker-build-centos8:
	$(DOCKER) run -v `pwd`:/go/src/github.com/jenningsloy318/redfish_exporter  -w /go/src/github.com/jenningsloy318/redfish_exporter docker.io/jenningsloy318/prom-builder:centos8  make build

docker-build:
	make docker-build-centos7
	make docker-build-centos8

rpm: | build
	@echo ">> building binaries"
	$(RPM)

docker-rpm-centos7:
	$(DOCKER) run -v `pwd`:/go/src/github.com/jenningsloy318/redfish_exporter  -w /go/src/github.com/jenningsloy318/redfish_exporter docker.io/jenningsloy318/prom-builder:centos7  make rpm

docker-rpm-centos8:
	$(DOCKER) run -v `pwd`:/go/src/github.com/jenningsloy318/redfish_exporter  -w /go/src/github.com/jenningsloy318/redfish_exporter docker.io/jenningsloy318/prom-builder:centos8  make rpm

docker-rpm:
	make docker-rpm-centos7
	make docker-rpm-centos8

fmt:
	@echo ">> format code style"
	$(GOFMT) -w $$(find . -path ./vendor -prune -o -name '*.go' -print) 

clean:
	rm -rf $(BIN_DIR)

.PHONY: all style check_license fmt build fmt build rpm docker-build docker-rpm 