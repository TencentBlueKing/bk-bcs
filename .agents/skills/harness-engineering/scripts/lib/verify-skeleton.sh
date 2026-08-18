#!/usr/bin/env bash
# F4: TODO skeletons must not appear in「当前项目选用」table.
# shellcheck shell=bash

verify_skeleton() {
  local target="$1"
  local readme="$target/docs/standards/README.md"
  local in_selected=0
  local line fname file_cell

  [[ -f "$readme" ]] || return 0

  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" == '## 当前项目选用'* ]]; then
      in_selected=1
      continue
    fi
    if [[ $in_selected -eq 1 && "$line" == '## '* ]]; then
      in_selected=0
      continue
    fi
    [[ $in_selected -eq 1 ]] || continue

    fname=""
    if [[ "$line" == *']('* ]]; then
      fname="${line#*](}"
      fname="${fname%%)*}"
    elif [[ "$line" == *\`*.md\`* ]]; then
      fname="${line#*\`}"
      fname="${fname%%\`*}"
    fi
    [[ -n "$fname" && "$fname" == *.md ]] || continue

    file_cell="$target/docs/standards/$fname"
    if [[ -f "$file_cell" ]] && grep -qE '<!--[[:space:]]*TODO' "$file_cell"; then
      verify_error "standards README 当前选用 lists $fname which contains <!-- TODO"
    fi
  done <"$readme"
}
