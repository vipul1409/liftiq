# LiftIQ

AI-powered elevator compliance monitoring platform. LiftIQ ingests real-time BMS telemetry from elevator systems, evaluates ASME A17.1 safety rules deterministically, and surfaces compliance status via a mobile app — enabling field inspectors to catch issues before they become violations.

## Architecture

```
elevator-simulator  →  telemetry-ingestor  →  TimescaleDB
                                                    ↓
                              liftiq-mobile  ←  compliance-engine
```

| Service | Stack | Description |
|---|---|---|
| `elevator-simulator` | Python / FastAPI / BAC0 | Simulates 3 elevators over BACnet/IP + HTTP fault-injection API |
| `telemetry-ingestor` | Go / pgx | Polls simulator every 5 s, bulk-inserts into TimescaleDB |
| `compliance-engine` | Go / pgx | Evaluates 20 ASME A17.1 rules against latest telemetry; REST API |
| `liftiq-mobile` | React Native / Expo | Inspector app — scan → voice-guided compliance checklist → summary |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Python 3.10+ (elevator-simulator)
- Go 1.25+ (telemetry-ingestor, compliance-engine)
- Node.js 18+ (liftiq-mobile)
- `brew install libpcap` (macOS, for BACnet)

### Local stack (Docker)

```bash
cp deploy/env/.env.prod.example deploy/env/.env.prod
# edit .env.prod if needed, then:
make compose-local
```

### Manual startup

```bash
# 1. Elevator simulator
cd elevator-simulator
./setup.sh
source .venv/bin/activate
python main.py --no-bacnet

# 2. TimescaleDB + telemetry ingestor
cd telemetry-ingestor
make docker-up
make run

# 3. Compliance engine
cd compliance-engine
make run

# 4. Mobile app (requires a development build for voice — see below)
cd liftiq-mobile
npm install
npx expo prebuild          # generates native iOS/Android projects
npx expo run:ios           # or: npx expo run:android
```

## Demo

See [DEMO.md](DEMO.md) for the full investor and ISP demo runbooks, including fault-injection scenarios that demonstrate near-threshold ASME violations on ELV-003.

```bash
make demo-up      # start full stack with health checks
make demo-status  # live compliance summary
make demo-down    # tear down
```

## Compliance Rules

The compliance engine evaluates 20 ASME A17.1 rules across five subsystems:

| Subsystem | Key threshold |
|---|---|
| Door system | Close force < 135 N (§2.13.4) |
| Brake system | Response < 80 ms (§8.6.4.1) |
| Leveling | Accuracy within 12.7 mm / ½ in (ADA) |
| Safety circuits | Must be continuous (§2.26.1) |
| Motor / drive | Current, temperature, vibration |

## Project Structure

```
liftiq/
├── elevator-simulator/    # Python BACnet simulator + HTTP API
├── telemetry-ingestor/    # Go ingestor → TimescaleDB
├── compliance-engine/     # Go rules engine + REST API
├── liftiq-mobile/         # React Native inspector app
├── deploy/                # Docker Compose overlays (local/dev/prod)
└── scripts/               # Demo lifecycle + DB reset scripts
```

See [CLAUDE.md](CLAUDE.md) for the full developer guide and [LiftIQ_Engineering_Plan_BMS_Integration.md](LiftIQ_Engineering_Plan_BMS_Integration.md) for the product and protocol blueprint.

## Mobile App — Voice Commands

The compliance screen supports hands-free inspection via push-to-talk voice commands. Tap the mic button in the floating VoiceBar, then speak:

| Command | Examples | Action |
|---|---|---|
| `pass` | "pass", "looks good", "ok" | Mark active rule as pass |
| `fail` | "fail", "no good", "bad" | Mark active rule as fail |
| `next` | "next", "continue" | Advance to next rule + read it aloud |
| `skip` | "skip", "ignore" | Skip active rule + advance |
| `photo` | "photo", "camera" | *(Week 6 stub — says "Opening camera.")* |
| `stop` | "stop", "done" | Stop listening |

The active rule is highlighted with a blue left border. TTS reads each rule description and status aloud when advancing. Requires a development build (not Expo Go) — `expo-speech-recognition` uses native iOS/Android speech APIs.

> **Note:** Speech recognition does not work in the iOS Simulator. Test on a physical device or Android Emulator.

## Build Phases

| Phase | Scope | Status |
|---|---|---|
| Phase 1 (Weeks 1–4) | Simulated environment, ingestor, compliance engine, mobile app | Complete |
| Phase 2 — Week 5 | Voice-to-command pipeline (STT + TTS + intent parser) | Complete |
| Phase 2 — Week 6 | Photo evidence capture | Not started |
| Phase 2 — Week 7 | PDF report generator | Not started |
| Phase 2 — Week 8 | End-to-end demo flow | Not started |
| Phase 3 (Weeks 9–14) | Real BMS / BACnet integration | Not started |
| Phase 4 (Weeks 12–16) | OEM RAG knowledge base | Not started |

## License

Proprietary. All rights reserved.
