#!/usr/bin/env bash
# demo-status.sh — Health check all LiftIQ services and show live data

set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check() {
  local name=$1 url=$2
  if curl -sf "$url" > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} $name  ($url)"
  else
    echo -e "  ${RED}✗${NC} $name  ($url)"
  fi
}

echo ""
echo -e "${BLUE}── Service health ───────────────────────────────${NC}"
check "Elevator simulator" "http://localhost:8000/health"
check "Compliance engine " "http://localhost:8080/health"

echo ""
echo -e "${BLUE}── Elevator units (compliance engine) ───────────${NC}"
curl -sf http://localhost:8080/units 2>/dev/null \
  | python3 -m json.tool 2>/dev/null \
  || echo -e "  ${YELLOW}(compliance engine not reachable)${NC}"

echo ""
echo -e "${BLUE}── Compliance summary per unit ──────────────────${NC}"
for tag in ELV-001 ELV-002 ELV-003; do
  result=$(curl -sf "http://localhost:8080/units/${tag}/compliance/summary" 2>/dev/null)
  if [[ -n "$result" ]]; then
    overall=$(echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['summary']['overall'])" 2>/dev/null)
    pass=$(echo "$result"    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['summary']['pass'])" 2>/dev/null)
    fail=$(echo "$result"    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['summary']['fail'])" 2>/dev/null)
    unknown=$(echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['summary']['unknown'])" 2>/dev/null)
    if [[ "$overall" == "pass" ]]; then
      color=$GREEN
    elif [[ "$overall" == "fail" ]]; then
      color=$RED
    else
      color=$YELLOW
    fi
    echo -e "  ${color}${tag}${NC}  overall=${color}${overall}${NC}  pass=${pass}  fail=${fail}  unknown=${unknown}"
  else
    echo -e "  ${YELLOW}${tag}  (no data yet — wait for first ingest poll)${NC}"
  fi
done

echo ""
echo -e "${BLUE}── Ingestor log (last 5 lines) ──────────────────${NC}"
if [[ -f /tmp/liftiq-ingestor.log ]]; then
  tail -5 /tmp/liftiq-ingestor.log
else
  echo "  (no ingestor log found)"
fi
echo ""
