#!/usr/bin/env bash
# Run the mock demeter API on the host, without the Home Assistant stack.
#
# Useful when testing demeter clients (the Go server, the web UI, scripts)
# against a deterministic topology:
#
#   ./dev/mock.sh                 # listens on http://localhost:8249
#   PORT=9000 ./dev/mock.sh       # listen somewhere else
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

exec python3 mock_demeter.py
