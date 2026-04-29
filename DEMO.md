# LiftIQ Phase 2 — Demo Runbooks

End-to-end demo of the full Phase 2 stack:
**elevator simulator → telemetry ingestor → TimescaleDB → compliance engine → report generator → frontend**

Two demo modes are available. Both use the same backend services — only the frontend differs.

---

## Web demo (recommended for quick demos)

**No mobile device or Expo setup required.** Open a browser on your laptop and go.

```bash
make demo-up
open http://localhost:5173
```

Full runbook: **[DEMO-WEB.md](DEMO-WEB.md)**

---

## Mobile demo (full native experience)

Uses the React Native app with native voice recognition, camera with GPS tagging,
and native share sheet for PDF delivery.

```bash
make demo-up
make mobile        # press 'i' for iOS Simulator
```

Full runbook: **[DEMO-MOBILE.md](DEMO-MOBILE.md)**

> Requires a one-time dev build: `cd liftiq-mobile && npx expo run:ios`

---

## Common backend commands

Both demo modes share the same backend. These commands work regardless of which frontend you use.

### Quick start / stop

```bash
make demo-up              # start all 6 services
make demo-down            # stop all services (prompts to delete data)
make demo-status          # health check + live compliance summary
```

### Fault injection

```bash
make fault-door           # door motor degradation on ELV-003 (→ 135 N)
make fault-brake          # brake wear on ELV-003 (→ 80 ms)
make fault-motor          # motor bearing wear on ELV-003 (→ 20 A)
make fault-safety         # safety circuit intermittent on ELV-003
make fault-leveling       # leveling drift on ELV-003 (→ 12.7 mm)
make clear-fault          # clear any active fault on ELV-003
```

### Logs

```bash
make logs-simulator       # elevator simulator
make logs-ingestor        # telemetry ingestor
make logs-compliance      # compliance engine
make logs-report          # report generator
make logs-web             # web app
```

### Database

```bash
make demo-down-reset      # stop + truncate data (keep schema)
make demo-down-reset-hard # stop + drop all tables
```

---

## Architecture

```text
Simulated BACnet elevator (Python/FastAPI)        :8000
  → Go telemetry ingestor → TimescaleDB           :5432
  → Go compliance engine (20 ASME A17.1 rules)    :8080
  → Go report generator (headless Chrome PDF)      :8082
  → Web app (React/Vite)                           :5173
  → Mobile app (React Native/Expo)                 Metro
```

---

## Choosing between web and mobile

| Consideration       | Web                              | Mobile                               |
| ------------------- | -------------------------------- | ------------------------------------ |
| Setup time          | Zero (just a browser)            | Dev build required (~5 min first run) |
| Voice commands      | Chrome/Edge/Safari only          | iOS device or Android Emulator       |
| Photo capture       | File picker (select existing)    | Native camera with GPS tagging       |
| Signature           | Mouse / trackpad drawing         | Finger drawing on touch screen       |
| PDF delivery        | Browser download                 | Native share sheet                   |
| Best for            | Quick demos, investor pitches    | Field-realistic ISP pilot demos      |
