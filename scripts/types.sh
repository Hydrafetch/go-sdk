#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
SPEC="${HYDRAFETCH_SPEC:-https://api.hydrafetch.com/openapi.json}"
GEN="go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2"
if [ "${1:-}" = check ]; then
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  sed "s#^output: .*#output: $tmp/generated.go#" oapi-codegen.yaml > "$tmp/config.yaml"
  $GEN -config "$tmp/config.yaml" "$SPEC"
  if ! cmp -s "$tmp/generated.go" generated.go; then
    echo "generated.go no longer matches $SPEC. Run scripts/types.sh, review the change, and release." >&2
    exit 1
  fi
  echo "Types match the published API."
  exit 0
fi
$GEN -config oapi-codegen.yaml "$SPEC"
if [[ "$SPEC" == http* ]]; then body="$(curl -fsS "$SPEC")"; else body="$(cat "$SPEC")"; fi
printf '%s' "$body" | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>console.log(require("crypto").createHash("sha256").update(JSON.stringify(JSON.parse(s))).digest("hex")))' > openapi.sha256
echo "Wrote generated.go and openapi.sha256 from $SPEC."
