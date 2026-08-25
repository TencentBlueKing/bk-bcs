---
name: tapd-iteration-init
slug: tapd-iteration-init
version: 2.0.0
description: |
  迭代执行流水线启动器。分三个阶段执行：信息检查（验证 git 环境、获取迭代信息、
  解析迭代分支）→ 恢复检测（已有状态文件时判定恢复策略）→ 新流程初始化（创建分支、
  迭代目录和 iteration-state.json）。
---

# 环境初始化与恢复检测

本 skill 是迭代执行流水线的第一个子 skill，负责准备执行环境。
支持 Linux / macOS (Bash) 和 Windows (PowerShell) 双平台。
执行流程分为三个阶段，按顺序进行：

```
阶段 1: 信息检查 ─→ 阶段 2: 恢复检测 ─→ 阶段 3: 新流程初始化
                           │
                           └─ 检测到已有状态 → 执行恢复流程（不进入阶段 3）
```

> **Windows 首次运行**：需先设置 UTF-8 编码：
> ```powershell
> [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
> ```

---

## 阶段 1: 信息检查

依次完成三项检查，任一失败则终止并告知用户。

### 1.1 确认本地 Git 项目

验证当前目录是一个有效的 git 仓库：

**Linux / macOS (Bash):**
```bash
git rev-parse --is-inside-work-tree
```

**Windows (PowerShell):**
```powershell
git rev-parse --is-inside-work-tree
```

失败时终止，告知用户"当前目录不是 git 仓库，请在项目根目录下执行"。

同时采集项目基础信息供后续使用：

**Linux / macOS (Bash):**
```bash
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@')
MAIN_BRANCH=${MAIN_BRANCH:-master}
PROJECT_PATH=$(git remote get-url origin | sed 's/.*://;s/\.git$//')
```

**Windows (PowerShell):**
```powershell
$MAIN_BRANCH = (git symbolic-ref refs/remotes/origin/HEAD 2>$null) -replace '^refs/remotes/origin/',''
if (-not $MAIN_BRANCH) { $MAIN_BRANCH = 'master' }
$PROJECT_PATH = (git remote get-url origin) -replace '.*:','' -replace '\.git$',''
```

### 1.2 获取迭代信息

需要两个参数来调用 TAPD MCP `iterations_get`。参数确定前，先检查 `project.json` 是否存在：

**workspace_id** — 按以下优先级确定：
1. 用户消息中显式指定 → 直接使用
2. 项目根目录 `project.json` 中的 `workspace_id` → 读取并告知用户
3. 以上均无 → 询问用户输入

**iteration_id** — 按以下优先级确定：
1. 用户消息中显式指定 → 直接使用
2. 以上无 → 询问用户输入（无默认值，必须提供）

两个参数就绪后，调用 TAPD MCP 获取迭代详情：

```
调用 TAPD MCP: iterations_get
参数: workspace_id, id=iteration_id
返回: 迭代名称、起止时间等信息
```

调用失败时终止，提示用户检查 TAPD MCP 连接或参数是否正确。

### 1.3 解析迭代分支名称

从迭代名称中`name`提取版本号作为分支名。迭代名称规范为 `iteration-vMAJOR.MINOR.x`。

**解析规则**：提取 `v` 开头的版本部分作为分支名和目录名。

| 迭代名称 | 解析出的版本（VERSION） | 分支名 | 迭代目录 |
|---------|----------------------|--------|---------|
| `iteration-v0.9.x` | `v0.9.x` | `v0.9.x` | `specs/v0.9.x/` |
| `iteration-v1.0.x` | `v1.0.x` | `v1.0.x` | `specs/v1.0.x/` |
| `iteration-v2.1.x` | `v2.1.x` | `v2.1.x` | `specs/v2.1.x/` |

**Linux / macOS (Bash):**
```bash
VERSION=$(echo "${ITERATION_NAME}" | grep -oP 'v[\d.]+x?' || echo "${ITERATION_NAME}")
```

**Windows (PowerShell):**
```powershell
$VERSION = if ($ITERATION_NAME -match 'v[\d.]+x?') { $Matches[0] } else { $ITERATION_NAME }
```

如果迭代名称不符合 `iteration-vMAJOR.MINOR.x` 规范，使用完整迭代名称作为 VERSION，
并提示用户确认。

### 补充信息采集

在信息检查阶段还需采集以下辅助信息：

**owner（当前用户）**：
1. 用户消息中显式指定 → 直接使用
2. `project.json` 中的 `owner` → 读取使用
3. 以上均无 → 询问用户

**agent_tool（Agent 工具）**：
从用户消息中提取工具名称（如"使用 claude 工具"→ `claude`），未指定时默认为 `agent`。

---

## 阶段 2: 恢复检测

检查 `specs/${VERSION}/iteration-state.json` 是否已存在。

### 不存在状态文件 → 进入阶段 3

状态文件不存在说明是全新迭代，直接进入阶段 3（新流程初始化）。

### 存在状态文件 → 执行恢复逻辑

读取已有的 `iteration-state.json`，根据状态决定后续动作：

| 已有 status | 用户意图 | 动作 |
|------------|---------|------|
| `completed` | 任意 | 告知用户该迭代已完成；询问是否重新开始（重新开始则删除旧状态，进入阶段 3） |
| 其他非终态 | 任意 | 读取 `references/recovery.md` 执行恢复流程 |

恢复流程的详细规则参见 `references/recovery.md`，核心逻辑为：
1. 展示当前进度摘要（已完成/进行中/未开始的需求数）
2. 检查状态一致性
3. 按 `sequence` 顺序找到第一个未完成需求，根据其 phase 跳转对应子 skill
4. 如用户消息已表达继续意图（"帮我继续"、"接着做"），视为已确认直接恢复

**注意**：走恢复流程时不进入阶段 3，恢复完成后直接告知编排层从哪个子 skill 继续。

---

## 阶段 3: 新流程初始化

仅在阶段 2 确认无已有状态（或用户要求重新开始）时执行。

### 3.1 创建迭代分支

从主分支创建新的版本分支并推送到远程：

**Linux / macOS (Bash):**
```bash
git checkout -b "${VERSION}" "${MAIN_BRANCH}"
git push -u origin "${VERSION}"
```

**Windows (PowerShell):**
```powershell
git checkout -b $VERSION $MAIN_BRANCH
git push -u origin $VERSION
```

如果分支已存在（如远程已有同名分支），切换到该分支而非报错：

**Linux / macOS (Bash):**
```bash
git checkout "${VERSION}" 2>/dev/null || git checkout -b "${VERSION}" "${MAIN_BRANCH}"
git push -u origin "${VERSION}" 2>/dev/null || true
```

**Windows (PowerShell):**
```powershell
git checkout $VERSION 2>$null
if ($LASTEXITCODE -ne 0) { git checkout -b $VERSION $MAIN_BRANCH }
git push -u origin $VERSION 2>$null
```

### 3.2 创建迭代目录

**Linux / macOS (Bash):**
```bash
mkdir -p "specs/${VERSION}"
```

**Windows (PowerShell):**
```powershell
New-Item -ItemType Directory -Force -Path "specs\$VERSION" | Out-Null
```

### 3.3 创建 iteration-state.json

在迭代目录下写入初始状态文件 `specs/${VERSION}/iteration-state.json`：

```json
{
  "iteration_id": "<从 1.2 获取>",
  "workspace_id": "<从 1.2 确定>",
  "iteration_name": "<从 1.2 获取的迭代名称>",
  "owner": "<补充信息采集确定>",
  "project_path": "<从 1.1 采集>",
  "iter_branch": "<从 1.3 采集 VERSION>",
  "agent_tool": "<补充信息采集确定，默认 agent>",
  "patch": 0,
  "started_at": "<当前时间 ISO 8601>",
  "end_at": "",
  "status": "initialized",
  "all_parents": [],
  "selected_story": "",
  "stories": {},
  "sequence": []
}
```

`all_parents`，`selected_story`，`stories` 和 `sequence` 在本阶段留空，由 tapd-iteration-analysis 负责填充。

---

## 产出

本 skill 完成后，保证以下条件成立：

| 产出项 | 新流程 | 恢复流程 |
|-------|--------|---------|
| 版本分支已创建并推送 | ✅ | ✅（已存在） |
| 迭代目录 `specs/${VERSION}/` 已创建 | ✅ | ✅（已存在） |
| `iteration-state.json` 已写入 | ✅（新建） | ✅（已读取） |
| 恢复起点已确定 | — | ✅ |

完成后告知编排层下一步进入哪个子 skill：
- **新流程** → `tapd-iteration-analysis`
- **恢复流程** → 由 `references/recovery.md` 映射表决定
