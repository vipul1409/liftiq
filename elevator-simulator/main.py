"""
main.py
LiftIQ Elevator Simulator — entry point.

Starts three simulated BACnet elevators (ELV-001, ELV-002, ELV-003) with
distinct wear profiles and runs them in a 1-second tick loop.

Concurrently serves:
  • BACnet/IP devices on UDP port 47808 (requires BAC0)
  • HTTP REST API on http://localhost:8000 (requires fastapi + uvicorn)

Usage:
    python main.py [--tick-interval SECONDS] [--http-port PORT] [--no-bacnet]

Environment variables:
    LIFTIQ_TICK_INTERVAL    seconds between simulation ticks (default: 1.0)
    LIFTIQ_HTTP_PORT        HTTP API port (default: 8000)
    LIFTIQ_BACNET_IP        BACnet local IP/prefix, e.g. 192.168.1.10/24
                            (default: 0.0.0.0/24)
"""
from __future__ import annotations

import argparse
import logging
import os
import signal
import sys
import threading
import time

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%dT%H:%M:%S",
)
logger = logging.getLogger("liftiq.simulator")

from elevator_state import make_elevator_fleet, ElevatorState
from bacnet_server import BACnetServer
import api as http_api


# ---------------------------------------------------------------------------
# Simulation loop
# ---------------------------------------------------------------------------

def run_simulation(
    elevators: list[ElevatorState],
    bacnet: BACnetServer,
    tick_interval: float,
    stop_event: threading.Event,
) -> None:
    """Main simulation loop. Ticks every elevator and pushes to BACnet."""
    logger.info("Simulation loop started (tick=%.1fs)", tick_interval)
    while not stop_event.is_set():
        start = time.monotonic()

        for elev in elevators:
            elev.tick(seconds=tick_interval)

        bacnet.update()

        # Periodic console snapshot every 10 ticks
        if int(time.time()) % 10 == 0:
            _log_snapshot(elevators)

        elapsed = time.monotonic() - start
        sleep_time = max(0.0, tick_interval - elapsed)
        stop_event.wait(timeout=sleep_time)

    logger.info("Simulation loop stopped.")


def _log_snapshot(elevators: list[ElevatorState]) -> None:
    for e in elevators:
        fault_tag = f" [FAULT:{e.injected_fault}]" if e.injected_fault else ""
        logger.info(
            "%s | floor=%-2d dir=%-4s | motor=%.1fA %.0f°C | "
            "door_force=%.0fN brake=%dms | safety=%s%s",
            e.unit_id,
            e.floor,
            e.direction,
            e.motor_current,
            e.motor_temp,
            e.door_close_force_n,
            e.brake_response_ms,
            "OK" if e.safety_circuit_ok else "FAIL",
            fault_tag,
        )


# ---------------------------------------------------------------------------
# HTTP server thread
# ---------------------------------------------------------------------------

def start_http_server(port: int, stop_event: threading.Event) -> threading.Thread:
    if http_api.app is None:
        logger.warning("FastAPI not available — HTTP server disabled. pip install fastapi uvicorn")
        return threading.Thread(target=lambda: None, daemon=True)

    try:
        import uvicorn  # type: ignore
    except ImportError:
        logger.warning("uvicorn not installed — HTTP server disabled. pip install uvicorn")
        return threading.Thread(target=lambda: None, daemon=True)

    config = uvicorn.Config(
        app=http_api.app,
        host="0.0.0.0",
        port=port,
        log_level="warning",
        loop="asyncio",
    )
    server = uvicorn.Server(config)

    def _run():
        server.run()

    # Wire stop_event → uvicorn shutdown
    def _watch_stop():
        stop_event.wait()
        server.should_exit = True

    t = threading.Thread(target=_run, name="http-server", daemon=True)
    w = threading.Thread(target=_watch_stop, name="http-watcher", daemon=True)
    t.start()
    w.start()
    logger.info("HTTP API listening on http://0.0.0.0:%d", port)
    return t


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="LiftIQ Elevator Simulator")
    p.add_argument(
        "--tick-interval",
        type=float,
        default=float(os.environ.get("LIFTIQ_TICK_INTERVAL", "1.0")),
        help="Seconds between simulation ticks (default: 1.0)",
    )
    p.add_argument(
        "--http-port",
        type=int,
        default=int(os.environ.get("LIFTIQ_HTTP_PORT", "8000")),
        help="HTTP API port (default: 8000)",
    )
    p.add_argument(
        "--bacnet-ip",
        default=os.environ.get("LIFTIQ_BACNET_IP", "0.0.0.0/24"),
        help="BACnet local IP/prefix (default: 0.0.0.0/24)",
    )
    p.add_argument(
        "--no-bacnet",
        action="store_true",
        help="Disable BACnet server (HTTP + stdout only)",
    )
    return p.parse_args()


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    args = parse_args()
    stop_event = threading.Event()

    # Graceful shutdown on SIGINT / SIGTERM
    def _handle_signal(sig, _frame):
        logger.info("Received signal %s — shutting down…", signal.Signals(sig).name)
        stop_event.set()

    signal.signal(signal.SIGINT, _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    # Build the elevator fleet
    elevators = make_elevator_fleet()
    logger.info(
        "Initialized %d elevators: %s",
        len(elevators),
        [e.unit_id for e in elevators],
    )

    # Register with HTTP API
    http_api.register_elevators(elevators)

    # Start BACnet server
    bacnet = BACnetServer(elevators, local_ip=args.bacnet_ip)
    if not args.no_bacnet:
        bacnet.start()
    else:
        logger.info("BACnet server disabled (--no-bacnet)")

    # Start HTTP server in background thread
    start_http_server(args.http_port, stop_event)

    # Run simulation loop (blocking until stop_event)
    try:
        run_simulation(elevators, bacnet, args.tick_interval, stop_event)
    finally:
        bacnet.stop()
        logger.info("Simulator exited cleanly.")


if __name__ == "__main__":
    main()
