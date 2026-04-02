VENV      := .venv
PYTHON    := python3
ACTIVATE  := source $(VENV)/bin/activate

.DEFAULT_GOAL := help

.PHONY: help setup run run-no-bacnet test test-cov fault-door fault-brake fault-motor fault-safety fault-leveling clear-fault clean

help:
	@echo ""
	@echo "LiftIQ Elevator Simulator"
	@echo ""
	@echo "  make setup            Create .venv and install dependencies"
	@echo "  make run              Run with BACnet + HTTP API"
	@echo "  make run-no-bacnet    Run with HTTP API only (no BACnet)"
	@echo ""
	@echo "  make test             Run unit tests"
	@echo "  make test-cov         Run tests with HTML coverage report"
	@echo ""
	@echo "  make fault-door       Inject door_motor_degradation on ELV-003"
	@echo "  make fault-brake      Inject brake_wear on ELV-003"
	@echo "  make fault-motor      Inject motor_bearing_wear on ELV-003"
	@echo "  make fault-safety     Inject safety_circuit_intermittent on ELV-003"
	@echo "  make fault-leveling   Inject leveling_drift on ELV-003"
	@echo "  make clear-fault      Clear injected fault on ELV-003"
	@echo ""
	@echo "  make clean            Remove .venv and __pycache__"
	@echo ""

setup:
	chmod +x setup.sh && ./setup.sh

$(VENV)/bin/activate: requirements.txt
	$(PYTHON) -m venv $(VENV)
	$(VENV)/bin/pip install --quiet --upgrade pip
	$(VENV)/bin/pip install --quiet -r requirements.txt

run: $(VENV)/bin/activate
	$(ACTIVATE) && python main.py

run-no-bacnet: $(VENV)/bin/activate
	$(ACTIVATE) && python main.py --no-bacnet

test: $(VENV)/bin/activate
	$(ACTIVATE) && pytest tests/ -v

test-cov: $(VENV)/bin/activate
	$(ACTIVATE) && pytest tests/ -v --cov=. --cov-report=term-missing --cov-report=html

# Fault injection helpers (simulator must be running in another terminal)
fault-door:
	curl -sf -X POST http://localhost:8000/elevators/ELV-003/fault \
	     -H "Content-Type: application/json" \
	     -d '{"fault":"door_motor_degradation"}' | python3 -m json.tool

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

clean:
	rm -rf $(VENV) __pycache__ **/__pycache__
