GO ?= go
BINARY := bin/geomind-api.exe

.PHONY: fmt vet test run build check

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/api

build:
	mkdir -p bin
	$(GO) build -o $(BINARY) ./cmd/api

check: fmt vet test build
