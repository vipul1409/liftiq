#!/usr/bin/env bash
# demo-start.sh — Bring up the full LiftIQ Phase 1 stack
#
# Services started (in order, with health checks between each):
#   1. TimescaleDB         (Docker)                    :5432
#   2. Elevator simulator  (Python/FastAPI)             :8000
#   3. Telemetry ingestor  (Go)                         → TimescaleDB
#   4. Compliance engine   (Go/HTTP)                    :8080
#
# Logs written to /tmp/liftiq-*.log
# PIDs written to /tmp/liftiq-*.pid

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="/tmp"
PID_DIR="/tmp"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[liftiq]${NC} $*"; }
success() { echo -e "${GREEN}[liftiq]${NC} $*"; }
warn()    { echo -e "${YELLOW}[liftiq]${NC} $*"; }
error()   { echo -e "${RED}[liftiq]${NC} $*" >&2; }

wait_for_port() {
  local name=$1 port=$2 timeout=${3:-30}
  info "Waiting for $name on port $port..."
  local i=0
  while ! curl -sf "http://localhost:${port}/health" > /dev/null 2>&1; do
    sleep 1
    i=$((i + 1))
    if [[ $i -ge $timeout ]]; then
      error "$name did not become healthy within ${timeout}s"
      error "Check log: $LOG_DIR/liftiq-${name}.log"
      exit 1
    fi
    printf '.'
  done
  echo ""
  success "$name is healthy on :${port}"
}

# ── 0. Pre-flight ───────────────────────────────────────────────────────────

if [[ -f "$PID_DIR/liftiq-simulator.pid" ]] || [[ -f "$PID_DIR/liftiq-compliance.pid" ]]; then
  warn "Demo may already be running. Run 'make demo-down' first, or check /tmp/liftiq-*.pid"
  exit 1
fi

# ── 1. TimescaleDB ──────────────────────────────────────────────────────────

info "Starting TimescaleDB..."
cd "$REPO_ROOT/telemetry-ingestor"
docker compose up -d
info "Waiting for TimescaleDB to be ready..."
until docker compose exec -T timescaledb pg_isready -U liftiq -d liftiq > /dev/null 2>&1; do
  printf '.'; sleep 1
done
echo ""
success "TimescaleDB is ready on :5432"

# ── 2. Elevator Simulator ───────────────────────────────────────────────────

info "Starting elevator simulator..."
cd "$REPO_ROOT/elevator-simulator"

if [[ ! -d ".venv" ]]; then
  info "No venv found — running setup..."
  ./setup.sh
fi

source .venv/bin/activate
nohup python main.py --no-bacnet \
  > "$LOG_DIR/liftiq-simulator.log" 2>&1 &
echo $! > "$PID_DIR/liftiq-simulator.pid"
deactivate

wait_for_port "simulator" 8000

# ── 3. Telemetry Ingestor ───────────────────────────────────────────────────

info "Starting telemetry ingestor..."
cd "$REPO_ROOT/telemetry-ingestor"
make build > /dev/null

nohup env \
  DATABASE_URL="postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable" \
  SIMULATOR_URL="http://localhost:8000" \
  POLL_INTERVAL_SECONDS=5 \
  LOG_LEVEL=info \
  ./bin/ingestd \
  > "$LOG_DIR/liftiq-ingestor.log" 2>&1 &
echo $! > "$PID_DIR/liftiq-ingestor.pid"
success "Telemetry ingestor started (PID $(cat "$PID_DIR/liftiq-ingestor.pid"))"
info "Ingestor will begin writing rows to TimescaleDB within 5 seconds"

# ── 4. Compliance Engine ────────────────────────────────────────────────────

info "Starting compliance engine..."
cd "$REPO_ROOT/compliance-engine"
make build > /dev/null

nohup env \
  DATABASE_URL="postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable" \
  HTTP_PORT=8080 \
  STALE_WINDOW_MINUTES=10 \
  LOG_LEVEL=info \
  ./bin/complianced \
  > "$LOG_DIR/liftiq-compliance.log" 2>&1 &
echo $! > "$PID_DIR/liftiq-compliance.pid"

wait_for_port "compliance" 8080

# ── Done ────────────────────────────────────────────────────────────────────

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        LiftIQ Phase 1 stack is up            ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Elevator simulator   http://localhost:8000/elevators"
echo -e "  Compliance engine    http://localhost:8080/units"
echo -e "  Simulator docs       http://localhost:8000/docs"
echo ""
echo -e "  Logs:"
echo -e "    Simulator   $LOG_DIR/liftiq-simulator.log"
echo -e "    Ingestor    $LOG_DIR/liftiq-ingestor.log"
echo -e "    Compliance  $LOG_DIR/liftiq-compliance.log"
echo ""
echo -e "  ${YELLOW}Wait ~10 seconds for the first telemetry rows to be ingested,${NC}"
echo -e "  ${YELLOW}then open the mobile app or run: make demo-status${NC}"
echo ""
