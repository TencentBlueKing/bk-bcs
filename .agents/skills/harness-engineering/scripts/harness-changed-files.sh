#!/usr/bin/env bash
# 获取本次巡检的变更文件列表
# 入参：$1=DIFF_BASE $2=REPORT_PATH
# 输出：变更文件列表（每行一个路径），或单行 UPGRADE_TO_FULL 信号
set -euo pipefail

DIFF_BASE="${1:-}"
REPORT_PATH="${2:-docs/harness/gardening-report.md}"

if [ -n "$DIFF_BASE" ]; then
    # 功能分支：基于切出点做 diff
    git diff --name-only "$DIFF_BASE"..HEAD
else
    # master / 非 issue 分支：读取 last-commit 字段
    if [ ! -f "$REPORT_PATH" ]; then
        echo "UPGRADE_TO_FULL"
        exit 0
    fi
    LAST=$(grep -oP '(?<=last-commit: )\S+' "$REPORT_PATH" || true)
    if [ -z "$LAST" ]; then
        echo "UPGRADE_TO_FULL"
        exit 0
    fi
    if ! git cat-file -e "${LAST}^{commit}" 2>/dev/null; then
        echo "UPGRADE_TO_FULL"
        exit 0
    fi
    git diff --name-only "$LAST"..HEAD
fi
