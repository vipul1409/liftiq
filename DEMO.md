# LiftIQ Phase 1 — Demo Runbook

End-to-end demo of the full Phase 1 stack:
**elevator simulator → telemetry ingestor → TimescaleDB → compliance engine → mobile app**

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Python | 3.10+ | Required by elevator simulator |
| Go | 1.25+ | Required by ingestor and compliance engine |
| Docker | Any recent | Required for TimescaleDB |
| Node.js | 18+ | Required for mobile app |
| Expo Go | Latest | Install on iOS/Android from the app store |

macOS only: `brew install libpcap` (required by BAC0 — harmless if BACnet is not used)

---

## Quick start (one command)

```bash
# From the repo root
make demo-up
```

This starts all four services in order with health checks between each step.
Logs are written to `/tmp/liftiq-*.log`.

Once the stack is up, start the mobile app:

```bash
make mobile
# Scan the QR code with Expo Go, or press 'i' to open in iOS Simulator
```

---

## What `make demo-up` does

| Step | Service | Port | Wait condition |
|---|---|---|---|
| 1 | TimescaleDB (Docker) | 5432 | `pg_isready` |
| 2 | Elevator simulator | 8000 | `GET /health` → 200 |
| 3 | Telemetry ingestor | — | process started (polls every 5 s) |
| 4 | Compliance engine | 8080 | `GET /health` → 200 |

After ~10 seconds the ingestor writes the first batch of rows and the compliance engine begins serving live results.

---

## Verify the stack is healthy

```bash
make demo-status
```

Expected output:

```
── Service health ───────────────────────────────
  ✓ Elevator simulator  (http://localhost:8000/health)
  ✓ Compliance engine   (http://localhost:8080/health)

── Elevator units (compliance engine) ───────────
{
    "units": ["ELV-001", "ELV-002", "ELV-003"]
}

── Compliance summary per unit ──────────────────
  ELV-001  overall=pass  pass=20  fail=0  unknown=0
  ELV-002  overall=pass  pass=20  fail=0  unknown=0
  ELV-003  overall=pass  pass=20  fail=0  unknown=0
```

If you see `unknown=20` the ingestor hasn't written data yet — wait a few more seconds and retry.

---

## Investor demo script

This is the recommended demo sequence for investors. It shows a healthy elevator, injects a fault, and watches the compliance engine flag it automatically.

### 1. Establish baseline (2 min)

Open the mobile app and "scan" ELV-003 (the high-mileage unit). Show that all 20 ASME checks are green — the system is live and reading real simulated telemetry.

```bash
# Also visible via curl
curl -s http://localhost:8080/units/ELV-003/compliance/summary | python3 -m json.tool
```

Point out:
- Values are real telemetry, not mock data (motor current 18+ A, door force 118 N — already near thresholds)
- The compliance engine is connected to TimescaleDB; data is persistent and queryable

### 2. Inject a door fault (2 min)

```bash
make fault-door
```

This tells the simulator that ELV-003's door motor is degrading. The `door_close_force_n` value will start climbing toward the 135 N ASME A17.1 limit.

```bash
# Watch the value rise (re-run every ~10 s)
make demo-status
```

Hit **Refresh** in the mobile app. Once `door_close_force_n` crosses 135 N you'll see:
- `ASME-006` flips from **PASS** to **FAIL** in the checklist
- Overall status banner turns **red**
- The summary shows **fail=1**

Explain: *"This is exactly what happens when a real elevator's door motor starts wearing out. The inspector would get an automatic flag before they even arrive on-site."*

### 3. Show manual override (1 min)

On the **Compliance** screen, tap **Override Pass** on `ASME-006`. The badge changes to **PASS (manual)**.

Explain: *"Some checks can't be verified from telemetry alone — the inspector physically tests the door force with a gauge. They override the result directly in the app. The manual result is captured alongside the auto-evaluated ones."*

### 4. Clear the fault and show recovery (1 min)

```bash
make clear-fault
```

Hit **Refresh** in the mobile app. The value drops back below 135 N and `ASME-006` returns to **PASS**. Explain: *"The system reflects real-time state. If maintenance fixes the issue before the inspection, the checklist auto-clears."*

### 5. Close (30 sec)

Point at the three-panel architecture: simulated BACnet elevator → Go ingestor → TimescaleDB → Go compliance engine → React Native app. In Phase 3 this same pipeline connects to a real building management system.

---

## ISP pilot demo script

For a technician audience, skip the fault injection and focus on the inspection workflow:

1. Walk up to elevator, tap "Scan Elevator Tag" → select unit
2. Show pre-filled checklist — 17 of 20 items auto-passed from telemetry
3. The 3 remaining items need a physical test (e.g., door force gauge) — tap Override
4. Tap **View Summary** — show the overall pass/fail card
5. Explain the time savings: *"You used to fill this out by hand after the inspection. Now it's pre-filled before you open the panel."*

---

## Fault types reference

| Command | Fault | ASME rule that trips | Threshold |
|---|---|---|---|
| `make fault-door` | Door motor degradation | ASME-006 | 135 N close force |
| `make fault-brake` | Brake wear | ASME-011 | 80 ms response time |
| `make fault-motor` | Motor bearing wear | ASME-001 | 20 A motor current |
| `make fault-safety` | Safety circuit intermittent | ASME-020 | circuit must be continuous |
| `make fault-leveling` | Leveling drift | ASME-014 | 12.7 mm (½ inch ADA) |

Clear any active fault:

```bash
make clear-fault
```

---

## Manual startup (step-by-step)

If you prefer to run services in separate terminals to see logs live:

**Terminal 1 — TimescaleDB**
```bash
cd telemetry-ingestor
make docker-up
```

**Terminal 2 — Elevator simulator**
```bash
cd elevator-simulator
make setup          # first time only
make run-no-bacnet
```

**Terminal 3 — Telemetry ingestor**
```bash
cd telemetry-ingestor
make run
```

**Terminal 4 — Compliance engine**
```bash
cd compliance-engine
make run
```

**Terminal 5 — Mobile app**
```bash
cd liftiq-mobile
make start
```

---

## Tear down

```bash
make demo-down
```

This stops all background processes and removes the TimescaleDB container (data is preserved in the Docker volume `timescaledb_data`).

To also delete the database volume:

```bash
cd telemetry-ingestor
docker compose down -v
```

---

## Troubleshooting

**`demo-status` shows `unknown` for all metrics**
The ingestor hasn't written data yet, or the compliance engine's stale window has expired. Check:
```bash
make logs-ingestor
```
Look for `"msg":"poll complete"` lines. If absent, the ingestor can't reach the simulator — confirm the simulator is running on port 8000.

**Compliance engine returns 404 for a unit**
The ingestor hasn't registered that unit yet. The `elevator_units` table is populated on the first poll. Wait 5–10 seconds and retry.

**`make demo-up` fails at TimescaleDB step**
Docker may not be running. Start Docker Desktop and retry.

**iOS Simulator can't reach the compliance engine**
The Expo app uses `http://localhost:8080` by default (correct for iOS Simulator). For Android Emulator, update `liftiq-mobile/.env`:
```
EXPO_PUBLIC_COMPLIANCE_API_URL=http://10.0.2.2:8080
```

**Simulator venv missing**
`demo-start.sh` runs `./setup.sh` automatically if `.venv` is absent. To rebuild manually:
```bash
cd elevator-simulator && rm -rf .venv && make setup
```
