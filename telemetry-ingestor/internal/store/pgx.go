package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// schema contains the full TimescaleDB DDL run on every startup.
// All statements are idempotent (IF NOT EXISTS / on_conflict guards).
// Each entry is executed as a separate Exec call because pgx v5 does not
// support multi-statement strings in a single Exec.
var migrationStmts = []string{
	// Extension
	`CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE`,

	// Elevator unit registry (unit_tag has a UNIQUE constraint for ON CONFLICT upserts)
	`CREATE TABLE IF NOT EXISTS elevator_units (
		id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
		unit_tag        VARCHAR(50)  NOT NULL UNIQUE,
		building_id     UUID,
		controller_make VARCHAR(100),
		controller_model VARCHAR(100),
		protocol        VARCHAR(20)  NOT NULL DEFAULT 'simulator',
		bms_address     VARCHAR(100),
		installed_date  DATE,
		created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`,

	// Telemetry hypertable (narrow/long format: one row per metric per poll)
	`CREATE TABLE IF NOT EXISTS telemetry (
		time    TIMESTAMPTZ      NOT NULL,
		unit_id UUID             NOT NULL REFERENCES elevator_units(id),
		metric  VARCHAR(50)      NOT NULL,
		value   DOUBLE PRECISION NOT NULL,
		quality VARCHAR(10)      NOT NULL DEFAULT 'good'
	)`,

	// Convert telemetry to a TimescaleDB hypertable partitioned by time
	`SELECT create_hypertable('telemetry', 'time', if_not_exists => TRUE)`,

	// Add space partition by unit_id (hash, 4 partitions).
	// Idempotent: the DO block checks timescaledb_information.dimensions first.
	// Space partitioning means data for each elevator is co-located within a
	// time chunk, making per-unit range queries more efficient.
	`DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM timescaledb_information.dimensions
			WHERE hypertable_name = 'telemetry' AND column_name = 'unit_id'
		) THEN
			PERFORM add_dimension('telemetry', 'unit_id', number_of_partitions => 4);
		END IF;
	END $$`,

	// Composite index: unit + metric + recency — the primary query pattern
	// for the compliance engine ("last N readings of brake_response_ms for ELV-003")
	`CREATE INDEX IF NOT EXISTS idx_telemetry_unit_metric
		ON telemetry (unit_id, metric, time DESC)`,
}

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

// Migrate runs all DDL statements required to bootstrap the schema.
// Safe to call on every startup — all statements are idempotent.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	for i, stmt := range migrationStmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			// Include a short prefix of the statement so the error is actionable
			preview := stmt
			if len(preview) > 60 {
				preview = preview[:60] + "…"
			}
			return fmt.Errorf("migration step %d (%s): %w", i+1, preview, err)
		}
	}
	return nil
}

// pgxStore is the production Store backed by a pgxpool.
type pgxStore struct {
	pool      *pgxpool.Pool
	unitCache sync.Map // string tag → uuid.UUID
}

// NewPGXStore returns a Store backed by the given connection pool.
func NewPGXStore(pool *pgxpool.Pool) Store {
	return &pgxStore{pool: pool}
}

// LookupUnit upserts the elevator_units row for tag on first call and caches
// the result so subsequent calls never hit the database.
func (s *pgxStore) LookupUnit(ctx context.Context, tag string) (uuid.UUID, error) {
	if v, ok := s.unitCache.Load(tag); ok {
		return v.(uuid.UUID), nil
	}

	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO elevator_units (unit_tag, protocol)
		VALUES ($1, 'simulator')
		ON CONFLICT (unit_tag) DO UPDATE SET unit_tag = EXCLUDED.unit_tag
		RETURNING id
	`, tag).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert elevator_units %q: %w", tag, err)
	}

	s.unitCache.Store(tag, id)
	return id, nil
}

// WriteRows bulk-inserts rows into the telemetry hypertable using pgx CopyFrom.
// CopyFrom bypasses the WAL parsing overhead of multi-row INSERT and is the
// fastest bulk-insert path available through pgx.
func (s *pgxStore) WriteRows(ctx context.Context, rows []Row) error {
	if len(rows) == 0 {
		return nil
	}

	cols := []string{"time", "unit_id", "metric", "value", "quality"}
	src := pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
		r := rows[i]
		return []any{r.Time, r.UnitID, r.Metric, r.Value, r.Quality}, nil
	})

	n, err := s.pool.CopyFrom(ctx, pgx.Identifier{"telemetry"}, cols, src)
	if err != nil {
		return fmt.Errorf("CopyFrom telemetry (%d rows): %w", len(rows), err)
	}
	if int(n) != len(rows) {
		return fmt.Errorf("CopyFrom wrote %d rows, expected %d", n, len(rows))
	}
	return nil
}
