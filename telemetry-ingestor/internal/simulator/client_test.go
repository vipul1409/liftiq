package simulator_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liftiq/telemetry-ingestor/internal/simulator"
)

// serve returns a test server that writes body as the JSON response.
func serve(t *testing.T, status int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/elevators" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(status)
		if body != nil {
			if err := json.NewEncoder(w).Encode(body); err != nil {
				t.Errorf("encode response: %v", err)
			}
		}
	}))
}

func TestFetchAll_ReturnsAllSnapshots(t *testing.T) {
	want := []simulator.ElevatorSnapshot{
		{UnitID: "ELV-001", Floor: 3, MotorCurrentA: 12.5},
		{UnitID: "ELV-002", Floor: 7, MotorCurrentA: 13.8},
	}
	srv := serve(t, http.StatusOK, want)
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d snapshots, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].UnitID != want[i].UnitID {
			t.Errorf("[%d] UnitID: got %q, want %q", i, got[i].UnitID, want[i].UnitID)
		}
		if got[i].MotorCurrentA != want[i].MotorCurrentA {
			t.Errorf("[%d] MotorCurrentA: got %f, want %f", i, got[i].MotorCurrentA, want[i].MotorCurrentA)
		}
	}
}

func TestFetchAll_EmptyArray(t *testing.T) {
	srv := serve(t, http.StatusOK, []simulator.ElevatorSnapshot{})
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d items, want 0", len(got))
	}
}

func TestFetchAll_HTTP500ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestFetchAll_InvalidJSONReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json {{{`))
	}))
	defer srv.Close()

	_, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestFetchAll_CancelledContextReturnsError(t *testing.T) {
	srv := serve(t, http.StatusOK, []simulator.ElevatorSnapshot{})
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before making the request

	_, err := simulator.NewHTTPClient(srv.URL).FetchAll(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestFetchAll_BoolFieldsRoundTrip(t *testing.T) {
	// Verify bool fields survive JSON decode correctly.
	// They are stored as 0.0/1.0 in the DB (handled by the mapper), but
	// the client must decode them as bool without loss.
	payload := []simulator.ElevatorSnapshot{{
		UnitID:          "ELV-001",
		DoorInterlockOk: false,
		GovernorOk:      true,
		SafetyCircuitOk: true,
	}}
	srv := serve(t, http.StatusOK, payload)
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].DoorInterlockOk != false {
		t.Error("DoorInterlockOk: expected false")
	}
	if !got[0].GovernorOk {
		t.Error("GovernorOk: expected true")
	}
	if !got[0].SafetyCircuitOk {
		t.Error("SafetyCircuitOk: expected true")
	}
}

func TestFetchAll_NullInjectedFaultDecodesAsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"unit_id":"ELV-001","injected_fault":null}]`))
	}))
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].InjectedFault != nil {
		t.Errorf("InjectedFault: expected nil, got %q", *got[0].InjectedFault)
	}
}

func TestFetchAll_NonNullInjectedFaultDecodesAsString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"unit_id":"ELV-001","injected_fault":"brake_wear"}]`))
	}))
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].InjectedFault == nil {
		t.Fatal("InjectedFault: expected non-nil")
	}
	if *got[0].InjectedFault != "brake_wear" {
		t.Errorf("InjectedFault: got %q, want %q", *got[0].InjectedFault, "brake_wear")
	}
}

func TestFetchAll_AllNumericFieldsPresent(t *testing.T) {
	payload := []simulator.ElevatorSnapshot{{
		UnitID:             "ELV-001",
		MotorCurrentA:      14.72,
		MotorTempC:         51.3,
		MotorRPM:           1447.0,
		MotorRunHours:      38900.5,
		TripCount:          2_891_000,
		DoorCycleCount:     3_780_000,
		DoorMotorAmps:      2.43,
		DoorCloseForceN:    118.2,
		DoorCloseTimeMs:    3250,
		BrakeEngagementCount: 2_891_000,
		BrakeCurrentA:      1.83,
		BrakeResponseMs:    71,
		LevelingAccuracyMm: 4.1,
		VibrationG:         0.07,
	}}
	srv := serve(t, http.StatusOK, payload)
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	g := got[0]
	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"MotorCurrentA", g.MotorCurrentA, 14.72},
		{"MotorTempC", g.MotorTempC, 51.3},
		{"MotorRPM", g.MotorRPM, 1447.0},
		{"MotorRunHours", g.MotorRunHours, 38900.5},
		{"DoorMotorAmps", g.DoorMotorAmps, 2.43},
		{"DoorCloseForceN", g.DoorCloseForceN, 118.2},
		{"BrakeCurrentA", g.BrakeCurrentA, 1.83},
		{"LevelingAccuracyMm", g.LevelingAccuracyMm, 4.1},
		{"VibrationG", g.VibrationG, 0.07},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %g, want %g", c.name, c.got, c.want)
		}
	}
}

func TestFetchAll_AbsentOrNullMetricIsMarkedMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// door_close_force_n renamed upstream; brake_response_ms explicitly null.
		_, _ = w.Write([]byte(`[{"unit_id":"ELV-003","door_close_force":118.0,"brake_response_ms":null,"motor_current_a":12.5}]`))
	}))
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	snap := got[0]
	for _, m := range []string{"door_close_force_n", "brake_response_ms"} {
		if !snap.IsMissing(m) {
			t.Errorf("%s: want missing", m)
		}
	}
	if snap.IsMissing("motor_current_a") {
		t.Error("motor_current_a: present in payload but reported missing")
	}
}

func TestFetchAll_ZeroValueIsNotMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"unit_id":"ELV-001","safety_circuit_ok":false,"door_obstruction_events":0}]`))
	}))
	defer srv.Close()

	got, err := simulator.NewHTTPClient(srv.URL).FetchAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].IsMissing("safety_circuit_ok") || got[0].IsMissing("door_obstruction_events") {
		t.Error("real zero/false readings reported missing")
	}
}
