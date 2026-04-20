.DEFAULT_GOAL := help

.PHONY: help demo-up demo-down demo-status \
        fault-door fault-brake fault-motor fault-safety fault-leveling clear-fault \
        logs-simulator logs-ingestor logs-compliance logs-report \
        mobile mobile-ios mobile-android \
        test-all \
        compose-local compose-dev compose-prod compose-down compose-build \
        db-reset db-reset-hard

# ── Help ────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "LiftIQ — Phase 2 orchestration"
	@echo ""
	@echo "  Local demo (native processes, fastest iteration)"
	@echo "  make demo-up          Start all 5 services (simulator, ingestor, compliance, report)"
	@echo "  make demo-down        Stop all services"
	@echo "  make demo-status      Health check + live compliance summary"
	@echo ""
	@echo "  Testing"
	@echo "  make test-all         Run all tests (Go services + mobile)"
	@echo ""
	@echo "  Docker environments"
	@echo "  make compose-local    Build + start full stack (local, ports exposed)"
	@echo "  make compose-dev      Build + start full stack (dev, detached)"
	@echo "  make compose-prod     Build + start full stack (prod, detached)"
	@echo "  make compose-down     Stop and remove all containers"
	@echo "  make compose-build    Build all Docker images without starting"
	@echo ""
	@echo "  Database"
	@echo "  make db-reset         Truncate all data (keep schema)"
	@echo "  make db-reset-hard    Drop all tables (schema reset)"
	@echo ""
	@echo "  Fault injection (simulator must be running)"
	@echo "  make fault-door       Inject door_motor_degradation on ELV-003"
	@echo "  make fault-brake      Inject brake_wear on ELV-003"
	@echo "  make fault-motor      Inject motor_bearing_wear on ELV-003"
	@echo "  make fault-safety     Inject safety_circuit_intermittent on ELV-003"
	@echo "  make fault-leveling   Inject leveling_drift on ELV-003"
	@echo "  make clear-fault      Clear injected fault on ELV-003"
	@echo ""
	@echo "  Logs"
	@echo "  make logs-simulator   Tail simulator log"
	@echo "  make logs-ingestor    Tail ingestor log"
	@echo "  make logs-compliance  Tail compliance engine log"
	@echo "  make logs-report      Tail report generator log"
	@echo ""
	@echo "  Mobile app (requires dev build — not Expo Go)"
	@echo "  make mobile           Start Expo Metro bundler"
	@echo "  make mobile-ios       Open directly in iOS Simulator"
	@echo "  make mobile-android   Open directly in Android Emulator"
	@echo ""

# ── Demo lifecycle ───────────────────────────────────────────────────────────

demo-up:
	./scripts/demo-start.sh

demo-down:
	./scripts/demo-stop.sh

demo-status:
	./scripts/demo-status.sh

# ── Fault injection ──────────────────────────────────────────────────────────

fault-door:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"door_motor_degradation"}' | python3 -m json.tool
	@echo ""
	@echo "Watch door_close_force_n climb toward 135 N:"
	@echo "  make demo-status   (poll every ~10 s)"

fault-brake:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"brake_wear"}' | python3 -m json.tool

fault-motor:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"motor_bearing_wear"}' | python3 -m json.tool

fault-safety:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"safety_circuit_intermittent"}' | python3 -m json.tool

fault-leveling:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"leveling_drift"}' | python3 -m json.tool

clear-fault:
	curl -sf -X DELETE http://localhost:8000/elevators/ELV-003/fault | python3 -m json.tool

# ── Logs ─────────────────────────────────────────────────────────────────────

logs-simulator:
	tail -f /tmp/liftiq-simulator.log

logs-ingestor:
	tail -f /tmp/liftiq-ingestor.log

logs-compliance:
	tail -f /tmp/liftiq-compliance.log

logs-report:
	tail -f /tmp/liftiq-report.log

# ── Docker environments ───────────────────────────────────────────────────────

COMPOSE_BASE := docker compose -f deploy/docker-compose.yml

compose-local:
	$(COMPOSE_BASE) -f deploy/docker-compose.local.yml \
	  --env-file deploy/env/.env.local \
	  up --build

compose-dev:
	$(COMPOSE_BASE) -f deploy/docker-compose.dev.yml \
	  --env-file deploy/env/.env.dev \
	  up --build -d

compose-prod:
	@test -f deploy/env/.env.prod || \
	  (echo "ERROR: deploy/env/.env.prod not found. Copy .env.prod.example and fill in secrets." && exit 1)
	$(COMPOSE_BASE) -f deploy/docker-compose.prod.yml \
	  --env-file deploy/env/.env.prod \
	  up --build -d

compose-down:
	$(COMPOSE_BASE) down

compose-build:
	$(COMPOSE_BASE) build

# ── Database ──────────────────────────────────────────────────────────────────

db-reset:
	./scripts/db-reset.sh

db-reset-hard:
	./scripts/db-reset.sh --hard

# ── Mobile ───────────────────────────────────────────────────────────────────
# Requires a development build (not Expo Go) — run `npx expo run:ios` once first.

mobile:
	cd liftiq-mobile && npx expo start

mobile-ios:
	cd liftiq-mobile && npx expo start --ios

mobile-android:
	cd liftiq-mobile && npx expo start --android

# ── Tests ─────────────────────────────────────────────────────────────────────

test-all:
	@echo "── telemetry-ingestor ──────────────────────────"
	cd telemetry-ingestor && go test ./internal/...
	@echo ""
	@echo "── compliance-engine ───────────────────────────"
	cd compliance-engine && go test ./internal/...
	@echo ""
	@echo "── report-generator ────────────────────────────"
	cd report-generator && go test ./internal/...
	@echo ""
	@echo "── liftiq-mobile ───────────────────────────────"
	cd liftiq-mobile && npm test
	@echo ""
	@echo "── elevator-simulator ──────────────────────────"
	cd elevator-simulator && source .venv/bin/activate && python -m pytest tests/ -q
