#!/usr/bin/env bash
# Start the dev stack with the real demeter server built at the repository root.
#
# The server runs the repo-root `demeter` binary (override with DEMETER_BINARY)
# and keeps its SQLite database in the demeter-data Docker volume.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

export DEMETER_BINARY="${DEMETER_BINARY:-../demeter}"
export HA_UID="${HA_UID:-$(id -u)}"
export HA_GID="${HA_GID:-$(id -g)}"

if [[ ! -f "${DEMETER_BINARY}" ]]; then
    echo "demeter binary not found at ${DEMETER_BINARY}" >&2
    echo "build it first from the repository root: task build" >&2
    exit 1
fi

if [[ ! -x "${DEMETER_BINARY}" ]]; then
    echo "${DEMETER_BINARY} is not executable" >&2
    exit 1
fi

# Pre-create the bind-mount target so Docker does not create it as root.
mkdir -p config/custom_components/demeter

docker compose --profile live up -d --build

cat <<EOF

Home Assistant: http://localhost:${HA_PORT:-8123}
demeter server: http://localhost:${DEMETER_SERVER_PORT:-8249}
Mock demeter API: http://localhost:${DEMETER_PORT:-18249}

Add the integration in Home Assistant with host:
  http://demeter:8249

The container also has the binary's sensor simulator:
  docker compose exec demeter /usr/local/bin/demeter simulate-sensor --addr demeter:8249

Stop the stack with:
  ./down.sh
EOF
