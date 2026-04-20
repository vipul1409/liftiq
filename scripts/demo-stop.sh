#!/usr/bin/env bash
# demo-stop.sh — Tear down the full LiftIQ Phase 1 stack

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="/tmp"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[liftiq]${NC} $*"; }
success() { echo -e "${GREEN}[liftiq]${NC} $*"; }

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

stop_pid compliance
stop_pid ingestor
stop_pid simulator

info "Stopping TimescaleDB…"
cd "$REPO_ROOT/telemetry-ingestor"
docker compose down
success "TimescaleDB stopped"

echo ""
success "All LiftIQ services stopped."
echo ""
