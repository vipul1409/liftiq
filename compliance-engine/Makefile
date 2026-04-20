BINARY    := complianced
BUILD_DIR := ./bin
CMD       := ./cmd/complianced

DATABASE_URL ?= postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable
HTTP_PORT    ?= 8080

.DEFAULT_GOAL := help

.PHONY: help build run test test-race vet clean

help:
	@echo ""
	@echo "LiftIQ Compliance Engine"
	@echo ""
	@echo "  make build          Compile binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  make run            Run compliance engine (requires TimescaleDB)"
	@echo "  make test           Run unit tests (no DB required)"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make vet            Run go vet"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Environment variables:"
	@echo "  DATABASE_URL            $(DATABASE_URL)"
	@echo "  HTTP_PORT               $(HTTP_PORT)  (default 8080)"
	@echo "  STALE_WINDOW_MINUTES    (default 10)"
	@echo "  LOG_LEVEL               debug|info|warn|error (default info)"
	@echo ""
	@echo "See root 'make demo-up' to start the full Phase 2 stack."
	@echo ""

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Built $(BUILD_DIR)/$(BINARY)"

run: build
	DATABASE_URL="$(DATABASE_URL)" \
	HTTP_PORT="$(HTTP_PORT)" \
	$(BUILD_DIR)/$(BINARY)

# Unit tests only — no database required.
# All external dependencies are replaced by in-process fakes.
test:
	go test ./internal/... -v

test-race:
	go test -race ./internal/... -v

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
