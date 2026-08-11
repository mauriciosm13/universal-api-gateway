#!/usr/bin/env bash
set -euo pipefail
go list -m all >/dev/null
echo "Bootstrap dependency inventory complete (Phase 4 will add vulnerability and license checks)."
