SHELL := /bin/zsh

TARGET := bin/mapil

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --short HEAD)
BUILD := $(shell date +%Y-%m-%dT%H:%M:%S%z)

LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
ifeq ($(GO_ENV),production)
LDFLAGS += -ldflags "-X=main.devMode=false"
else
LDFLAGS += -ldflags "-X=main.devMode=true"
endif

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean install uninstall fmt simplify check run test test-all

all: check install

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)
	@rm -rf bin

install:
	@go install $(LDFLAGS)

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l ${SRC} | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@golangci-lint run ./...
	@go tool vet ${SRC}

run: install
	@$(TARGET)

test: clean
	@go test ./...

test-all: clean fmt build test
