.PHONY: build lint deps dev_deps

BUILDER := "unknown"
VERSION := "unknown"

ifeq ($(origin IMAGES_BUILDER),undefined)
	BUILDER = $(shell git config --get user.name);
else
	BUILDER = ${IMAGES_BUILDER};
endif

ifeq ($(origin IMAGES_VERSION),undefined)
	VERSION = $(shell git rev-parse HEAD);
else
	VERSION = ${IMAGES_VERSION};
endif

build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o bin/images_processor .

lint:
	staticcheck ./...
	go vet ./...
	golangci-lint run

format:
	gofmt -s -w .

deps:
	go mod download

dev_deps:
	go install honnef.co/go/tools/cmd/staticcheck@v0.3.1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@master # TODO fix this version, currently master to lint go1.18 files

test:
	go test -count=1 -cover ./...

clean:
	rm -rf bin
