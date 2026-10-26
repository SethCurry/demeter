#!/usr/bin/env bash
# Start Home Assistant and the mock demeter API for manual testing.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

export HA_UID="${HA_UID:-$(id -u)}"
export HA_GID="${HA_GID:-$(id -g)}"

# Pre-create the bind-mount target so Docker does not create it as root.
mkdir -p config/custom_components/demeter

docker compose up -d --build

cat <<EOF

Home Assistant: http://localhost:${HA_PORT:-8123}
Mock demeter API: http://localhost:${DEMETER_PORT:-18249}

Add the integration in Home Assistant with host:
  http://demeter-mock:8249

Follow the integration logs with:
  docker compose logs -f homeassistant

Stop the stack with:
  ./down.sh
EOF
