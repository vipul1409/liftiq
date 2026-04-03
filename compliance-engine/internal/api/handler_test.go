package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/liftiq/compliance-engine/internal/api"
	"github.com/liftiq/compliance-engine/internal/rules"
	"github.com/liftiq/compliance-engine/internal/store"
)

// ── Test double ───────────────────────────────────────────────────────────────

type fakeStore struct {
	units    []string
	readings map[string][]store.MetricReading // tag → readings
	listErr  error
	fetchErr error
}

func (f *fakeStore) ListUnits(_ context.Context) ([]string, error) {
	return f.units, f.listErr
}

func (f *fakeStore) FetchLatest(_ context.Context, tag string, _ time.Duration) ([]store.MetricReading, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	readings, ok := f.readings[tag]
	if !ok {
		return nil, store.ErrUnitNotFound
	}
	return readings, nil
}

// allGoodReadings returns one MetricReading per ASME rule at a passing value.
func allGoodReadings() []store.MetricReading {
	vals := map[string]float64{
		"motor_current_a":         10.0,
		"motor_temp_c":            60.0,
		"motor_rpm":               1200.0,
		"motor_run_hours":         5000.0,
		"trip_count":              100000.0,
		"door_close_force_n":      90.0,
		"door_motor_amps":         2.0,
		"door_close_time_ms":      2500.0,
		"door_cycle_count":        500000.0,
		"door_obstruction_events": 2.0,
		"brake_response_ms":       50.0,
		"brake_current_a":         1.5,
		"brake_engagement_count":  100000.0,
		"leveling_accuracy_mm":    5.0,
		"vibration_g":             0.05,
		"door_interlock_ok":       1.0,
		"governor_ok":             1.0,
		"buffer_ok":               1.0,
		"pit_switch_ok":           1.0,
		"safety_circuit_ok":       1.0,
	}
	now := time.Now().UTC()
	out := make([]store.MetricReading, 0, len(vals))
	for metric, value := range vals {
		out = append(out, store.MetricReading{Metric: metric, Value: value, Time: now})
	}
	return out
}

func newHandler(st store.Store) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return api.New(st, 10*time.Minute, log)
}

// ── Request helpers ───────────────────────────────────────────────────────────

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, w.Body.String())
	}
}

// ── Health ────────────────────────────────────────────────────────────────────

func TestHealth_Returns200(t *testing.T) {
	w := get(t, newHandler(&fakeStore{}), "/health")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestHealth_ReturnsOKStatus(t *testing.T) {
	w := get(t, newHandler(&fakeStore{}), "/health")
	var body map[string]string
	decodeJSON(t, w, &body)
	if body["status"] != "ok" {
		t.Errorf("body.status: got %q, want %q", body["status"], "ok")
	}
}

func TestHealth_ContentTypeIsJSON(t *testing.T) {
	w := get(t, newHandler(&fakeStore{}), "/health")
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

// ── List units ────────────────────────────────────────────────────────────────

func TestListUnits_Returns200(t *testing.T) {
	st := &fakeStore{units: []string{"ELV-001", "ELV-002"}}
	w := get(t, newHandler(st), "/units")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestListUnits_ReturnsUnitSlice(t *testing.T) {
	st := &fakeStore{units: []string{"ELV-001", "ELV-002", "ELV-003"}}
	w := get(t, newHandler(st), "/units")
	var body map[string][]string
	decodeJSON(t, w, &body)
	if len(body["units"]) != 3 {
		t.Errorf("units count: got %d, want 3", len(body["units"]))
	}
}

func TestListUnits_EmptyListReturnsEmptyArray(t *testing.T) {
	st := &fakeStore{units: nil}
	w := get(t, newHandler(st), "/units")
	var body map[string][]string
	decodeJSON(t, w, &body)
	if len(body["units"]) != 0 {
		t.Errorf("expected empty array, got %v", body["units"])
	}
}

func TestListUnits_StoreError_Returns500(t *testing.T) {
	st := &fakeStore{listErr: errors.New("db down")}
	w := get(t, newHandler(st), "/units")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── Compliance (full results) ─────────────────────────────────────────────────

func TestCompliance_AllPass_Returns200(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestCompliance_Returns20Results(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if len(body.Results) != 20 {
		t.Errorf("results count: got %d, want 20", len(body.Results))
	}
}

func TestCompliance_AllPassSummary(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.Summary.Pass != 20 || body.Summary.Fail != 0 || body.Summary.Unknown != 0 {
		t.Errorf("summary: got pass=%d fail=%d unknown=%d, want 20/0/0",
			body.Summary.Pass, body.Summary.Fail, body.Summary.Unknown)
	}
	if body.Summary.Overall != rules.Pass {
		t.Errorf("overall: got %s, want pass", body.Summary.Overall)
	}
}

func TestCompliance_UnitTagInResponse(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-002": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-002/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.UnitTag != "ELV-002" {
		t.Errorf("unit_tag: got %q, want ELV-002", body.UnitTag)
	}
}

func TestCompliance_FailingMetric_OverallFail(t *testing.T) {
	readings := allGoodReadings()
	for i, r := range readings {
		if r.Metric == "door_close_force_n" {
			readings[i].Value = 200.0 // exceeds 135 N limit
		}
	}
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-003": readings}}
	w := get(t, newHandler(st), "/units/ELV-003/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.Summary.Overall != rules.Fail {
		t.Errorf("overall: got %s, want fail", body.Summary.Overall)
	}
	if body.Summary.Fail < 1 {
		t.Errorf("fail count: got %d, want ≥ 1", body.Summary.Fail)
	}
}

func TestCompliance_UnknownUnit_Returns404(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{}}
	w := get(t, newHandler(st), "/units/ELV-GHOST/compliance")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestCompliance_StoreError_Returns500(t *testing.T) {
	st := &fakeStore{fetchErr: errors.New("db down")}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestCompliance_EmptyReadings_AllUnknown(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": {}}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.Summary.Unknown != 20 {
		t.Errorf("unknown count: got %d, want 20", body.Summary.Unknown)
	}
	if body.Summary.Overall != rules.Unknown {
		t.Errorf("overall: got %s, want unknown", body.Summary.Overall)
	}
}

func TestCompliance_StaleWindowInResponse(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.StaleWindowMinutes != 10.0 {
		t.Errorf("stale_window_minutes: got %g, want 10", body.StaleWindowMinutes)
	}
}

func TestCompliance_AsOfIsSet(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	before := time.Now().UTC().Add(-time.Second)
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	after := time.Now().UTC().Add(time.Second)
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.AsOf.Before(before) || body.AsOf.After(after) {
		t.Errorf("as_of %v outside window [%v, %v]", body.AsOf, before, after)
	}
}

// ── Safety circuit fault (the most critical auto-fail) ───────────────────────

func TestCompliance_SafetyCircuitFault_AutoFail(t *testing.T) {
	readings := allGoodReadings()
	for i, r := range readings {
		if r.Metric == "safety_circuit_ok" {
			readings[i].Value = 0.0
		}
	}
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": readings}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance")
	var body api.ComplianceResponse
	decodeJSON(t, w, &body)
	if body.Summary.Overall != rules.Fail {
		t.Errorf("overall: got %s, want fail", body.Summary.Overall)
	}
}

// ── Compliance summary ────────────────────────────────────────────────────────

func TestComplianceSummary_Returns200(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance/summary")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestComplianceSummary_CorrectCounts(t *testing.T) {
	readings := allGoodReadings()
	for i, r := range readings {
		if r.Metric == "safety_circuit_ok" {
			readings[i].Value = 0.0 // one failure
		}
	}
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": readings}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance/summary")
	var body api.SummaryResponse
	decodeJSON(t, w, &body)
	if body.Summary.Fail != 1 {
		t.Errorf("fail count: got %d, want 1", body.Summary.Fail)
	}
	if body.Summary.Pass != 19 {
		t.Errorf("pass count: got %d, want 19", body.Summary.Pass)
	}
}

func TestComplianceSummary_UnknownUnit_Returns404(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{}}
	w := get(t, newHandler(st), "/units/ELV-GHOST/compliance/summary")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestComplianceSummary_DoesNotIncludeFullResults(t *testing.T) {
	st := &fakeStore{readings: map[string][]store.MetricReading{"ELV-001": allGoodReadings()}}
	w := get(t, newHandler(st), "/units/ELV-001/compliance/summary")
	// Decode as raw map to verify no "results" key is present.
	var raw map[string]json.RawMessage
	decodeJSON(t, w, &raw)
	if _, ok := raw["results"]; ok {
		t.Error("summary response should not include 'results' key")
	}
}

// ── Route method and path guard ───────────────────────────────────────────────

func TestUnknownPath_Returns404(t *testing.T) {
	w := get(t, newHandler(&fakeStore{}), "/nonexistent")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestWrongMethod_Returns405(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	newHandler(&fakeStore{}).ServeHTTP(w, req)
	// Go 1.22+ ServeMux returns 405 for wrong method on a registered path.
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", w.Code)
	}
}
