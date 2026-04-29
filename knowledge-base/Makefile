.DEFAULT_GOAL := help

BUILD_DIR := ./bin

.PHONY: help build build-ingest run test test-race vet clean seed

help:
	@echo ""
	@echo "knowledge-base — OEM knowledge base with RAG (Ollama + pgvector)"
	@echo ""
	@echo "  make build         Compile knowledged binary"
	@echo "  make build-ingest  Compile ingest CLI binary"
	@echo "  make run           Build + start knowledge base service"
	@echo "  make seed          Seed demo data (requires Ollama + TimescaleDB)"
	@echo "  make test          Run unit tests"
	@echo "  make test-race     Run with race detector"
	@echo "  make vet           Run go vet"
	@echo "  make clean         Remove build artifacts"
	@echo ""

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/knowledged ./cmd/knowledged
	@echo "Built $(BUILD_DIR)/knowledged"

build-ingest:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/ingest ./cmd/ingest
	@echo "Built $(BUILD_DIR)/ingest"

run: build
	DATABASE_URL="postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable" \
	HTTP_PORT=8085 \
	LOG_LEVEL=debug \
	$(BUILD_DIR)/knowledged

seed: build-ingest
	$(BUILD_DIR)/ingest \
		--seed testdata/smartrise_c4_seed.json \
		--database-url "postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable"

test:
	go test ./internal/... -v

test-race:
	go test -race ./internal/... -v

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
