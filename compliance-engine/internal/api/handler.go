package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/liftiq/compliance-engine/internal/rules"
	"github.com/liftiq/compliance-engine/internal/store"
)

// Handler serves the compliance HTTP API.
type Handler struct {
	store       store.Store
	staleWindow time.Duration
	log         *slog.Logger
}

// New returns an http.Handler wired to all compliance routes.
func New(st store.Store, staleWindow time.Duration, log *slog.Logger) http.Handler {
	h := &Handler{store: st, staleWindow: staleWindow, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /units", h.listUnits)
	mux.HandleFunc("GET /units/{tag}/compliance", h.compliance)
	mux.HandleFunc("GET /units/{tag}/compliance/summary", h.complianceSummary)
	return mux
}

// ── Response types ────────────────────────────────────────────────────────────

// ComplianceResponse is the JSON body for GET /units/{tag}/compliance.
type ComplianceResponse struct {
	UnitTag            string         `json:"unit_tag"`
	AsOf               time.Time      `json:"as_of"`
	StaleWindowMinutes float64        `json:"stale_window_minutes"`
	Results            []rules.Result `json:"results"`
	Summary            rules.Summary  `json:"summary"`
}

// SummaryResponse is the JSON body for GET /units/{tag}/compliance/summary.
type SummaryResponse struct {
	UnitTag            string        `json:"unit_tag"`
	AsOf               time.Time     `json:"as_of"`
	StaleWindowMinutes float64       `json:"stale_window_minutes"`
	Summary            rules.Summary `json:"summary"`
}

// ── Route handlers ─────────────────────────────────────────────────────────────

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listUnits(w http.ResponseWriter, r *http.Request) {
	tags, err := h.store.ListUnits(r.Context())
	if err != nil {
		h.log.Error("list units failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to list units"))
		return
	}
	if tags == nil {
		tags = []string{}
	}
	writeJSON(w, http.StatusOK, map[string][]string{"units": tags})
}

func (h *Handler) compliance(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	asOf := time.Now().UTC()

	results, ok := h.fetchAndEvaluate(w, r, tag)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, ComplianceResponse{
		UnitTag:            tag,
		AsOf:               asOf,
		StaleWindowMinutes: h.staleWindow.Minutes(),
		Results:            results,
		Summary:            rules.Summarise(results),
	})
}

func (h *Handler) complianceSummary(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	asOf := time.Now().UTC()

	results, ok := h.fetchAndEvaluate(w, r, tag)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, SummaryResponse{
		UnitTag:            tag,
		AsOf:               asOf,
		StaleWindowMinutes: h.staleWindow.Minutes(),
		Summary:            rules.Summarise(results),
	})
}

// ── Internal helpers ───────────────────────────────────────────────────────────

// fetchAndEvaluate fetches latest telemetry for tag, runs all 20 ASME rules,
// and returns the results. On error it writes the error response to w and
// returns false; callers must not write any further response in that case.
func (h *Handler) fetchAndEvaluate(w http.ResponseWriter, r *http.Request, tag string) ([]rules.Result, bool) {
	readings, err := h.store.FetchLatest(r.Context(), tag, h.staleWindow)
	if errors.Is(err, store.ErrUnitNotFound) {
		writeJSON(w, http.StatusNotFound, errBody("unit not found: "+tag))
		return nil, false
	}
	if err != nil {
		h.log.Error("fetch latest telemetry failed", "unit", tag, "err", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to fetch telemetry"))
		return nil, false
	}

	// Only good readings are evaluated. When the latest reading for a metric is
	// not good (e.g. missing), leaving it out makes its rule Unknown immediately
	// rather than evaluating a placeholder value.
	values := make(rules.MetricValues, len(readings))
	for _, rd := range readings {
		if rd.Quality != store.QualityGood {
			continue
		}
		values[rd.Metric] = rd.Value
	}
	return rules.EvaluateAll(values), true
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func errBody(msg string) map[string]string { return map[string]string{"error": msg} }
