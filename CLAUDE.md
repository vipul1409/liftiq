# LiftIQ — Claude Code Guide

This document is the authoritative reference for working in this repository.
Read it before making changes.

---

## Repository layout

```
liftiq/
├── CLAUDE.md                              ← you are here
├── LiftIQ_Engineering_Plan_BMS_Integration.md   ← engineering blueprint
└── elevator-simulator/                    ← Phase 1 Week 1 deliverable
    ├── elevator_state.py                  ← ElevatorState dataclass + simulation
    ├── bacnet_server.py                   ← BACnet/IP server (BAC0 wrapper)
    ├── api.py                             ← FastAPI HTTP fault-injection API
    ├── main.py                            ← entry point / orchestrator
    ├── requirements.txt                   ← Python dependencies
    ├── setup.sh                           ← venv creation script
    └── Makefile                           ← developer shortcuts
```

The overall product plan and protocol details live in
`LiftIQ_Engineering_Plan_BMS_Integration.md`.
Check that document before adding new components or changing data-point mappings.

---

## elevator-simulator

### Requirements

| Requirement | Version |
|---|---|
| Python | **3.10 or newer** (BAC0 hard requirement) |
| BAC0 | 22.x (BACnet server — optional, see below) |
| FastAPI + uvicorn | Any recent version |

On macOS, BAC0 depends on `libpcap`:

```bash
brew install libpcap
```

### First-time setup

```bash
cd elevator-simulator

# Option A — automated (recommended)
chmod +x setup.sh
./setup.sh

# Option B — manual
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

`setup.sh` will:
1. Verify Python ≥ 3.10.
2. Create `.venv/` inside `elevator-simulator/`.
3. Install all dependencies from `requirements.txt`.
4. Run a quick import check and warn if BAC0 is missing.

To rebuild the venv from scratch:

```bash
rm -rf .venv && ./setup.sh
```

### Activating the venv

Always activate before running any Python commands:

```bash
source .venv/bin/activate   # fish: source .venv/bin/activate.fish
```

Deactivate when done:

```bash
deactivate
```

### Running the simulator

```bash
# With BACnet/IP server + HTTP API (requires BAC0 + libpcap)
python main.py

# HTTP API only — no BACnet (works without BAC0)
python main.py --no-bacnet

# Custom tick interval and port
python main.py --no-bacnet --tick-interval 0.5 --http-port 9000
```

#### Environment variable overrides

| Variable | Default | Description |
|---|---|---|
| `LIFTIQ_TICK_INTERVAL` | `1.0` | Seconds between simulation ticks |
| `LIFTIQ_HTTP_PORT` | `8000` | HTTP API port |
| `LIFTIQ_BACNET_IP` | `0.0.0.0/24` | BACnet local IP/prefix (e.g. `192.168.1.10/24`) |

#### Makefile shortcuts

```bash
make setup            # same as ./setup.sh
make run              # BACnet + HTTP
make run-no-bacnet    # HTTP only
make clean            # remove .venv and __pycache__
```

### HTTP API (fault injection & monitoring)

The API runs at `http://localhost:8000` by default.
Interactive docs: `http://localhost:8000/docs`

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `GET` | `/elevators` | All elevator telemetry snapshots |
| `GET` | `/elevators/{unit_id}` | Single elevator state |
| `GET` | `/elevators/{unit_id}/bacnet` | BACnet point map |
| `GET` | `/faults` | List available fault types |
| `POST` | `/elevators/{unit_id}/fault` | Inject a fault `{"fault": "<type>"}` |
| `DELETE` | `/elevators/{unit_id}/fault` | Clear injected fault |

#### Available fault types

| Fault | What it simulates | ASME threshold to watch |
|---|---|---|
| `door_motor_degradation` | Rising door motor current + accelerated close-force increase | Close force < 135 N (ASME 2.13.4) |
| `brake_wear` | Brake response held near 75 ms | Response < 80 ms (ASME 8.6.4.1) |
| `motor_bearing_wear` | Accelerated motor current drift + elevated vibration | — |
| `safety_circuit_intermittent` | Random 5% chance of safety circuit trip per tick | Circuit must be continuous (ASME 2.26.1) |
| `leveling_drift` | Leveling accuracy near 11 mm | Within 12.7 mm / ½ inch (ADA) |

#### Example fault-injection workflow (demo script)

```bash
# Terminal 1 — start simulator
source .venv/bin/activate
python main.py --no-bacnet

# Terminal 2 — inject fault on the high-mileage elevator
curl -s http://localhost:8000/elevators/ELV-003 | python3 -m json.tool   # baseline

curl -X POST http://localhost:8000/elevators/ELV-003/fault \
     -H "Content-Type: application/json" \
     -d '{"fault":"door_motor_degradation"}'

# Watch door_close_force_n climb toward 135 N over ~10 minutes
watch -n 2 "curl -s http://localhost:8000/elevators/ELV-003 | python3 -m json.tool"

# Clear the fault
curl -X DELETE http://localhost:8000/elevators/ELV-003/fault
```

Or use the Makefile targets:

```bash
make fault-door       # inject door_motor_degradation on ELV-003
make fault-brake      # inject brake_wear on ELV-003
make clear-fault      # clear fault on ELV-003
```

### BACnet/IP details

When BAC0 is available, each elevator is exposed as a separate BACnet device:

| Elevator | BACnet device ID | Description |
|---|---|---|
| ELV-001 | 1001 | New unit — baseline wear |
| ELV-002 | 1002 | Mid-life unit — moderate wear |
| ELV-003 | 1003 | High-mileage unit — near thresholds |

Default port: **UDP 47808** (standard BACnet/IP).

To read an elevator from another BAC0 client on the same network:

```python
import BAC0
bacnet = BAC0.connect(ip='192.168.1.100/24')
elevator = BAC0.device('192.168.1.50', 1001, bacnet)
motor_current = elevator['analogInput 0'].presentValue
door_force    = elevator['analogInput 12'].presentValue
```

Full BACnet object map is in `elevator_state.py → to_bacnet_points()`.

### Simulated elevators

`elevator_state.py → make_elevator_fleet()` creates three units:

| Unit | Profile | Notable values |
|---|---|---|
| ELV-001 | New | Motor drift 0.0, door force 67 N, brake 45 ms |
| ELV-002 | Mid-life | Motor drift 0.12, door force 98 N, brake 58 ms |
| ELV-003 | High-mileage | Motor drift 0.35, door force 118 N, brake 71 ms |

ELV-003 is the best elevator to fault-inject for demos — its baselines are already
close to ASME thresholds, so injected faults become visible quickly.

---

## Code conventions

- **Simulation logic** belongs in `elevator_state.py`. Keep `ElevatorState` as a
  plain dataclass with no I/O.
- **BACnet I/O** belongs in `bacnet_server.py`. All BAC0 imports are guarded by
  try/except so the rest of the code works without BACnet installed.
- **HTTP I/O** belongs in `api.py`. All FastAPI imports are guarded similarly.
- **Wiring** happens only in `main.py`.
- Do not add LLM calls or probabilistic logic to the compliance rule engine
  (future `compliance/rules.go`). Safety-critical pass/fail must be deterministic
  and auditable per the engineering plan.

---

## Build phases (from engineering plan)

| Phase | Weeks | Status |
|---|---|---|
| Phase 1 — Simulated environment | 1–4 | Week 1 complete (this simulator) |
| Phase 2 — Voice + report generation | 5–8 | Not started |
| Phase 3 — Real BMS integration | 9–14 | Not started |
| Phase 4 — OEM RAG knowledge base | 12–16 | Not started |

Week 2 deliverable: Go telemetry ingestion service polling this simulator
every 5 seconds and writing to TimescaleDB.
