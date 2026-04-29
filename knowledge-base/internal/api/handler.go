package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/liftiq/knowledge-base/internal/embed"
	"github.com/liftiq/knowledge-base/internal/llm"
	"github.com/liftiq/knowledge-base/internal/store"
)

// Handler serves the knowledge base HTTP API.
type Handler struct {
	store     store.Store
	embedder  embed.Embedder
	generator llm.Generator
	log       *slog.Logger
}

// New creates an HTTP handler with all routes registered.
func New(st store.Store, emb embed.Embedder, gen llm.Generator, log *slog.Logger) http.Handler {
	h := &Handler{store: st, embedder: emb, generator: gen, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /query", h.query)
	mux.HandleFunc("GET /documents", h.listDocuments)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// QueryRequest is the JSON body for POST /query.
type QueryRequest struct {
	Question        string `json:"question"`
	ControllerMake  string `json:"controller_make,omitempty"`
	ControllerModel string `json:"controller_model,omitempty"`
	TopK            int    `json:"top_k,omitempty"`
}

// QueryResult is a single search result in the response.
type QueryResult struct {
	ID              string  `json:"id"`
	ControllerMake  string  `json:"controller_make"`
	ControllerModel string  `json:"controller_model"`
	Section         string  `json:"section"`
	Content         string  `json:"content"`
	PageNumber      int     `json:"page_number"`
	DocumentName    string  `json:"document_name"`
	Similarity      float64 `json:"similarity"`
}

// QueryResponse is the JSON response for POST /query.
type QueryResponse struct {
	Query   string        `json:"query"`
	Answer  string        `json:"answer"`
	Results []QueryResult `json:"results"`
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errBody(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Question == "" {
		errBody(w, http.StatusBadRequest, "question is required")
		return
	}
	if req.TopK <= 0 {
		req.TopK = 3
	}

	// Embed the question.
	vec, err := h.embedder.Embed(r.Context(), req.Question)
	if err != nil {
		h.log.Error("embedding failed", "err", err)
		errBody(w, http.StatusInternalServerError, "failed to embed question")
		return
	}

	// Search pgvector for similar chunks.
	results, err := h.store.Search(r.Context(), vec, req.ControllerMake, req.ControllerModel, req.TopK)
	if err != nil {
		h.log.Error("search failed", "err", err)
		errBody(w, http.StatusInternalServerError, "search failed")
		return
	}

	// Build response results.
	qResults := make([]QueryResult, len(results))
	chunks := make([]string, len(results))
	for i, sr := range results {
		qResults[i] = QueryResult{
			ID:              sr.ID,
			ControllerMake:  sr.ControllerMake,
			ControllerModel: sr.ControllerModel,
			Section:         sr.Section,
			Content:         sr.Content,
			PageNumber:      sr.PageNumber,
			DocumentName:    sr.DocumentName,
			Similarity:      sr.Similarity,
		}
		chunks[i] = sr.Content
	}

	// Generate a synthesized answer via LLM.
	var answer string
	if len(chunks) > 0 {
		answer, err = h.generator.Generate(r.Context(), req.Question, chunks)
		if err != nil {
			h.log.Warn("LLM generation failed, returning chunks only", "err", err)
			answer = ""
		}
	}

	resp := QueryResponse{
		Query:   req.Question,
		Answer:  answer,
		Results: qResults,
	}
	if resp.Results == nil {
		resp.Results = []QueryResult{}
	}
	writeJSON(w, http.StatusOK, resp)
}

// DocumentResponse is a single document in the list response.
type DocumentResponse struct {
	ID              string `json:"id"`
	Filename        string `json:"filename"`
	ControllerMake  string `json:"controller_make"`
	ControllerModel string `json:"controller_model"`
	ChunkCount      int    `json:"chunk_count"`
	IngestedAt      string `json:"ingested_at"`
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.store.ListDocuments(r.Context())
	if err != nil {
		h.log.Error("list documents failed", "err", err)
		errBody(w, http.StatusInternalServerError, "failed to list documents")
		return
	}

	resp := make([]DocumentResponse, len(docs))
	for i, d := range docs {
		resp[i] = DocumentResponse{
			ID:              d.ID,
			Filename:        d.Filename,
			ControllerMake:  d.ControllerMake,
			ControllerModel: d.ControllerModel,
			ChunkCount:      d.ChunkCount,
			IngestedAt:      d.IngestedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, map[string][]DocumentResponse{"documents": resp})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func errBody(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
