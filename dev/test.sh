#!/usr/bin/env bash
# Run the end-to-end Home Assistant smoke test.
#
# Set KEEP=1 to leave the containers running after the test (useful for
# inspecting the UI or logs on failure).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

export HA_URL="${HA_URL:-http://localhost:${HA_PORT:-8123}}"
export HA_UID="${HA_UID:-$(id -u)}"
export HA_GID="${HA_GID:-$(id -g)}"

# Pre-create the bind-mount target so Docker does not create it as root.
mkdir -p config/custom_components/demeter

cleanup() {
    local status=$?

    if [[ ${status} -ne 0 ]]; then
        echo
        echo "--- container logs (last 200 lines) ---" >&2
        docker compose logs --no-color --tail=200 >&2 || true
    fi

    if [[ "${KEEP:-0}" != "1" ]]; then
        docker compose down --remove-orphans --volumes || true
    else
        echo
        echo "KEEP=1 set; leaving containers running."
    fi

    exit "${status}"
}
trap cleanup EXIT

docker compose up -d --build
python3 smoke_test.py
