"""
test_api.py
Integration tests for the FastAPI HTTP API (api.py).

Uses FastAPI's TestClient (backed by httpx) to exercise all routes
without starting a real server. The shared elevator registry is reset
before each test via register_elevators() to ensure isolation.
"""
import pytest

pytest.importorskip("fastapi", reason="fastapi not installed — skipping API tests")
pytest.importorskip("httpx",   reason="httpx not installed — skipping API tests")

from fastapi.testclient import TestClient  # noqa: E402

import api  # noqa: E402
from elevator_state import ElevatorState, FAULT_TYPES  # noqa: E402


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture()
def elevators():
    """Two fresh ElevatorState objects re-created for every test."""
    return [ElevatorState(unit_id="ELV-001"), ElevatorState(unit_id="ELV-002")]


@pytest.fixture(autouse=True)
def reset_registry(elevators):
    """Populate the shared registry before each test and clear it after."""
    api.register_elevators(elevators)
    yield
    api.register_elevators([])


@pytest.fixture()
def client():
    return TestClient(api.app)


# ---------------------------------------------------------------------------
# GET /health
# ---------------------------------------------------------------------------

class TestHealth:
    def test_returns_200(self, client):
        assert client.get("/health").status_code == 200

    def test_status_is_ok(self, client):
        assert client.get("/health").json()["status"] == "ok"

    def test_lists_registered_elevator_ids(self, client):
        ids = client.get("/health").json()["elevators"]
        assert "ELV-001" in ids
        assert "ELV-002" in ids


# ---------------------------------------------------------------------------
# GET /elevators
# ---------------------------------------------------------------------------

class TestListElevators:
    def test_returns_200(self, client):
        assert client.get("/elevators").status_code == 200

    def test_returns_all_elevators(self, client):
        data = client.get("/elevators").json()
        assert len(data) == 2

    def test_unit_ids_match_registry(self, client):
        data = client.get("/elevators").json()
        unit_ids = {e["unit_id"] for e in data}
        assert unit_ids == {"ELV-001", "ELV-002"}

    def test_empty_registry_returns_empty_list(self, client):
        api.register_elevators([])
        data = client.get("/elevators").json()
        assert data == []


# ---------------------------------------------------------------------------
# GET /elevators/{unit_id}
# ---------------------------------------------------------------------------

class TestGetElevator:
    def test_returns_200_for_known_unit(self, client):
        assert client.get("/elevators/ELV-001").status_code == 200

    def test_returns_correct_unit_id(self, client):
        data = client.get("/elevators/ELV-001").json()
        assert data["unit_id"] == "ELV-001"

    def test_returns_404_for_unknown_unit(self, client):
        assert client.get("/elevators/ELV-UNKNOWN").status_code == 404

    def test_response_contains_motor_current(self, client):
        data = client.get("/elevators/ELV-001").json()
        assert "motor_current_a" in data

    def test_response_contains_safety_fields(self, client):
        data = client.get("/elevators/ELV-001").json()
        for field in ("door_interlock_ok", "governor_ok", "safety_circuit_ok"):
            assert field in data

    def test_response_injected_fault_is_none_initially(self, client):
        data = client.get("/elevators/ELV-001").json()
        assert data["injected_fault"] is None


# ---------------------------------------------------------------------------
# GET /elevators/{unit_id}/bacnet
# ---------------------------------------------------------------------------

class TestGetBACnetPoints:
    def test_returns_200(self, client):
        assert client.get("/elevators/ELV-001/bacnet").status_code == 200

    def test_returns_404_for_unknown_unit(self, client):
        assert client.get("/elevators/ELV-UNKNOWN/bacnet").status_code == 404

    def test_contains_motor_current_key(self, client):
        pts = client.get("/elevators/ELV-001/bacnet").json()
        assert "analogInput:0" in pts

    def test_contains_safety_binary_inputs(self, client):
        pts = client.get("/elevators/ELV-001/bacnet").json()
        for key in ("binaryInput:0", "binaryInput:1", "binaryInput:2",
                    "binaryInput:3", "binaryInput:4"):
            assert key in pts

    def test_contains_floor_analog_value(self, client):
        pts = client.get("/elevators/ELV-001/bacnet").json()
        assert "analogValue:0" in pts


# ---------------------------------------------------------------------------
# GET /faults
# ---------------------------------------------------------------------------

class TestListFaults:
    def test_returns_200(self, client):
        assert client.get("/faults").status_code == 200

    def test_returns_all_fault_types(self, client):
        data = client.get("/faults").json()
        returned = set(data["available_faults"])
        assert returned == FAULT_TYPES


# ---------------------------------------------------------------------------
# POST /elevators/{unit_id}/fault
# ---------------------------------------------------------------------------

class TestInjectFault:
    def test_returns_200_for_valid_fault(self, client):
        r = client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        assert r.status_code == 200

    def test_response_contains_injected_fault(self, client):
        r = client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        assert r.json()["injected_fault"] == "brake_wear"

    def test_response_contains_unit_id(self, client):
        r = client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        assert r.json()["unit_id"] == "ELV-001"

    def test_fault_reflected_in_get_elevator(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        data = client.get("/elevators/ELV-001").json()
        assert data["injected_fault"] == "brake_wear"

    def test_invalid_fault_returns_400(self, client):
        r = client.post("/elevators/ELV-001/fault", json={"fault": "explode_everything"})
        assert r.status_code == 400

    def test_invalid_fault_does_not_change_state(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        client.post("/elevators/ELV-001/fault", json={"fault": "bad_fault"})
        data = client.get("/elevators/ELV-001").json()
        assert data["injected_fault"] == "brake_wear"

    def test_unknown_unit_returns_404(self, client):
        r = client.post("/elevators/ELV-UNKNOWN/fault", json={"fault": "brake_wear"})
        assert r.status_code == 404

    def test_faults_on_different_elevators_are_independent(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        data_002 = client.get("/elevators/ELV-002").json()
        assert data_002["injected_fault"] is None

    @pytest.mark.parametrize("fault_type", sorted(FAULT_TYPES))
    def test_all_fault_types_accepted(self, client, fault_type):
        r = client.post("/elevators/ELV-001/fault", json={"fault": fault_type})
        assert r.status_code == 200
        assert r.json()["injected_fault"] == fault_type


# ---------------------------------------------------------------------------
# DELETE /elevators/{unit_id}/fault
# ---------------------------------------------------------------------------

class TestClearFault:
    def test_returns_200(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        r = client.delete("/elevators/ELV-001/fault")
        assert r.status_code == 200

    def test_response_contains_cleared_fault(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        r = client.delete("/elevators/ELV-001/fault")
        assert r.json()["cleared_fault"] == "brake_wear"

    def test_fault_removed_from_state(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        client.delete("/elevators/ELV-001/fault")
        data = client.get("/elevators/ELV-001").json()
        assert data["injected_fault"] is None

    def test_clear_when_no_fault_returns_none(self, client):
        r = client.delete("/elevators/ELV-001/fault")
        assert r.status_code == 200
        assert r.json()["cleared_fault"] is None

    def test_unknown_unit_returns_404(self, client):
        r = client.delete("/elevators/ELV-UNKNOWN/fault")
        assert r.status_code == 404

    def test_clear_does_not_affect_other_elevator(self, client):
        client.post("/elevators/ELV-001/fault", json={"fault": "brake_wear"})
        client.post("/elevators/ELV-002/fault", json={"fault": "leveling_drift"})
        client.delete("/elevators/ELV-001/fault")
        data_002 = client.get("/elevators/ELV-002").json()
        assert data_002["injected_fault"] == "leveling_drift"
