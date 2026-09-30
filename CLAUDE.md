# LiftIQ — Claude Code Guide

This document is the authoritative reference for working in this repository.
Read it before making changes.

---

## Repository layout

```
liftiq/
├── CLAUDE.md                              ← you are here
├── DEMO.md                                ← demo index (links to web + mobile runbooks)
├── DEMO-WEB.md                            ← web demo runbook (laptop-only, browser-based)
├── DEMO-MOBILE.md                         ← mobile demo runbook (React Native / Expo)
├── Makefile                               ← root orchestration (demo-up/down/status, fault injection)
├── CONTEXT.md                             ← domain glossary (Inspection, Override, Rule catalogue…)
├── docs/adr/                              ← architecture decision records — read before restructuring
├── contracts/                             ← cross-language fixtures (telemetry metric names, simulator snapshot)
├── scripts/
│   ├── demo-start.sh                      ← start all services with health checks
│   ├── demo-stop.sh                       ← stop all services
│   ├── demo-status.sh                     ← health check + live compliance summary
│   └── db-reset.sh                        ← truncate or drop all TimescaleDB data
├── deploy/
│   ├── docker-compose.yml                 ← base service definitions
│   ├── docker-compose.local.yml           ← local overlay (ports, debug)
│   ├── docker-compose.dev.yml             ← dev overlay (restart, debug)
│   ├── docker-compose.prod.yml            ← prod overlay (limits, warn logging)
│   └── env/
│       ├── .env.local                     ← local defaults
│       ├── .env.dev                       ← dev defaults
│       └── .env.prod.example              ← prod template (copy → .env.prod)
├── LiftIQ_Engineering_Plan_BMS_Integration.md   ← engineering blueprint
├── liftiq-web/                            ← Web demo app (laptop-friendly e2e demo)
│   ├── index.html                         ← Vite entry point
│   ├── vite.config.ts                     ← dev proxy → compliance :8080 + reports :8082
│   ├── Makefile                           ← dev, build, preview targets
│   └── src/
│       ├── main.tsx                       ← ReactDOM + BrowserRouter + InspectionProvider
│       ├── App.tsx                        ← routes: / → Scan, /inspect/:tag, /summary, /sign
│       ├── api/                           ← fetch wrappers (compliance, report)
│       ├── context/InspectionContext.tsx   ← cross-screen session state
│       ├── hooks/                         ← useUnits, useCompliance, useOverrides, usePhotos, useVoice
│       ├── components/                    ← StatusBadge, RuleRow, PhotoStrip, VoiceBar, SignatureCanvas
│       └── screens/                       ← ScanScreen, ComplianceScreen, SummaryScreen, SignatureScreen
├── elevator-simulator/                    ← Phase 1 Week 1 deliverable
│   ├── elevator_state.py                  ← ElevatorState dataclass + simulation
│   ├── bacnet_server.py                   ← BACnet/IP server (BAC0 wrapper)
│   ├── api.py                             ← FastAPI HTTP fault-injection API
│   ├── main.py                            ← entry point / orchestrator
│   ├── requirements.txt                   ← Python dependencies
│   ├── setup.sh                           ← venv creation script
│   ├── Makefile                           ← developer shortcuts
│   └── tests/                             ← pytest test suite (102 tests)
├── telemetry-ingestor/                    ← Phase 1 Week 2 deliverable
    ├── cmd/ingestd/main.go                ← binary entry point
    ├── internal/
    │   ├── config/config.go               ← env-var config
    │   ├── simulator/
    │   │   ├── client.go                  ← HTTP client + Client interface
    │   │   └── model.go                   ← ElevatorSnapshot JSON shape
    │   ├── store/
    │   │   ├── store.go                   ← Store interface + Row type
    │   │   └── pgx.go                     ← pgxpool impl, Connect, Migrate
    │   └── ingest/
    │       ├── mapper.go                  ← pure snapshot → []store.Row conversion
    │       └── poller.go                  ← 5-second poll loop
    ├── migrations/001_schema.sql          ← TimescaleDB DDL (reference)
    ├── docker-compose.yml                 ← TimescaleDB for local dev
    ├── go.mod / go.sum
    └── Makefile
├── compliance-engine/                     ← Phase 1 Week 3 deliverable
│   ├── cmd/complianced/main.go            ← binary entry point (port 8080)
│   ├── internal/
│   │   ├── api/handler.go                 ← HTTP routes + response types
│   │   ├── config/config.go               ← env-var config
│   │   ├── rules/
│   │   │   ├── rule.go                    ← Rule + Result types
│   │   │   ├── registry.go                ← 20 ASME A17.1 rule definitions
│   │   │   └── evaluator.go               ← EvaluateAll function
│   │   └── store/
│   │       ├── store.go                   ← Store interface + MetricReading type
│   │       └── pgx.go                     ← pgxpool impl (DISTINCT ON latest-per-metric)
│   ├── go.mod / go.sum
│   └── Makefile
└── liftiq-mobile/                         ← Phase 1 Week 4 + Phase 2 Week 5 deliverable
    ├── App.tsx                            ← entry point (NavigationContainer)
    ├── app.json                           ← Expo config (bundleId, permissions, plugins)
    ├── .env                               ← EXPO_PUBLIC_COMPLIANCE_API_URL
    ├── Makefile                           ← developer shortcuts
    ├── src/
    │   ├── api/
    │   │   ├── client.ts                  ← base fetch wrapper (ApiError, timeout)
    │   │   └── compliance.ts              ← getUnits, getCompliance wrappers
    │   ├── types/
    │   │   ├── compliance.ts              ← TypeScript interfaces (mirrors Go API shapes)
    │   │   └── voice.ts                   ← VoiceIntent union + VoiceState interface
    │   ├── utils/
    │   │   └── intentParser.ts            ← pure keyword → VoiceIntent function
    │   ├── screens/
    │   │   ├── ScanScreen.tsx             ← NFC stub + unit picker
    │   │   ├── ComplianceScreen.tsx       ← 20-rule checklist, overrides, voice cursor
    │   │   └── SummaryScreen.tsx          ← pass/fail banner + counts
    │   ├── components/
    │   │   ├── RuleRow.tsx                ← single rule card + override buttons + isActive highlight
    │   │   ├── StatusBadge.tsx            ← pass/fail/unknown colored pill
    │   │   ├── UnitPicker.tsx             ← modal bottom sheet for unit selection
    │   │   └── VoiceBar.tsx               ← floating mic button + pulse/speaking-dots animation
    │   ├── hooks/
    │   │   ├── useCompliance.ts           ← fetch Compliance snapshot + refetch
    │   │   ├── useUnits.ts                ← fetch unit list
    │   │   └── useVoice.ts                ← STT lifecycle, intent dispatch, TTS readback
    │   ├── store/overrides.ts             ← in-memory manual pass/fail overrides
    │   └── navigation/AppNavigator.tsx    ← native-stack: Scan → Compliance → Summary
    └── package.json
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
| ELV-003 | High-mileage | Motor drift 0.28, door force 118 N, brake 71 ms |

ELV-003 is the best elevator to fault-inject for demos — its baselines are already
close to ASME thresholds, so injected faults become visible quickly.

---

## telemetry-ingestor

### Requirements

| Requirement | Version |
|---|---|
| Go | 1.25+ |
| TimescaleDB | 2.x on PostgreSQL 16 |
| Docker | For local TimescaleDB |

### First-time setup

```bash
cd telemetry-ingestor

# Start TimescaleDB (waits for pg_isready before returning)
make docker-up

# Build binary
make build
```

### Running

```bash
# Defaults: simulator at localhost:8000, DB at localhost:5432
make run

# Or with explicit config
DATABASE_URL="postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable" \
SIMULATOR_URL="http://localhost:8000" \
POLL_INTERVAL_SECONDS=5 \
LOG_LEVEL=debug \
./bin/ingestd
```

Schema migrations run automatically on every start (`store.Migrate`). All DDL is idempotent — safe to re-run.

### End-to-end startup (both services)

```bash
# Terminal 1 — elevator simulator
cd elevator-simulator && make run-no-bacnet

# Terminal 2 — TimescaleDB + ingestor
cd telemetry-ingestor
make docker-up   # one-time: starts TimescaleDB container
make run         # builds and starts ingestd
```

Rows appear within 5 seconds of the first poll. Verify with the queries below.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | *(required)* | PostgreSQL DSN for TimescaleDB |
| `SIMULATOR_URL` | `http://localhost:8000` | Elevator simulator base URL |
| `POLL_INTERVAL_SECONDS` | `5` | Seconds between polls (float OK) |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |

### Makefile targets

| Target | Description |
|---|---|
| `make build` | Compile binary to `./bin/ingestd` |
| `make run` | Build + run (requires TimescaleDB and simulator running) |
| `make test` | Unit tests for `simulator`, `ingest`, `config` packages (no DB) |
| `make test-race` | Same with Go race detector (`-race`) |
| `make vet` | Run `go vet ./...` |
| `make docker-up` | Start TimescaleDB container; poll until `pg_isready` |
| `make docker-down` | Stop and remove TimescaleDB container |
| `make docker-logs` | Tail TimescaleDB container logs |
| `make clean` | Remove `./bin/` |

### Data flow

```
Simulator  GET /elevators  (every POLL_INTERVAL_SECONDS)
  → []ElevatorSnapshot  (one per elevator, 20 measurable fields each)
  → ingest.LookupUnit   (upsert elevator_units; cached after first hit)
  → ingest.MapSnapshot  → []store.Row  (20 rows × N elevators per poll)
  → store.WriteRows     → pgx.CopyFrom → telemetry hypertable
```

At default settings (3 elevators, 5 s interval): **60 rows/poll → 720 rows/min → ~1 M rows/day**.

### Stored metrics (20 per poll per elevator)

| Metric name | Source field | Go type |
|---|---|---|
| `motor_current_a` | `MotorCurrentA` | float64 |
| `motor_temp_c` | `MotorTempC` | float64 |
| `motor_rpm` | `MotorRPM` | float64 |
| `motor_run_hours` | `MotorRunHours` | float64 |
| `trip_count` | `TripCount` | int → float64 |
| `door_cycle_count` | `DoorCycleCount` | int → float64 |
| `door_motor_amps` | `DoorMotorAmps` | float64 |
| `door_close_force_n` | `DoorCloseForceN` | float64 |
| `door_close_time_ms` | `DoorCloseTimeMs` | int → float64 |
| `door_obstruction_events` | `DoorObstructionEvents` | int → float64 |
| `brake_engagement_count` | `BrakeEngagementCount` | int → float64 |
| `brake_current_a` | `BrakeCurrentA` | float64 |
| `brake_response_ms` | `BrakeResponseMs` | int → float64 |
| `leveling_accuracy_mm` | `LevelingAccuracyMm` | float64 |
| `vibration_g` | `VibrationG` | float64 |
| `door_interlock_ok` | `DoorInterlockOk` | bool → 1.0/0.0 |
| `governor_ok` | `GovernorOk` | bool → 1.0/0.0 |
| `buffer_ok` | `BufferOk` | bool → 1.0/0.0 |
| `pit_switch_ok` | `PitSwitchOk` | bool → 1.0/0.0 |
| `safety_circuit_ok` | `SafetyCircuitOk` | bool → 1.0/0.0 |

Excluded from storage: `Direction`, `DoorStatus`, `MaxFloors`, `InjectedFault` (operational/categorical fields).

### TimescaleDB schema

**`elevator_units`** — unit registry; auto-upserted on first encounter:

```sql
CREATE TABLE elevator_units (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_tag         VARCHAR(50)  NOT NULL UNIQUE,   -- e.g. "ELV-001"
    building_id      UUID,                            -- nullable for simulator
    controller_make  VARCHAR(100),
    controller_model VARCHAR(100),
    protocol         VARCHAR(20)  NOT NULL DEFAULT 'simulator',
    bms_address      VARCHAR(100),
    installed_date   DATE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

**`telemetry`** — narrow/long hypertable (one row per metric per poll):

```sql
CREATE TABLE telemetry (
    time    TIMESTAMPTZ      NOT NULL,
    unit_id UUID             NOT NULL REFERENCES elevator_units(id),
    metric  VARCHAR(50)      NOT NULL,
    value   DOUBLE PRECISION NOT NULL,
    quality VARCHAR(10)      NOT NULL DEFAULT 'good'   -- 'good' | 'stale' | 'missing'
);
```

Partitioning:
- **Time** partition: 7-day chunks via `create_hypertable('telemetry', 'time')`
- **Space** partition: hash by `unit_id` into 4 buckets via `add_dimension` — co-locates all data for one elevator within a chunk, speeding per-unit range queries
- **Index**: `(unit_id, metric, time DESC)` — the primary compliance-engine access pattern

### Key design decisions

**pgx CopyFrom** — `WriteRows` uses `pgx.CopyFrom` (PostgreSQL COPY protocol) rather than multi-row INSERT. CopyFrom bypasses WAL parsing overhead and is the fastest bulk-insert path available through pgx. At 60 rows per poll this is overkill, but it keeps headroom when elevator count scales.

**sync.Map unit cache** — `LookupUnit` caches tag→UUID in a `sync.Map` after the first DB round-trip. Subsequent polls for the same elevator never hit the database. The upsert SQL (`ON CONFLICT (unit_tag) DO UPDATE ... RETURNING id`) is idempotent so a cold restart is safe.

**pgx v5 multi-statement limitation** — pgx v5 does not execute multiple SQL statements in a single `Exec` call. Migrations are stored as a `[]string` slice and executed one statement at a time. The `add_dimension` idempotency is handled with a PL/pgSQL `DO` block (no `IF NOT EXISTS` equivalent in standard TimescaleDB SQL).

**Interface injection** — `simulator.Client` and `store.Store` are interfaces defined in their own packages. All tests use in-process fakes (`fakeClient`, `fakeStore`) with `sync.Mutex` protection; no real network or database is required. The `store` package is excluded from unit tests (`make test`) because it needs a live TimescaleDB.

### Common verification queries

```sql
-- Row count per elevator per minute (last 10 minutes)
SELECT
    u.unit_tag,
    time_bucket('1 minute', t.time) AS bucket,
    count(*)                        AS rows
FROM telemetry t
JOIN elevator_units u ON u.id = t.unit_id
WHERE t.time > NOW() - INTERVAL '10 minutes'
GROUP BY 1, 2
ORDER BY 2 DESC, 1;

-- Latest value of every metric for ELV-003
SELECT t.metric, t.value, t.time
FROM telemetry t
JOIN elevator_units u ON u.id = t.unit_id
WHERE u.unit_tag = 'ELV-003'
  AND t.time > NOW() - INTERVAL '1 minute'
ORDER BY t.metric, t.time DESC;

-- ASME A17.1 threshold check (last reading per metric per elevator)
SELECT
    u.unit_tag,
    t.metric,
    last(t.value, t.time) AS latest_value
FROM telemetry t
JOIN elevator_units u ON u.id = t.unit_id
WHERE t.metric IN ('door_close_force_n', 'brake_response_ms', 'leveling_accuracy_mm')
  AND t.time > NOW() - INTERVAL '10 minutes'
GROUP BY 1, 2
ORDER BY 1, 2;
```

### Packages

| Package | Responsibility |
|---|---|
| `config` | Load env vars; fail fast if `DATABASE_URL` missing |
| `simulator` | `Client` interface + `httpClient` impl; `ElevatorSnapshot` JSON model |
| `store` | `Store` interface; `pgxStore` (CopyFrom + upsert cache); `Connect`; `Migrate` |
| `ingest` | `MapSnapshot` (pure, no I/O); `Poller.Run` + `PollOnce` (poll loop) |

---

## Code conventions

- **Simulation logic** belongs in `elevator_state.py`. Keep `ElevatorState` as a
  plain dataclass with no I/O.
- **BACnet I/O** belongs in `bacnet_server.py`. All BAC0 imports are guarded by
  try/except so the rest of the code works without BACnet installed.
- **HTTP I/O** belongs in `api.py`. All FastAPI imports are guarded similarly.
- **Wiring** happens only in `main.py` / `cmd/ingestd/main.go`.
- Do not add LLM calls or probabilistic logic to the compliance rule engine
  (future `compliance/rules.go`). Safety-critical pass/fail must be deterministic
  and auditable per the engineering plan.
- Interfaces are defined in the package that owns them (`simulator.Client`,
  `store.Store`). Concrete implementations satisfy them without explicit declaration.

---

## Docker deployment

Every backend service has a `Dockerfile`. Environment-specific compose overlays live in `deploy/`.

### Directory layout

```
deploy/
├── docker-compose.yml          ← base: shared service definitions (no ports/restart)
├── docker-compose.local.yml    ← local: all ports exposed, debug logging, faster ticks
├── docker-compose.dev.yml      ← dev: restart on-failure, debug logging
├── docker-compose.prod.yml     ← prod: always-restart, resource limits, warn logging
└── env/
    ├── .env.local               ← local defaults (safe to commit)
    ├── .env.dev                 ← dev defaults (fill in before deploying)
    └── .env.prod.example        ← prod template (copy to .env.prod, never commit)
```

### Makefile targets

| Target | Description |
|---|---|
| `make compose-local` | Build + start full stack (foreground, all ports) |
| `make compose-dev` | Build + start full stack (detached) |
| `make compose-prod` | Build + start full stack (detached, requires `.env.prod`) |
| `make compose-down` | Stop and remove all containers |
| `make compose-build` | Build all images without starting |

### First-time prod setup

```bash
cp deploy/env/.env.prod.example deploy/env/.env.prod
# edit .env.prod — set DB_PASSWORD to a strong value
make compose-prod
```

### Environment variables

All three environments share the same variable names. Defaults are set in the base compose file (`${VAR:-default}`).

| Variable | Local | Dev | Prod |
|---|---|---|---|
| `DB_PASSWORD` | `liftiq` | set in `.env.dev` | set in `.env.prod` |
| `LOG_LEVEL` | `debug` | `debug` | `warn` |
| `STALE_WINDOW` | `60` min | `10` min | `5` min |
| `POLL_INTERVAL` | `5` s | `5` s | `5` s |
| `TICK_INTERVAL` | `0.5` s | `1.0` s | `1.0` s |

### Database reset

```bash
make db-reset         # truncate all data, keep schema
make db-reset-hard    # drop all tables (ingestor re-migrates on next start)

# Against a remote host
DATABASE_URL="postgres://liftiq:pass@myhost:5432/liftiq?sslmode=disable" \
  ./scripts/db-reset.sh
```

---

## liftiq-mobile

### Requirements

| Requirement | Version |
|---|---|
| Node.js | 18+ |
| Expo CLI | via `npx expo` (no global install needed) |

### First-time setup

```bash
cd liftiq-mobile
npm install
npx expo prebuild      # generates ios/ and android/ native projects (required for voice)
```

`prebuild` is required because `expo-speech-recognition` links native modules (iOS `SFSpeechRecognizer`, Android `SpeechRecognizer`). Expo Go does not support it.

### Running

```bash
# Development build (required — Expo Go does not support expo-speech-recognition)
make ios            # builds and runs on iOS Simulator / connected device
make android        # builds and runs on Android Emulator / connected device
```

> **iOS Simulator caveat:** `SFSpeechRecognizer` does not function in the simulator. Voice commands require a physical iOS device or Android Emulator.

### Environment

Edit `.env` to point at the compliance engine:

```
EXPO_PUBLIC_COMPLIANCE_API_URL=http://localhost:8080       # iOS device on same network
EXPO_PUBLIC_COMPLIANCE_API_URL=http://10.0.2.2:8080        # Android Emulator
EXPO_PUBLIC_COMPLIANCE_API_URL=http://192.168.x.x:8080     # physical device (LAN IP)
```

### App flow

1. **Scan screen** — tap "Scan Elevator Tag" (NFC stub) → picker shows all units from `GET /units`
2. **Compliance screen** — displays all 20 ASME A17.1 rules grouped by subsystem; active rule highlighted with blue border; VoiceBar for push-to-talk inspection
3. **Summary screen** — overall pass/fail banner, per-category counts

### Voice-to-command (Phase 2 Week 5)

The VoiceBar floats above the footer on the Compliance screen. Tap the mic button to start push-to-talk recognition.

| Intent | Trigger phrases | Action |
|---|---|---|
| `pass` | "pass", "passed", "looks good", "ok" | `setOverride(activeRule, 'pass')` + TTS "Marked pass." |
| `fail` | "fail", "failed", "no good", "bad" | `setOverride(activeRule, 'fail')` + TTS "Marked fail." |
| `next` | "next", "continue", "move on" | Advance cursor + TTS reads next rule |
| `skip` | "skip", "ignore", "next item" | TTS "Skipped." + advance cursor |
| `photo` | "photo", "camera", "picture" | TTS stub "Opening camera." (Week 6 wires camera) |
| `stop` | "stop", "done", "cancel" | Stop listening |
| `unknown` | anything else | TTS "Didn't catch that." |

**Key files:**
- `src/utils/intentParser.ts` — pure keyword matcher, multi-word phrases checked before single tokens
- `src/hooks/useVoice.ts` — permission request, STT lifecycle (`expo-speech-recognition`), TTS readback (`expo-speech`); pauses mic while speaking to prevent feedback
- `src/components/VoiceBar.tsx` — pulsing ring animation while listening, bouncing dots while speaking
- `src/screens/ComplianceScreen.tsx` — `activeRuleId` cursor state, `handleIntent` dispatch, `advanceCursor`

**Dependencies:** `expo-speech ~13.0.0`, `expo-speech-recognition ~3.1.2`

### NFC note

Real NFC (via `expo-nfc-manager`) requires a development build. The stub is intentional for Phase 1 — wire up actual NFC in Phase 3 when testing with real hardware.

### Makefile targets

| Target | Description |
|---|---|
| `make start` | Start Expo dev server (Metro only — no native build) |
| `make ios` | `expo run:ios` — build + run development client |
| `make android` | `expo run:android` — build + run development client |
| `make typecheck` | Run `tsc --noEmit` |

---

## liftiq-web

### Requirements

| Requirement | Version |
|---|---|
| Node.js | 18+ |

### First-time setup

```bash
cd liftiq-web
npm install
```

### Running

```bash
make web        # start dev server at http://localhost:5173
# or
make web-build  # production build → dist/
```

The dev server proxies API requests to the backend services:
- `/api/compliance/*` → `http://localhost:8080` (compliance engine)
- `/api/reports/*` → `http://localhost:8082` (report generator)

### App flow

1. **Scan screen** (`/`) — tap "Scan Elevator" → select unit from picker
2. **Compliance screen** (`/inspect/:tag`) — 20 ASME A17.1 rules grouped by subsystem; override buttons, photo upload, voice commands (Web Speech API)
3. **Summary screen** (`/summary`) — pass/fail banner, per-category counts
4. **Signature screen** (`/sign`) — HTML5 Canvas signature → SVG data URI, "Sign & Generate PDF" → blob download

### Voice commands (Chrome/Edge/Safari)

Uses the Web Speech API. Same intent patterns as the mobile app:
- **pass/fail** — override active rule
- **next/skip** — advance cursor
- **photo** — open file picker
- **stop** — stop listening

Falls back gracefully on browsers without `SpeechRecognition` (e.g. Firefox).

### Key differences from mobile app

| Feature | Mobile (`liftiq-mobile`) | Web (`liftiq-web`) |
|---|---|---|
| Framework | React Native / Expo | React + Vite |
| Voice | expo-speech-recognition | Web Speech API |
| Photo capture | expo-image-picker + GPS | File input + Geolocation API |
| Signature | PanResponder + dots | HTML5 Canvas |
| PDF delivery | expo-file-system + Sharing | Blob URL download |
| State passing | React Navigation route params | InspectionContext |

### Makefile targets

| Target | Description |
|---|---|
| `make dev` | Start Vite dev server with HMR |
| `make build` | TypeScript check + production build |
| `make preview` | Serve production build locally |
| `make typecheck` | Run `tsc --noEmit` |

---

## Build phases (from engineering plan)

| Phase | Week | Deliverable | Status |
|---|---|---|---|
| Phase 1 | 1 | Elevator simulator (BACnet/IP + HTTP fault injection) | Complete |
| Phase 1 | 2 | Telemetry ingestor (Go → TimescaleDB) | Complete |
| Phase 1 | 3 | Compliance engine (20 ASME A17.1 rules, Go REST API) | Complete |
| Phase 1 | 4 | Mobile app skeleton (scan → checklist → summary) | Complete |
| Phase 2 | 5 | Voice-to-command pipeline (STT + TTS + intent parser) | Complete |
| Phase 2 | 6 | Photo evidence capture | Complete |
| Phase 2 | 7 | PDF report generator | Complete |
| Phase 2 | 8 | End-to-end demo flow | Complete |
| — | — | Web demo app (laptop-friendly e2e) | Complete |
| Phase 3 | 9–14 | Real BMS / BACnet integration | Not started |
| Phase 4 | 12–16 | OEM RAG knowledge base | Not started |

---

## Agent skills

### Issue tracker

GitHub Issues on `vipul1409/liftiq`, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
