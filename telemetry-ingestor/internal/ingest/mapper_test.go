package ingest_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liftiq/telemetry-ingestor/internal/ingest"
	"github.com/liftiq/telemetry-ingestor/internal/simulator"
	"github.com/liftiq/telemetry-ingestor/internal/store"
)

// toMap converts a []store.Row slice into a metric→value lookup for assertions.
func toMap(rows []store.Row) map[string]float64 {
	m := make(map[string]float64, len(rows))
	for _, r := range rows {
		m[r.Metric] = r.Value
	}
	return m
}

func TestMapSnapshot_Produces20Rows(t *testing.T) {
	rows := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uuid.New(), time.Now())
	if len(rows) != 20 {
		t.Errorf("expected 20 rows, got %d", len(rows))
	}
}

func TestMapSnapshot_AllTimestampsEqualAt(t *testing.T) {
	at := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	rows := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uuid.New(), at)
	for _, r := range rows {
		if !r.Time.Equal(at) {
			t.Errorf("metric %q: time %v, want %v", r.Metric, r.Time, at)
		}
	}
}

func TestMapSnapshot_AllUnitIDsEqualInput(t *testing.T) {
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	rows := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uid, time.Now())
	for _, r := range rows {
		if r.UnitID != uid {
			t.Errorf("metric %q: unit_id %v, want %v", r.Metric, r.UnitID, uid)
		}
	}
}

func TestMapSnapshot_AllQualityIsGood(t *testing.T) {
	rows := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uuid.New(), time.Now())
	for _, r := range rows {
		if r.Quality != "good" {
			t.Errorf("metric %q: quality %q, want %q", r.Metric, r.Quality, "good")
		}
	}
}

func TestMapSnapshot_MetricNamesAreUnique(t *testing.T) {
	rows := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uuid.New(), time.Now())
	seen := make(map[string]bool)
	for _, r := range rows {
		if seen[r.Metric] {
			t.Errorf("duplicate metric name: %q", r.Metric)
		}
		seen[r.Metric] = true
	}
}

func TestMapSnapshot_FloatFields(t *testing.T) {
	snap := simulator.ElevatorSnapshot{
		UnitID:             "ELV-001",
		MotorCurrentA:      14.5,
		MotorTempC:         52.3,
		MotorRPM:           1450.0,
		MotorRunHours:      12450.1,
		DoorMotorAmps:      2.15,
		DoorCloseForceN:    98.7,
		BrakeCurrentA:      1.82,
		LevelingAccuracyMm: 4.1,
		VibrationG:         0.06,
	}

	m := toMap(ingest.MapSnapshot(snap, uuid.New(), time.Now()))

	cases := []struct {
		metric string
		want   float64
	}{
		{"motor_current_a", 14.5},
		{"motor_temp_c", 52.3},
		{"motor_rpm", 1450.0},
		{"motor_run_hours", 12450.1},
		{"door_motor_amps", 2.15},
		{"door_close_force_n", 98.7},
		{"brake_current_a", 1.82},
		{"leveling_accuracy_mm", 4.1},
		{"vibration_g", 0.06},
	}
	for _, tc := range cases {
		got, ok := m[tc.metric]
		if !ok {
			t.Errorf("metric %q missing", tc.metric)
			continue
		}
		if got != tc.want {
			t.Errorf("metric %q: got %g, want %g", tc.metric, got, tc.want)
		}
	}
}

func TestMapSnapshot_IntFieldsConvertedToFloat64(t *testing.T) {
	snap := simulator.ElevatorSnapshot{
		UnitID:                "ELV-001",
		TripCount:             847231,
		DoorCycleCount:        1_204_500,
		DoorCloseTimeMs:       3200,
		DoorObstructionEvents: 12,
		BrakeEngagementCount:  847231,
		BrakeResponseMs:       45,
	}

	m := toMap(ingest.MapSnapshot(snap, uuid.New(), time.Now()))

	cases := []struct {
		metric string
		want   float64
	}{
		{"trip_count", 847231},
		{"door_cycle_count", 1_204_500},
		{"door_close_time_ms", 3200},
		{"door_obstruction_events", 12},
		{"brake_engagement_count", 847231},
		{"brake_response_ms", 45},
	}
	for _, tc := range cases {
		got, ok := m[tc.metric]
		if !ok {
			t.Errorf("metric %q missing", tc.metric)
			continue
		}
		if got != tc.want {
			t.Errorf("metric %q: got %g, want %g", tc.metric, got, tc.want)
		}
	}
}

func TestMapSnapshot_TrueBoolsBecome1(t *testing.T) {
	snap := simulator.ElevatorSnapshot{
		UnitID:          "ELV-001",
		DoorInterlockOk: true,
		GovernorOk:      true,
		BufferOk:        true,
		PitSwitchOk:     true,
		SafetyCircuitOk: true,
	}
	m := toMap(ingest.MapSnapshot(snap, uuid.New(), time.Now()))

	for _, metric := range []string{
		"door_interlock_ok", "governor_ok", "buffer_ok", "pit_switch_ok", "safety_circuit_ok",
	} {
		if got := m[metric]; got != 1.0 {
			t.Errorf("metric %q (true): got %g, want 1.0", metric, got)
		}
	}
}

func TestMapSnapshot_FalseBoolsBecome0(t *testing.T) {
	snap := simulator.ElevatorSnapshot{
		UnitID:          "ELV-001",
		DoorInterlockOk: false,
		GovernorOk:      false,
		BufferOk:        false,
		PitSwitchOk:     false,
		SafetyCircuitOk: false,
	}
	m := toMap(ingest.MapSnapshot(snap, uuid.New(), time.Now()))

	for _, metric := range []string{
		"door_interlock_ok", "governor_ok", "buffer_ok", "pit_switch_ok", "safety_circuit_ok",
	} {
		if got := m[metric]; got != 0.0 {
			t.Errorf("metric %q (false): got %g, want 0.0", metric, got)
		}
	}
}

func TestMapSnapshot_ExcludedOperationalFields(t *testing.T) {
	// direction, door_status, max_floors, injected_fault must NOT appear as metrics.
	fault := "brake_wear"
	snap := simulator.ElevatorSnapshot{
		UnitID:        "ELV-001",
		Direction:     "up",
		DoorStatus:    "open",
		MaxFloors:     10,
		InjectedFault: &fault,
	}
	m := toMap(ingest.MapSnapshot(snap, uuid.New(), time.Now()))

	for _, excluded := range []string{"direction", "door_status", "max_floors", "injected_fault"} {
		if _, found := m[excluded]; found {
			t.Errorf("excluded field %q found in metric output", excluded)
		}
	}
}

func TestMapSnapshot_AllExpectedMetricsPresent(t *testing.T) {
	expected := []string{
		"motor_current_a", "motor_temp_c", "motor_rpm", "motor_run_hours",
		"trip_count",
		"door_cycle_count", "door_motor_amps", "door_close_force_n",
		"door_close_time_ms", "door_obstruction_events",
		"brake_engagement_count", "brake_current_a", "brake_response_ms",
		"leveling_accuracy_mm", "vibration_g",
		"door_interlock_ok", "governor_ok", "buffer_ok", "pit_switch_ok", "safety_circuit_ok",
	}

	m := toMap(ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001"}, uuid.New(), time.Now()))

	for _, metric := range expected {
		if _, ok := m[metric]; !ok {
			t.Errorf("expected metric %q missing from output", metric)
		}
	}
}

func TestMapSnapshot_TwoSnapshotsProduceIndependentRows(t *testing.T) {
	uid1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	uid2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	at1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	at2 := time.Date(2026, 1, 1, 0, 0, 5, 0, time.UTC)

	rows1 := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-001", MotorCurrentA: 10.0}, uid1, at1)
	rows2 := ingest.MapSnapshot(simulator.ElevatorSnapshot{UnitID: "ELV-002", MotorCurrentA: 20.0}, uid2, at2)

	m1 := toMap(rows1)
	m2 := toMap(rows2)

	if m1["motor_current_a"] == m2["motor_current_a"] {
		t.Error("two snapshots with different values produced the same motor_current_a")
	}
	if rows1[0].UnitID == rows2[0].UnitID {
		t.Error("rows from different snapshots share the same unit_id")
	}
	if rows1[0].Time.Equal(rows2[0].Time) {
		t.Error("rows from different polls share the same timestamp")
	}
}

func TestMapSnapshot_MissingFieldWrittenAsMissingNotGood(t *testing.T) {
	var snap simulator.ElevatorSnapshot
	// door_close_force_n absent (e.g. renamed upstream); everything else present.
	payload := `{"unit_id":"ELV-003","motor_current_a":12.5,"motor_temp_c":50,"motor_rpm":1400,
		"motor_run_hours":100,"trip_count":1,"door_cycle_count":1,"door_motor_amps":2,
		"door_close_time_ms":3000,"door_obstruction_events":0,"brake_engagement_count":1,
		"brake_current_a":1.5,"brake_response_ms":60,"leveling_accuracy_mm":4,"vibration_g":0.05,
		"door_interlock_ok":true,"governor_ok":true,"buffer_ok":true,"pit_switch_ok":true,
		"safety_circuit_ok":true}`
	if err := json.Unmarshal([]byte(payload), &snap); err != nil {
		t.Fatal(err)
	}

	rows := ingest.MapSnapshot(snap, uuid.New(), time.Now())
	if len(rows) != 20 {
		t.Fatalf("got %d rows, want 20", len(rows))
	}
	for _, r := range rows {
		want := "good"
		if r.Metric == "door_close_force_n" {
			want = "missing"
		}
		if r.Quality != want {
			t.Errorf("%s: quality %q, want %q", r.Metric, r.Quality, want)
		}
	}
}
