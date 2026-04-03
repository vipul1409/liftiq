package ingest_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liftiq/telemetry-ingestor/internal/ingest"
	"github.com/liftiq/telemetry-ingestor/internal/simulator"
	"github.com/liftiq/telemetry-ingestor/internal/store"
)

// ---------------------------------------------------------------------------
// Test doubles — both are safe for concurrent use.
// The poller goroutine reads these while tests may mutate them between ticks.
// ---------------------------------------------------------------------------

type fakeClient struct {
	mu        sync.Mutex
	snapshots []simulator.ElevatorSnapshot
	err       error
	callCount int
}

func (f *fakeClient) FetchAll(_ context.Context) ([]simulator.ElevatorSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	return f.snapshots, f.err
}

func (f *fakeClient) setErr(err error) {
	f.mu.Lock()
	f.err = err
	f.mu.Unlock()
}

func (f *fakeClient) setSnapshots(snaps []simulator.ElevatorSnapshot) {
	f.mu.Lock()
	f.snapshots = snaps
	f.mu.Unlock()
}

func (f *fakeClient) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

type fakeStore struct {
	mu       sync.Mutex
	units    map[string]uuid.UUID
	unitErr  error
	rows     []store.Row
	writeErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		units: map[string]uuid.UUID{
			"ELV-001": uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			"ELV-002": uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			"ELV-003": uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		},
	}
}

func (f *fakeStore) LookupUnit(_ context.Context, tag string) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unitErr != nil {
		return uuid.Nil, f.unitErr
	}
	id, ok := f.units[tag]
	if !ok {
		return uuid.Nil, errors.New("unknown unit tag: " + tag)
	}
	return id, nil
}

func (f *fakeStore) WriteRows(_ context.Context, rows []store.Row) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.writeErr != nil {
		return f.writeErr
	}
	f.rows = append(f.rows, rows...)
	return nil
}

// rowCount returns the number of rows written so far.
func (f *fakeStore) rowCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.rows)
}

// allRows returns a copy of all written rows.
func (f *fakeStore) allRows() []store.Row {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]store.Row, len(f.rows))
	copy(out, f.rows)
	return out
}

// ---------------------------------------------------------------------------
// Logger helpers
// ---------------------------------------------------------------------------

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func bufLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func threeElevatorSnapshots() []simulator.ElevatorSnapshot {
	return []simulator.ElevatorSnapshot{
		{UnitID: "ELV-001", MotorCurrentA: 12.5, SafetyCircuitOk: true},
		{UnitID: "ELV-002", MotorCurrentA: 13.8, SafetyCircuitOk: true},
		{UnitID: "ELV-003", MotorCurrentA: 15.1, SafetyCircuitOk: true},
	}
}

// ---------------------------------------------------------------------------
// PollOnce tests
// ---------------------------------------------------------------------------

func TestPollOnce_Writes20RowsPerElevator(t *testing.T) {
	snaps := threeElevatorSnapshots()
	client := &fakeClient{snapshots: snaps}
	st := newFakeStore()

	ingest.New(client, st, time.Second, discardLogger()).PollOnce(context.Background())

	want := len(snaps) * 20
	if st.rowCount() != want {
		t.Errorf("got %d rows, want %d", st.rowCount(), want)
	}
}

func TestPollOnce_FetchErrorDoesNotWrite(t *testing.T) {
	var buf bytes.Buffer
	client := &fakeClient{}
	client.setErr(errors.New("simulator down"))
	st := newFakeStore()

	ingest.New(client, st, time.Second, bufLogger(&buf)).PollOnce(context.Background())

	if st.rowCount() != 0 {
		t.Errorf("expected 0 rows on fetch error, got %d", st.rowCount())
	}
	if !strings.Contains(buf.String(), "fetch failed") {
		t.Errorf("expected 'fetch failed' log line, got: %s", buf.String())
	}
}

func TestPollOnce_WriteErrorIsLogged(t *testing.T) {
	var buf bytes.Buffer
	client := &fakeClient{snapshots: []simulator.ElevatorSnapshot{{UnitID: "ELV-001"}}}
	st := newFakeStore()
	st.writeErr = errors.New("db down")

	ingest.New(client, st, time.Second, bufLogger(&buf)).PollOnce(context.Background())

	if !strings.Contains(buf.String(), "write failed") {
		t.Errorf("expected 'write failed' log line, got: %s", buf.String())
	}
}

func TestPollOnce_UnknownUnitSkippedOthersStillWritten(t *testing.T) {
	snaps := []simulator.ElevatorSnapshot{
		{UnitID: "ELV-001"},
		{UnitID: "ELV-UNKNOWN"},
		{UnitID: "ELV-003"},
	}
	var buf bytes.Buffer
	client := &fakeClient{snapshots: snaps}
	st := newFakeStore()

	ingest.New(client, st, time.Second, bufLogger(&buf)).PollOnce(context.Background())

	want := 2 * 20 // ELV-001 + ELV-003, ELV-UNKNOWN skipped
	if st.rowCount() != want {
		t.Errorf("got %d rows, want %d", st.rowCount(), want)
	}
	if !strings.Contains(buf.String(), "unit lookup failed") {
		t.Errorf("expected 'unit lookup failed' log line, got: %s", buf.String())
	}
}

func TestPollOnce_AllRowsShareCapturedAtTimestamp(t *testing.T) {
	client := &fakeClient{snapshots: threeElevatorSnapshots()}
	st := newFakeStore()

	before := time.Now().UTC().Truncate(time.Second)
	ingest.New(client, st, time.Second, discardLogger()).PollOnce(context.Background())
	after := time.Now().UTC().Truncate(time.Second).Add(time.Second)

	for _, r := range st.allRows() {
		if r.Time.Before(before) || r.Time.After(after) {
			t.Errorf("row time %v outside window [%v, %v]", r.Time, before, after)
		}
	}
}

func TestPollOnce_AllRowsHaveGoodQuality(t *testing.T) {
	client := &fakeClient{snapshots: threeElevatorSnapshots()}
	st := newFakeStore()

	ingest.New(client, st, time.Second, discardLogger()).PollOnce(context.Background())

	for _, r := range st.allRows() {
		if r.Quality != "good" {
			t.Errorf("metric %q: quality %q, want %q", r.Metric, r.Quality, "good")
		}
	}
}

func TestPollOnce_UnitIDsMatchLookup(t *testing.T) {
	client := &fakeClient{snapshots: []simulator.ElevatorSnapshot{
		{UnitID: "ELV-001"},
		{UnitID: "ELV-002"},
	}}
	st := newFakeStore()

	ingest.New(client, st, time.Second, discardLogger()).PollOnce(context.Background())

	uid1 := st.units["ELV-001"]
	uid2 := st.units["ELV-002"]
	for _, r := range st.allRows() {
		if r.UnitID != uid1 && r.UnitID != uid2 {
			t.Errorf("row has unexpected unit_id %v", r.UnitID)
		}
	}
}

func TestPollOnce_EmptySnapshotListWritesZeroRows(t *testing.T) {
	client := &fakeClient{snapshots: []simulator.ElevatorSnapshot{}}
	st := newFakeStore()

	ingest.New(client, st, time.Second, discardLogger()).PollOnce(context.Background())

	if st.rowCount() != 0 {
		t.Errorf("got %d rows for empty snapshot list, want 0", st.rowCount())
	}
}

func TestPollOnce_ConsecutivePollsAccumulateRows(t *testing.T) {
	client := &fakeClient{snapshots: []simulator.ElevatorSnapshot{{UnitID: "ELV-001"}}}
	st := newFakeStore()
	p := ingest.New(client, st, time.Second, discardLogger())

	p.PollOnce(context.Background())
	p.PollOnce(context.Background())

	want := 2 * 20
	if st.rowCount() != want {
		t.Errorf("after 2 polls: got %d rows, want %d", st.rowCount(), want)
	}
}

// ---------------------------------------------------------------------------
// Run loop tests
// ---------------------------------------------------------------------------

func TestRun_StopsOnContextCancel(t *testing.T) {
	client := &fakeClient{snapshots: []simulator.ElevatorSnapshot{}}
	st := newFakeStore()

	p := ingest.New(client, st, 100*time.Millisecond, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned non-nil error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop within 2s after context cancellation")
	}
}

func TestRun_PollsImmediatelyOnStart(t *testing.T) {
	// interval is 10s — if Run polls immediately, we see a call within 200ms
	client := &fakeClient{}
	client.setSnapshots([]simulator.ElevatorSnapshot{})
	st := newFakeStore()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ingest.New(client, st, 10*time.Second, discardLogger()).Run(ctx)

	time.Sleep(100 * time.Millisecond)
	if client.calls() == 0 {
		t.Error("expected at least one FetchAll call immediately on Run start")
	}
}

func TestRun_TickerFiresMultipleTimes(t *testing.T) {
	client := &fakeClient{}
	client.setSnapshots([]simulator.ElevatorSnapshot{})
	st := newFakeStore()

	ctx, cancel := context.WithCancel(context.Background())

	go ingest.New(client, st, 30*time.Millisecond, discardLogger()).Run(ctx)

	// Allow ~4 ticks (1 immediate + ~3 ticker)
	time.Sleep(120 * time.Millisecond)
	cancel()

	if client.calls() < 3 {
		t.Errorf("expected ≥ 3 polls in 120ms at 30ms interval, got %d", client.calls())
	}
}

func TestRun_ContinuesAfterFetchError(t *testing.T) {
	// Simulator starts down; after 50ms it recovers. The poller must survive
	// the errors and successfully write rows once the simulator is healthy.
	client := &fakeClient{}
	client.setErr(errors.New("transient error"))

	st := newFakeStore()
	ctx, cancel := context.WithCancel(context.Background())

	go ingest.New(client, st, 30*time.Millisecond, discardLogger()).Run(ctx)

	// Recover: clear the error and provide a snapshot.
	time.Sleep(50 * time.Millisecond)
	client.setErr(nil)
	client.setSnapshots([]simulator.ElevatorSnapshot{{UnitID: "ELV-001"}})

	time.Sleep(80 * time.Millisecond)
	cancel()

	if client.calls() < 2 {
		t.Errorf("expected ≥ 2 FetchAll calls (error then success), got %d", client.calls())
	}
	if st.rowCount() == 0 {
		t.Error("expected rows to be written after error cleared, got 0")
	}
}
