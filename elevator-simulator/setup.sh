#!/usr/bin/env bash
# setup.sh — Create and populate the elevator-simulator virtual environment.
# Run once from the elevator-simulator/ directory:
#   chmod +x setup.sh && ./setup.sh

set -euo pipefail

VENV_DIR=".venv"
PYTHON="${PYTHON:-python3}"
MIN_PYTHON_MINOR=10   # BAC0 requires Python 3.10+

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

green()  { printf '\033[0;32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[0;33m%s\033[0m\n' "$*"; }
red()    { printf '\033[0;31m%s\033[0m\n' "$*"; }
die()    { red "ERROR: $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Python version check
# ---------------------------------------------------------------------------

PY_VERSION=$("$PYTHON" -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
PY_MINOR=$("$PYTHON"   -c "import sys; print(sys.version_info.minor)")

echo "Using Python $PY_VERSION ($("$PYTHON" -c 'import sys; print(sys.executable)'))"

if [ "$PY_MINOR" -lt "$MIN_PYTHON_MINOR" ]; then
    die "Python 3.${MIN_PYTHON_MINOR}+ required (found $PY_VERSION). " \
        "Install a newer Python and re-run, or set PYTHON=/path/to/python3.x"
fi

# ---------------------------------------------------------------------------
# Create virtual environment
# ---------------------------------------------------------------------------

if [ -d "$VENV_DIR" ]; then
    yellow "Virtual environment already exists at $VENV_DIR — skipping creation."
    yellow "To rebuild from scratch: rm -rf $VENV_DIR && ./setup.sh"
else
    echo "Creating virtual environment in $VENV_DIR …"
    "$PYTHON" -m venv "$VENV_DIR"
    green "Virtual environment created."
fi

# ---------------------------------------------------------------------------
# Activate and install dependencies
# ---------------------------------------------------------------------------

# shellcheck source=/dev/null
source "$VENV_DIR/bin/activate"

echo "Upgrading pip …"
pip install --quiet --upgrade pip

echo "Installing dependencies from requirements.txt …"
pip install --quiet -r requirements.txt

green "Dependencies installed."

# ---------------------------------------------------------------------------
# Verify key imports
# ---------------------------------------------------------------------------

echo "Verifying imports …"

python - <<'PYCHECK'
errors = []

try:
    import fastapi
    import uvicorn
    import pydantic
except ImportError as e:
    errors.append(f"  MISSING: {e}")

try:
    import BAC0
except ImportError:
    # BAC0 has system-level deps (libpcap) that may not be present everywhere.
    # Mark as a warning, not a hard error.
    print("  WARNING: BAC0 import failed — BACnet server will be disabled at runtime.")
    print("           On macOS, try: brew install libpcap")

if errors:
    print("Import errors:")
    for err in errors:
        print(err)
    raise SystemExit(1)
else:
    print("  OK: fastapi, uvicorn, pydantic")
PYCHECK

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------

green ""
green "Setup complete. To start the simulator:"
green ""
green "  source $VENV_DIR/bin/activate"
green "  python main.py --no-bacnet"
green ""
green "Or use the Makefile shortcuts (make run, make run-no-bacnet)."
