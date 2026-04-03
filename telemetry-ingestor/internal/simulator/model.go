package simulator

// ElevatorSnapshot mirrors the JSON shape of one item in GET /elevators.
// Field names match the Python simulator's to_dict() output exactly.
//
// Operational fields (direction, door_status, max_floors, injected_fault) are
// decoded but intentionally excluded from metric storage — they are display state,
// not sensor readings for the compliance engine.
type ElevatorSnapshot struct {
	UnitID    string  `json:"unit_id"`
	Floor     int     `json:"floor"`
	MaxFloors int     `json:"max_floors"`
	Direction string  `json:"direction"`  // "up" | "down" | "idle"
	DoorStatus string `json:"door_status"` // "open" | "closed" | "opening" | "closing"

	// Motor telemetry
	MotorCurrentA float64 `json:"motor_current_a"`
	MotorTempC    float64 `json:"motor_temp_c"`
	MotorRPM      float64 `json:"motor_rpm"`
	MotorRunHours float64 `json:"motor_run_hours"`

	// Trip counters
	TripCount int `json:"trip_count"`

	// Door telemetry
	DoorCycleCount        int     `json:"door_cycle_count"`
	DoorMotorAmps         float64 `json:"door_motor_amps"`
	DoorCloseForceN       float64 `json:"door_close_force_n"`
	DoorCloseTimeMs       int     `json:"door_close_time_ms"`
	DoorObstructionEvents int     `json:"door_obstruction_events"`

	// Brake telemetry
	BrakeEngagementCount int     `json:"brake_engagement_count"`
	BrakeCurrentA        float64 `json:"brake_current_a"`
	BrakeResponseMs      int     `json:"brake_response_ms"`

	// Ride quality
	LevelingAccuracyMm float64 `json:"leveling_accuracy_mm"`
	VibrationG         float64 `json:"vibration_g"`

	// Safety circuit (stored as 0.0 / 1.0)
	DoorInterlockOk bool `json:"door_interlock_ok"`
	GovernorOk      bool `json:"governor_ok"`
	BufferOk        bool `json:"buffer_ok"`
	PitSwitchOk     bool `json:"pit_switch_ok"`
	SafetyCircuitOk bool `json:"safety_circuit_ok"`

	// Simulator-only metadata — not stored as metrics
	InjectedFault *string `json:"injected_fault"`
}
