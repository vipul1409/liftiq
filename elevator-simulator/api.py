"""
api.py
FastAPI HTTP server providing:
  GET  /elevators                    → list all elevator states
  GET  /elevators/{unit_id}          → single elevator state
  POST /elevators/{unit_id}/fault    → inject a fault  {"fault": "door_motor_degradation"}
  DELETE /elevators/{unit_id}/fault  → clear injected fault
  GET  /elevators/{unit_id}/bacnet   → BACnet point map for the elevator
  GET  /health                       → liveness probe

Run alongside the simulation loop in a background thread.
"""
from __future__ import annotations

import threading
from typing import Optional

try:
    from fastapi import FastAPI, HTTPException  # type: ignore
    from pydantic import BaseModel             # type: ignore
    _FASTAPI_AVAILABLE = True
except ImportError:
    _FASTAPI_AVAILABLE = False

from elevator_state import ElevatorState, FAULT_TYPES

# Shared elevator registry — populated by main.py before the server starts
_registry: dict[str, ElevatorState] = {}
_lock = threading.Lock()


def register_elevators(elevators: list[ElevatorState]) -> None:
    """Register the shared elevator list with the API module."""
    with _lock:
        _registry.clear()
        for elev in elevators:
            _registry[elev.unit_id] = elev


def _get_elevator(unit_id: str) -> ElevatorState:
    with _lock:
        elev = _registry.get(unit_id)
    if elev is None:
        raise HTTPException(status_code=404, detail=f"Elevator '{unit_id}' not found")
    return elev


if _FASTAPI_AVAILABLE:
    app = FastAPI(
        title="LiftIQ Elevator Simulator",
        description=(
            "HTTP API for the elevator simulator. Use the /fault endpoints to inject "
            "faults for demo and testing purposes."
        ),
        version="0.1.0",
    )

    class FaultRequest(BaseModel):
        fault: str

    @app.get("/health")
    def health():
        return {"status": "ok", "elevators": list(_registry.keys())}

    @app.get("/elevators")
    def list_elevators():
        with _lock:
            return [elev.to_dict() for elev in _registry.values()]

    @app.get("/elevators/{unit_id}")
    def get_elevator(unit_id: str):
        return _get_elevator(unit_id).to_dict()

    @app.get("/elevators/{unit_id}/bacnet")
    def get_bacnet_points(unit_id: str):
        return _get_elevator(unit_id).to_bacnet_points()

    @app.post("/elevators/{unit_id}/fault")
    def inject_fault(unit_id: str, body: FaultRequest):
        if body.fault not in FAULT_TYPES:
            raise HTTPException(
                status_code=400,
                detail=f"Unknown fault '{body.fault}'. Valid faults: {sorted(FAULT_TYPES)}",
            )
        elev = _get_elevator(unit_id)
        elev.inject_fault(body.fault)
        return {"unit_id": unit_id, "injected_fault": body.fault}

    @app.delete("/elevators/{unit_id}/fault")
    def clear_fault(unit_id: str):
        elev = _get_elevator(unit_id)
        prev = elev.injected_fault
        elev.inject_fault(None)
        return {"unit_id": unit_id, "cleared_fault": prev}

    @app.get("/faults")
    def list_fault_types():
        return {"available_faults": sorted(FAULT_TYPES)}

else:
    # Stub so main.py can still import this module without FastAPI
    app = None  # type: ignore
