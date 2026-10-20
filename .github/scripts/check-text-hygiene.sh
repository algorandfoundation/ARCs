#!/usr/bin/env bash
set -euo pipefail

status=0
cr=$'\r'

for file in "$@"; do
  if LC_ALL=C grep -n "$cr" "$file" >/dev/null; then
    while IFS=: read -r line_number _; do
      printf '%s:%s: CRLF or CR line ending\n' "$file" "$line_number"
    done < <(LC_ALL=C grep -n "$cr" "$file")
    status=1
  fi

  if [[ "$file" == *.md ]]; then
    trailing_lines=$(
      awk '
        match($0, /[ \t]+$/) {
          trailing = substr($0, RSTART, RLENGTH)
          if (trailing ~ /^  +$/) {
            next
          }
          print NR
        }
      ' "$file"
    )
  else
    trailing_lines=$(
      awk '
        match($0, /[ \t]+$/) {
          print NR
        }
      ' "$file"
    )
  fi

  if [[ -n "$trailing_lines" ]]; then
    while IFS= read -r line_number; do
      printf '%s:%s: trailing whitespace\n' "$file" "$line_number"
    done <<<"$trailing_lines"
    status=1
  fi

  size=$(wc -c < "$file" | tr -d '[:space:]')
  if [[ "$size" -eq 0 ]]; then
    continue
  fi

  if [[ "$(tail -c 1 "$file" | wc -l | tr -d '[:space:]')" -ne 1 ]]; then
    printf '%s: missing final newline\n' "$file"
    status=1
    continue
  fi

  if [[ "$(tail -c 2 "$file" | wc -l | tr -d '[:space:]')" -eq 2 ]]; then
    printf '%s: expected exactly one newline at end of file\n' "$file"
    status=1
  fi
done

exit "$status"
