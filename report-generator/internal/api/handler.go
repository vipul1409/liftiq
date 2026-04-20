package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/liftiq/report-generator/internal/report"
)

// PDFRenderer is satisfied by report.ChromePDFRenderer and by test fakes.
type PDFRenderer interface {
	Render(ctx context.Context, req report.Request) ([]byte, error)
}

// Handler serves the report generation HTTP API.
type Handler struct {
	renderer PDFRenderer
	log      *slog.Logger
}

// New returns an http.Handler wired to all report routes.
func New(renderer PDFRenderer, log *slog.Logger) http.Handler {
	h := &Handler{renderer: renderer, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /reports", h.generateReport)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) generateReport(w http.ResponseWriter, r *http.Request) {
	var req report.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(fmt.Sprintf("invalid request body: %v", err)))
		return
	}

	if req.UnitTag == "" {
		writeJSON(w, http.StatusBadRequest, errBody("unit_tag is required"))
		return
	}
	if len(req.Results) == 0 {
		writeJSON(w, http.StatusBadRequest, errBody("results must not be empty"))
		return
	}

	pdfBytes, err := h.renderer.Render(r.Context(), req)
	if err != nil {
		h.log.Error("pdf render failed", "unit", req.UnitTag, "err", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to generate report"))
		return
	}

	filename := fmt.Sprintf("liftiq-report-%s.pdf", req.UnitTag)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func errBody(msg string) map[string]string { return map[string]string{"error": msg} }
