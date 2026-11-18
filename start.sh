#!/usr/bin/env bash
# Fixed launcher entrypoint, following Orchestrator.
exec bash "$(dirname "$(readlink -f "$0")")/scripts/start-videolab.sh" "$@"
