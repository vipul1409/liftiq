"""
test_elevator_state.py
Unit tests for ElevatorState simulation logic and make_elevator_fleet().

Tests use random.seed() for deterministic results when exercising
stochastic behaviour, or run enough iterations that statistical
properties hold with overwhelming probability.
"""
import random
import unittest.mock as mock

import pytest

from elevator_state import ElevatorState, FAULT_TYPES, make_elevator_fleet

# ---------------------------------------------------------------------------
# ASME / ADA safety thresholds (mirrors compliance/rules.go targets)
# ---------------------------------------------------------------------------
DOOR_FORCE_LIMIT_N = 135.0   # ASME A17.1 § 2.13.4
BRAKE_RESPONSE_LIMIT_MS = 80  # ASME A17.1 § 8.6.4.1
LEVELING_LIMIT_MM = 12.7     # ADA (½ inch)


# ===========================================================================
# Initialisation
# ===========================================================================

class TestInit:
    def test_motor_current_set_from_baseline(self):
        e = ElevatorState(motor_current_baseline=15.0)
        assert e.motor_current == 15.0

    def test_motor_temp_set_from_baseline(self):
        e = ElevatorState(motor_temp_baseline=50.0)
        assert e.motor_temp == 50.0

    def test_default_no_fault(self):
        e = ElevatorState()
        assert e.injected_fault is None

    def test_default_safety_circuit_healthy(self):
        e = ElevatorState()
        assert e.door_interlock_ok is True
        assert e.governor_ok is True
        assert e.buffer_ok is True
        assert e.pit_switch_ok is True
        assert e.safety_circuit_ok is True


# ===========================================================================
# tick() — monotonic / bounded invariants
# ===========================================================================

class TestTickInvariants:
    """Properties that must hold on every tick regardless of random seed."""

    def test_motor_current_always_in_range(self):
        random.seed(0)
        e = ElevatorState()
        for _ in range(1_000):
            e.tick(1.0)
            assert 0.0 <= e.motor_current <= 25.0, (
                f"motor_current {e.motor_current} out of [0, 25]"
            )

    def test_motor_drift_always_increases(self):
        random.seed(42)
        e = ElevatorState(motor_current_drift=0.0)
        initial = e.motor_current_drift
        for _ in range(200):
            e.tick(1.0)
        assert e.motor_current_drift > initial

    def test_trip_count_non_decreasing(self):
        random.seed(1)
        e = ElevatorState(trip_count=0)
        prev = 0
        for _ in range(200):
            e.tick(1.0)
            assert e.trip_count >= prev
            prev = e.trip_count

    def test_door_cycle_count_non_decreasing(self):
        random.seed(2)
        e = ElevatorState(door_cycle_count=0)
        prev = 0
        for _ in range(200):
            e.tick(1.0)
            assert e.door_cycle_count >= prev
            prev = e.door_cycle_count

    def test_motor_run_hours_non_decreasing(self):
        random.seed(3)
        e = ElevatorState(motor_run_hours=0.0)
        prev = 0.0
        for _ in range(200):
            e.tick(1.0)
            assert e.motor_run_hours >= prev
            prev = e.motor_run_hours

    def test_floor_always_within_bounds(self):
        random.seed(5)
        e = ElevatorState(max_floors=10)
        for _ in range(500):
            e.tick(1.0)
            assert 1 <= e.floor <= e.max_floors

    def test_direction_always_valid(self):
        valid = {"up", "down", "idle"}
        random.seed(7)
        e = ElevatorState()
        for _ in range(200):
            e.tick(1.0)
            assert e.direction in valid

    def test_door_status_always_valid(self):
        valid = {"open", "closed", "opening", "closing"}
        random.seed(8)
        e = ElevatorState()
        for _ in range(200):
            e.tick(1.0)
            assert e.door_status in valid

    def test_leveling_accuracy_non_negative(self):
        random.seed(9)
        e = ElevatorState()
        for _ in range(200):
            e.tick(1.0)
            assert e.leveling_accuracy_mm >= 0.0

    def test_vibration_non_negative(self):
        random.seed(10)
        e = ElevatorState()
        for _ in range(200):
            e.tick(1.0)
            assert e.vibration_g >= 0.0


# ===========================================================================
# tick() — directional / RPM coupling
# ===========================================================================

class TestTickRPM:
    def test_idle_direction_gives_zero_rpm(self):
        e = ElevatorState(direction="idle")
        # Patch random so no trip fires (random.random returns > 0.1)
        with mock.patch("elevator_state.random.random", return_value=0.5):
            with mock.patch("elevator_state.random.uniform", return_value=1.0):
                with mock.patch("elevator_state.random.gauss", return_value=0.0):
                    e.tick(1.0)
        assert e.motor_rpm == 0.0

    def test_moving_direction_gives_nonzero_rpm(self):
        random.seed(42)
        # Run many ticks; at some point direction will be non-idle
        e = ElevatorState()
        non_idle_rpm_seen = False
        for _ in range(200):
            e.tick(1.0)
            if e.direction != "idle" and e.motor_rpm > 0:
                non_idle_rpm_seen = True
                break
        assert non_idle_rpm_seen


# ===========================================================================
# tick() — door force drifts upward on average
# ===========================================================================

class TestTickDoorForceDrift:
    def test_door_force_increases_over_time(self):
        """Average drift is ~0.002 N/s; after 500 ticks it should be detectable."""
        random.seed(42)
        e = ElevatorState(door_close_force_n=67.0)
        for _ in range(500):
            e.tick(1.0)
        assert e.door_close_force_n > 67.0


# ===========================================================================
# inject_fault()
# ===========================================================================

class TestInjectFault:
    def test_sets_fault(self):
        e = ElevatorState()
        e.inject_fault("brake_wear")
        assert e.injected_fault == "brake_wear"

    def test_clears_fault_with_none(self):
        e = ElevatorState()
        e.inject_fault("brake_wear")
        e.inject_fault(None)
        assert e.injected_fault is None

    def test_raises_on_unknown_fault(self):
        e = ElevatorState()
        with pytest.raises(ValueError, match="Unknown fault"):
            e.inject_fault("nonexistent_fault")

    def test_raises_does_not_change_state(self):
        e = ElevatorState()
        e.inject_fault("brake_wear")
        with pytest.raises(ValueError):
            e.inject_fault("bad_fault")
        assert e.injected_fault == "brake_wear"  # unchanged

    @pytest.mark.parametrize("fault", sorted(FAULT_TYPES))
    def test_all_known_faults_accepted(self, fault):
        e = ElevatorState()
        e.inject_fault(fault)
        assert e.injected_fault == fault


# ===========================================================================
# Fault effects — statistical assertions
# ===========================================================================

class TestFaultEffects:
    """Each fault must produce a measurable change in the target telemetry."""

    def test_brake_wear_raises_response_ms(self):
        random.seed(42)
        e = ElevatorState(brake_response_ms=45)
        e.inject_fault("brake_wear")
        samples = [e.tick(1.0) or e.brake_response_ms for _ in range(100)]
        mean = sum(samples) / len(samples)
        # Fault holds response around 75 ms; healthy baseline is 45 ms
        assert mean > 60

    def test_motor_bearing_wear_accelerates_drift(self):
        random.seed(42)
        e_healthy = ElevatorState(motor_current_drift=0.0)
        e_faulty  = ElevatorState(motor_current_drift=0.0)
        e_faulty.inject_fault("motor_bearing_wear")
        for _ in range(50):
            e_healthy.tick(1.0)
            e_faulty.tick(1.0)
        assert e_faulty.motor_current_drift > e_healthy.motor_current_drift * 5

    def test_motor_bearing_wear_raises_vibration(self):
        random.seed(42)
        e = ElevatorState()
        e.inject_fault("motor_bearing_wear")
        samples = [e.tick(1.0) or e.vibration_g for _ in range(80)]
        mean = sum(samples) / len(samples)
        # Baseline ~0.05 g; fault targets ~0.18 g
        assert mean > 0.10

    def test_leveling_drift_raises_accuracy(self):
        random.seed(42)
        e = ElevatorState(leveling_accuracy_mm=3.2)
        e.inject_fault("leveling_drift")
        samples = [e.tick(1.0) or e.leveling_accuracy_mm for _ in range(80)]
        mean = sum(samples) / len(samples)
        # Baseline ~3.2 mm; fault targets ~11 mm
        assert mean > 7.0

    def test_door_motor_degradation_increases_force_faster(self):
        random.seed(42)
        e_normal = ElevatorState(door_close_force_n=67.0)
        e_faulty = ElevatorState(door_close_force_n=67.0)
        e_faulty.inject_fault("door_motor_degradation")
        for _ in range(100):
            e_normal.tick(1.0)
            e_faulty.tick(1.0)
        assert e_faulty.door_close_force_n > e_normal.door_close_force_n

    def test_door_motor_degradation_raises_door_motor_amps(self):
        e = ElevatorState(door_cycle_count=5_000)
        e.inject_fault("door_motor_degradation")
        e.tick(1.0)
        # Formula: 2.1 + (door_cycle_count % 10000) * 0.0003
        expected = 2.1 + (5_000 % 10_000) * 0.0003
        assert abs(e.door_motor_amps - expected) < 0.01

    def test_safety_circuit_intermittent_can_trip(self):
        """With enough seeds the fault must produce at least one trip."""
        tripped = False
        for seed in range(200):
            random.seed(seed)
            e = ElevatorState()
            e.inject_fault("safety_circuit_intermittent")
            e.tick(1.0)
            if not e.safety_circuit_ok or not e.door_interlock_ok:
                tripped = True
                break
        assert tripped, "Expected at least one safety-circuit trip in 200 random seeds"

    def test_safety_circuit_healthy_without_fault(self):
        """Without fault, safety circuit stays True across many ticks."""
        random.seed(42)
        e = ElevatorState()
        for _ in range(200):
            e.tick(1.0)
        # No fault injected — safety_circuit_ok must never be touched by tick()
        assert e.safety_circuit_ok is True


# ===========================================================================
# to_bacnet_points()
# ===========================================================================

class TestToBACnetPoints:
    EXPECTED_KEYS = {
        "analogInput:0",  "analogInput:1",  "analogInput:2",  "analogInput:3",
        "analogInput:10", "analogInput:11", "analogInput:12", "analogInput:13",
        "analogInput:14",
        "analogInput:20", "analogInput:21", "analogInput:22",
        "analogInput:30", "analogInput:31",
        "analogInput:40",
        "binaryInput:0",  "binaryInput:1",  "binaryInput:2",
        "binaryInput:3",  "binaryInput:4",
        "multiStateInput:0", "multiStateInput:1",
        "analogValue:0",
    }

    def test_returns_all_expected_keys(self):
        pts = ElevatorState().to_bacnet_points()
        assert set(pts.keys()) == self.EXPECTED_KEYS

    def test_motor_current_matches_state(self):
        e = ElevatorState()
        e.motor_current = 14.75
        assert e.to_bacnet_points()["analogInput:0"] == round(14.75, 2)

    def test_floor_matches_state(self):
        e = ElevatorState(floor=7)
        assert e.to_bacnet_points()["analogValue:0"] == 7

    @pytest.mark.parametrize("direction,code", [("up", 1), ("down", 2), ("idle", 3)])
    def test_direction_encoding(self, direction, code):
        e = ElevatorState(direction=direction)
        assert e.to_bacnet_points()["multiStateInput:0"] == code

    @pytest.mark.parametrize("status,code", [
        ("open", 1), ("closed", 2), ("opening", 3), ("closing", 4)
    ])
    def test_door_status_encoding(self, status, code):
        e = ElevatorState(door_status=status)
        assert e.to_bacnet_points()["multiStateInput:1"] == code

    def test_binary_safety_inputs_reflect_state(self):
        e = ElevatorState(
            door_interlock_ok=True,
            governor_ok=False,
            buffer_ok=True,
            pit_switch_ok=False,
            safety_circuit_ok=True,
        )
        pts = e.to_bacnet_points()
        assert pts["binaryInput:0"] is True
        assert pts["binaryInput:1"] is False
        assert pts["binaryInput:2"] is True
        assert pts["binaryInput:3"] is False
        assert pts["binaryInput:4"] is True

    def test_brake_response_ms_matches_state(self):
        e = ElevatorState(brake_response_ms=55)
        # brake_response_ms is set to 55 at init but overwritten by tick();
        # without ticking, it should reflect the init value
        e.brake_response_ms = 55
        assert e.to_bacnet_points()["analogInput:22"] == 55

    def test_door_cycle_count_matches_state(self):
        e = ElevatorState(door_cycle_count=999_999)
        assert e.to_bacnet_points()["analogInput:10"] == 999_999


# ===========================================================================
# to_dict()
# ===========================================================================

class TestToDict:
    def test_unit_id_present(self):
        e = ElevatorState(unit_id="ELV-TEST")
        assert e.to_dict()["unit_id"] == "ELV-TEST"

    def test_injected_fault_present_when_set(self):
        e = ElevatorState()
        e.inject_fault("brake_wear")
        assert e.to_dict()["injected_fault"] == "brake_wear"

    def test_injected_fault_none_when_not_set(self):
        e = ElevatorState()
        assert e.to_dict()["injected_fault"] is None

    def test_contains_motor_current_key(self):
        e = ElevatorState()
        assert "motor_current_a" in e.to_dict()

    def test_motor_current_rounded_to_two_places(self):
        e = ElevatorState()
        e.motor_current = 12.3456789
        val = e.to_dict()["motor_current_a"]
        assert val == round(12.3456789, 2)

    def test_all_safety_booleans_present(self):
        d = ElevatorState().to_dict()
        for key in ("door_interlock_ok", "governor_ok", "buffer_ok",
                    "pit_switch_ok", "safety_circuit_ok"):
            assert key in d


# ===========================================================================
# make_elevator_fleet()
# ===========================================================================

class TestMakeElevatorFleet:
    def test_returns_three_elevators(self):
        assert len(make_elevator_fleet()) == 3

    def test_unit_ids_are_unique(self):
        fleet = make_elevator_fleet()
        ids = [e.unit_id for e in fleet]
        assert len(ids) == len(set(ids))

    def test_expected_unit_ids_present(self):
        fleet = make_elevator_fleet()
        ids = {e.unit_id for e in fleet}
        assert ids == {"ELV-001", "ELV-002", "ELV-003"}

    def test_wear_profiles_ordered_by_drift(self):
        fleet = {e.unit_id: e for e in make_elevator_fleet()}
        assert (
            fleet["ELV-001"].motor_current_drift
            < fleet["ELV-002"].motor_current_drift
            < fleet["ELV-003"].motor_current_drift
        )

    def test_wear_profiles_ordered_by_door_force(self):
        fleet = {e.unit_id: e for e in make_elevator_fleet()}
        assert (
            fleet["ELV-001"].door_close_force_n
            < fleet["ELV-002"].door_close_force_n
            < fleet["ELV-003"].door_close_force_n
        )

    def test_wear_profiles_ordered_by_brake_response(self):
        fleet = {e.unit_id: e for e in make_elevator_fleet()}
        assert (
            fleet["ELV-001"].brake_response_ms
            < fleet["ELV-002"].brake_response_ms
            < fleet["ELV-003"].brake_response_ms
        )

    def test_all_elevators_below_asme_door_force_limit(self):
        for e in make_elevator_fleet():
            assert e.door_close_force_n < DOOR_FORCE_LIMIT_N, (
                f"{e.unit_id} door force {e.door_close_force_n} N exceeds {DOOR_FORCE_LIMIT_N} N"
            )

    def test_all_elevators_below_asme_brake_limit(self):
        for e in make_elevator_fleet():
            assert e.brake_response_ms < BRAKE_RESPONSE_LIMIT_MS, (
                f"{e.unit_id} brake {e.brake_response_ms} ms exceeds {BRAKE_RESPONSE_LIMIT_MS} ms"
            )

    def test_all_elevators_below_ada_leveling_limit(self):
        for e in make_elevator_fleet():
            assert e.leveling_accuracy_mm < LEVELING_LIMIT_MM, (
                f"{e.unit_id} leveling {e.leveling_accuracy_mm} mm exceeds {LEVELING_LIMIT_MM} mm"
            )

    def test_each_call_returns_independent_instances(self):
        fleet_a = make_elevator_fleet()
        fleet_b = make_elevator_fleet()
        fleet_a[0].inject_fault("brake_wear")
        assert fleet_b[0].injected_fault is None
