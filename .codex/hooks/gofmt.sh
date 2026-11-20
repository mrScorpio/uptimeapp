#!/usr/bin/env bash

set -euo pipefail

mapfile -d '' candidate_files < <(
  {
    git diff --name-only -z HEAD -- '*.go'
    git ls-files --others --exclude-standard -z -- '*.go'
  } | sort -zu
)

changed_files=()
for file in "${candidate_files[@]}"; do
  [[ -f "$file" ]] && changed_files+=("$file")
done

if ((${#changed_files[@]} == 0)); then
  exit 0
fi

gofmt -w -- "${changed_files[@]}"
