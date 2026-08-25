#!/usr/bin/env bash
# F1: false "已就绪" / dangling "权威" path claims.
# shellcheck shell=bash
# Expects: TARGET, verify_error, verify_list_md from caller.

verify_facts() {
  local target="$1"
  local f line lineno

  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    lineno=0
    while IFS= read -r line || [[ -n "$line" ]]; do
      lineno=$((lineno + 1))
      if [[ "$line" != *已就绪* ]]; then
        continue
      fi
      if [[ "$line" == *project.json* ]] && [[ ! -f "$target/project.json" ]]; then
        verify_error "$f:$lineno claims project.json 已就绪 but file missing"
      fi
      if [[ "$line" == *graph.json* ]] && [[ ! -f "$target/docs/dev-map/graph.json" ]]; then
        verify_error "$f:$lineno claims graph.json 已就绪 but file missing"
      fi
      # 声称 graph.json 已入库/已提交，但已被 ignore → ERROR（结果默认不入库）
      if [[ "$line" == *graph.json* ]] && [[ "$line" =~ 已入库|已提交|纳入[[:space:]]*git|git[[:space:]]*add ]]; then
        if [[ -f "$target/docs/dev-map/.gitignore" ]]; then
          local gi="$target/docs/dev-map/.gitignore"
          # 枚举式 graph.json 行，或白名单式 * + !README.md
          if grep -qE '^[[:space:]]*graph\.json[[:space:]]*$' "$gi" \
            || { grep -qE '^[[:space:]]*\*[[:space:]]*$' "$gi" \
              && grep -qE '^[[:space:]]*!README\.md[[:space:]]*$' "$gi"; }; then
            verify_error "$f:$lineno claims graph.json committed/tracked but docs/dev-map/.gitignore ignores it"
          fi
        fi
      fi
      # AGENTS.md as readiness claim (table cell / backtick), not every mention
      if [[ "$line" =~ \`AGENTS\.md\`|根目录.*AGENTS\.md|入口.*AGENTS\.md ]] \
        && [[ ! -f "$target/AGENTS.md" ]]; then
        verify_error "$f:$lineno claims AGENTS.md 已就绪/入口 but file missing"
      fi
    done <"$f"
  done < <(verify_list_md "$target")

  # 「权威来源」指向仓内不存在的 skill/tool-dependencies 路径
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    lineno=0
    while IFS= read -r line || [[ -n "$line" ]]; do
      lineno=$((lineno + 1))
      if [[ "$line" != *权威* ]]; then
        continue
      fi
      if [[ "$line" == *tool-dependencies.md* ]]; then
        if [[ ! -f "$target/tool-dependencies.md" ]] \
          && [[ ! -f "$target/.codebuddy/skills/harness-engineering/references/tool-dependencies.md" ]] \
          && [[ ! -f "$target/.cursor/skills/harness-engineering/references/tool-dependencies.md" ]] \
          && [[ ! -f "$target/.agents/skills/harness-engineering/references/tool-dependencies.md" ]]; then
          # Only error when also claiming local authority / 已就绪 context in nearby sense:
          # if the line says 权威来源 or 权威清单 as in-repo path
          if [[ "$line" == *权威来源* || "$line" == *权威清单* || "$line" == *为权威* ]]; then
            verify_error "$f:$lineno cites tool-dependencies.md as authority but path not in repo"
          fi
        fi
      fi
    done <"$f"
  done < <(verify_list_md "$target")
}
