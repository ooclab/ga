# Go parameters
TOPDIR=$(PWD)
MODULE_PATH=~/.ga/middlewares
GOCMD=go
GOBUILD=$(GOCMD) build -mod=mod
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w

BUILD_DATE := "$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')"
GIT_COMMIT := "$(shell git rev-parse HEAD)"
VERSION :="$(shell git describe --tags --abbrev=0 | tr -d '\n')"
VERSION_COMMIT :="$(shell git describe --tags --abbrev=0 | tr -d '\n')-$(shell git rev-parse HEAD | tr -d '\n')"

# LDFLAGS=-ldflags "-s"
LDFLAGS=-ldflags "-s -X github.com/ooclab/ga/version.buildDate=$(BUILD_DATE) -X github.com/ooclab/ga/version.gitCommit=$(VERSION_COMMIT) -X github.com/ooclab/ga/version.gitVersion=$(VERSION)"
STATIC_LDFLAGS=-a -installsuffix cgo $(LDFLAGS)"

PROGRAM_NAME=ga
SUBDIRS := $(wildcard middlewares/*/.)

all: build $(SUBDIRS)

build:
	$(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)

$(SUBDIRS):
	mkdir -pv $(MODULE_PATH)
	cd $@ && go build -buildmode=plugin && cp -v *.so $(MODULE_PATH)

build-static:
	CGO_ENABLED=0 $(GOBUILD) -v $(STATIC_LDFLAGS) -o $(PROGRAM_NAME)

test:
	go test -v --cover ./...

clean:
	rm -f *.so ga


.PHONY: all $(SUBDIRS) build test
