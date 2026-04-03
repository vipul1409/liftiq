package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Row is one metric reading for one elevator at one point in time.
// It is the shared data contract between the ingest and store packages.
type Row struct {
	Time    time.Time
	UnitID  uuid.UUID
	Metric  string
	Value   float64
	Quality string // "good" | "stale" | "missing"
}

// Store is the write surface for all elevator telemetry persistence.
type Store interface {
	// LookupUnit returns the UUID for the given simulator unit tag (e.g. "ELV-001").
	// Implementations must upsert the unit row on first encounter and cache results
	// to avoid a DB round-trip on every poll.
	LookupUnit(ctx context.Context, tag string) (uuid.UUID, error)

	// WriteRows bulk-inserts telemetry rows into the TimescaleDB hypertable.
	// All rows in a single call are written in one CopyFrom transaction.
	WriteRows(ctx context.Context, rows []Row) error
}
