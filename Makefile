BINARY  := binman
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/thaitrn/binman/cmd.Version=$(VERSION)
GOFLAGS ?=

.PHONY: build run test install uninstall tidy fmt vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run -ldflags "$(LDFLAGS)" . $(ARGS)

test:
	go test ./...

install: build
	@dest=$(shell go env GOPATH)/bin; \
	echo "installing $$dest/$(BINARY)"; \
	install -d $$dest && install -m 0755 $(BINARY) $$dest/

uninstall:
	rm -f $(shell go env GOPATH)/bin/$(BINARY)

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
