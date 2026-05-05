BINARY_NAME = dbvault

.PHONY: all build install tidy fmt vet test

all: build

build:
	go build -o $(BINARY_NAME) ./...

install:
	go install ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...
