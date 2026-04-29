package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaGenerator calls the Ollama /api/generate endpoint.
type OllamaGenerator struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaGenerator creates a generator backed by a local Ollama instance.
func NewOllamaGenerator(baseURL, model string) *OllamaGenerator {
	return &OllamaGenerator{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

func (g *OllamaGenerator) Generate(ctx context.Context, question string, chunks []string) (string, error) {
	context_ := strings.Join(chunks, "\n\n---\n\n")

	prompt := fmt.Sprintf(`You are LiftIQ, an elevator maintenance assistant. Answer the technician's question using ONLY the reference material provided below. Be concise and practical — include specific part numbers, wire pins, and repair steps when available. If the reference material does not contain enough information to answer, say so.

Reference material:
%s

Technician's question: %s

Answer:`, context_, question)

	body, err := json.Marshal(ollamaGenerateRequest{
		Model:  g.model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal generate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama generate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama generate returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode generate response: %w", err)
	}

	return strings.TrimSpace(result.Response), nil
}
