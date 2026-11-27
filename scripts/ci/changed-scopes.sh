#!/usr/bin/env bash
# Outputs GitHub Actions booleans given git diff --name-only on stdin.
set -euo pipefail
backend=false
desktop=false
extension=false
persistence=false
while IFS= read -r path; do
  [[ -n "$path" ]] || continue
  case "$path" in
    backend/*) backend=true ;;
    apps/desktop/*) desktop=true ;;
    apps/browser-extension/*) extension=true ;;
    docs/*|research/*|.ai/*|*.md|.gitignore) ;;
    *) backend=true; desktop=true; extension=true ;;
  esac
  case "$path" in
    backend/migrations/*|backend/internal/collections/*|backend/internal/notifications/*|backend/internal/meli/order_notifications*|backend/internal/meli/order_coordination*|backend/internal/meli/order_write*|backend/internal/meli/order_shipment*|backend/internal/meli/orders.go|backend/internal/meli/order_billing*) persistence=true ;;
    .github/workflows/*|scripts/ci/*) backend=true; desktop=true; extension=true ;;
  esac
done
printf 'backend=%s\ndesktop=%s\nextension=%s\npersistence=%s\n' \
  "$backend" "$desktop" "$extension" "$persistence"
