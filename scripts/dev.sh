#!/usr/bin/env bash
set -Eeuo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo 'Set DATABASE_URL before starting the backend.' >&2
  exit 1
fi
jwt_secret="${JWT_SECRET:-}"
if [[ ${#jwt_secret} -lt 32 ]]; then
  echo 'Set JWT_SECRET to at least 32 characters before starting the backend.' >&2
  exit 1
fi
if [[ ! -d "$project_dir/frontend/node_modules" ]]; then
  echo 'Install frontend dependencies first: cd frontend && npm install' >&2
  exit 1
fi

build_dir="$(mktemp -d)"
backend_pid=''
frontend_pid=''

cleanup() {
  trap - EXIT INT TERM
  for pid in "$backend_pid" "$frontend_pid"; do
    if [[ -n "$pid" ]]; then
      kill -- "-$pid" 2>/dev/null || kill -- "$pid" 2>/dev/null || true
    fi
  done
  for pid in "$backend_pid" "$frontend_pid"; do
    if [[ -n "$pid" ]]; then
      wait "$pid" 2>/dev/null || true
    fi
  done
  rm -rf -- "$build_dir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo 'Building backend...'
(cd "$project_dir/backend" && go build -o "$build_dir/uptimeapp-api" ./cmd/api)

echo 'Starting backend on :8080 and frontend on :3000...'
setsid "$build_dir/uptimeapp-api" &
backend_pid=$!
(cd "$project_dir/frontend" && exec setsid npm run dev) &
frontend_pid=$!

set +e
wait -n "$backend_pid" "$frontend_pid"
status=$?
set -e
echo 'A server stopped; shutting down both servers.' >&2
exit "$status"
