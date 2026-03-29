.PHONY: install test

VERSION ?= dev
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

install:
	go build $(LDFLAGS) -o $(HOME)/.local/bin/ai-collector ./cmd/ai-collector/
	@echo "Installed: ~/.local/bin/ai-collector"
	@echo "Run 'ai-collector init' to get started."

build:
	go build $(LDFLAGS) -o ai-collector ./cmd/ai-collector/

test:
	go test ./... -count=1
