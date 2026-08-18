#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERIFY="$ROOT/scripts/harness-verify.sh"
FIX="$ROOT/tests/fixtures"

chmod +x "$VERIFY"

expect_fail() {
  local name="$1"
  local path="$2"
  if "$VERIFY" "$path" >/tmp/harness-verify-out.txt 2>/tmp/harness-verify-err.txt; then
    echo "EXPECTED_FAIL: $name should fail" >&2
    cat /tmp/harness-verify-err.txt >&2 || true
    exit 1
  fi
  echo "OK fail: $name"
}

expect_ok() {
  local name="$1"
  local path="$2"
  if ! "$VERIFY" "$path" >/tmp/harness-verify-out.txt 2>/tmp/harness-verify-err.txt; then
    echo "EXPECTED_OK: $name should pass" >&2
    cat /tmp/harness-verify-err.txt >&2 || true
    exit 1
  fi
  echo "OK pass: $name"
}

expect_fail bad-ready-claim "$FIX/bad-ready-claim"
expect_fail bad-todo-selected "$FIX/bad-todo-selected"
expect_fail bad-vitest-conflict "$FIX/bad-vitest-conflict"
expect_fail bad-workflow-force "$FIX/bad-workflow-force"
expect_fail bad-tooling-ready-column "$FIX/bad-tooling-ready-column"
expect_fail bad-abs-path "$FIX/bad-abs-path"
expect_fail bad-standards-no-gate "$FIX/bad-standards-no-gate"
expect_fail bad-standards-orphan-frontend "$FIX/bad-standards-orphan-frontend"
expect_ok good-minimal "$FIX/good-minimal"
expect_ok good-desensitized-min "$FIX/good-desensitized-min"
expect_ok good-project-owned-gittracked "$FIX/good-project-owned-gittracked"
expect_ok good-standards-gate "$FIX/good-standards-gate"

SYNC="$ROOT/scripts/sync-standards-rules.sh"
chmod +x "$SYNC"
SYNC_OUT="$(mktemp)"
"$SYNC" --dry-run "$FIX/good-standards-gate" >"$SYNC_OUT"
if ! grep -q 'standards-frontend' "$SYNC_OUT"; then
  echo "EXPECTED: sync-standards-rules dry-run mentions standards-frontend" >&2
  cat "$SYNC_OUT" >&2
  exit 1
fi
rm -f "$SYNC_OUT"
echo "OK sync-standards-rules: dry-run"

# Claude 实体目录：写入 .claude/rules（paths frontmatter），不写 .codex/rules
CLAUDE_FIX="$(mktemp -d)"
mkdir -p "$CLAUDE_FIX/docs/standards" "$CLAUDE_FIX/.claude"
cp "$FIX/good-standards-gate/docs/standards/README.md" "$CLAUDE_FIX/docs/standards/"
cp "$FIX/good-standards-gate/AGENTS.md" "$CLAUDE_FIX/" 2>/dev/null || true
# minimal 选用表（若 fixture README 已有前端选用则直接用）
if ! grep -q 'frontend-' "$CLAUDE_FIX/docs/standards/README.md"; then
  echo "EXPECTED: good-standards-gate README has frontend selection" >&2
  exit 1
fi
"$SYNC" "$CLAUDE_FIX" >/tmp/sync-claude-out.txt
if [[ ! -f "$CLAUDE_FIX/.claude/rules/standards-frontend.md" ]]; then
  echo "EXPECTED: .claude/rules/standards-frontend.md written" >&2
  cat /tmp/sync-claude-out.txt >&2
  exit 1
fi
if ! grep -q '^paths:' "$CLAUDE_FIX/.claude/rules/standards-frontend.md"; then
  echo "EXPECTED: Claude rule has paths frontmatter" >&2
  exit 1
fi
if [[ -d "$CLAUDE_FIX/.codex/rules" ]]; then
  echo "UNEXPECTED: must not create .codex/rules for standards sync" >&2
  exit 1
fi
rm -rf "$CLAUDE_FIX"
echo "OK sync-standards-rules: claude entity layout"

DOCTOR="$ROOT/scripts/harness-doctor.sh"
chmod +x "$DOCTOR" "$VERIFY"
if ! "$DOCTOR" "$FIX/good-minimal" >/tmp/harness-doctor-out.txt 2>/tmp/harness-doctor-err.txt; then
  echo "EXPECTED_OK: harness-doctor should exit 0" >&2
  cat /tmp/harness-doctor-err.txt >&2 || true
  exit 1
fi
echo "OK doctor: good-minimal"

if ! "$DOCTOR" "$FIX/good-project-owned-gittracked" >/tmp/harness-doctor-out.txt 2>/tmp/harness-doctor-err.txt; then
  echo "EXPECTED_OK: harness-doctor project-owned should exit 0" >&2
  exit 1
fi
if ! grep -qE 'project_owned[[:space:]]+my-tool[[:space:]]+present' /tmp/harness-doctor-out.txt; then
  echo "EXPECTED: doctor reports project_owned my-tool present" >&2
  cat /tmp/harness-doctor-out.txt >&2
  exit 1
fi
echo "OK doctor: project-owned my-tool"

if ! "$DOCTOR" "$FIX/good-project-owned-nested" >/tmp/harness-doctor-nested.txt 2>/tmp/harness-doctor-nested.err; then
  echo "EXPECTED_OK: harness-doctor nested project-owned should exit 0" >&2
  cat /tmp/harness-doctor-nested.err >&2 || true
  exit 1
fi
if ! grep -qE 'project_owned[[:space:]]+nested-tool[[:space:]]+present' /tmp/harness-doctor-nested.txt; then
  echo "EXPECTED: doctor reports project_owned nested-tool present (monorepo subtree)" >&2
  cat /tmp/harness-doctor-nested.txt >&2
  exit 1
fi
echo "OK doctor: project-owned nested-tool"

if ! grep -qE 'monorepo|apps/\*/\.agents' "$ROOT/references/project-owned-tools.md"; then
  echo "EXPECTED: project-owned-tools.md must document monorepo subtree scan" >&2
  exit 1
fi
echo "OK project-owned-tools: monorepo subtree documented"

# F2b: high-risk presets must not ship hard-coded project paths / forced vitest
STD="$ROOT/assets/standards"
for f in frontend-vue3.md api-grpc-gateway.md backend-gin.md api-swagger.md; do
  if grep -nE 'web/src/|npx vitest run|每次提交前必须全部通过' "$STD/$f"; then
    echo "EXPECTED_CLEAN: $f still has forbidden hard-codes" >&2
    exit 1
  fi
done
if ! grep -q '启用前提' "$STD/backend-trpc-go.md"; then
  echo "EXPECTED: backend-trpc-go.md must state 启用前提" >&2
  exit 1
fi
for f in backend-gin.md api-swagger.md; do
  if [[ ! -f "$STD/$f" ]]; then
    echo "EXPECTED: missing $f" >&2
    exit 1
  fi
done
if ! grep -qE 'id:[[:space:]]*gin' "$STD/index.yaml" || ! grep -qE 'id:[[:space:]]*swagger' "$STD/index.yaml"; then
  echo "EXPECTED: index.yaml registers gin + swagger" >&2
  exit 1
fi
# Strategy B: no *-generic.md in assets; fallback is skip_unmatched
if compgen -G "$STD/*-generic.md" >/dev/null 2>&1; then
  echo "EXPECTED_GONE: assets/standards/*-generic.md must be removed" >&2
  ls "$STD"/*-generic.md >&2 || true
  exit 1
fi
if ! grep -q 'skip_unmatched' "$STD/index.yaml"; then
  echo "EXPECTED: index.yaml fallback.strategy=skip_unmatched" >&2
  exit 1
fi
echo "OK sanitize: standards presets (no generic)"

# R1/R2: detect-standards monorepo + no false trpc / no false grpc-gateway
DETECT="$ROOT/scripts/detect-standards.sh"
chmod +x "$DETECT"
MONO="$FIX/monorepo-gin-vue"
TRPC="$FIX/trpc-no-proto"
OUT="$(mktemp)"
"$DETECT" --json "$MONO" >"$OUT"
if ! grep -q '"id": "vue3"' "$OUT" || ! grep -q '"id": "gin"' "$OUT" || ! grep -q '"id": "swagger"' "$OUT"; then
  echo "EXPECTED: monorepo-gin-vue → vue3 + gin + swagger" >&2
  cat "$OUT" >&2
  exit 1
fi
if ! grep -q 'apps/ui' "$OUT" || ! grep -q 'apps/server' "$OUT"; then
  echo "EXPECTED: monorepo roots apps/ui + apps/server" >&2
  cat "$OUT" >&2
  exit 1
fi
"$DETECT" --json "$TRPC" >"$OUT"
if grep -q '"id": "trpc-go"' "$OUT"; then
  echo "EXPECTED: trpc-no-proto must NOT match trpc-go" >&2
  cat "$OUT" >&2
  exit 1
fi
if ! grep -q '"id": "gin"' "$OUT"; then
  echo "EXPECTED: trpc-no-proto with gin require matches gin" >&2
  cat "$OUT" >&2
  exit 1
fi
rm -f "$OUT"
echo "OK detect-standards: monorepo + trpc-no-proto"

# L3: thin entry + companion volume dirs for split presets
SPLIT_PRESETS=(
  api-grpc-gateway
  backend-trpc-go
  backend-trpc-agent-go
  backend-go-micro
  quality-code-review
)
for stem in "${SPLIT_PRESETS[@]}"; do
  entry="$STD/${stem}.md"
  vol_dir="$STD/${stem}"
  if [[ ! -f "$entry" ]]; then
    echo "EXPECTED: missing entry $entry" >&2
    exit 1
  fi
  lines="$(wc -l <"$entry" | tr -d ' ')"
  if [[ "$lines" -gt 120 ]]; then
    echo "EXPECTED: $entry has $lines lines (>120)" >&2
    exit 1
  fi
  if [[ ! -d "$vol_dir" ]] || ! compgen -G "$vol_dir/*.md" >/dev/null; then
    echo "EXPECTED: companion dir $vol_dir with at least one .md" >&2
    exit 1
  fi
  if ! grep -qE '分册目录|\./'"${stem}"'/' "$entry"; then
    echo "EXPECTED: $entry must contain 分册目录 or ./${stem}/ links" >&2
    exit 1
  fi
done
echo "OK split-presets: thin entries + companion dirs"

# agents-merge.md：先理解再裁决（反描点）+ 工作单元
AM="$ROOT/references/agents-merge.md"
WU="$ROOT/references/agents-work-units.md"
if [[ ! -f "$AM" ]]; then
  echo "EXPECTED: missing $AM" >&2
  exit 1
fi
if [[ ! -f "$WU" ]]; then
  echo "EXPECTED: missing $WU" >&2
  exit 1
fi
for needle in '先理解' '禁止' '描点' 'RETAIN-ENTRY' '充分条件' '工作单元' '四件套' 'nearest'; do
  if ! grep -q "$needle" "$AM"; then
    echo "EXPECTED: agents-merge.md must contain: $needle" >&2
    exit 1
  fi
done
for needle in '工作单元' 'git ls-files' '局部约定优先' '不覆写'; do
  if ! grep -q "$needle" "$WU"; then
    echo "EXPECTED: agents-work-units.md must contain: $needle" >&2
    exit 1
  fi
done
if ! grep -q 'agents-merge.md' "$ROOT/harness-generating/SKILL.md"; then
  echo "EXPECTED: harness-generating SKILL.md must reference agents-merge.md" >&2
  exit 1
fi
if ! grep -qE 'agents-work-units|工作单元 AGENTS|git ls-files' "$ROOT/harness-generating/SKILL.md"; then
  echo "EXPECTED: harness-generating SKILL.md must reference work-unit scan/index" >&2
  exit 1
fi
if ! grep -q '工作单元 AGENTS' "$ROOT/harness-generating/references/report-template.md"; then
  echo "EXPECTED: report-template.md must have 工作单元 AGENTS section" >&2
  exit 1
fi
DIM12="$ROOT/harness-gardening/references/detection-dimensions.md"
if ! grep -q '维度 12' "$DIM12"; then
  echo "EXPECTED: detection-dimensions.md must define 维度 12" >&2
  exit 1
fi
for needle in '工作单元' '未索引' '误改'; do
  if ! grep -q "$needle" "$DIM12"; then
    echo "EXPECTED: detection-dimensions.md dim12 must contain: $needle" >&2
    exit 1
  fi
done
echo "OK agents-merge: reference + work-units + generating + gardening dim12"

# harness-ide-setup.sh：语法、落盘路径（同 sync-standards-rules）、绝对 GRAPHIFY_OUT
IDE_SETUP="$ROOT/scripts/harness-ide-setup.sh"
chmod +x "$IDE_SETUP"
if grep -qE '^(<<<<<<<|=======|>>>>>>>)' "$IDE_SETUP"; then
  echo "UNEXPECTED: unresolved merge conflict in $IDE_SETUP" >&2
  exit 1
fi
if ! bash -n "$IDE_SETUP"; then
  echo "EXPECTED: harness-ide-setup.sh must be valid bash" >&2
  exit 1
fi

abs_graphify_out_ok() {
  local f="$1"
  grep -q 'git rev-parse --show-toplevel' "$f" \
    && grep -q 'GRAPHIFY_OUT="$_root/docs/dev-map"' "$f" \
    && ! grep -qE 'export GRAPHIFY_OUT=docs/dev-map' "$f"
}

IDE_FIX="$(mktemp -d)"
mkdir -p "$IDE_FIX/.cursor"
"$IDE_SETUP" "$IDE_FIX" >/tmp/ide-setup-cursor.txt
if [[ ! -f "$IDE_FIX/.cursor/rules/graphify.mdc" ]]; then
  echo "EXPECTED: .cursor/rules/graphify.mdc written" >&2
  cat /tmp/ide-setup-cursor.txt >&2
  exit 1
fi
if ! abs_graphify_out_ok "$IDE_FIX/.cursor/rules/graphify.mdc"; then
  echo "EXPECTED: cursor graphify rule uses absolute GRAPHIFY_OUT" >&2
  exit 1
fi
if [[ -d "$IDE_FIX/.codebuddy" ]]; then
  echo "UNEXPECTED: cursor-only layout must not create .codebuddy" >&2
  exit 1
fi
if [[ -d "$IDE_FIX/.codex/rules" ]]; then
  echo "UNEXPECTED: must not create .codex/rules" >&2
  exit 1
fi
rm -rf "$IDE_FIX"
echo "OK harness-ide-setup: cursor entity layout"

IDE_FIX="$(mktemp -d)"
mkdir -p "$IDE_FIX/.cursor" "$IDE_FIX/.codebuddy" "$IDE_FIX/.claude"
"$IDE_SETUP" "$IDE_FIX" >/tmp/ide-setup-multi.txt
for f in \
  "$IDE_FIX/.cursor/rules/graphify.mdc" \
  "$IDE_FIX/.codebuddy/rules/graphify.md" \
  "$IDE_FIX/.claude/rules/graphify.md"; do
  if [[ ! -f "$f" ]]; then
    echo "EXPECTED: missing $f" >&2
    cat /tmp/ide-setup-multi.txt >&2
    exit 1
  fi
  if ! abs_graphify_out_ok "$f"; then
    echo "EXPECTED: $f uses absolute GRAPHIFY_OUT" >&2
    exit 1
  fi
done
if [[ ! -f "$IDE_FIX/.codebuddy/settings.json" ]]; then
  echo "EXPECTED: .codebuddy/settings.json hook-guard" >&2
  exit 1
fi
if ! grep -q 'hook-guard search' "$IDE_FIX/.codebuddy/settings.json" \
  || ! grep -q 'hook-guard read' "$IDE_FIX/.codebuddy/settings.json"; then
  echo "EXPECTED: settings.json has search+read hook-guard" >&2
  cat "$IDE_FIX/.codebuddy/settings.json" >&2
  exit 1
fi
if ! grep -q 'git rev-parse --show-toplevel' "$IDE_FIX/.codebuddy/settings.json"; then
  echo "EXPECTED: hook-guard GRAPHIFY_OUT is absolute via git rev-parse" >&2
  cat "$IDE_FIX/.codebuddy/settings.json" >&2
  exit 1
fi
if [[ -d "$IDE_FIX/.codex/rules" ]]; then
  echo "UNEXPECTED: must not create .codex/rules" >&2
  exit 1
fi
rm -rf "$IDE_FIX"
echo "OK harness-ide-setup: multi-IDE + hook-guard"

IDE_FIX="$(mktemp -d)"
mkdir -p "$IDE_FIX/.agents"
ln -s .agents "$IDE_FIX/.cursor"
ln -s .agents "$IDE_FIX/.codebuddy"
"$IDE_SETUP" "$IDE_FIX" >/tmp/ide-setup-fallback.txt
if [[ ! -f "$IDE_FIX/.agents/rules/graphify.mdc" ]] || [[ ! -f "$IDE_FIX/.agents/rules/graphify.md" ]]; then
  echo "EXPECTED: fallback writes .agents/rules dual format" >&2
  cat /tmp/ide-setup-fallback.txt >&2
  exit 1
fi
if [[ ! -f "$IDE_FIX/.agents/settings.json" ]]; then
  echo "EXPECTED: fallback hook-guard via .codebuddy → .agents/settings.json" >&2
  exit 1
fi
rm -rf "$IDE_FIX"
echo "OK harness-ide-setup: symlink fallback"

echo "all harness-verify fixture tests passed"
