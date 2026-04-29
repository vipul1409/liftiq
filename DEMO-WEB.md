# LiftIQ — Web Demo Runbook

End-to-end demo using the **browser-based web app** — no mobile device or Expo setup required.

**Stack**: elevator simulator → telemetry ingestor → TimescaleDB → compliance engine → report generator → **web app (React)**

The web app supports the complete inspection workflow:
scan → voice-guided checklist → photo evidence → digital signature → PDF report download

---

## Prerequisites

| Tool            | Version    | Notes                                               |
| --------------- | ---------- | --------------------------------------------------- |
| Python          | 3.10+      | Required by elevator simulator                      |
| Go              | 1.25+      | Required by ingestor, compliance engine, report gen |
| Docker          | Any recent | Required for TimescaleDB                            |
| Node.js         | 18+        | Required for web app                                |
| Chrome/Chromium | Latest     | Required by report generator for PDF rendering      |

macOS only: `brew install libpcap` (required by BAC0 — harmless if BACnet is not used)

> **No mobile device needed.** The web app runs in any modern browser at `http://localhost:5173`.
> Voice commands work in Chrome, Edge, and Safari (via Web Speech API).

---

## Quick start

```bash
# From the repo root
make demo-up
```

This starts all six services in order with health checks between each step.
Once complete, open your browser:

```text
http://localhost:5173
```

That's it — the web app is ready. No additional setup needed.

---

## What `make demo-up` starts

| Step | Service              | Port | Wait condition                    |
| ---- | -------------------- | ---- | --------------------------------- |
| 1    | TimescaleDB (Docker) | 5432 | `pg_isready`                      |
| 2    | Elevator simulator   | 8000 | `GET /health` → 200               |
| 3    | Telemetry ingestor   | —    | process started (polls every 5 s) |
| 4    | Compliance engine    | 8080 | `GET /health` → 200               |
| 5    | Report generator     | 8082 | `GET /health` → 200               |
| 6    | Web app (Vite)       | 5173 | `GET /` → 200                     |

After ~10 seconds the ingestor writes the first batch of rows and the compliance engine begins
serving live results.

---

## Verify the stack is healthy

```bash
make demo-status
```

Expected output:

```text
── Service health ───────────────────────────────
  ✓ Elevator simulator   (http://localhost:8000/health)
  ✓ Compliance engine    (http://localhost:8080/health)
  ✓ Report generator     (http://localhost:8082/health)

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

Recommended sequence for investors. Shows the full end-to-end workflow:
live telemetry → fault → voice inspection → PDF report with digital signature.

### 1. Establish baseline (2 min)

Open `http://localhost:5173` in Chrome. Click **Scan Elevator** — the button pulses while
"reading" the NFC tag (1 s simulated scan) then presents the unit picker.
Select **ELV-003** (the high-mileage unit).

Show that all 20 ASME A17.1 checks are green — live telemetry, not mock data
(motor current 18+ A, door force 118 N — already near thresholds).

The rules are grouped into six subsystems with sticky section headers:
Motor, Trip / Usage, Door Operator, Brake System, Ride Quality, Safety Circuits.

```bash
# Also visible via curl
curl -s http://localhost:8080/units/ELV-003/compliance/summary | python3 -m json.tool
```

### 2. Inject a door fault (2 min)

In a terminal:

```bash
make fault-door
```

The `door_close_force_n` value climbs toward the 135 N ASME A17.1 limit.

```bash
make demo-status   # poll every ~10 s to watch the value rise
```

Click **Refresh** in the web app. Once `door_close_force_n` crosses 135 N:

- `ASME-006` flips from **PASS** to **FAIL**
- The footer summary updates to show **fail=1**
- The "View Summary" button badge turns red

> *"This is exactly what happens when a real door motor starts wearing out. The inspector gets an automatic flag before they open the panel."*

### 3. Voice-guided inspection (2 min)

Click the **mic button** in the floating voice bar at the bottom of the Compliance screen.

> **Note:** Voice commands require Chrome, Edge, or Safari. The browser will prompt for
> microphone permission on first use.

Speak commands aloud:

- **"next"** → advances to the next rule (reads rule aloud via TTS)
- **"pass"** → marks current rule pass
- **"fail"** → marks current rule fail
- **"take photo"** → opens file picker for evidence photo
- **"skip"** → skips without recording
- **"stop"** → deactivates voice mode

The active rule is highlighted with a blue left border. Voice navigation moves through all 20
rules in order, reading each one aloud.

> *"The technician never touches the screen — hands-free compliance logging while physically inspecting the elevator."*

### 4. Photo evidence (1 min)

On `ASME-006`, click **Photo** (or say "take photo"). A file picker opens. Select any image
file — it appears as a thumbnail attached to the rule.

Multiple photos can be attached to a single rule. Each photo shows a small thumbnail in the
evidence strip below the rule.

> **Tip for demos:** Prepare a few elevator photos on the desktop ahead of time for quick selection.

### 5. Manual override (1 min)

Some checks need a physical gauge (e.g., door close force). Click **Override Pass** on `ASME-006`.
The badge changes to **PASS (manual)**.

Click **Override Fail** on a passing rule to demonstrate the reverse. Click **Clear** to revert
to the telemetry status.

> *"The inspector overrides directly in the app. Manual results are captured alongside auto-evaluated ones and appear in the final report."*

### 6. Generate PDF report with digital signature (2 min)

Click **View Summary** in the sticky footer → review pass/fail/unknown counts.

Click **Proceed to Sign**.

On the signature screen, draw a signature with your mouse (or trackpad). Click **Sign & Generate PDF**.

The app:

1. POSTs the full inspection (rules, overrides, photos, signature) to the report generator
2. The report generator renders an ASME A17.1-formatted HTML page
3. Headless Chrome prints it to PDF
4. The browser downloads the PDF automatically

Open the downloaded PDF and show:

- LiftIQ header, unit tag, inspection date
- Pass/fail banner with rule counts
- Full 20-rule table grouped by subsystem (Motor, Door Operator, Brake System, etc.)
- Photo evidence section with GPS coordinates (if available)
- **Technician Certification** section with the drawn signature

### 7. Close (30 sec)

Point at the six-layer architecture:

```text
Simulated BACnet elevator
  → Go telemetry ingestor → TimescaleDB
  → Go compliance engine (20 ASME A17.1 rules)
  → React web app (voice + photo evidence)
  → Go report generator (headless Chrome PDF)
```

In Phase 3 this same pipeline connects to a real building management system.

---

## ISP pilot demo script

For a technician audience, focus on time savings and workflow:

1. Open `http://localhost:5173` → click **Scan Elevator** → select unit
2. Show pre-filled checklist — 17 of 20 items auto-evaluated from live telemetry
3. Activate voice mode: say **"next"** to advance, **"pass"** / **"fail"** to record, **"take photo"** for evidence
4. For items needing physical testing, click **Override Pass/Fail**
5. Click **View Summary** → **Proceed to Sign** → draw signature → **Sign & Generate PDF**
6. PDF downloads automatically — share via email or print

> *"You used to fill this out by hand after the inspection and type it up later. Now the checklist
> is pre-filled, you inspect hands-free with voice, photos are attached on-site, and the signed
> PDF downloads in seconds."*

---

## Voice command reference

| What you say                   | Intent  | Action                                    |
| ------------------------------ | ------- | ----------------------------------------- |
| "pass" / "looks good" / "ok"  | `pass`  | Mark current rule pass                    |
| "fail" / "failed" / "no good" | `fail`  | Mark fail                                 |
| "next" / "continue"           | `next`  | Advance to next rule (read aloud)         |
| "skip" / "next item"          | `skip`  | Skip current rule, advance                |
| "take photo" / "photo"        | `photo` | Open file picker for current rule         |
| "stop" / "done"               | `stop`  | Deactivate voice mode                     |

> **Browser support:** Voice requires `SpeechRecognition` API — Chrome, Edge, Safari 14.1+.
> On Firefox, the mic button shows "Voice not supported in this browser" and is disabled.

---

## Fault types reference

| Command               | Fault                       | ASME rule that trips | Threshold                  |
| --------------------- | --------------------------- | -------------------- | -------------------------- |
| `make fault-door`     | Door motor degradation      | ASME-006             | 135 N close force          |
| `make fault-brake`    | Brake wear                  | ASME-011             | 80 ms response time        |
| `make fault-motor`    | Motor bearing wear          | ASME-001             | 20 A motor current         |
| `make fault-safety`   | Safety circuit intermittent | ASME-020             | circuit must be continuous |
| `make fault-leveling` | Leveling drift              | ASME-014             | 12.7 mm (half inch ADA)   |

```bash
make clear-fault   # clear any active fault on ELV-003
```

---

## Manual startup (step-by-step)

If you prefer separate terminals to see logs live:

### Terminal 1 — TimescaleDB

```bash
cd telemetry-ingestor
make docker-up
```

### Terminal 2 — Elevator simulator

```bash
cd elevator-simulator
make setup          # first time only
make run-no-bacnet
```

### Terminal 3 — Telemetry ingestor

```bash
cd telemetry-ingestor
make run
```

### Terminal 4 — Compliance engine

```bash
cd compliance-engine
make run
```

### Terminal 5 — Report generator

```bash
cd report-generator
make run
```

### Terminal 6 — Web app

```bash
cd liftiq-web
npm install         # first time only
make dev            # starts at http://localhost:5173
```

---

## Standalone web development

To iterate on the web app without `make demo-up`:

```bash
# Start backends individually (or via make demo-up)
# Then in a separate terminal:
make web            # starts Vite dev server with HMR at :5173
```

The Vite dev server proxies API requests:
- `/api/compliance/*` → `http://localhost:8080`
- `/api/reports/*` → `http://localhost:8082`

---

## Tear down

```bash
make demo-down
```

This stops all background processes (including the web app) and shuts down the TimescaleDB
container (data is preserved in the Docker volume `timescaledb_data`).

To also delete the database volume:

```bash
cd telemetry-ingestor
docker compose down -v
```

---

## Troubleshooting

### Web app shows "Connection Error"

The compliance engine is not reachable. Check that the backend is running:

```bash
make demo-status
```

If the backends are running but the web app can't reach them, the Vite proxy may not be working.
Check the web app log:

```bash
make logs-web
```

### `demo-status` shows `unknown` for all metrics

The ingestor hasn't written data yet. Check:

```bash
make logs-ingestor
```

Look for `"msg":"poll complete"` lines. If absent, confirm the simulator is running on port 8000.

### Report generator returns 500

Chrome/Chromium must be installed and discoverable. On macOS:

```bash
brew install --cask google-chrome
```

The `chromedp` library auto-discovers Chrome on `$PATH` and in standard install locations.

### Voice commands not working in browser

Voice commands require Chrome, Edge, or Safari 14.1+. Firefox does not support the
Web Speech API. The mic button will show "Voice not supported in this browser" and be disabled.

If using Chrome and voice still doesn't work:
1. Check that microphone permission is granted (click the lock icon in the address bar)
2. Make sure you're accessing the app via `localhost` (Web Speech API may not work on plain IPs)

### PDF download doesn't start

The report generator must be running on port 8082. Check:

```bash
curl -s http://localhost:8082/health | python3 -m json.tool
```

Also check the browser's developer console (F12) for errors.

### `make demo-down` fails with Docker API error

Docker Desktop may not be running, or the socket version may have changed after an upgrade.
Start Docker Desktop and retry. If it still fails, stop the TimescaleDB container manually:

```bash
cd telemetry-ingestor && docker compose down
```

### Simulator venv missing

`demo-start.sh` runs `./setup.sh` automatically if `.venv` is absent. To rebuild manually:

```bash
cd elevator-simulator && rm -rf .venv && make setup
```
