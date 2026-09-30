package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/liftiq/report-generator/internal/report"
)

// ── Test double ───────────────────────────────────────────────────────────────

type fakePDFRenderer struct {
	err    error
	called bool
	last   report.Inspection
}

func (f *fakePDFRenderer) Render(_ context.Context, in report.Inspection) ([]byte, error) {
	f.called = true
	f.last = in
	if f.err != nil {
		return nil, f.err
	}
	// Minimal valid PDF header so tests can assert on Content-Type.
	return []byte("%PDF-1.4 fake"), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newHandler(r *fakePDFRenderer) http.Handler {
	return New(r, noopLogger())
}

func post(t *testing.T, handler http.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func validRequest() report.Request {
	return report.Request{
		UnitTag:     "ELV-003",
		InspectedAt: time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC),
		Technician:  "J. Smith",
		Results:     fakeResults(),
	}
}

func fakeResults() []report.RuleResult {
	return []report.RuleResult{
		{
			RuleID: "ASME-001", Description: "Motor current within limit",
			ASMERef: "ASME A17.1 2.10.1", Metric: "motor_current_a",
			Value: 15.0, Threshold: 20.0, Unit: "A",
			Status: report.StatusPass, Message: "within limit",
		},
	}
}

// ── GET /health ───────────────────────────────────────────────────────────────

func TestHealth_Returns200(t *testing.T) {
	rr := get(t, newHandler(&fakePDFRenderer{}), "/health")
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestHealth_ReturnsStatusOK(t *testing.T) {
	rr := get(t, newHandler(&fakePDFRenderer{}), "/health")
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestHealth_ContentTypeIsJSON(t *testing.T) {
	rr := get(t, newHandler(&fakePDFRenderer{}), "/health")
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

// ── POST /reports — success ───────────────────────────────────────────────────

func TestGenerateReport_Returns200OnSuccess(t *testing.T) {
	rr := post(t, newHandler(&fakePDFRenderer{}), validRequest())
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestGenerateReport_ContentTypeIsPDF(t *testing.T) {
	rr := post(t, newHandler(&fakePDFRenderer{}), validRequest())
	ct := rr.Header().Get("Content-Type")
	if ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
}

func TestGenerateReport_ContentDispositionContainsUnitTag(t *testing.T) {
	rr := post(t, newHandler(&fakePDFRenderer{}), validRequest())
	cd := rr.Header().Get("Content-Disposition")
	if cd == "" {
		t.Fatal("missing Content-Disposition header")
	}
	if !containsStr(cd, "ELV-003") {
		t.Errorf("Content-Disposition %q does not contain unit tag ELV-003", cd)
	}
}

func TestGenerateReport_BodyContainsPDFBytes(t *testing.T) {
	rr := post(t, newHandler(&fakePDFRenderer{}), validRequest())
	body := rr.Body.Bytes()
	if len(body) == 0 {
		t.Fatal("expected non-empty PDF body")
	}
}

func TestGenerateReport_CallsRenderer(t *testing.T) {
	fake := &fakePDFRenderer{}
	post(t, newHandler(fake), validRequest())
	if !fake.called {
		t.Error("renderer was not called")
	}
}

func TestGenerateReport_PassesUnitTagToRenderer(t *testing.T) {
	fake := &fakePDFRenderer{}
	post(t, newHandler(fake), validRequest())
	if fake.last.UnitTag != "ELV-003" {
		t.Errorf("renderer received UnitTag = %q, want ELV-003", fake.last.UnitTag)
	}
}

func TestGenerateReport_PassesResultsToRenderer(t *testing.T) {
	fake := &fakePDFRenderer{}
	post(t, newHandler(fake), validRequest())
	if len(fake.last.Rows) != 1 {
		t.Errorf("renderer received %d rows, want 1", len(fake.last.Rows))
	}
}

func TestGenerateReport_OverrideFailReachesRendererSummary(t *testing.T) {
	fake := &fakePDFRenderer{}
	req := validRequest()
	req.Overrides = map[string]report.Status{"ASME-001": report.StatusFail}
	rr := post(t, newHandler(fake), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	want := report.Summary{Pass: 0, Fail: 1, Unknown: 0, Overall: report.StatusFail}
	if fake.last.Summary != want {
		t.Errorf("renderer summary = %+v, want %+v", fake.last.Summary, want)
	}
	row := fake.last.Rows[0]
	if row.Effective != report.StatusFail || !row.Overridden || row.Status != report.StatusPass {
		t.Errorf("row = effective %q overridden %v telemetry %q; want fail/true/pass",
			row.Effective, row.Overridden, row.Status)
	}
}

func TestGenerateReport_IgnoresClientSuppliedSummary(t *testing.T) {
	fake := &fakePDFRenderer{}
	body := map[string]any{
		"unit_tag": "ELV-003",
		"results":  fakeResults(),
		"summary":  map[string]any{"pass": 0, "fail": 5, "unknown": 0, "overall": "fail"},
	}
	post(t, newHandler(fake), body)
	if fake.last.Summary.Overall != report.StatusPass {
		t.Errorf("renderer summary overall = %q, want pass (derived from results)", fake.last.Summary.Overall)
	}
}

// ── POST /reports — validation errors ────────────────────────────────────────

func TestGenerateReport_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newHandler(&fakePDFRenderer{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestGenerateReport_Returns400WhenUnitTagMissing(t *testing.T) {
	req := validRequest()
	req.UnitTag = ""
	rr := post(t, newHandler(&fakePDFRenderer{}), req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestGenerateReport_Returns400WhenResultsEmpty(t *testing.T) {
	req := validRequest()
	req.Results = nil
	rr := post(t, newHandler(&fakePDFRenderer{}), req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestGenerateReport_Returns400WhenOverrideNamesUnknownRule(t *testing.T) {
	fake := &fakePDFRenderer{}
	req := validRequest()
	req.Overrides = map[string]report.Status{"ASME-999": report.StatusPass}
	rr := post(t, newHandler(fake), req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	if fake.called {
		t.Error("renderer called for an invalid request")
	}
}

func TestGenerateReport_Returns400WhenOverrideStatusInvalid(t *testing.T) {
	req := validRequest()
	req.Overrides = map[string]report.Status{"ASME-001": report.StatusUnknown}
	rr := post(t, newHandler(&fakePDFRenderer{}), req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestGenerateReport_ErrorBodyContainsErrorKey(t *testing.T) {
	req := validRequest()
	req.UnitTag = ""
	rr := post(t, newHandler(&fakePDFRenderer{}), req)
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body["error"] == "" {
		t.Error("error body missing 'error' key")
	}
}

// ── POST /reports — renderer error ───────────────────────────────────────────

func TestGenerateReport_Returns500WhenRendererFails(t *testing.T) {
	fake := &fakePDFRenderer{err: errFakeRender}
	rr := post(t, newHandler(fake), validRequest())
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestGenerateReport_RendererErrorBodyHasErrorKey(t *testing.T) {
	fake := &fakePDFRenderer{err: errFakeRender}
	rr := post(t, newHandler(fake), validRequest())
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body["error"] == "" {
		t.Error("error body missing 'error' key")
	}
}

// ── Route guards ──────────────────────────────────────────────────────────────

func TestUnknownPath_Returns404(t *testing.T) {
	rr := get(t, newHandler(&fakePDFRenderer{}), "/nonexistent")
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestWrongMethod_Returns405(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/reports", nil)
	rr := httptest.NewRecorder()
	newHandler(&fakePDFRenderer{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rr.Code)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var errFakeRender = errors.New("fake render error")

func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}
