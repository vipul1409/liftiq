#!/usr/bin/env bash
# demo-stop.sh — Tear down the full LiftIQ Phase 2 stack
#
# Usage:
#   ./scripts/demo-stop.sh                 # stop services, keep data
#   ./scripts/demo-stop.sh --reset         # stop services, truncate data (keep schema)
#   ./scripts/demo-stop.sh --reset-hard    # stop services, drop all tables
#   ./scripts/demo-stop.sh --reset-volume  # stop services, delete Docker volume (full wipe)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="/tmp"

RESET_MODE=""
for arg in "$@"; do
  case "$arg" in
    --reset)        RESET_MODE="truncate" ;;
    --reset-hard)   RESET_MODE="hard" ;;
    --reset-volume) RESET_MODE="volume" ;;
  esac
done

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[liftiq]${NC} $*"; }
success() { echo -e "${GREEN}[liftiq]${NC} $*"; }
warn()    { echo -e "${YELLOW}[liftiq]${NC} $*"; }

stop_pid() {
  local name=$1 pidfile="$PID_DIR/liftiq-${1}.pid"
  if [[ -f "$pidfile" ]]; then
    local pid
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid"
      info "Stopped $name (PID $pid)"
    else
      info "$name (PID $pid) was not running"
    fi
    rm -f "$pidfile"
  else
    info "No PID file for $name — skipping"
  fi
}

stop_pid report
stop_pid compliance
stop_pid ingestor
stop_pid simulator

# ── Database reset (runs while TimescaleDB is still up) ─────────────────────

if [[ -z "$RESET_MODE" ]]; then
  echo ""
  read -r -p "$(echo -e "${YELLOW}Delete database data? [y/N]:${NC} ")" answer
  case "$answer" in
    [yY]|[yY][eE][sS]) RESET_MODE="truncate" ;;
  esac
fi

if [[ "$RESET_MODE" == "truncate" || "$RESET_MODE" == "hard" ]]; then
  if docker info > /dev/null 2>&1 && docker ps --format '{{.Names}}' | grep -q liftiq-timescaledb; then
    if [[ "$RESET_MODE" == "hard" ]]; then
      info "Dropping all tables…"
      docker exec -i liftiq-timescaledb psql -U liftiq -d liftiq -c "DROP TABLE IF EXISTS telemetry CASCADE;"
      docker exec -i liftiq-timescaledb psql -U liftiq -d liftiq -c "DROP TABLE IF EXISTS elevator_units CASCADE;"
      success "Tables dropped. Ingestor will re-migrate on next start."
    else
      info "Truncating data…"
      docker exec -i liftiq-timescaledb psql -U liftiq -d liftiq -c "TRUNCATE TABLE telemetry;"
      docker exec -i liftiq-timescaledb psql -U liftiq -d liftiq -c "TRUNCATE TABLE elevator_units CASCADE;"
      success "All data cleared. Schema intact."
    fi
  else
    warn "TimescaleDB container not running — skipping database reset"
  fi
fi

# ── Stop TimescaleDB ────────────────────────────────────────────────────────

info "Stopping TimescaleDB…"
cd "$REPO_ROOT/telemetry-ingestor"
if docker info > /dev/null 2>&1; then
  if [[ "$RESET_MODE" == "volume" ]]; then
    docker compose down -v
    success "TimescaleDB stopped and volume deleted."
  else
    docker compose down
    success "TimescaleDB stopped."
  fi
else
  warn "Docker is not running — skipping TimescaleDB shutdown (container will stop when Docker starts)"
fi

echo ""
success "All LiftIQ services stopped."
echo ""
