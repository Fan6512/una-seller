#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
expect() {
  local files="$1" desired="$2" actual
  actual=$(printf '%s\n' "$files" | bash changed-scopes.sh)
  if [[ "$actual" != "$desired" ]]; then
    printf 'FAIL: %s\nexpected:\n%s\ngot:\n%s\n' "$files" "$desired" "$actual" >&2
    exit 1
  fi
}
expect 'docs/architecture/DELIVERY.md' $'backend=false\ndesktop=false\nextension=false\npersistence=false'
expect 'apps/desktop/src/App.tsx' $'backend=false\ndesktop=true\nextension=false\npersistence=false'
expect 'backend/internal/httpapi/orders.go' $'backend=true\ndesktop=false\nextension=false\npersistence=false'
expect 'backend/migrations/0005_meli_order_purchases.sql' $'backend=true\ndesktop=false\nextension=false\npersistence=true'
expect 'backend/internal/collections/collections.go' $'backend=true\ndesktop=false\nextension=false\npersistence=true'
expect 'apps/browser-extension/content.js' $'backend=false\ndesktop=false\nextension=true\npersistence=false'
expect $'apps/desktop/src/orders.ts\nbackend/internal/meli/orders.go' $'backend=true\ndesktop=true\nextension=false\npersistence=true'
expect '.github/workflows/ci.yml' $'backend=true\ndesktop=true\nextension=true\npersistence=false'
expect 'scripts/ci/changed-scopes.sh' $'backend=true\ndesktop=true\nextension=true\npersistence=false'
expect 'unknown-root-config.toml' $'backend=true\ndesktop=true\nextension=true\npersistence=false'
expect '' $'backend=false\ndesktop=false\nextension=false\npersistence=false'
expect 'backend/internal/notifications/worker.go' $'backend=true\ndesktop=false\nextension=false\npersistence=true'
expect 'backend/internal/meli/order_notifications.go' $'backend=true\ndesktop=false\nextension=false\npersistence=true'
expect 'backend/internal/meli/order_coordination.go' $'backend=true\ndesktop=false\nextension=false\npersistence=true'
printf '14 scope checks PASS\n'
