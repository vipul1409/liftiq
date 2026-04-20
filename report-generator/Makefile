BIN     := bin/reportd
CMD     := ./cmd/reportd
HTTP_PORT ?= 8082

.DEFAULT_GOAL := help

.PHONY: help build run test test-race vet clean

help:
	@echo ""
	@echo "LiftIQ Report Generator"
	@echo ""
	@echo "  make build          Compile binary to $(BIN)"
	@echo "  make run            Run report generator (requires Chrome on PATH)"
	@echo "  make test           Run unit tests (no Chrome required)"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make vet            Run go vet"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Environment variables:"
	@echo "  HTTP_PORT    $(HTTP_PORT)  (default 8082)"
	@echo "  LOG_LEVEL    debug|info|warn|error (default info)"
	@echo ""
	@echo "Chrome discovery: chromedp auto-detects Chrome/Chromium on PATH."
	@echo "macOS: brew install --cask google-chrome"
	@echo ""

build:
	@mkdir -p bin
	go build -o $(BIN) $(CMD)
	@echo "Built $(BIN)"

run: build
	HTTP_PORT="$(HTTP_PORT)" \
	./$(BIN)

test:
	go test ./internal/...

test-race:
	go test -race ./internal/...

vet:
	go vet ./...

clean:
	rm -rf bin/
