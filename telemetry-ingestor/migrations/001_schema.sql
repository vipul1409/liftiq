-- LiftIQ Telemetry Ingestion Service — TimescaleDB Schema
-- Version 1 | Phase 1 Week 2
--
-- All statements are idempotent; safe to re-run on every service start.
-- Run order matters: elevator_units must exist before telemetry (FK dependency),
-- and create_hypertable must be called before add_dimension.

-- ============================================================
-- 1. TimescaleDB extension
-- ============================================================
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ============================================================
-- 2. Elevator unit registry
--    unit_tag is the simulator's string ID (e.g. "ELV-001") and
--    doubles as the NFC tag identifier in production deployments.
--    UNIQUE constraint is required for ON CONFLICT upserts in the
--    ingestion service's LookupUnit() call.
-- ============================================================
CREATE TABLE IF NOT EXISTS elevator_units (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_tag         VARCHAR(50)  NOT NULL UNIQUE,
    building_id      UUID,                          -- nullable for simulator / prototype
    controller_make  VARCHAR(100),
    controller_model VARCHAR(100),
    protocol         VARCHAR(20)  NOT NULL DEFAULT 'simulator',
    bms_address      VARCHAR(100),
    installed_date   DATE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 3. Telemetry hypertable (narrow/long format)
--    One row per metric per elevator per poll cycle.
--    At 20 metrics × 3 elevators × 5-second polls:
--      ~720 rows/minute, ~1,037,000 rows/day.
--
--    quality values: "good" | "missing" (staleness is handled by the
--    compliance engine's stale window, not stored as a quality)
-- ============================================================
CREATE TABLE IF NOT EXISTS telemetry (
    time    TIMESTAMPTZ      NOT NULL,
    unit_id UUID             NOT NULL REFERENCES elevator_units(id),
    metric  VARCHAR(50)      NOT NULL,
    value   DOUBLE PRECISION NOT NULL,
    quality VARCHAR(10)      NOT NULL DEFAULT 'good'
);

-- ============================================================
-- 4. Hypertable: partition by time (7-day chunks by default)
-- ============================================================
SELECT create_hypertable('telemetry', 'time', if_not_exists => TRUE);

-- ============================================================
-- 5. Space partition: hash by unit_id into 4 buckets
--    Co-locates all data for one elevator within a time chunk,
--    making per-unit range queries (the primary compliance-engine
--    access pattern) significantly more efficient.
--    Guarded by a DO block so it is idempotent.
-- ============================================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM   timescaledb_information.dimensions
        WHERE  hypertable_name = 'telemetry'
        AND    column_name     = 'unit_id'
    ) THEN
        PERFORM add_dimension('telemetry', 'unit_id', number_of_partitions => 4);
    END IF;
END $$;

-- ============================================================
-- 6. Composite index for the primary query pattern:
--    "last N readings of <metric> for <unit_id>"
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_telemetry_unit_metric
    ON telemetry (unit_id, metric, time DESC);
