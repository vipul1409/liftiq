package ingest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liftiq/telemetry-ingestor/internal/ingest"
	"github.com/liftiq/telemetry-ingestor/internal/simulator"
)

// readContract loads a file from the repo-level contracts/ directory shared
// with the Python simulator and the compliance engine.
func readContract(t *testing.T, name string, v any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "contracts", name))
	if err != nil {
		t.Fatalf("read contract %s: %v", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("decode contract %s: %v", name, err)
	}
}

// The golden simulator snapshot must decode and map to exactly the contract's
// metrics, every one with quality "good". A renamed json tag or mapper entry
// shows up here as a missing or unexpected metric.
// Each mapped value must equal the golden snapshot's raw JSON value. This
// catches a json tag that no longer matches the simulator key: the field then
// decodes as a zero value and would otherwise be stored as a real reading.
func TestContract_GoldenSnapshotValuesSurviveMapping(t *testing.T) {
	var raw map[string]any
	readContract(t, "simulator-snapshot.json", &raw)
	var snap simulator.ElevatorSnapshot
	readContract(t, "simulator-snapshot.json", &snap)

	for _, r := range ingest.MapSnapshot(snap, uuid.New(), time.Now()) {
		var want float64
		switch v := raw[r.Metric].(type) {
		case float64:
			want = v
		case bool:
			if v {
				want = 1
			}
		default:
			t.Errorf("%s: golden snapshot value %v (%T) is not numeric or bool", r.Metric, raw[r.Metric], raw[r.Metric])
			continue
		}
		if r.Value != want {
			t.Errorf("%s: mapped value %v, golden snapshot has %v", r.Metric, r.Value, want)
		}
	}
}

func TestContract_GoldenSnapshotMapsToContractMetrics(t *testing.T) {
	var want []string
	readContract(t, "telemetry-metrics.json", &want)
	var snap simulator.ElevatorSnapshot
	readContract(t, "simulator-snapshot.json", &snap)

	rows := ingest.MapSnapshot(snap, uuid.New(), time.Now())
	var got []string
	for _, r := range rows {
		got = append(got, r.Metric)
		if r.Quality != "good" {
			t.Errorf("%s: quality %q from golden snapshot, want good", r.Metric, r.Quality)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("mapped metrics %v, contract %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("mapped metric %q, contract has %q", got[i], want[i])
		}
	}
}
