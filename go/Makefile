.PHONY: build lint deps dev_deps

BUILDER := "unknown"
VERSION := "unknown"

ifeq ($(origin IMAGES_BUILDER),undefined)
	BUILDER = $(shell git config --get user.name)
else
	BUILDER = ${IMAGES_BUILDER}
endif

ifeq ($(origin IMAGES_VERSION),undefined)
	VERSION = $(shell git rev-parse HEAD)
else
	VERSION = ${IMAGES_VERSION}
endif

build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o out/image_processor cmd/*.go

lint:
	golangci-lint run --go=1.18

format:
	gofmt -s -w .

deps:
	go mod download

dev_deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@master # TODO fix this version, currently master to lint go1.18 files

test:
	LD_LIBRARY_PATH=$$(realpath $$(pwd)/../cpp/out/lib):$$LD_LIBRARY_PATH \
	PATH=$$(realpath $$(pwd)/../cpp/out/bin):$$PATH \
		go test \
			-count=1 \
			-cover \
			-parallel $$(nproc) \
			-timeout 160s \
			-race ./...

work:
	echo "go 1.18\n\nuse (\n\t.\n\t../../Common\n\t../../message-queue/go\n)" > go.work
	go mod tidy

clean:
	rm -rf bin
