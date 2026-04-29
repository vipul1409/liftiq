"""
elevator_state.py
ElevatorState dataclass and simulation logic.
Simulates a BACnet-enabled elevator controller with realistic telemetry
including gradual drift patterns, intermittent faults, and noise.
"""
import math
import random
import time
from dataclasses import dataclass, field
from typing import Optional


FAULT_TYPES = {
    "door_motor_degradation",
    "brake_wear",
    "motor_bearing_wear",
    "safety_circuit_intermittent",
    "leveling_drift",
}


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
    motor_run_hours: float = 5200.0
    trip_count: int = 125_000

    # Door telemetry
    door_cycle_count: int = 480_000
    door_motor_amps: float = 2.1
    door_close_force_n: float = 67.0       # ASME limit is 135N
    door_close_time_ms: int = 3200
    door_obstruction_events: int = 2

    # Brake telemetry
    brake_engagement_count: int = 125_000
    brake_current: float = 1.8
    brake_response_ms: int = 45

    # Ride quality
    leveling_accuracy_mm: float = 3.2      # ADA limit is 12.7mm (0.5 inch)
    vibration_g: float = 0.05

    # Safety circuit
    door_interlock_ok: bool = True
    governor_ok: bool = True
    buffer_ok: bool = True
    pit_switch_ok: bool = True
    safety_circuit_ok: bool = True

    # Internal computed fields (updated on each tick)
    motor_current: float = field(init=False)
    motor_temp: float = field(init=False)

    # Fault injection
    injected_fault: Optional[str] = None

    def __post_init__(self):
        self.motor_current = self.motor_current_baseline
        self.motor_temp = self.motor_temp_baseline

    def tick(self, seconds: float = 1.0) -> None:
        """Advance simulation by `seconds`. Call this in your main loop."""
        # Simulate normal operation: random floor changes
        if random.random() < 0.1 * seconds:  # 10% chance of trip per second
            self.direction = random.choice(["up", "down"])
            self.floor = random.randint(1, self.max_floors)
            self.trip_count += 1
            self.brake_engagement_count += 1
            self.door_cycle_count += 2  # one open + one close per trip
            self.motor_run_hours += 0.002 * seconds

        # Door status cycling
        if random.random() < 0.05 * seconds:
            self.door_status = random.choice(["open", "closed", "opening", "closing"])

        # Motor current: baseline + drift + noise + load variation
        load_factor = random.uniform(0.3, 1.2)  # 30–120% rated load
        noise = random.gauss(0, 0.15)
        self.motor_current_drift += random.gauss(0.0001, 0.00005) * seconds
        current = (self.motor_current_baseline + self.motor_current_drift) * load_factor + noise
        self.motor_current = max(0.0, min(current, 25.0))

        # Motor temp: varies with load and ambient
        ambient_cycle = 2.0 * math.sin(time.time() / 3600 * math.pi)  # day/night cycle
        self.motor_temp = (
            self.motor_temp_baseline
            + (load_factor * 8)
            + ambient_cycle
            + random.gauss(0, 0.5)
        )

        # Motor RPM (idle when direction == "idle")
        if self.direction == "idle":
            self.motor_rpm = 0.0
        else:
            self.motor_rpm = random.gauss(1450, 20)

        # Door close force: gradual increase simulates track wear
        self.door_close_force_n += random.gauss(0.002, 0.001) * seconds
        self.door_close_time_ms = int(3200 + random.gauss(0, 50))

        # Brake: gradual slowdown simulates pad wear
        self.brake_response_ms = int(
            45 + self.motor_current_drift * 100 + random.gauss(0, 2)
        )

        # Leveling accuracy drift
        self.leveling_accuracy_mm = abs(random.gauss(3.2, 0.8) + self.motor_current_drift * 5)

        # Vibration
        self.vibration_g = abs(random.gauss(0.05, 0.01) + self.motor_current_drift * 0.1)

        # Apply injected faults
        self._apply_fault()

    def _apply_fault(self) -> None:
        if self.injected_fault == "door_motor_degradation":
            # Door motor current rises as door cycles accumulate
            self.door_motor_amps = 2.1 + (self.door_cycle_count % 10000) * 0.0003
            self.door_close_force_n += 0.05  # faster force increase
        elif self.injected_fault == "brake_wear":
            # Brake response approaching limit (80 ms threshold)
            self.brake_response_ms = int(75 + random.gauss(0, 3))
        elif self.injected_fault == "motor_bearing_wear":
            # Accelerated motor current drift and vibration
            self.motor_current_drift += 0.005
            self.vibration_g = abs(random.gauss(0.18, 0.03))
        elif self.injected_fault == "safety_circuit_intermittent":
            # Random intermittent safety circuit trip
            if random.random() < 0.05:
                self.door_interlock_ok = False
                self.safety_circuit_ok = False
            else:
                self.door_interlock_ok = True
                self.safety_circuit_ok = True
        elif self.injected_fault == "leveling_drift":
            # Leveling creeping beyond ADA 12.7 mm threshold
            self.leveling_accuracy_mm = abs(random.gauss(11.0, 1.5))

    def inject_fault(self, fault_type: Optional[str]) -> None:
        """Set or clear a fault. Pass None to clear."""
        if fault_type is not None and fault_type not in FAULT_TYPES:
            raise ValueError(f"Unknown fault type '{fault_type}'. Valid: {FAULT_TYPES}")
        self.injected_fault = fault_type

    def to_bacnet_points(self) -> dict:
        """Return a flat dict of BACnet object key → present value."""
        return {
            # Motor
            "analogInput:0": round(self.motor_current, 2),
            "analogInput:1": round(self.motor_temp, 1),
            "analogInput:2": round(self.motor_rpm, 0),
            "analogInput:3": round(self.motor_run_hours, 1),
            # Door
            "analogInput:10": self.door_cycle_count,
            "analogInput:11": round(self.door_motor_amps, 2),
            "analogInput:12": round(self.door_close_force_n, 1),
            "analogInput:13": self.door_close_time_ms,
            "analogInput:14": self.door_obstruction_events,
            # Brake
            "analogInput:20": self.brake_engagement_count,
            "analogInput:21": round(self.brake_current, 2),
            "analogInput:22": self.brake_response_ms,
            # Ride quality
            "analogInput:30": round(self.leveling_accuracy_mm, 2),
            "analogInput:31": round(self.vibration_g, 4),
            # Trips
            "analogInput:40": self.trip_count,
            # Safety circuit (binary inputs)
            "binaryInput:0": self.door_interlock_ok,
            "binaryInput:1": self.governor_ok,
            "binaryInput:2": self.buffer_ok,
            "binaryInput:3": self.pit_switch_ok,
            "binaryInput:4": self.safety_circuit_ok,
            # Multi-state: direction (1=up, 2=down, 3=idle)
            "multiStateInput:0": {"up": 1, "down": 2, "idle": 3}[self.direction],
            # Multi-state: door status (1=open, 2=closed, 3=opening, 4=closing)
            "multiStateInput:1": {"open": 1, "closed": 2, "opening": 3, "closing": 4}[
                self.door_status
            ],
            # Analog value: current floor
            "analogValue:0": self.floor,
        }

    def to_dict(self) -> dict:
        """Flat dict of all state fields for the HTTP API."""
        return {
            "unit_id": self.unit_id,
            "floor": self.floor,
            "max_floors": self.max_floors,
            "direction": self.direction,
            "door_status": self.door_status,
            "motor_current_a": round(self.motor_current, 2),
            "motor_temp_c": round(self.motor_temp, 1),
            "motor_rpm": round(self.motor_rpm, 0),
            "motor_run_hours": round(self.motor_run_hours, 1),
            "trip_count": self.trip_count,
            "door_cycle_count": self.door_cycle_count,
            "door_motor_amps": round(self.door_motor_amps, 2),
            "door_close_force_n": round(self.door_close_force_n, 1),
            "door_close_time_ms": self.door_close_time_ms,
            "door_obstruction_events": self.door_obstruction_events,
            "brake_engagement_count": self.brake_engagement_count,
            "brake_current_a": round(self.brake_current, 2),
            "brake_response_ms": self.brake_response_ms,
            "leveling_accuracy_mm": round(self.leveling_accuracy_mm, 2),
            "vibration_g": round(self.vibration_g, 4),
            "door_interlock_ok": self.door_interlock_ok,
            "governor_ok": self.governor_ok,
            "buffer_ok": self.buffer_ok,
            "pit_switch_ok": self.pit_switch_ok,
            "safety_circuit_ok": self.safety_circuit_ok,
            "injected_fault": self.injected_fault,
        }


def make_elevator_fleet() -> list[ElevatorState]:
    """Create 3 elevators with distinct wear profiles as specified in the Phase 1 plan.

    All three start GREEN (all 20 checks pass).  Counter/service thresholds:
      motor_run_hours < 20,000    trip_count < 500,000
      door_cycle_count < 2,000,000    brake_engagement_count < 500,000
      door_obstruction_events ≤ 10    brake_response_ms ≤ 80
    ELV-003 is tuned *just below* the limits so fault injection quickly trips them.
    """
    return [
        ElevatorState(
            unit_id="ELV-001",
            # New unit — comfortable margin on everything
            motor_current_baseline=12.5,
            motor_current_drift=0.0,
            motor_temp_baseline=45.0,
            motor_run_hours=5_200.0,
            trip_count=125_000,
            door_cycle_count=480_000,
            door_close_force_n=67.0,
            door_obstruction_events=2,
            brake_engagement_count=125_000,
            brake_response_ms=45,
            floor=1,
            max_floors=10,
        ),
        ElevatorState(
            unit_id="ELV-002",
            # Moderately worn unit — elevated baselines but still passing
            motor_current_baseline=13.8,
            motor_current_drift=0.12,
            motor_temp_baseline=48.5,
            motor_run_hours=14_800.0,
            trip_count=320_000,
            door_cycle_count=1_450_000,
            door_close_force_n=98.0,
            door_obstruction_events=6,
            brake_engagement_count=340_000,
            brake_response_ms=58,
            floor=5,
            max_floors=10,
        ),
        ElevatorState(
            unit_id="ELV-003",
            # High-mileage unit — just below thresholds; fault injection pushes over
            motor_current_baseline=15.1,
            motor_current_drift=0.28,       # kept below 0.35 so brake stays ≤80
            motor_temp_baseline=52.0,
            motor_run_hours=19_200.0,       # approaching 20,000h limit
            trip_count=480_000,             # approaching 500,000 limit
            door_cycle_count=1_920_000,     # approaching 2,000,000 limit
            door_close_force_n=118.0,       # approaching 135N limit
            door_obstruction_events=8,      # approaching 10 limit
            brake_engagement_count=475_000, # approaching 500,000 limit
            brake_response_ms=71,           # approaching 80ms limit
            floor=8,
            max_floors=10,
        ),
    ]
