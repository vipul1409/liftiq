package ingest

import (
	"time"

	"github.com/google/uuid"
	"github.com/liftiq/telemetry-ingestor/internal/simulator"
	"github.com/liftiq/telemetry-ingestor/internal/store"
)

// MapSnapshot converts one ElevatorSnapshot into a slice of store.Row values.
//
// It is a pure function with no I/O — all randomness and time are supplied by
// the caller, making it fully deterministic and straightforwardly testable.
//
// 20 metric rows are produced per snapshot:
//   - 15 float-valued sensor readings
//   - 5 boolean safety-circuit fields (stored as 1.0 / 0.0)
//
// Fields absent or null in the decoded payload produce rows with Quality
// "missing" instead of "good".
//
// Excluded fields: direction, door_status, max_floors, injected_fault.
// These are operational/display metadata, not sensor readings.
func MapSnapshot(snap simulator.ElevatorSnapshot, unitID uuid.UUID, at time.Time) []store.Row {
	b2f := func(b bool) float64 {
		if b {
			return 1.0
		}
		return 0.0
	}

	// Named struct keeps metric name and value co-located, making the mapping
	// table easy to audit against the engineering plan's data-point table.
	type entry struct {
		name  string
		value float64
	}

	entries := []entry{
		// Motor telemetry
		{"motor_current_a", snap.MotorCurrentA},
		{"motor_temp_c", snap.MotorTempC},
		{"motor_rpm", snap.MotorRPM},
		{"motor_run_hours", snap.MotorRunHours},
		// Trip counters
		{"trip_count", float64(snap.TripCount)},
		// Door telemetry
		{"door_cycle_count", float64(snap.DoorCycleCount)},
		{"door_motor_amps", snap.DoorMotorAmps},
		{"door_close_force_n", snap.DoorCloseForceN},
		{"door_close_time_ms", float64(snap.DoorCloseTimeMs)},
		{"door_obstruction_events", float64(snap.DoorObstructionEvents)},
		// Brake telemetry
		{"brake_engagement_count", float64(snap.BrakeEngagementCount)},
		{"brake_current_a", snap.BrakeCurrentA},
		{"brake_response_ms", float64(snap.BrakeResponseMs)},
		// Ride quality
		{"leveling_accuracy_mm", snap.LevelingAccuracyMm},
		{"vibration_g", snap.VibrationG},
		// Safety circuit (bool → 1.0 / 0.0)
		{"door_interlock_ok", b2f(snap.DoorInterlockOk)},
		{"governor_ok", b2f(snap.GovernorOk)},
		{"buffer_ok", b2f(snap.BufferOk)},
		{"pit_switch_ok", b2f(snap.PitSwitchOk)},
		{"safety_circuit_ok", b2f(snap.SafetyCircuitOk)},
	}

	rows := make([]store.Row, len(entries))
	for i, e := range entries {
		// A field absent or null in the simulator payload is recorded as
		// "missing" (value 0) so the compliance engine reports Unknown rather
		// than evaluating a fabricated zero.
		quality := store.QualityGood
		if snap.IsMissing(e.name) {
			quality = store.QualityMissing
		}
		rows[i] = store.Row{
			Time:    at,
			UnitID:  unitID,
			Metric:  e.name,
			Value:   e.value,
			Quality: quality,
		}
	}
	return rows
}
