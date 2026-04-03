BINARY     := ingestd
BUILD_DIR  := ./bin
CMD        := ./cmd/ingestd

DATABASE_URL ?= postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable
SIMULATOR_URL ?= http://localhost:8000

.DEFAULT_GOAL := help

.PHONY: help build run test test-race lint vet docker-up docker-down docker-logs clean

help:
	@echo ""
	@echo "LiftIQ Telemetry Ingestor"
	@echo ""
	@echo "  make build          Compile binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  make run            Run ingestor (requires TimescaleDB + simulator)"
	@echo "  make test           Run unit tests (no DB required)"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make vet            Run go vet"
	@echo ""
	@echo "  make docker-up      Start TimescaleDB container"
	@echo "  make docker-down    Stop and remove TimescaleDB container"
	@echo "  make docker-logs    Tail TimescaleDB logs"
	@echo ""
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Environment variables:"
	@echo "  DATABASE_URL        $(DATABASE_URL)"
	@echo "  SIMULATOR_URL       $(SIMULATOR_URL)"
	@echo "  POLL_INTERVAL_SECONDS  (default 5)"
	@echo "  LOG_LEVEL              debug|info|warn|error (default info)"
	@echo ""

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Built $(BUILD_DIR)/$(BINARY)"

run: build
	DATABASE_URL="$(DATABASE_URL)" \
	SIMULATOR_URL="$(SIMULATOR_URL)" \
	$(BUILD_DIR)/$(BINARY)

# Unit tests only — no database required.
# The store package is excluded because it requires a live TimescaleDB.
test:
	go test ./internal/simulator/... ./internal/ingest/... ./internal/config/... -v

test-race:
	go test -race ./internal/simulator/... ./internal/ingest/... ./internal/config/... -v

vet:
	go vet ./...

docker-up:
	docker compose up -d
	@echo "Waiting for TimescaleDB to be ready..."
	@until docker compose exec timescaledb pg_isready -U liftiq -d liftiq > /dev/null 2>&1; do \
		printf '.'; sleep 1; \
	done
	@echo ""
	@echo "TimescaleDB is ready at postgres://liftiq:liftiq@localhost:5432/liftiq"

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f timescaledb

clean:
	rm -rf $(BUILD_DIR)
