# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
# LDFLAGS=-ldflags "-s"
LDFLAGS=-ldflags "-s -X main.buildstamp=`date -u '+%Y-%m-%dT%H:%M:%SZ'` -X main.githash=`git rev-parse HEAD | cut -c1-8`"
STATIC_LDFLAGS=-a -installsuffix cgo -ldflags "-s -X main.buildstamp=`date -u '+%Y-%m-%dT%H:%M:%SZ'` -X main.githash=`git rev-parse HEAD | cut -c1-8`"

PROGRAM_NAME=ga

build:
	$(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)

build-static:
	CGO_ENABLED=0 $(GOBUILD) -v $(STATIC_LDFLAGS) -o $(PROGRAM_NAME)

test:
	go test -v --cover ./...
