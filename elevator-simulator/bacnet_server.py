"""
bacnet_server.py
Wraps multiple ElevatorState instances and exposes them as BACnet/IP devices
on the local network using BAC0.

Each elevator is a separate BACnet device instance (device IDs 1001, 1002, 1003).
BACnet object IDs match the mapping defined in elevator_state.to_bacnet_points().

Usage:
    server = BACnetServer(elevators, local_ip="0.0.0.0/24")
    server.start()      # starts background thread
    server.update()     # push current elevator states to BACnet objects
    server.stop()
"""
import logging
import threading

from elevator_state import ElevatorState

logger = logging.getLogger(__name__)

# BACnet object type tags used in to_bacnet_points()
_ANALOG_INPUT = "analogInput"
_BINARY_INPUT = "binaryInput"
_MULTI_STATE_INPUT = "multiStateInput"
_ANALOG_VALUE = "analogValue"

# Base BACnet device instance ID for the first elevator
_DEVICE_ID_BASE = 1001


class BACnetServer:
    """
    Creates one BACnet/IP device per elevator and keeps their object values
    synchronised with the live ElevatorState on every call to update().

    Requires BAC0 >= 22.x and bacpypes3 to be installed.
    If BAC0 is not available the server silently degrades to a no-op so the
    rest of the simulator (HTTP API, stdout logging) still works.
    """

    def __init__(
        self,
        elevators: list[ElevatorState],
        local_ip: str = "0.0.0.0/24",
        port: int = 47808,
    ) -> None:
        self.elevators = elevators
        self.local_ip = local_ip
        self.port = port

        self._apps: dict = {}   # unit_id → BAC0 app/device
        self._available = False
        self._lock = threading.Lock()

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def start(self) -> None:
        """Initialise BACnet devices for every elevator."""
        try:
            import BAC0  # type: ignore
        except ImportError:
            logger.warning(
                "BAC0 not installed — BACnet server disabled. "
                "Install with: pip install BAC0"
            )
            return

        for idx, elev in enumerate(self.elevators):
            device_id = _DEVICE_ID_BASE + idx
            try:
                app = self._create_device(BAC0, elev, device_id)
                self._apps[elev.unit_id] = app
                logger.info(
                    "BACnet device started — unit=%s device_id=%d",
                    elev.unit_id,
                    device_id,
                )
            except Exception as exc:  # noqa: BLE001
                logger.error(
                    "Failed to create BACnet device for %s: %s", elev.unit_id, exc
                )

        self._available = bool(self._apps)

    def stop(self) -> None:
        """Disconnect all BACnet devices."""
        for unit_id, app in self._apps.items():
            try:
                app.disconnect()
                logger.info("BACnet device stopped — unit=%s", unit_id)
            except Exception as exc:  # noqa: BLE001
                logger.warning("Error stopping BACnet device %s: %s", unit_id, exc)
        self._apps.clear()
        self._available = False

    def update(self) -> None:
        """Push current elevator state values into BACnet objects."""
        if not self._available:
            return

        with self._lock:
            for elev in self.elevators:
                app = self._apps.get(elev.unit_id)
                if app is None:
                    continue
                points = elev.to_bacnet_points()
                for obj_key, value in points.items():
                    try:
                        self._write_point(app, obj_key, value)
                    except Exception as exc:  # noqa: BLE001
                        logger.debug("BACnet write error %s[%s]: %s", elev.unit_id, obj_key, exc)

    @property
    def is_available(self) -> bool:
        return self._available

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _create_device(self, BAC0, elev: ElevatorState, device_id: int):
        """
        Create a BAC0 application that acts as a BACnet server device.

        BAC0's server (local device) API varies across versions. This
        implementation targets BAC0 >= 22.x. For older versions the device
        creation may need adjustment — see BAC0 docs for your version.
        """
        # BAC0.lite() creates a minimal BACnet stack (client + server).
        # Passing `localObjName` and `deviceId` sets the device properties.
        app = BAC0.lite(
            ip=self.local_ip,
            port=self.port,
            localObjName=f"LiftIQ-{elev.unit_id}",
            deviceId=device_id,
        )

        # Pre-create all BACnet objects so they are discoverable via Who-Has / Read-Property.
        initial_points = elev.to_bacnet_points()
        for obj_key, value in initial_points.items():
            obj_type, obj_instance = obj_key.split(":")
            obj_instance = int(obj_instance)
            self._create_object(app, obj_type, obj_instance, value, elev.unit_id)

        return app

    def _create_object(self, app, obj_type: str, instance: int, initial_value, unit_id: str):
        """Add a new BACnet object to the local device."""
        name = f"{unit_id}-{obj_type}-{instance}"
        try:
            if obj_type == _ANALOG_INPUT:
                app.create_object(
                    objectType="analogInput",
                    instance=instance,
                    objectName=name,
                    presentValue=float(initial_value),
                    units="noUnits",
                )
            elif obj_type == _BINARY_INPUT:
                app.create_object(
                    objectType="binaryInput",
                    instance=instance,
                    objectName=name,
                    presentValue="active" if initial_value else "inactive",
                )
            elif obj_type == _MULTI_STATE_INPUT:
                app.create_object(
                    objectType="multiStateInput",
                    instance=instance,
                    objectName=name,
                    presentValue=int(initial_value),
                    numberOfStates=4,
                )
            elif obj_type == _ANALOG_VALUE:
                app.create_object(
                    objectType="analogValue",
                    instance=instance,
                    objectName=name,
                    presentValue=float(initial_value),
                    units="noUnits",
                )
        except Exception as exc:  # noqa: BLE001
            # Object may already exist on reconnect
            logger.debug("create_object skipped %s: %s", name, exc)

    def _write_point(self, app, obj_key: str, value) -> None:
        """Update the presentValue of a BACnet object in the local device."""
        obj_type, obj_instance = obj_key.split(":")
        obj_instance = int(obj_instance)

        if obj_type == _BINARY_INPUT:
            pv = "active" if value else "inactive"
        elif obj_type == _MULTI_STATE_INPUT:
            pv = int(value)
        else:
            pv = float(value)

        # BAC0 local device write — API differs slightly by version
        try:
            app[f"{obj_type} {obj_instance}"].presentValue = pv
        except (KeyError, AttributeError):
            # Fallback for alternative BAC0 APIs
            app.set_point_name(f"{obj_type} {obj_instance}", pv)
