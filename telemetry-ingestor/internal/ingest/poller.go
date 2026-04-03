package ingest

import (
	"context"
	"log/slog"
	"time"

	"github.com/liftiq/telemetry-ingestor/internal/simulator"
	"github.com/liftiq/telemetry-ingestor/internal/store"
)

// Poller fetches elevator snapshots on a fixed interval and writes them to
// the TimescaleDB hypertable via the Store interface.
type Poller struct {
	client   simulator.Client
	store    store.Store
	interval time.Duration
	log      *slog.Logger
}

// New creates a Poller. client and st must be non-nil.
func New(client simulator.Client, st store.Store, interval time.Duration, log *slog.Logger) *Poller {
	return &Poller{
		client:   client,
		store:    st,
		interval: interval,
		log:      log,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
// An initial poll is fired immediately before the first ticker tick.
// Always returns nil — errors during individual polls are logged and the loop
// continues, matching the plan's requirement for a resilient ingestion service.
func (p *Poller) Run(ctx context.Context) error {
	p.log.Info("poller started", "interval", p.interval)

	// Poll immediately so the operator doesn't wait one full interval for data.
	p.PollOnce(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("poller stopped")
			return nil
		case <-ticker.C:
			p.PollOnce(ctx)
		}
	}
}

// PollOnce performs a single fetch-map-write cycle.
// It is exported so it can be called directly in tests without triggering
// the full ticker loop.
func (p *Poller) PollOnce(ctx context.Context) {
	start := time.Now()

	snapshots, err := p.client.FetchAll(ctx)
	if err != nil {
		p.log.Error("fetch failed", "err", err)
		return
	}

	// All rows in this poll share a single timestamp, truncated to the second.
	// This makes it easy to reconstruct a consistent "snapshot at T" query.
	capturedAt := start.UTC().Truncate(time.Second)

	var rows []store.Row
	for _, snap := range snapshots {
		unitID, err := p.store.LookupUnit(ctx, snap.UnitID)
		if err != nil {
			// One bad unit should not abort the whole poll — log and skip.
			p.log.Error("unit lookup failed", "unit_tag", snap.UnitID, "err", err)
			continue
		}
		rows = append(rows, MapSnapshot(snap, unitID, capturedAt)...)
	}

	if err := p.store.WriteRows(ctx, rows); err != nil {
		p.log.Error("write failed", "rows", len(rows), "err", err)
		return
	}

	p.log.Info("poll complete",
		"elevators", len(snapshots),
		"rows", len(rows),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}
