# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
# LDFLAGS=-ldflags "-s"
LDFLAGS=-ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"
STATIC_LDFLAGS=-a -installsuffix cgo -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"

PROGRAM_NAME=ga

build:
	$(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)

build-static:
	CGO_ENABLED=0 $(GOBUILD) -v $(STATIC_LDFLAGS) -o $(PROGRAM_NAME)

test:
	go test -v --cover ./...
