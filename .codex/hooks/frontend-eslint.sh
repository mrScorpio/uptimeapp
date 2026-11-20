#!/usr/bin/env bash

set -euo pipefail

mapfile -d '' candidate_files < <(
  {
    git diff --name-only -z HEAD -- frontend
    git ls-files --others --exclude-standard -z -- frontend
  } | sort -zu
)

frontend_files=()
for file in "${candidate_files[@]}"; do
  [[ -f "$file" ]] || continue
  [[ "$file" == "frontend/next-env.d.ts" ]] && continue
  case "$file" in
    frontend/*.js|frontend/*.jsx|frontend/*.mjs|frontend/*.cjs|frontend/*.ts|frontend/*.tsx)
      frontend_files+=("${file#frontend/}")
      ;;
  esac
done

if ((${#frontend_files[@]} == 0)); then
  exit 0
fi

npm --prefix frontend exec -- eslint "${frontend_files[@]}"
