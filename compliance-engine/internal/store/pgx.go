package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgxpool connection to dsn and verifies it with a ping.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

type pgxStore struct {
	pool *pgxpool.Pool
}

// NewPGXStore returns a Store backed by the given connection pool.
func NewPGXStore(pool *pgxpool.Pool) Store {
	return &pgxStore{pool: pool}
}

// ListUnits returns all unit tags in the elevator_units table, alphabetically.
func (s *pgxStore) ListUnits(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT unit_tag FROM elevator_units ORDER BY unit_tag`,
	)
	if err != nil {
		return nil, fmt.Errorf("query elevator_units: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan unit_tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// FetchLatest returns the most recent reading within [now-window, now] for each
// metric belonging to unitTag. Metrics with no row in that window are absent
// from the returned slice (callers treat them as Unknown).
// Returns ErrUnitNotFound if unitTag has no row in elevator_units.
func (s *pgxStore) FetchLatest(ctx context.Context, unitTag string, window time.Duration) ([]MetricReading, error) {
	// Resolve the unit UUID; fail fast if unknown.
	var unitID [16]byte
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM elevator_units WHERE unit_tag = $1`,
		unitTag,
	).Scan(&unitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lookup elevator_units %q: %w", unitTag, err)
	}

	// DISTINCT ON (metric) picks the row with the greatest time per metric.
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (metric) metric, value, time, quality
		FROM telemetry
		WHERE unit_id = $1
		  AND time > NOW() - $2::interval
		ORDER BY metric, time DESC
	`, unitID, window.String())
	if err != nil {
		return nil, fmt.Errorf("query telemetry for %q: %w", unitTag, err)
	}
	defer rows.Close()

	var readings []MetricReading
	for rows.Next() {
		var r MetricReading
		if err := rows.Scan(&r.Metric, &r.Value, &r.Time, &r.Quality); err != nil {
			return nil, fmt.Errorf("scan telemetry row: %w", err)
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}
