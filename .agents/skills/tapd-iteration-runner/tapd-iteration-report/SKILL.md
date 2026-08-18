---
name: tapd-iteration-report
slug: tapd-iteration-report
version: 1.0.0
description: |
  迭代执行流水线的收尾阶段。当迭代内所有需求 phase 均为 committed 后触发，
  扫描 commit.md 生成 changelog，汇总研发效能数据写入 summary.md，
  清理临时文件，递增 patch 版本号，最终提交并推送至远程仓库。
---

# 迭代信息汇总

## 定位

本子 skill 是迭代执行流水线的最后一环——在所有需求完成开发并提交后，
对整个迭代进行总结性收尾：构建版本 changelog、汇总研发效能数据、清理临时文件、
递增 patch 版本号，并将汇总结果提交推送至远程仓库。

## 前置条件

编排层在调用本子 skill 前需确认以下条件全部满足：

1. **所有需求 phase 为 `committed`**：`iteration-state.json` 中 `sequence` 列出的所有需求，
   其 phase 字段必须全部为 `committed`
2. **所有需求目录存在 `commit.md`**：`specs/${VERSION}/${ID}/commit.md` 文件必须全部存在
3. **迭代 status 为 `completed`**：`iteration-state.json.status` 为 `completed`

如果任一条件不满足，输出明确的错误信息并终止执行，由编排层决定后续处理。

## 状态文件

从 `specs/${VERSION}/iteration-state.json` 读取以下关键信息：

| 字段 | 用途 |
|------|------|
| `patch` | 当前 patch 版本号，用于构建版本标题（如 v0.9.3） |
| `iter_branch` | 迭代分支名（如 v0.9.x），用于推送 |
| `started_at` | 迭代开始时间 |
| `end_at` | 迭代结束时间 |
| `stories` | 所有需求及其统计字段 |
| `sequence` | 需求执行顺序列表 |

**版本号构成规则**：
- 迭代目录名即为版本前缀（如 `v0.9.x` → `v0.9.`）
- 完整版本号 = 版本前缀去掉末尾的 `x` + patch 值（如 `v0.9.` + `3` → `v0.9.3`）

## 执行流程

### 1. 读取迭代状态

```bash
VERSION="<迭代目录名>"  # 如 v0.9.x
STATE_FILE="specs/${VERSION}/iteration-state.json"

# 读取关键字段
PATCH=$(jq -r '.patch' "$STATE_FILE")
iter_branch=$(jq -r '.iter_branch' "$STATE_FILE")
STARTED_AT=$(jq -r '.started_at' "$STATE_FILE")
END_AT=$(jq -r '.end_at' "$STATE_FILE")

# 构建版本号：将目录名末尾的 x 替换为 patch 值
# 例如 v0.9.x → v0.9.3
VERSION_TAG=$(echo "$VERSION" | sed "s/x$/${PATCH}/")
```

### 2. 生成 Changelog

扫描本迭代目录下所有需求的 `commit.md` 文件，提取 commit 信息，汇总为版本 changelog。

逻辑：
1. 如果 `changelog.md` 不存在，创建并写入标题 `# Changelog`
2. 追加本次版本章节标题 `## ${VERSION_TAG}` 和发布时间
3. 按 `sequence` 顺序遍历每个需求：
   - 从 `iteration-state.json.stories` 中提取需求名称
   - 从 `${ID}/commit.md` 中提取 `## Commit Message` 章节内容
   - 追加到 changelog

详细脚本实现见 `../../references/report-scripts.md` §1。

### 3. 汇总研发效能数据

从 `iteration-state.json` 的 `stories` 中提取所有需求的统计字段进行汇总。

**代码变更指标**：

| 指标 | 来源字段 | 汇总方式 |
|------|---------|---------|
| 总变更行数 | `stats.total` | 求和 |
| 新增代码行数 | `stats.add_code` | 求和 |
| 删除代码行数 | `stats.delete_code` | 求和 |
| 逻辑代码行数 | `stats.logic_code` | 求和 |
| 测试代码行数 | `stats.test_code` | 求和 |
| 文档变更行数 | `stats.docs` | 求和 |
| 变更文件数 | `stats.files` | 求和 |
| 需求实现数量 | — | 计数 sequence 长度 |
| 每个需求实现时间 | `stats.duration_sec` | runner 记录的壁钟耗时，直接读取 |
| 平均实现时间 | — | 求平均 |
| 最大实现时间 | — | 取最大值 |
| 最小实现时间 | — | 取最小值 |

**成本指标**（从各需求的 `stats.cost` 聚合，分三层统计展示）：

#### 数据来源与计算规则

| 数据项 | 来源 | 说明 |
|--------|------|------|
| 需求总耗时 | `stats.duration_sec` | runner 记录的壁钟耗时（committed 时刻 − started_at，含等待用户决策的间隔）|
| 需求总 credit | `per_stage.*.credit` 求和 | 执行层各阶段求和 |
| 需求总 tokens | `per_stage.*.total_*_tokens` 求和 | 执行层各阶段求和 |
| 迭代总耗时 | 各需求 `stats.duration_sec` 求和 | |
| 迭代总 credit | 各需求的总 credit 求和 | |
| 迭代总 tokens | 各需求的总 tokens 求和 | |

> **关键语义**：
> - `stats.duration_sec` 为 runner 在 loop 中记录的壁钟耗时（需求首次开始到 committed），包含等待用户决策的间隔
> - credit 和 tokens 仅来自执行层 `per_stage`

**输出文件**：`specs/${VERSION}/summary.md`，追加模式（不覆盖历史记录）。

summary.md 的成本汇总章节格式：

```markdown
### 成本汇总

#### 第一层：迭代总体

| 指标 | 值 |
|------|-----|
| 迭代总耗时 | <sum_duration> |
| 迭代总 credit | <sum_credit> |
| 迭代总 tokens（输入） | <sum_input_tokens> |
| 迭代总 tokens（输出） | <sum_output_tokens> |
| 迭代总 tokens（缓存） | <sum_cache_tokens> |
| 迭代 speckit 调用次数 | <sum_speckit_calls> |
| 需求数量 | <story_count> |
| 单需求平均耗时 | <avg_duration> |
| 单需求平均 credit | <avg_credit> |
| 单需求最高 credit | <max_credit>（需求 #<ID>）|

#### 第二层：各需求独立统计

| 需求 ID | 需求名称 | 耗时 | Credit | 输入 tokens | 输出 tokens | 缓存 tokens | 总 tokens | 调用次数 |
|---------|---------|------|--------|------------|------------|------------|----------|---------|
| #<ID_1> | <name> | <dur> | <credit> | <input> | <output> | <cache> | <total> | <calls> |
| #<ID_2> | <name> | <dur> | <credit> | <input> | <output> | <cache> | <total> | <calls> |
| ... | ... | ... | ... | ... | ... | ... | ... | ... |
| **合计** | — | <sum> | <sum> | <sum> | <sum> | <sum> | <sum> | <sum> |

> 每行 tokens = `per_stage` 各阶段 `*_tokens` 求和
> 每行 Credit = `per_stage` 各阶段 `.credit` 求和
> 每行 耗时 = `stats.duration_sec`

#### 第三层：各阶段成本统计（执行层）

| 阶段 | 总 credit | 占比 | 输入 tokens | 输出 tokens | 缓存 tokens | 调用次数 |
|------|-----------|------|------------|------------|------------|---------|
| clarify | <x> | x% | <x> | <x> | <x> | <x> |
| specify | <x> | x% | <x> | <x> | <x> | <x> |
| plan | <x> | x% | <x> | <x> | <x> | <x> |
| tasks | <x> | x% | <x> | <x> | <x> | <x> |
| implement | <x> | x% | <x> | <x> | <x> | <x> |
| validate | <x> | x% | <x> | <x> | <x> | <x> |
| commit | <x> | x% | <x> | <x> | <x> | <x> |
| **执行层合计** | <sum> | 100% | <sum> | <sum> | <sum> | <sum> |
```

详细脚本实现见 `../../references/report-scripts.md` §2。

> **注意**：`children` 字段需兼容 null / `{}` / 不存在三种空值形式。
> `stats.cost` 字段需兼容不存在的情况——缺失时该需求的 cost 视为 0。
> `stats.cost.per_stage` 字段需兼容不存在 / `{}` 的情况——缺失时各阶段 cost 视为 0。
> `stats.duration_sec` 字段需兼容不存在的情况——缺失时该需求耗时视为 0。

### 4. 清理临时目录

删除本迭代目录下的 `.tmp/` 临时目录：

**Linux / macOS:**
```bash
rm -rf "specs/${VERSION}/.tmp/"
```

**Windows (PowerShell):**
```powershell
if (Test-Path "specs\$VERSION\.tmp") {
    Remove-Item "specs\$VERSION\.tmp" -Recurse -Force
}
```

### 5. 递增 patch 版本号

将 `iteration-state.json` 中的 `patch` 字段自增 1，为下一轮迭代做准备：

**Linux / macOS:**
```bash
jq '.patch += 1' "$STATE_FILE" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"
```

**Windows (PowerShell):**
```powershell
$state = Get-Content $STATE_FILE | ConvertFrom-Json
$state.patch = $state.patch + 1
$state | ConvertTo-Json -Depth 10 | Set-Content $STATE_FILE -Encoding UTF8
```

### 6. 提交并推送

将迭代汇总结果提交到迭代分支，并推送至远程仓库：

**Linux / macOS:**
```bash
git add -A
git commit -m "docs(迭代): ${VERSION_TAG} 变更日志与效能报告

- 从各需求提交记录生成版本变更日志
- 汇总迭代研发效能数据
- 清理临时文件
- 递增 patch 版本号至 $(jq -r '.patch' "$STATE_FILE")

迭代版本: ${VERSION_TAG}"

git push origin "${iter_branch}"
```

**Windows (PowerShell):**
```powershell
git add -A
git commit -m "docs(迭代): $VERSION_TAG 变更日志与效能报告`n`n- 从各需求提交记录生成版本变更日志`n- 汇总迭代研发效能数据`n- 清理临时文件`n- 递增 patch 版本号至 $($state.patch)`n`n迭代版本: $VERSION_TAG"

git push origin $iter_branch
```

## 产出

| 产出物 | 路径 | 说明 |
|--------|------|------|
| Changelog | `specs/${VERSION}/changelog.md` | 版本变更记录，追加模式 |
| 效能报告 | `specs/${VERSION}/summary.md` | 研发效能数据汇总，追加模式 |
| 状态更新 | `specs/${VERSION}/iteration-state.json` | patch 已自增 |
| Git 提交 | 迭代分支 | 已提交并推送至远程 |
| 清理 | `specs/${VERSION}/.tmp/` 已删除 | 临时文件已清理 |

## 异常处理

| 异常场景 | 处理方式 |
|---------|---------|
| 存在未 committed 的需求 | 输出未完成需求列表，终止执行 |
| commit.md 文件缺失 | 输出缺失文件列表，终止执行 |
| git push 失败 | 提示用户检查远程仓库权限和网络，保留本地提交 |
| jq 命令不可用 | 使用编程方式直接解析 JSON（编排层处理） |
| changelog.md/summary.md 已存在 | 追加新版本章节，不覆盖历史记录 |
