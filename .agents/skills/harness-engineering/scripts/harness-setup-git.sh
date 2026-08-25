#!/usr/bin/env bash
# 配置 graph.json merge driver（仅当用户显式入库 graph.json 时有用）与 gardening-report merge 策略
# 默认 harness 策略：graph.json 由 docs/dev-map/.gitignore 忽略、不提交；merge driver 为 opt-in 预留
# 幂等可重复执行，输出 GITATTR_MODIFIED=true/false 供调用方决定是否 git add
set -euo pipefail

GITATTR="${1:-.gitattributes}"
MODIFIED=false

ensure_gitattr() {
    grep -qF "$1" "$GITATTR" 2>/dev/null || {
        echo "$1" >> "$GITATTR"
        MODIFIED=true
    }
}

[ -f "$GITATTR" ] || touch "$GITATTR"
# opt-in：若仓库显式跟踪 graph.json，union-merge 可避免无意义冲突
ensure_gitattr "docs/dev-map/graph.json merge=graphify"
ensure_gitattr "docs/harness/gardening-report.md merge=ours"

if ! git config merge.graphify.driver >/dev/null 2>&1; then
    git config merge.graphify.name 'graphify graph.json union merge'
    git config merge.graphify.driver 'graphify merge-driver %O %A %B'
fi

echo "GITATTR_MODIFIED=$MODIFIED"
