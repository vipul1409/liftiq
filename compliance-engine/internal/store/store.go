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
	Metric  string
	Value   float64
	Time    time.Time
	Quality string // QualityGood or QualityMissing; only QualityGood values are real readings
}

// Metric quality values, as written by the telemetry-ingestor. Staleness is
// not a quality: FetchLatest's window already leaves out old readings.
const (
	QualityGood    = "good"
	QualityMissing = "missing"
)

// Store is the read-only interface the compliance engine uses to pull telemetry
// from TimescaleDB. Tests replace it with an in-process fake.
type Store interface {
	// ListUnits returns all known unit tags ordered alphabetically.
	ListUnits(ctx context.Context) ([]string, error)

	// FetchLatest returns the most recent reading for each metric that has at
	// least one row within [now-window, now] for the given unit, whatever its
	// quality — callers must not evaluate a reading whose Quality is not "good".
	// Returns ErrUnitNotFound if the unit tag is not in elevator_units.
	FetchLatest(ctx context.Context, unitTag string, window time.Duration) ([]MetricReading, error)
}
