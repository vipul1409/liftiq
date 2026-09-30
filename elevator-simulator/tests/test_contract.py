"""
test_contract.py
The simulator's HTTP payload must match the checked-in contract that the
Go telemetry-ingestor decodes (contracts/simulator-snapshot.json) and that
the compliance engine's rules reference (contracts/telemetry-metrics.json).
A renamed key here fails this test instead of silently becoming a zero
reading downstream.
"""
import json
from pathlib import Path

from elevator_state import make_elevator_fleet

CONTRACTS = Path(__file__).resolve().parents[2] / "contracts"


def _load(name):
    return json.loads((CONTRACTS / name).read_text())


def test_to_dict_keys_match_snapshot_contract():
    snapshot_keys = set(_load("simulator-snapshot.json"))
    for elevator in make_elevator_fleet():
        assert set(elevator.to_dict()) == snapshot_keys


def test_every_contract_metric_is_emitted_and_numeric():
    metrics = _load("telemetry-metrics.json")
    payload = make_elevator_fleet()[0].to_dict()
    for m in metrics:
        assert m in payload, f"simulator does not emit metric {m!r}"
        assert isinstance(payload[m], (int, float, bool)), f"{m} is {type(payload[m]).__name__}"
