#!/usr/bin/env bash
set -Eeuo pipefail

# Ubuntu/Linux redeploy helper
# - Builds frontend (pnpm install + pnpm build)
# - Then runs backend (go run .)
#
# Why this order?
#   "go run ." will start the backend and usually blocks (keeps running),
#   so if we run it first the frontend build step would never execute.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/src/backend"
FRONTEND_DIR="$ROOT_DIR/src/frontend"

PID_FILE="$ROOT_DIR/.backend-go-run.pid"
LOG_FILE="$ROOT_DIR/.backend-go-run.log"

usage() {
  cat <<'EOF'
Usage:
  ./scripts/redeploy.sh [--daemon]

Options:
  --daemon   Run backend in background (nohup) and write pid/log in repo root.

Default behavior:
  1) cd src/frontend  -> pnpm install  -> pnpm build
  2) cd src/backend   -> go run .      (foreground, blocking)
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[ERROR] Missing command: $1" >&2
    return 1
  fi
}

stop_existing_backend_if_any() {
  if [[ -f "$PID_FILE" ]]; then
    local pid
    pid="$(cat "$PID_FILE" || true)"
    if [[ -n "${pid}" ]] && kill -0 "$pid" >/dev/null 2>&1; then
      echo "[INFO] Stopping existing backend process (pid=$pid) ..."
      kill "$pid" >/dev/null 2>&1 || true
      # Give it a moment to exit gracefully
      for _ in {1..30}; do
        if ! kill -0 "$pid" >/dev/null 2>&1; then
          break
        fi
        sleep 0.2
      done
      if kill -0 "$pid" >/dev/null 2>&1; then
        echo "[WARN] Backend still running, sending SIGKILL (pid=$pid) ..." >&2
        kill -9 "$pid" >/dev/null 2>&1 || true
      fi
    fi
    rm -f "$PID_FILE" || true
  fi
}

DAEMON=0
case "${1:-}" in
  --daemon)
    DAEMON=1
    ;;
  "" )
    ;;
  -h|--help)
    usage
    exit 0
    ;;
  *)
    echo "[ERROR] Unknown option: $1" >&2
    usage >&2
    exit 2
    ;;
esac

require_cmd go
require_cmd pnpm

echo "[INFO] Building frontend..."
cd "$FRONTEND_DIR"
if [[ "${CI:-}" == "true" || "${CI:-}" == "1" ]]; then
  pnpm install --frozen-lockfile
else
  pnpm install
fi
pnpm build

echo "[INFO] Starting backend..."
cd "$BACKEND_DIR"

if [[ "$DAEMON" -eq 1 ]]; then
  stop_existing_backend_if_any

  : > "$LOG_FILE"
  echo "[INFO] Backend logs: $LOG_FILE"
  echo "[INFO] Starting backend in background (nohup)..."

  # shellcheck disable=SC2091
  nohup go run . >>"$LOG_FILE" 2>&1 &
  echo $! > "$PID_FILE"
  echo "[INFO] Backend started. pid=$(cat "$PID_FILE")"
else
  echo "[INFO] Backend running in foreground (Ctrl+C to stop)."
  go run .
fi
