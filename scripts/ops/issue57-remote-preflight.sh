#!/usr/bin/env bash
# Issue #57 remote Go API READ-ONLY preflight. Never print credentials.
set -euo pipefail

BASE="https://mkd.186512.xyz"
printf '[info] Una Seller issue #57 preflight (READ ONLY)\n'
printf '[info] host=%s arch=%s\n' "$(hostname -s)" "$(uname -m)"
for cmd in curl systemctl ss nginx sha256sum; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    printf '[warn] missing command: %s\n' "$cmd"
  fi
done

if command -v systemctl >/dev/null 2>&1; then
  printf '[info] una-seller-api service: '
  systemctl is-active una-seller-api 2>/dev/null || true
  # Neither environment contents nor journal logs are inspected.
  systemctl show una-seller-api --property=FragmentPath --property=MainPID --no-pager 2>/dev/null || true
fi

if command -v ss >/dev/null 2>&1; then
  printf '[info] loopback API listener (expected 127.0.0.1:8787):\n'
  ss -lnt 2>/dev/null | grep -E '(^State|127\.0\.0\.1:8787|0\.0\.0\.0:8787|\[::\]:8787)' || true
fi

if command -v nginx >/dev/null 2>&1; then
  printf '[info] nginx configuration check (no reload):\n'
  if nginx -t 2>&1; then true; else
    printf '[warn] nginx -t requires sudo or Nginx configuration is invalid; inspect privately\n'
  fi
fi

if [ -f /opt/una-seller/una-seller-api ] && command -v sha256sum >/dev/null 2>&1; then
  printf '[info] installed Go API sha256 (identification only, not a version claim):\n'
  sha256sum /opt/una-seller/una-seller-api
else
  printf '[warn] Go API binary not found at template path; locate service-owned binary before upgrade\n'
fi

probe() {
  local url="$1" status
  status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' --max-time 8 "$url" 2>/dev/null || true)"
  [ -n "$status" ] || status=000
  printf '[probe] %s => HTTP %s\n' "$url" "$status"
}
probe 'http://127.0.0.1:8787/healthz'
probe "$BASE/healthz"
probe "$BASE/api/v1/collections"
probe "$BASE/api/v1/products"
probe "$BASE/api/v1/categories/predict?q=Winter%20Plant%20Cleaner"
printf '[info] Expected AFTER upgrade: health=200, unauthenticated API routes=401.\n'
printf '[info] Before upgrade /products and /categories/predict may be 404; this is not a DB check.\n'
printf '[info] No DB queries, migrations, Nginx reloads or service restarts were executed.\n'
