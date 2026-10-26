#!/usr/bin/env bash
# Stop the Home Assistant test stack.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

docker compose down --remove-orphans --volumes
