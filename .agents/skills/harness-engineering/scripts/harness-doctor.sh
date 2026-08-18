#!/usr/bin/env bash
# harness-doctor.sh — runtime probe for harness tooling (stdout only, never writes repo files).
# Usage: harness-doctor.sh [--json] [--help] [workspace_root]
# Exit 0 when probing completes; missing tools are WARN rows, not script failure.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
JSON=0
ROOT=""

usage() {
  cat <<'EOF'
Usage: harness-doctor.sh [--json] [workspace_root]

Probe CLI / Skill install root / project.json / project-owned tools.
Prints a human-readable table (or one JSON object with --json).
Never writes files under the workspace. Exit 0 when probing finishes.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *)
      if [[ -n "$ROOT" ]]; then
        echo "ERROR: unexpected argument: $1" >&2
        exit 2
      fi
      ROOT="$1"
      shift
      ;;
  esac
done

ROOT="${ROOT:-.}"
ROOT="$(cd "$ROOT" && pwd)"

# Rows: kind|name|status|detail  (status: present|absent|skip|info)
ROWS=()

add_row() {
  ROWS+=("$1|$2|$3|$4")
}

resolve_skill_install_root() {
  # Prefer .agents/skills, then skills/ (dev), then IDE dirs / symlinks.
  if [[ -f "$ROOT/.agents/skills/harness-engineering/SKILL.md" ]] || \
     compgen -G "$ROOT/.agents/skills/*/SKILL.md" >/dev/null 2>&1; then
    echo ".agents/skills"
    return
  fi
  if [[ -f "$ROOT/skills/harness-engineering/SKILL.md" ]] || \
     compgen -G "$ROOT/skills/*/SKILL.md" >/dev/null 2>&1; then
    echo "skills"
    return
  fi
  local ide
  for ide in .cursor .codebuddy .claude .codex; do
    local skills_dir="$ROOT/$ide/skills"
    if [[ -L "$ROOT/$ide" ]]; then
      local target
      target="$(readlink "$ROOT/$ide" || true)"
      if [[ "$target" == ".agents" || "$target" == */.agents ]]; then
        if [[ -d "$ROOT/.agents/skills" ]]; then
          echo ".agents/skills"
          return
        fi
      fi
    fi
    if [[ -f "$skills_dir/harness-engineering/SKILL.md" ]] || \
       compgen -G "$skills_dir/*/SKILL.md" >/dev/null 2>&1; then
      echo "${ide}/skills"
      return
    fi
  done
  echo ""
}

probe_cli() {
  local name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    add_row "cli" "$name" "present" "$(command -v "$name")"
  else
    add_row "cli" "$name" "absent" "command -v $name failed"
  fi
}

probe_path() {
  local kind="$1" name="$2" rel="$3"
  if [[ -e "$ROOT/$rel" ]]; then
    add_row "$kind" "$name" "present" "$rel"
  else
    add_row "$kind" "$name" "absent" "missing: $rel"
  fi
}

# --- baseline CLI ---
probe_cli git
probe_cli bash
probe_cli jq
if [[ -f "$ROOT/go.mod" ]]; then
  probe_cli go
fi
if [[ -f "$ROOT/package.json" ]]; then
  probe_cli node
fi

# --- Skill install root + harness-engineering ---
SKILL_INSTALL_ROOT="$(resolve_skill_install_root)"
if [[ -z "$SKILL_INSTALL_ROOT" ]]; then
  add_row "skill_root" "SKILL_INSTALL_ROOT" "absent" "no skills layout under .agents/skills|skills|IDE"
else
  add_row "skill_root" "SKILL_INSTALL_ROOT" "present" "$SKILL_INSTALL_ROOT"
  probe_path "skill" "harness-engineering" "$SKILL_INSTALL_ROOT/harness-engineering/SKILL.md"
fi

# --- project.json (report only) ---
if [[ -f "$ROOT/project.json" ]]; then
  add_row "config" "project.json" "present" "project.json"
else
  add_row "config" "project.json" "absent" "optional unless tooling marks required"
fi

# --- baseline MCP: advisory only ---
add_row "mcp" "baseline" "info" "probe MCP in Agent session per tool-dependencies.md §一 (set HARNESS_DOCTOR_MCP=1 later for scripted probes)"

# --- project-owned tools (required when section or git-tracked layout exists) ---
TOOLING="$ROOT/docs/harness/tooling.md"
PROJECT_OWNED_COUNT=0

# 仓根 + monorepo 子树（…/.(agents|cursor|…)/skills/<name>/SKILL.md）
list_project_owned_skill_files() {
  if [[ -d "$ROOT/.git" ]] || git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$ROOT" ls-files 2>/dev/null | grep -E '(^|/)\.(agents|cursor|codebuddy|claude|codex)/skills/[^/]+/SKILL\.md$' || true
    git -C "$ROOT" ls-files -- 'skills/*/SKILL.md' 2>/dev/null || true
    return 0
  fi
  # 非 git：仅扫常见布局，排除 node_modules
  find "$ROOT" \( -path '*/node_modules/*' -o -path '*/.git/*' \) -prune -o \
    \( -path '*/.agents/skills/*/SKILL.md' -o -path '*/.cursor/skills/*/SKILL.md' \
       -o -path '*/.codebuddy/skills/*/SKILL.md' -o -path '*/.claude/skills/*/SKILL.md' \
       -o -path '*/.codex/skills/*/SKILL.md' -o -path '*/skills/*/SKILL.md' \) -type f -print 2>/dev/null \
    | sed "s|^$ROOT/||" || true
}

resolve_project_owned_skill_path() {
  local name="$1"
  local cand f
  for cand in \
    ".agents/skills/$name/SKILL.md" \
    ".cursor/skills/$name/SKILL.md" \
    ".codebuddy/skills/$name/SKILL.md" \
    ".claude/skills/$name/SKILL.md" \
    ".codex/skills/$name/SKILL.md" \
    "skills/$name/SKILL.md"; do
    if [[ -f "$ROOT/$cand" ]]; then
      echo "$cand"
      return 0
    fi
  done
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    if [[ "$f" =~ (^|/)\.(agents|cursor|codebuddy|claude|codex)/skills/${name}/SKILL\.md$ ]] \
      || [[ "$f" == "skills/$name/SKILL.md" ]]; then
      if [[ -f "$ROOT/$f" ]]; then
        echo "$f"
        return 0
      fi
    fi
  done < <(list_project_owned_skill_files)
  # 磁盘回退（fixture / 未 init git）：find 子树
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    echo "$f"
    return 0
  done < <(find "$ROOT" \( -path '*/node_modules/*' -o -path '*/.git/*' \) -prune -o \
    \( -path "*/.agents/skills/$name/SKILL.md" -o -path "*/.cursor/skills/$name/SKILL.md" \
       -o -path "*/.codebuddy/skills/$name/SKILL.md" -o -path "*/.claude/skills/$name/SKILL.md" \
       -o -path "*/.codex/skills/$name/SKILL.md" \) -type f -print 2>/dev/null | sed "s|^$ROOT/||" | head -1)
  return 1
}

probe_project_owned_name() {
  local name="$1"
  local path typ
  PROJECT_OWNED_COUNT=$((PROJECT_OWNED_COUNT + 1))
  # MCP server name: look in tracked/local mcp.json
  for path in .agents/mcp.json .cursor/mcp.json; do
    if [[ -f "$ROOT/$path" ]] && grep -qE "\"$name\"" "$ROOT/$path" 2>/dev/null; then
      add_row "project_owned" "$name" "present" "$path"
      return
    fi
  done
  if path="$(resolve_project_owned_skill_path "$name")"; then
    add_row "project_owned" "$name" "present" "$path"
  else
    add_row "project_owned" "$name" "absent" "no install-layout SKILL.md or mcp.json key for $name"
  fi
}

probe_project_owned_from_tooling() {
  local in_section=0
  local line name typ path col3 col4
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ^#+[[:space:]]*([0-9]+(\.[0-9]+)*[[:space:]]+)?项目自有工具 ]]; then
      in_section=1
      continue
    fi
    if [[ "$in_section" -eq 1 && "$line" =~ ^#+[[:space:]] ]]; then
      break
    fi
    if [[ "$in_section" -eq 1 && "$line" =~ ^\| ]]; then
      # skip header / separator (do not match「用途」in data cells)
      if [[ "$line" =~ ^\|[[:space:]]*名称[[:space:]]*\| ]] || [[ "$line" == *"----"* ]]; then
        continue
      fi
      name="$(echo "$line" | awk -F'|' '{gsub(/^ +| +$/,"",$2); print $2}')"
      col3="$(echo "$line" | awk -F'|' '{gsub(/^ +| +$/,"",$3); print $3}')"
      col4="$(echo "$line" | awk -F'|' '{gsub(/^ +| +$/,"",$4); print $4}')"
      col4="${col4//\`/}"
      [[ -z "$name" || "$name" == "（无）" ]] && continue
      # Legacy: | name | type | path | ...
      if [[ -n "$col4" && ( "$col4" == *SKILL.md || "$col4" == *mcp.json* ) ]]; then
        typ="$col3"
        path="$col4"
        PROJECT_OWNED_COUNT=$((PROJECT_OWNED_COUNT + 1))
        if [[ "$typ" =~ [Mm][Cc][Pp] ]]; then
          if [[ -f "$ROOT/$path" ]] && grep -qE "\"$name\"" "$ROOT/$path" 2>/dev/null; then
            add_row "project_owned" "$name" "present" "$path"
          elif [[ -f "$ROOT/$path" ]]; then
            add_row "project_owned" "$name" "absent" "mcp.json present but key missing: $name"
          else
            add_row "project_owned" "$name" "absent" "missing: $path"
          fi
        else
          if [[ -f "$ROOT/$path" ]]; then
            add_row "project_owned" "$name" "present" "$path"
          else
            add_row "project_owned" "$name" "absent" "missing: $path"
          fi
        fi
        continue
      fi
      # Current: | name | purpose |
      probe_project_owned_name "$name"
    fi
  done < "$TOOLING"
}

probe_project_owned_from_git() {
  [[ -d "$ROOT/.git" ]] || git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1 || return 0
  local f name
  local -A seen_names=()
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    name="$(basename "$(dirname "$f")")"
    [[ -n "${seen_names[$name]+x}" ]] && continue
    seen_names[$name]=1
    PROJECT_OWNED_COUNT=$((PROJECT_OWNED_COUNT + 1))
    probe_path "project_owned" "$name" "$f"
  done < <(list_project_owned_skill_files | sort -u)

  local mcp
  for mcp in .agents/mcp.json .cursor/mcp.json; do
    if git -C "$ROOT" ls-files --error-unmatch -- "$mcp" >/dev/null 2>&1; then
      PROJECT_OWNED_COUNT=$((PROJECT_OWNED_COUNT + 1))
      if [[ -f "$ROOT/$mcp" ]]; then
        add_row "project_owned" "mcp:$mcp" "present" "$mcp"
      else
        add_row "project_owned" "mcp:$mcp" "absent" "tracked but missing: $mcp"
      fi
    fi
  done
}

if [[ -f "$TOOLING" ]] && grep -qE '^#+[[:space:]]*([0-9]+(\.[0-9]+)*[[:space:]]+)?项目自有工具' "$TOOLING"; then
  probe_project_owned_from_tooling
elif [[ -d "$ROOT/.git" ]] || git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  probe_project_owned_from_git
fi

if [[ "$PROJECT_OWNED_COUNT" -eq 0 ]]; then
  add_row "project_owned" "(none)" "skip" "no project-owned section and no git-tracked install-layout skills/mcp"
fi

# --- Skill install root layout WARN (F14 / R3) ---
if [[ -n "$SKILL_INSTALL_ROOT" && "$SKILL_INSTALL_ROOT" != ".agents/skills" && "$SKILL_INSTALL_ROOT" != "skills" ]]; then
  add_row "skill_root" "layout" "info" "WARN: recommend .agents/skills + IDE symlink; see skill-install-root.md (current=$SKILL_INSTALL_ROOT)"
fi

# --- graph.json local artifact (R6) ---
if [[ -f "$ROOT/docs/dev-map/graph.json" ]]; then
  add_row "dev_map" "graph.json" "present" "docs/dev-map/graph.json (local; default not committed)"
elif [[ -d "$ROOT/docs/dev-map" ]]; then
  add_row "dev_map" "graph.json" "absent" "run: GRAPHIFY_OUT=docs/dev-map GRAPHIFY_NO_BACKUP=1 graphify update ."
else
  add_row "dev_map" "graph.json" "skip" "docs/dev-map/ not present"
fi

# --- output ---
if [[ "$JSON" -eq 1 ]]; then
  printf '{'
  printf '"workspace":"%s",' "$ROOT"
  printf '"skill_install_root":"%s",' "$SKILL_INSTALL_ROOT"
  printf '"items":['
  first=1
  for row in "${ROWS[@]}"; do
    IFS='|' read -r kind name status detail <<<"$row"
    detail="${detail//\\/\\\\}"
    detail="${detail//\"/\\\"}"
    name="${name//\"/\\\"}"
    if [[ "$first" -eq 1 ]]; then first=0; else printf ','; fi
    printf '{"kind":"%s","name":"%s","status":"%s","detail":"%s"}' "$kind" "$name" "$status" "$detail"
  done
  printf ']}\n'
else
  echo "harness-doctor workspace=$ROOT"
  echo "SKILL_INSTALL_ROOT=${SKILL_INSTALL_ROOT:-"(unresolved)"}"
  printf '%-16s %-28s %-8s %s\n' "KIND" "NAME" "STATUS" "DETAIL"
  printf '%-16s %-28s %-8s %s\n' "----" "----" "------" "------"
  for row in "${ROWS[@]}"; do
    IFS='|' read -r kind name status detail <<<"$row"
    printf '%-16s %-28s %-8s %s\n' "$kind" "$name" "$status" "$detail"
    if [[ "$status" == "absent" && "$kind" != "config" ]]; then
      echo "WARN: $kind/$name → $detail" >&2
    fi
  done
fi

exit 0
