package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client fetches elevator snapshots from the simulator HTTP API.
type Client interface {
	FetchAll(ctx context.Context) ([]ElevatorSnapshot, error)
}

type httpClient struct {
	baseURL string
	http    *http.Client
}

// NewHTTPClient returns a Client backed by the simulator at baseURL.
func NewHTTPClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchAll calls GET /elevators and returns the full elevator snapshot list.
func (c *httpClient) FetchAll(ctx context.Context) ([]ElevatorSnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/elevators", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /elevators: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("simulator returned HTTP %d", resp.StatusCode)
	}

	var snapshots []ElevatorSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshots); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return snapshots, nil
}
