.DEFAULT_GOAL := help

.PHONY: help demo-up demo-down demo-status \
        fault-door fault-brake fault-motor fault-safety fault-leveling clear-fault \
        logs-simulator logs-ingestor logs-compliance \
        mobile

# ── Help ────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "LiftIQ — Phase 1 demo orchestration"
	@echo ""
	@echo "  make demo-up          Start all services (TimescaleDB, simulator, ingestor, compliance engine)"
	@echo "  make demo-down        Stop all services"
	@echo "  make demo-status      Health check + live compliance summary for all elevators"
	@echo ""
	@echo "  Fault injection (ELV-003 must be ingesting — run demo-up first)"
	@echo "  make fault-door       Inject door_motor_degradation"
	@echo "  make fault-brake      Inject brake_wear"
	@echo "  make fault-motor      Inject motor_bearing_wear"
	@echo "  make fault-safety     Inject safety_circuit_intermittent"
	@echo "  make fault-leveling   Inject leveling_drift"
	@echo "  make clear-fault      Clear injected fault on ELV-003"
	@echo ""
	@echo "  Logs"
	@echo "  make logs-simulator   Tail simulator log"
	@echo "  make logs-ingestor    Tail ingestor log"
	@echo "  make logs-compliance  Tail compliance engine log"
	@echo ""
	@echo "  Mobile app"
	@echo "  make mobile           Start Expo dev server for liftiq-mobile"
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

# ── Mobile ───────────────────────────────────────────────────────────────────

mobile:
	cd liftiq-mobile && npx expo start
