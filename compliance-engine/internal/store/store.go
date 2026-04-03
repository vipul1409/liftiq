package store

import (
	"context"
	"errors"
	"time"
)

// ErrUnitNotFound is returned by FetchLatest when the unit tag has no record
// in the elevator_units table.
var ErrUnitNotFound = errors.New("unit not found")

// MetricReading is one metric value captured at a specific point in time.
type MetricReading struct {
	Metric string
	Value  float64
	Time   time.Time
}

// Store is the read-only interface the compliance engine uses to pull telemetry
// from TimescaleDB. Tests replace it with an in-process fake.
type Store interface {
	// ListUnits returns all known unit tags ordered alphabetically.
	ListUnits(ctx context.Context) ([]string, error)

	// FetchLatest returns the most recent reading for each metric that has at
	// least one row within [now-window, now] for the given unit.
	// Returns ErrUnitNotFound if the unit tag is not in elevator_units.
	FetchLatest(ctx context.Context, unitTag string, window time.Duration) ([]MetricReading, error)
}
