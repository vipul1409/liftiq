#!/usr/bin/env bash
# db-reset.sh — Clear data from the LiftIQ TimescaleDB database
#
# Usage:
#   ./scripts/db-reset.sh              # truncate data only (keeps schema)
#   ./scripts/db-reset.sh --hard       # drop all tables (full schema reset)
#
# By default connects via Docker exec to the running liftiq-timescaledb container.
# Set DATABASE_URL to connect to a remote host instead:
#   DATABASE_URL="postgres://liftiq:liftiq@myhost:5432/liftiq?sslmode=disable" \
#     ./scripts/db-reset.sh

set -euo pipefail

HARD=false
for arg in "$@"; do
  [[ "$arg" == "--hard" ]] && HARD=true
done

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ── Build psql command ────────────────────────────────────────────────────────

run_sql() {
  local sql=$1
  if [[ -n "${DATABASE_URL:-}" ]]; then
    psql "$DATABASE_URL" -c "$sql"
  else
    docker exec -i liftiq-timescaledb \
      psql -U liftiq -d liftiq -c "$sql"
  fi
}

# ── Confirm before destructive action ────────────────────────────────────────

if [[ "$HARD" == true ]]; then
  echo -e "${RED}WARNING: --hard will drop all tables and recreate the schema.${NC}"
  echo -e "${RED}All data AND schema will be destroyed. The ingestor will re-migrate on next start.${NC}"
else
  echo -e "${YELLOW}This will DELETE all rows from telemetry and elevator_units.${NC}"
  echo -e "${YELLOW}The schema (tables, indexes, hypertable config) will be preserved.${NC}"
fi

echo ""
read -r -p "Type 'yes' to continue: " confirm
if [[ "$confirm" != "yes" ]]; then
  echo "Aborted."
  exit 0
fi

# ── Execute ───────────────────────────────────────────────────────────────────

if [[ "$HARD" == true ]]; then
  echo ""
  echo "Dropping tables..."
  run_sql "DROP TABLE IF EXISTS telemetry CASCADE;"
  run_sql "DROP TABLE IF EXISTS elevator_units CASCADE;"
  echo -e "${GREEN}Tables dropped. Start the ingestor to re-apply migrations.${NC}"
else
  echo ""
  echo "Truncating telemetry..."
  run_sql "TRUNCATE TABLE telemetry;"
  echo "Truncating elevator_units..."
  run_sql "TRUNCATE TABLE elevator_units CASCADE;"
  echo -e "${GREEN}All data cleared. Schema intact.${NC}"
fi
