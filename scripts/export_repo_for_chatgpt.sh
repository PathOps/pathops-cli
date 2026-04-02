#!/usr/bin/env bash
set -euo pipefail

REPO_NAME="$(basename "$(git rev-parse --show-toplevel)" | tr '.' '-')"
OUTPUT="${REPO_NAME}_dump.txt"

git ls-files | while read -r file; do
  if file --mime "$file" | grep -q 'charset='; then
    echo "===== FILE: $file ====="
    cat "$file"
    echo
  fi
done > "$OUTPUT"

echo "Export completed -> $OUTPUT"