# LiftIQ Engineering Plan: BMS Integration & Zero-Cost Prototype

**Version 1.0 | March 2026 | Technical Architecture Document**

---

## Executive summary

This document is a hands-on engineering blueprint for building LiftIQ's BMS integration layer and a functional prototype at near-zero cost. It covers the three elevator communication protocols you'll encounter in the field (BACnet/IP, Modbus TCP, CAN bus), provides a concrete tech stack using only open-source tools, and lays out a phased build plan that gets you from a simulated elevator on your laptop to a live pilot in a real building — all before spending a dollar on proprietary licenses.

---

## 1. Understanding the elevator data landscape

Before writing a single line of code, you need to understand how elevator data actually flows in a building. There are three layers between the elevator's physical components and your application.

**Layer 1 — The elevator controller** is the brain of each elevator unit. It manages motor control, door operations, floor dispatch, and safety circuits. Major controller manufacturers include Smartrise, Virginia Controls, Motion Control Engineering (MCE), and the OEM-integrated controllers from Otis, KONE, Schindler, and TK Elevator. Controllers expose data over proprietary serial buses internally (CAN bus is most common) and increasingly over standard protocols externally.

**Layer 2 — The BMS gateway** sits between the elevator controller and the building management system. This is where proprietary elevator data gets translated into standard building automation protocols. The three protocols you will encounter are BACnet/IP (dominant in commercial buildings, ~60% of installations), Modbus TCP (common in older and mid-market buildings, ~25%), and native CAN bus or serial (remaining ~15%, mostly older installations without a gateway).

**Layer 3 — The building management system** is where Honeywell, Siemens, Johnson Controls, or Schneider Electric software aggregates data from all building systems (HVAC, lighting, fire, elevators). LiftIQ plugs in at this layer for buildings that have a BMS, or directly at Layer 2 for buildings that don't.

### Data points available from a modern elevator controller

| Category | Data points | Protocol exposure | Inspection value |
|---|---|---|---|
| Motor / traction | Current draw (amps), voltage, speed (RPM), temperature, run hours, trip count | BACnet AnalogInput / Modbus holding registers | Bearing wear detection, winding degradation, overload trending |
| Door operator | Open/close cycle count, motor amperage, close force (Newtons), timing (ms), obstruction events | BACnet AnalogInput + BinaryInput | #1 cause of service calls — detectable 2-6 weeks before failure |
| Brake system | Engagement count, current draw, pad thickness (if sensor equipped), response time (ms) | BACnet AnalogInput | Safety-critical — ASME A17.1 Section 8.6 requires periodic verification |
| Safety circuit | Door interlock status, governor trip, buffer status, pit switch, car-top switch | BACnet BinaryInput / Modbus discrete inputs | Continuous monitoring catches intermittent faults invisible to periodic inspection |
| Ride quality | Leveling accuracy (mm), vibration (g-force), acceleration/deceleration profiles | BACnet AnalogInput (if IoT-equipped) | ADA compliance — leveling must be within ±½ inch |
| Controller | Fault codes, firmware version, diagnostic logs, safety circuit continuity | BACnet CharacterString / Modbus input registers | Real-time fault classification and OEM-specific troubleshooting |

---

## 2. Protocol-by-protocol integration plan

### 2.1 BACnet/IP (priority 1 — covers ~60% of commercial buildings)

BACnet is an ASHRAE/ANSI/ISO standard (ASHRAE 135) with no royalties or licensing restrictions. It uses an object-oriented data model where every data point is a "BACnet object" with typed properties.

**Open-source stack: BAC0 (Python)**

BAC0 is an async Python 3.10+ library built on BACpypes3 that provides a high-level API for BACnet/IP communication. It is the fastest path to reading elevator data from a BACnet-enabled BMS.

```python
# Install
pip install BAC0

# Connect to BACnet network and discover devices
import BAC0

bacnet = BAC0.connect(ip='192.168.1.100/24')  # your network interface
devices = bacnet.whois()                        # discover all BACnet devices
print(devices)
# Output: [('192.168.1.50', 1001), ('192.168.1.51', 1002)]  # (IP, device_id)

# Read a specific elevator controller's data points
elevator = BAC0.device('192.168.1.50', 1001, bacnet)

# Read motor current (AnalogInput object, instance 0)
motor_current = elevator['analogInput 0'].presentValue
# Read door cycle count (AnalogInput object, instance 10)
door_cycles = elevator['analogInput 10'].presentValue
# Read door interlock status (BinaryInput object, instance 3)
interlock_ok = elevator['binaryInput 3'].presentValue
```

**BACnet object mapping for elevators:** The BACnet standard (Addendum 135-2012bj) actually defines specific elevator object types. Modern controllers may expose `Elevator Group Object` (type 57) and `Escalator Object` (type 58) with standardized properties like `car-position`, `car-direction`, `car-door-status`, and `fault-signals`. Older installations expose raw analog/binary inputs that you'll need to map manually per controller model.

**What you need for the prototype:** Zero hardware cost. BAC0 runs on any laptop. For testing without a real BMS, use the BACnet stack's built-in server simulator to create a virtual elevator controller on your local network.

### 2.2 Modbus TCP (priority 2 — covers ~25% of buildings)

Modbus is simpler than BACnet but less self-describing. Data is organized as numbered registers (16-bit words) with no built-in metadata. You need the controller manufacturer's register map to know what each register means.

**Open-source stack: pymodbus (Python)**

```python
# Install
pip install pymodbus

from pymodbus.client import ModbusTcpClient

client = ModbusTcpClient('192.168.1.60', port=502)
client.connect()

# Read holding registers 0-9 (motor data block)
result = client.read_holding_registers(address=0, count=10, slave=1)
motor_current_raw = result.registers[0]   # Raw 16-bit value
motor_voltage_raw = result.registers[1]
motor_temp_raw = result.registers[2]

# Convert raw to engineering units (scale factors from register map)
motor_current_amps = motor_current_raw * 0.01  # e.g., register value 1523 = 15.23A
motor_temp_celsius = motor_temp_raw * 0.1       # e.g., register value 452 = 45.2°C

# Read discrete inputs (safety circuit status)
safety = client.read_discrete_inputs(address=0, count=8, slave=1)
door_interlock = safety.bits[0]
governor_ok = safety.bits[1]
pit_switch_ok = safety.bits[2]

client.close()
```

**Challenge with Modbus:** Every controller manufacturer uses different register layouts. Smartrise puts motor current at register 40001; MCE might put it at register 30010. You need to build a **register map library** per controller model. This is tedious but becomes a moat — once you've mapped 10-15 popular controllers, competitors can't easily replicate that work.

### 2.3 CAN bus / Serial (priority 3 — legacy installations)

For older controllers without a BMS gateway, you'll encounter direct CAN bus (Controller Area Network) or RS-485 serial connections. These require a physical adapter (USB-to-CAN or USB-to-RS485, ~$30-50) and reverse-engineering of the controller's message format.

**Open-source stack: python-can**

```python
pip install python-can

import can

# Connect via USB-CAN adapter (e.g., PEAK PCAN-USB)
bus = can.interface.Bus(bustype='socketcan', channel='can0', bitrate=250000)

# Listen for messages
for msg in bus:
    # CAN messages have an arbitration ID and up to 8 bytes of data
    if msg.arbitration_id == 0x180:  # Motor status message (varies by controller)
        current = int.from_bytes(msg.data[0:2], 'big') * 0.01
        rpm = int.from_bytes(msg.data[2:4], 'big')
        temp = int.from_bytes(msg.data[4:6], 'big') * 0.1
        print(f"Motor: {current}A, {rpm}RPM, {temp}°C")
```

**For the prototype, skip CAN bus entirely.** Focus on BACnet and Modbus, which cover ~85% of the addressable market and require no hardware purchase.

---

## 3. The zero-cost prototype architecture

Here's how to build a working LiftIQ prototype without spending any money on hardware, cloud services, or licenses.

### 3.1 Tech stack (all free/open-source)

| Component | Technology | Cost | Why this choice |
|---|---|---|---|
| Elevator simulator | Python script generating realistic telemetry | $0 | No real elevator needed. Simulates BACnet objects with drift, faults, and noise. |
| Telemetry ingestion | Go service (Gin) polling simulator via BAC0 | $0 | You already know Go + Gin. Single binary, low resource usage. |
| Data store | PostgreSQL + TimescaleDB extension | $0 | Time-series optimized. Free tier on Timescale Cloud or self-hosted. |
| Embedding / RAG | Python FastAPI + `all-MiniLM-L6-v2` (local) | $0 | Reuse your AlignIQ embedding approach. OEM manuals indexed locally. |
| Compliance engine | Rule engine in Go mapping ASME A17.1 items to data points | $0 | Deterministic rules, not LLM. Safety-critical logic must be auditable. |
| Report generator | Go template engine → PDF via `chromedp` or `wkhtmltopdf` | $0 | Jurisdiction-specific templates rendered to PDF. |
| Mobile app | React Native (Expo) | $0 | Single codebase for iOS + Android. Free Expo dev builds. |
| Voice interaction | Web Speech API (browser) or Whisper (local) | $0 | Browser-native speech recognition for prototype. Whisper for offline mode later. |
| NFC identification | Phone's native NFC reader via React Native NFC library | $0 | Most modern phones have NFC built in. |
| Hosting (prototype) | Local machine or free tier: Railway / Fly.io / Supabase | $0 | No cloud spend until pilot customers. |

### 3.2 System architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    LiftIQ Prototype Architecture                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐    BACnet/IP     ┌──────────────────────┐     │
│  │   Elevator    │◄───────────────►│  Telemetry Ingestion │     │
│  │  Simulator    │    Modbus TCP   │    Service (Go)      │     │
│  │  (Python)     │◄───────────────►│                      │     │
│  └──────────────┘                  └──────────┬───────────┘     │
│                                               │                  │
│                                               ▼                  │
│                                    ┌──────────────────────┐     │
│                                    │   PostgreSQL +        │     │
│                                    │   TimescaleDB         │     │
│                                    │   (time-series store) │     │
│                                    └──────────┬───────────┘     │
│                                               │                  │
│                          ┌────────────────────┼──────────┐      │
│                          ▼                    ▼          ▼      │
│               ┌───────────────┐  ┌────────────────┐  ┌──────┐  │
│               │  Compliance   │  │   RAG Engine    │  │ Fleet│  │
│               │  Engine (Go)  │  │  (FastAPI +     │  │ Ana- │  │
│               │  ASME A17.1   │  │  MiniLM-L6-v2) │  │lytics│  │
│               │  rule mapping │  │  OEM manuals    │  │ (Go) │  │
│               └──────┬────────┘  └───────┬────────┘  └──┬───┘  │
│                      │                   │              │       │
│                      └───────────┬───────┘              │       │
│                                  ▼                      │       │
│                       ┌──────────────────┐              │       │
│                       │   API Gateway    │◄─────────────┘       │
│                       │   (Go + Gin)     │                      │
│                       └────────┬─────────┘                      │
│                                │                                 │
│                                ▼                                 │
│                    ┌──────────────────────┐                     │
│                    │   Mobile App         │                     │
│                    │   (React Native)     │                     │
│                    │   - NFC tap to start │                     │
│                    │   - Voice interaction│                     │
│                    │   - Photo capture    │                     │
│                    │   - Report sign-off  │                     │
│                    └──────────────────────┘                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 The elevator simulator (the key to a $0 prototype)

This is the most important piece. Instead of needing a real elevator, you build a Python script that behaves like a BACnet-enabled elevator controller — emitting realistic telemetry data including gradual drift patterns, intermittent faults, and noise.

```python
"""
elevator_simulator.py
Simulates a BACnet-enabled elevator controller with realistic telemetry.
Runs as a BACnet/IP server on your local network.
"""
import BAC0
import time
import random
import math
from dataclasses import dataclass, field
from typing import Optional

@dataclass
class ElevatorState:
    """Mutable state for one simulated elevator unit."""
    unit_id: str = "ELV-001"
    floor: int = 1
    max_floors: int = 10
    direction: str = "idle"       # "up", "down", "idle"
    door_status: str = "closed"   # "open", "closed", "opening", "closing"

    # Motor telemetry (baseline + drift)
    motor_current_baseline: float = 12.5   # amps at rated load
    motor_current_drift: float = 0.0       # accumulates over time (simulates wear)
    motor_temp_baseline: float = 45.0      # celsius
    motor_rpm: float = 0.0
    motor_run_hours: float = 12450.0
    trip_count: int = 847231

    # Door telemetry
    door_cycle_count: int = 1_204_500
    door_motor_amps: float = 2.1
    door_close_force_n: float = 67.0       # ASME limit is 135N
    door_close_time_ms: int = 3200
    door_obstruction_events: int = 12

    # Brake telemetry
    brake_engagement_count: int = 847231
    brake_current: float = 1.8
    brake_response_ms: int = 45

    # Safety circuit
    door_interlock_ok: bool = True
    governor_ok: bool = True
    buffer_ok: bool = True
    pit_switch_ok: bool = True
    safety_circuit_ok: bool = True

    # Fault injection
    injected_fault: Optional[str] = None

    def tick(self, seconds: float = 1.0):
        """Advance simulation by `seconds`. Call this in your main loop."""
        # Simulate normal operation: random floor changes
        if random.random() < 0.1:  # 10% chance of trip per tick
            self.direction = random.choice(["up", "down"])
            target = random.randint(1, self.max_floors)
            self.floor = target
            self.trip_count += 1
            self.door_cycle_count += 2  # open + close
            self.motor_run_hours += 0.002

        # Motor current: baseline + drift + noise + load variation
        load_factor = random.uniform(0.3, 1.2)  # 30-120% rated load
        noise = random.gauss(0, 0.15)
        self.motor_current_drift += random.gauss(0.0001, 0.00005)  # slow upward drift
        current = (self.motor_current_baseline + self.motor_current_drift) * load_factor + noise
        # Clamp to realistic range
        self.motor_current = max(0, min(current, 25.0))

        # Motor temp: varies with load and ambient
        ambient_cycle = 2.0 * math.sin(time.time() / 3600 * math.pi)  # day/night
        self.motor_temp = self.motor_temp_baseline + (load_factor * 8) + ambient_cycle + random.gauss(0, 0.5)

        # Door close force: gradual increase simulates track wear
        self.door_close_force_n += random.gauss(0.002, 0.001)
        self.door_close_time_ms = int(3200 + random.gauss(0, 50))

        # Brake response: gradual slowdown simulates pad wear
        self.brake_response_ms = int(45 + self.motor_current_drift * 100 + random.gauss(0, 2))

        # Fault injection
        if self.injected_fault == "door_motor_degradation":
            self.door_motor_amps = 2.1 + (self.door_cycle_count % 1000) * 0.003
        elif self.injected_fault == "brake_wear":
            self.brake_response_ms = int(45 + 30 + random.gauss(0, 5))  # 75ms (threshold is 80ms)
        elif self.injected_fault == "motor_bearing_wear":
            self.motor_current_drift += 0.005  # accelerated drift

    def to_bacnet_points(self) -> dict:
        """Return a dict of BACnet object values for the server to expose."""
        return {
            "analogInput:0": round(getattr(self, 'motor_current', self.motor_current_baseline), 2),
            "analogInput:1": round(self.motor_temp, 1),
            "analogInput:2": round(self.motor_rpm, 0),
            "analogInput:3": round(self.motor_run_hours, 1),
            "analogInput:10": self.door_cycle_count,
            "analogInput:11": round(self.door_motor_amps, 2),
            "analogInput:12": round(self.door_close_force_n, 1),
            "analogInput:13": self.door_close_time_ms,
            "analogInput:20": self.brake_engagement_count,
            "analogInput:21": round(self.brake_current, 2),
            "analogInput:22": self.brake_response_ms,
            "analogInput:30": self.trip_count,
            "binaryInput:0": self.door_interlock_ok,
            "binaryInput:1": self.governor_ok,
            "binaryInput:2": self.buffer_ok,
            "binaryInput:3": self.pit_switch_ok,
            "binaryInput:4": self.safety_circuit_ok,
            "multiStateInput:0": {"up": 1, "down": 2, "idle": 3}[self.direction],
            "analogValue:0": self.floor,
        }
```

**To simulate multiple elevators in a building:** Instantiate multiple `ElevatorState` objects with different `unit_id` values and slightly different baselines. Run them in a single process with a shared BACnet server exposing different device instances.

**To simulate faults for demo purposes:** Set `injected_fault` to one of the defined fault types and run the simulator for 10-15 minutes. The telemetry will show gradual degradation that the LiftIQ compliance engine should detect and flag.

---

## 4. Build plan — from laptop to live pilot

### Phase 1: Simulated environment (Weeks 1-4, cost: $0)

**Goal:** End-to-end data flow from simulated elevator → ingestion → storage → checklist pre-fill → mobile app display.

| Week | Deliverable | Technical details |
|---|---|---|
| 1 | Elevator simulator running as BACnet/IP server | Python script with BAC0 server mode. 3 simulated elevators with distinct profiles. Fault injection API. |
| 2 | Telemetry ingestion service | Go service polling simulator every 5 seconds. Writes to TimescaleDB hypertable partitioned by `unit_id` and `timestamp`. |
| 3 | Compliance engine v0.1 | Go rule engine with 20 core ASME A17.1 checklist items mapped to telemetry thresholds (e.g., door close force < 135N = auto-pass, brake response < 80ms = auto-pass). |
| 4 | Mobile app skeleton | React Native app with NFC scan stub, checklist display, and pass/fail buttons. No voice yet. |

**Key architecture decisions for Phase 1:**

The ingestion service should use a **poll-based model** (not subscription/COV) for the prototype. BACnet supports Change of Value (COV) subscriptions, but poll-based is simpler to implement, debug, and reason about. Move to COV in Phase 3 when you need real-time responsiveness.

The compliance engine must be **deterministic, not LLM-based.** Safety-critical pass/fail logic must be auditable and reproducible. Use a rule engine pattern:

```go
// compliance/rules.go
type Rule struct {
    ID          string
    ASMESection string
    Description string
    DataPoint   string
    Operator    string  // "lt", "gt", "eq", "between"
    Threshold   float64
    ThresholdHi float64 // for "between" operator
    AutoPassable bool   // can this be verified from telemetry alone?
}

var CoreRules = []Rule{
    {
        ID:          "DOOR-001",
        ASMESection: "2.13.4",
        Description: "Door closing force shall not exceed 135 Newtons",
        DataPoint:   "door_close_force_n",
        Operator:    "lt",
        Threshold:   135.0,
        AutoPassable: true,
    },
    {
        ID:          "BRAKE-001",
        ASMESection: "8.6.4.1",
        Description: "Brake shall engage within 80ms of signal",
        DataPoint:   "brake_response_ms",
        Operator:    "lt",
        Threshold:   80.0,
        AutoPassable: true,
    },
    {
        ID:          "SAFETY-001",
        ASMESection: "2.26.1",
        Description: "All safety circuit contacts shall be continuous",
        DataPoint:   "safety_circuit_ok",
        Operator:    "eq",
        Threshold:   1.0,
        AutoPassable: true,
    },
    {
        ID:          "LEVEL-001",
        ASMESection: "2.26.1.4",
        Description: "Car leveling accuracy within 1/2 inch of floor",
        DataPoint:   "leveling_accuracy_mm",
        Operator:    "lt",
        Threshold:   12.7, // 1/2 inch = 12.7mm
        AutoPassable: false, // requires physical verification
    },
}
```

### Phase 2: Voice interaction and report generation (Weeks 5-8, cost: $0)

**Goal:** Technician can complete a full inspection via voice commands and receive a generated PDF report.

| Week | Deliverable | Technical details |
|---|---|---|
| 5 | Voice-to-command pipeline | Web Speech API in React Native WebView for recognition. Simple intent parser: "pass", "fail", "skip", "next", "take photo". |
| 6 | Photo evidence capture | React Native camera integration. Photos tagged with checklist item ID, GPS, and timestamp. Stored as base64 in PostgreSQL (prototype) or S3 (production). |
| 7 | Report generator v0.1 | Go HTML template → PDF via headless Chrome (`chromedp`). ASME A17.1 format with telemetry data, pass/fail determinations, and embedded photos. |
| 8 | End-to-end demo flow | NFC tap → pre-filled checklist → voice-guided inspection → photo capture → PDF report → digital signature. |

**Voice interaction architecture for prototype:**

```
Technician speaks → Web Speech API (browser-native, free)
    → Raw transcript: "pass"
    → Intent parser (simple keyword matching, no LLM needed):
        "pass" / "passed" / "looks good" → PASS current item
        "fail" / "failed" / "no good"   → FAIL current item + prompt for photo
        "skip"                           → SKIP, move to next
        "next"                           → Confirm and advance
        "photo" / "picture"              → Open camera
    → Update checklist state
    → Agent reads next item aloud via Web Speech Synthesis API
```

For the prototype, this keyword-based approach is faster and more reliable than an LLM. You can add Whisper + LLM intent parsing in Phase 3 for handling natural language like "the brake pads look worn but within tolerance, I'll pass it for now."

### Phase 3: Real BMS integration (Weeks 9-14, cost: $0-500)

**Goal:** Connect to a real elevator's BMS in a partner building and validate that the data pipeline works with live telemetry.

| Week | Deliverable | Details |
|---|---|---|
| 9-10 | BMS partner secured | Approach 1: Find a friendly ISP through NAEC network who will give you read-only BMS access to one building. Approach 2: Contact a BMS vendor's developer program (Honeywell, Siemens, JCI all have them — often free for startups). |
| 11 | Controller register mapping | On-site visit to document which BACnet objects or Modbus registers correspond to which data points for that specific controller model. This is manual work — bring a laptop with BAC0 and Wireshark. |
| 12 | Live data ingestion | Point the ingestion service at the real BMS IP instead of the simulator. Validate data freshness, accuracy, and completeness. |
| 13-14 | Side-by-side validation | Run a traditional paper inspection alongside LiftIQ. Compare the pre-filled checklist against the technician's manual findings. Measure agreement rate. |

**How to get BMS access for free:**

1. **BMS vendor developer programs:** Honeywell (Niagara Framework) has a free developer license for testing. Siemens (Desigo CC) offers a sandbox environment. Johnson Controls (Metasys) has a partner program. Apply as a technology partner, not a customer.

2. **BACnet simulator tools:** If you can't get real BMS access yet, use YABE (Yet Another BACnet Explorer) — a free, open-source Windows tool that can simulate a full BACnet network with configurable objects. Or use Softdel's BOSS simulator.

3. **University buildings:** Many university facilities departments run BACnet-enabled BMS systems and are willing to give academic/research access to their networks. Approach the building engineering department.

4. **The USB-CAN adapter exception:** If your pilot building has an older controller with CAN bus only, you'll need a ~$30-50 USB-CAN adapter (PEAK PCAN-USB Basic or Canable). This is the only hardware cost in the entire prototype phase.

### Phase 4: OEM knowledge base and RAG (Weeks 12-16, cost: $0)

**Goal:** Build the domain-specific RAG system that maps fault codes to repair procedures.

This runs in parallel with Phase 3. The approach reuses your AlignIQ architecture:

```
OEM manual PDFs → Chunk by section → Embed with all-MiniLM-L6-v2 → Store in pgvector

Technician query ("what does fault code E-47 mean on a Smartrise C4 controller?")
    → Embed query → Cosine similarity search in pgvector
    → Top 3 chunks → Feed to LLM with context for natural language answer
```

**Where to get OEM manuals for free:**

- Smartrise publishes technical manuals on their dealer portal (registration is free for licensed elevator companies).
- Virginia Controls manuals are available through distributor relationships.
- Generic maintenance procedures from ASME A17.1 and the NAEC Maintenance Control Program (MCP) are publicly referenced.
- Many ISPs have binders full of OEM manuals they'd happily let you scan.

**Important legal note:** OEM manuals are typically copyrighted. For the prototype, use them under fair use for internal R&D. For production, negotiate data licensing agreements with each OEM, or build your own knowledge base from publicly available ASME standards, NAEC technical bulletins, and technician-contributed content.

---

## 5. Data model (PostgreSQL + TimescaleDB)

```sql
-- Core tables for the prototype

-- Elevator units
CREATE TABLE elevator_units (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_tag        VARCHAR(50) NOT NULL,          -- NFC tag ID
    building_id     UUID NOT NULL,
    controller_make VARCHAR(100),                   -- "Smartrise", "MCE", etc.
    controller_model VARCHAR(100),                  -- "C4", "iMotion", etc.
    protocol        VARCHAR(20) DEFAULT 'bacnet',   -- "bacnet", "modbus", "can"
    bms_address     VARCHAR(100),                   -- IP:port or device ID
    installed_date  DATE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Time-series telemetry (TimescaleDB hypertable)
CREATE TABLE telemetry (
    time            TIMESTAMPTZ NOT NULL,
    unit_id         UUID NOT NULL REFERENCES elevator_units(id),
    metric          VARCHAR(50) NOT NULL,           -- "motor_current", "door_force", etc.
    value           DOUBLE PRECISION NOT NULL,
    quality         VARCHAR(10) DEFAULT 'good'      -- "good", "stale", "missing"
);
SELECT create_hypertable('telemetry', 'time');
CREATE INDEX idx_telemetry_unit_metric ON telemetry (unit_id, metric, time DESC);

-- Inspections
CREATE TABLE inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id         UUID NOT NULL REFERENCES elevator_units(id),
    technician_id   UUID NOT NULL,
    jurisdiction    VARCHAR(20) DEFAULT 'ASME_A17_1',
    status          VARCHAR(20) DEFAULT 'draft',    -- "draft", "in_progress", "completed", "submitted"
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    signed_at       TIMESTAMPTZ,
    report_pdf_url  TEXT
);

-- Checklist items (per inspection)
CREATE TABLE checklist_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inspection_id   UUID NOT NULL REFERENCES inspections(id),
    rule_id         VARCHAR(20) NOT NULL,           -- "DOOR-001", "BRAKE-001", etc.
    asme_section    VARCHAR(20),
    description     TEXT,
    auto_result     VARCHAR(10),                    -- "pass", "fail", "manual" (pre-filled from telemetry)
    tech_result     VARCHAR(10),                    -- "pass", "fail", "skip" (technician's determination)
    telemetry_value DOUBLE PRECISION,               -- the value that was evaluated
    threshold       DOUBLE PRECISION,
    photo_urls      TEXT[],                          -- array of photo evidence URLs
    voice_note      TEXT,                            -- transcribed voice note
    reviewed_at     TIMESTAMPTZ
);

-- OEM knowledge base (for RAG)
CREATE TABLE oem_knowledge (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    controller_make VARCHAR(100),
    controller_model VARCHAR(100),
    section         TEXT,
    content         TEXT,
    embedding       vector(384)                     -- all-MiniLM-L6-v2 output dimension
);
CREATE INDEX idx_oem_embedding ON oem_knowledge USING ivfflat (embedding vector_cosine_ops);
```

---

## 6. Cost summary

| Phase | Duration | Hard costs | What you get |
|---|---|---|---|
| Phase 1: Simulated environment | Weeks 1-4 | $0 | End-to-end data flow with simulated elevators |
| Phase 2: Voice + reports | Weeks 5-8 | $0 | Full inspection workflow with voice and PDF generation |
| Phase 3: Real BMS integration | Weeks 9-14 | $0-50 (USB-CAN adapter if needed) | Live data from a real building |
| Phase 4: OEM RAG knowledge base | Weeks 12-16 | $0 | Fault code lookup and repair guidance |
| **Total prototype cost** | **16 weeks** | **$0-50** | **Demo-ready product for investor meetings and first ISP pilot** |

The only scenario where you spend money is if your pilot building has a legacy CAN bus controller without a BACnet gateway. Even then, a $30 USB-CAN adapter covers it.

---

## 7. What to demo to investors vs. what to demo to ISPs

**For investors:** Demo the simulator with fault injection. Show a "healthy" elevator, then inject a door motor degradation fault. Watch the checklist auto-flag it. Walk through the voice inspection. Generate the PDF report. This demonstrates the full vision without needing a real building.

**For ISP pilot customers:** Demo with their real BMS data. This requires Phase 3 to be complete. The live data is what converts skeptics — when a technician sees their actual elevator's motor current trending upward on screen, they immediately understand the value.

---

## 8. Risks and mitigations specific to BMS integration

| Risk | Impact | Mitigation |
|---|---|---|
| Controller manufacturer blocks data access | High | Focus on BMS-level integration (Layer 2-3), not controller-level. BMS data is owned by the building, not the OEM. |
| BACnet object IDs vary wildly between buildings | Medium | Build a "discovery and mapping" workflow into the onboarding process. Technician confirms data point mappings on first connection. |
| Telemetry data is too noisy for reliable anomaly detection | Medium | Use rolling averages (1-hour, 24-hour, 7-day) and compare against unit-specific baselines, not absolute thresholds. The simulator should include realistic noise levels. |
| Offline buildings with no network connectivity | Medium | Offline-first architecture: the mobile app stores the pre-filled checklist locally. Telemetry sync happens when the technician returns to connectivity. |
| BMS firewall blocks LiftIQ access | Low | Provide a lightweight on-premise gateway (Docker container on a Raspberry Pi, ~$50) that sits inside the building network and pushes data outbound over HTTPS. No inbound firewall holes needed. |

---

*This is a living document. Update it as you learn from each phase — especially Phase 3, where real-world BMS quirks will teach you things no spec document can.*
