#!/usr/bin/env bash
# 确定报告路径和 diff 基线，输出供 eval 的环境变量：REPORT_PATH 和 DIFF_BASE
set -euo pipefail

BRANCH=$(git branch --show-current)
ISSUE_NUM=$(echo "$BRANCH" | grep -oP '(?:feature/)?issue-\K[0-9]+' || true)

if [ -n "$ISSUE_NUM" ]; then
    REPORT_PATH="workflows/issue-$ISSUE_NUM/gardening-report.md"
    mkdir -p "workflows/issue-$ISSUE_NUM"

    # 沿 HEAD 向回遍历，找第一个被其他远端分支包含的 commit（即切出点）
    DIFF_BASE=$(git log HEAD --pretty=%H | while read -r sha; do
        if git branch -r --contains "$sha" 2>/dev/null \
            | grep -v "/$BRANCH\$" | grep -q .; then
            echo "$sha"
            break
        fi
    done)

    # fallback：无远端信息时退而使用 master merge-base
    if [ -z "${DIFF_BASE:-}" ]; then
        DIFF_BASE=$(git merge-base HEAD origin/master 2>/dev/null \
                 || git merge-base HEAD master 2>/dev/null \
                 || true)
    fi

    echo "REPORT_PATH=$REPORT_PATH"
    echo "DIFF_BASE=${DIFF_BASE:-}"
else
    echo "REPORT_PATH=docs/harness/gardening-report.md"
    echo "DIFF_BASE="
fi
