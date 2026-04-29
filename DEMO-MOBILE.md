# LiftIQ — Mobile Demo Runbook

End-to-end demo using the **React Native mobile app** on iOS or Android.

**Stack**: elevator simulator → telemetry ingestor → TimescaleDB → compliance engine → report generator → **mobile app (Expo)**

The mobile app supports the complete inspection workflow:
NFC tap → voice-guided checklist → photo evidence → digital signature → PDF report

---

## Prerequisites

| Tool            | Version    | Notes                                                    |
| --------------- | ---------- | -------------------------------------------------------- |
| Python          | 3.10+      | Required by elevator simulator                           |
| Go              | 1.25+      | Required by ingestor, compliance engine, report gen      |
| Docker          | Any recent | Required for TimescaleDB                                 |
| Node.js         | 18+        | Required for mobile app                                  |
| Chrome/Chromium | Latest     | Required by report generator for PDF rendering           |
| Expo Dev Client | Latest     | **Not** Expo Go — dev build required for voice + camera  |

macOS only: `brew install libpcap` (required by BAC0 — harmless if BACnet is not used)

> **Dev build required**: The mobile app uses `expo-speech-recognition` (native STT) and the
> camera/location APIs. Expo Go does not support these. Run `npx expo run:ios` or
> `npx expo run:android` once to install the dev build, then use `make mobile` for subsequent starts.

---

## Quick start

```bash
# From the repo root — start all backend services
make demo-up
```

This starts all six services in order with health checks between each step.
Logs are written to `/tmp/liftiq-*.log`.

Once the stack is up, start the mobile app:

```bash
make mobile
# Opens Metro bundler — press 'i' for iOS Simulator or scan QR for device
```

For direct launch:

```bash
make mobile-ios       # builds + runs on iOS Simulator / connected device
make mobile-android   # builds + runs on Android Emulator / connected device
```

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

> The web app also starts as part of `make demo-up`. For a mobile-only demo, you can ignore it.

---

## First-time mobile setup

```bash
cd liftiq-mobile
npm install
npx expo prebuild       # generates ios/ and android/ native projects
npx expo run:ios        # builds and installs dev client (first time takes a few minutes)
```

`prebuild` is required because `expo-speech-recognition` links native modules (iOS
`SFSpeechRecognizer`, Android `SpeechRecognizer`). Expo Go does not support it.

### Environment configuration

Edit `liftiq-mobile/.env` to point at the compliance and report engines:

```bash
# iOS Simulator (uses localhost correctly)
EXPO_PUBLIC_COMPLIANCE_API_URL=http://localhost:8080
EXPO_PUBLIC_REPORT_API_URL=http://localhost:8082

# Android Emulator (10.0.2.2 maps to host localhost)
EXPO_PUBLIC_COMPLIANCE_API_URL=http://10.0.2.2:8080
EXPO_PUBLIC_REPORT_API_URL=http://10.0.2.2:8082

# Physical device (use your machine's LAN IP)
EXPO_PUBLIC_COMPLIANCE_API_URL=http://192.168.x.x:8080
EXPO_PUBLIC_REPORT_API_URL=http://192.168.x.x:8082
```

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

Open the mobile app. Tap **Scan Elevator Tag** — the button pulses while "reading" the NFC tag
(1 s simulated scan) then presents the unit picker. Select **ELV-003** (the high-mileage unit).

Show that all 20 ASME A17.1 checks are green — live telemetry, not mock data
(motor current 18+ A, door force 118 N — already near thresholds).

```bash
# Also visible via curl
curl -s http://localhost:8080/units/ELV-003/compliance/summary | python3 -m json.tool
```

### 2. Inject a door fault (2 min)

```bash
make fault-door
```

The `door_close_force_n` value climbs toward the 135 N ASME A17.1 limit.

```bash
make demo-status   # poll every ~10 s to watch the value rise
```

Tap **Refresh** in the app. Once `door_close_force_n` crosses 135 N:

- `ASME-006` flips from **PASS** to **FAIL**
- Overall banner turns **red**
- Summary shows **fail=1**

> *"This is exactly what happens when a real door motor starts wearing out. The inspector gets an automatic flag before they open the panel."*

### 3. Voice-guided inspection (2 min)

Tap the **mic button** at the bottom of the Compliance screen to activate voice mode.

> **iOS Simulator caveat:** `SFSpeechRecognizer` does not function in the iOS Simulator.
> Voice commands require a physical iOS device or Android Emulator.

Say commands aloud:

- **"next"** → advances to the next rule (reads rule aloud via TTS)
- **"pass"** → marks current rule pass
- **"fail"** → marks fail + opens camera for photo evidence
- **"take photo"** → captures photo with GPS tag for the current rule
- **"skip"** → skips without recording

> *"The technician never touches the screen — hands-free compliance logging while physically inspecting the elevator."*

### 4. Photo evidence (1 min)

On `ASME-006`, tap **Photo** (or say "take photo"). The camera opens. Capture a photo of the
door mechanism. It appears as a GPS-tagged thumbnail attached to the rule.

If the rule is failed, the camera opens automatically after the voice "fail" command.

### 5. Manual override (1 min)

Some checks need a physical gauge (e.g., door close force). Tap **Override Pass** on `ASME-006`.
The badge changes to **PASS (manual)**.

> *"The inspector overrides directly in the app. Manual results are captured alongside auto-evaluated ones and appear in the final report."*

### 6. Generate PDF report with digital signature (2 min)

Tap **View Summary** → **Proceed to Sign**.

On the signature screen, draw a signature with your finger. Tap **Sign & Generate PDF**.

The app:

1. Converts all photos to base64
2. POSTs the full inspection to the report generator (`localhost:8082`)
3. The report generator renders an ASME A17.1-formatted HTML page via `html/template`
4. Headless Chrome (chromedp) prints it to PDF
5. The PDF is saved and the system share sheet appears

Open the PDF and show:

- LiftIQ header, unit tag, inspection date
- Pass/fail banner with rule counts
- Full 20-rule table grouped by subsystem (Motor, Door Operator, Brake System, etc.)
- Photo evidence section with GPS coordinates
- **Technician Certification** section with the drawn signature

### 7. Close (30 sec)

Point at the six-layer architecture:

```text
Simulated BACnet elevator
  → Go telemetry ingestor → TimescaleDB
  → Go compliance engine (20 ASME A17.1 rules)
  → React Native app (voice + camera)
  → Go report generator (headless Chrome PDF)
```

In Phase 3 this same pipeline connects to a real building management system.

---

## ISP pilot demo script

For a technician audience, focus on time savings and workflow:

1. Tap **Scan Elevator Tag** → select unit (NFC simulated)
2. Show pre-filled checklist — 17 of 20 items auto-evaluated from live telemetry
3. Activate voice mode: say **"next"** to advance, **"pass"** / **"fail"** to record, **"take photo"** for evidence
4. For items needing physical testing, tap **Override Pass/Fail**
5. Tap **View Summary** → **Proceed to Sign** → draw signature → **Sign & Generate PDF**
6. Share the PDF instantly from the device

> *"You used to fill this out by hand after the inspection and type it up later. Now the checklist
> is pre-filled, you inspect hands-free with voice, photos are GPS-tagged automatically, and the
> signed PDF is generated on-site in seconds."*

---

## Voice command reference

| What you say                   | Intent  | Action                            |
| ------------------------------ | ------- | --------------------------------- |
| "pass" / "looks good" / "ok"  | `pass`  | Mark current rule pass            |
| "fail" / "failed" / "no good" | `fail`  | Mark fail + open camera           |
| "next" / "continue"           | `next`  | Advance to next rule (read aloud) |
| "skip" / "next item"          | `skip`  | Skip current rule, advance        |
| "take photo" / "photo"        | `photo` | Open camera for current rule      |
| "stop" / "done"               | `stop`  | Deactivate voice mode             |

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

### Terminal 6 — Mobile app

```bash
cd liftiq-mobile
npm install         # first time only
npx expo prebuild   # first time only — generates native projects
make ios            # or: make android
```

For subsequent starts (after the dev build is installed):

```bash
make mobile         # starts Metro bundler — press 'i' for iOS
```

---

## Environment variables

| Variable                         | Service                     | Default                 | Description                             |
| -------------------------------- | --------------------------- | ----------------------- | --------------------------------------- |
| `HTTP_PORT`                      | compliance-engine           | `8080`                  | HTTP listen port                        |
| `HTTP_PORT`                      | report-generator            | `8082`                  | HTTP listen port                        |
| `DATABASE_URL`                   | compliance-engine, ingestor | —                       | PostgreSQL DSN                          |
| `STALE_WINDOW_MINUTES`           | compliance-engine           | `10`                    | Age at which telemetry is flagged stale |
| `LOG_LEVEL`                      | all Go services             | `info`                  | `debug` / `info` / `warn` / `error`    |
| `EXPO_PUBLIC_COMPLIANCE_API_URL` | mobile                      | `http://localhost:8080` | Compliance engine base URL              |
| `EXPO_PUBLIC_REPORT_API_URL`     | mobile                      | `http://localhost:8082` | Report generator base URL               |

For Android Emulator, update `liftiq-mobile/.env`:

```bash
EXPO_PUBLIC_COMPLIANCE_API_URL=http://10.0.2.2:8080
EXPO_PUBLIC_REPORT_API_URL=http://10.0.2.2:8082
```

---

## Tear down

```bash
make demo-down
```

This stops all background processes and shuts down the TimescaleDB container (data is preserved
in the Docker volume `timescaledb_data`).

To also delete the database volume:

```bash
cd telemetry-ingestor
docker compose down -v
```

---

## Troubleshooting

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

### `make demo-down` fails with Docker API error

Docker Desktop may not be running, or the socket version may have changed after an upgrade.
Start Docker Desktop and retry. If it still fails, stop the TimescaleDB container manually:

```bash
cd telemetry-ingestor && docker compose down
```

### `make demo-up` fails at TimescaleDB step

Docker may not be running. Start Docker Desktop and retry.

### Voice commands not recognised

The STT engine requires a native dev build. If the mic button does nothing, rebuild:

```bash
cd liftiq-mobile && npx expo run:ios
```

Grant microphone permission when prompted.

> **iOS Simulator caveat:** `SFSpeechRecognizer` does not function in the iOS Simulator.
> Voice commands require a physical iOS device or Android Emulator.

### iOS Simulator can't reach backend services

iOS Simulator uses `localhost` correctly. For a physical device, set both API URLs in
`liftiq-mobile/.env` to your machine's LAN IP (e.g., `http://192.168.1.x:8080`).

### PDF share sheet not appearing

`expo-sharing` requires a physical device or simulator with share targets. On bare iOS Simulator
use AirDrop or Mail to open the PDF, or check the saved path logged in the console.

### Camera not opening

Camera requires a physical device. The iOS Simulator does not have a camera. On Android Emulator,
you can use the simulated camera (configure in AVD settings).

### Simulator venv missing

`demo-start.sh` runs `./setup.sh` automatically if `.venv` is absent. To rebuild manually:

```bash
cd elevator-simulator && rm -rf .venv && make setup
```
