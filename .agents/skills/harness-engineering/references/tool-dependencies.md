# Agent 工具依赖清单

> 本文件记录本项目所有 Agent 和 Skill 运行所需的工具依赖。
> harness-generating 规范生成完成后，会对照本清单检查当前环境的工具完备性。
>
> **MCP 逻辑名说明**：表中 `tapd` 为逻辑名；实际 IDE 中的 Server 标识以各 skill 的 `meta.json` 或 Agent frontmatter（如 `tapd_mcp`）为准。

## 路径上下文说明

本项目（ai-practice）是 **skill/agent 开发仓**，其他项目通过将本仓内容部署到 IDE 配置目录来接入。两种上下文下的 skill 根目录不同，按以下优先级依次检测：

| 上下文 | 项目类型 | 项目级 skill 根目录（`{SKILL_ROOT}`） | 识别方式 |
|--------|---------|--------------------------------------|---------|
| **开发仓** | ai-practice 本仓 | `skills` | `skills/harness-engineering/SKILL.md` 存在 |
| **接入仓** | 其他业务项目（CodeBuddy） | `.codebuddy/skills` | `.codebuddy/skills/harness-engineering/SKILL.md` 存在 |
| **接入仓** | 其他业务项目（Claude） | `.claude/skills` | `.claude/skills/harness-engineering/SKILL.md` 存在 |
| **接入仓** | 其他业务项目（Cursor） | `.cursor/skills` | `.cursor/skills/harness-engineering/SKILL.md` 存在 |

本文件中凡以 **`{SKILL_ROOT}`** 标注的路径，均需按上述规则替换为实际前缀。

> **不受影响的路径**：
> - 通用框架 skill（systematic-debugging、brainstorming 等）始终位于对应 IDE 配置目录的 `skills/` 下，路径不变。
> - `.specify/`、`.agents/commands/` 等目标业务仓库自身结构，路径不变。

## 工具分类说明

| 类型 | 说明 | 检查方式 |
|------|------|---------|
| **MCP** | Model Context Protocol 服务 | 自行调用该 MCP 的任意只读接口；响应成功则就绪，报错则为缺口。**严禁要求用户在 IDE 中手动确认是否已连接** |
| **CLI** | 命令行工具 | `which <cmd>` 或 `command -v <cmd>` |
| **自定义脚本** | 项目内脚本 | 检查文件是否存在且可执行 |
| **配置** | 项目或 skill 级 JSON | 检查目标路径文件是否存在且字段齐全 |

---

## 一、MCP 服务依赖

### tapd（TAPD 项目管理）

| 依赖方 | 所需接口 | 必需 | 说明 |
|--------|---------|------|------|
| tapd-product-discovery | `stories_get`, `stories_create`, `stories_update`, `tapd_id_get` | 是 | 产品父单读取/创建/回写，角色子单创建 |
| tapd-story-clarification | `stories_get`, `stories_update` | 是 | 需求拉取与回写 |
| tapd-story-evaluation | `stories_get`, `stories_create`, `tapd_id_get` | 是 | 需求评估与子需求创建 |
| tapd-story-breakdown | `tapd_id_get` | 是 | 短 ID 转长 ID |
| tapd-story-score | — | 否 | 仅写本地需求文档；TAPD 回写由主 skill 负责 |
| tapd-iteration-plan | `iterations_get`, `iterations_create`, `stories_get`, `stories_update` | 是 | 迭代创建与需求规划 |
| tapd-iteration-runner | `stories_get`, `iterations_get` | 是 | 迭代需求拉取 |
| tapd-iteration-init | `iterations_get` | 是 | 迭代信息获取 |
| tapd-iteration-analysis | `stories_get` | 是 | 需求详情拉取 |
| tapd-story-pipeline | `stories_get` | 是 | 需求内容兜底拉取 |
| tapd-story-commit | `stories_update` | 是 | 提交后更新需求状态 |
| tapd-story-review | `stories_get`, `stories_update`, `comments_*` | 是 | 需求评审与评论汇总 |
| tapd-story-govern-pipeline | `stories_get`, `stories_update` | 是 | 前置治理编排（澄清→评审→评估） |
| graph-engineering | `stories_get`, `stories_update`（按 flow） | 条件 | 实验性编排；story flow 需 TAPD |
| tapd-bug-clarification | `bugs_get`, `bugs_update`, `tapd_field_detail_get` | 是 | 缺陷详情拉取、根因澄清结果回写 |
| tapd-bug-evaluation | `bugs_get`, `bugs_update`, `tapd_field_detail_get` | 是 | 缺陷工时/规模评估结果回写 |

**环境检查**：自行调用 `stories_get`（取 1 条，最小参数），响应成功则就绪，报错则记为缺口。

### gongfeng（工蜂 Git）

| 依赖方 | 所需接口 | 必需 | 说明 |
|--------|---------|------|------|
| flow-steward / develop-flow | Issue / MR / 提交查询 | 条件 | 本仓迭代 / develop-flow 涉及工蜂 Issue 时 |
| tapd-story-agent | Issue / MR 查询 | 条件 | 需求实现关联工蜂时 |
| speckit-execution-agent | 代码 / MR 查询 | 条件 | 执行器按需 |
| issue-feasibility-analysis | `get_issue_detail` 等 | 是 | 拉取工蜂 Issue 详情做可行性分析 |
| issue-batch-analysis | Issue 查询 + git | 是 | 多 Issue 批量前置 |

**环境检查**：调用任意只读接口（如 `get_current_user`），响应成功则就绪。

### bkm-bkte（监控与可观测性）

| 依赖方 | 所需服务 | 必需 | 说明 |
|--------|---------|------|------|
| sre-engineer | metrics / logs / dashboards / tracing / alarms / events / metadata / dashboard-edit | 条件 | 上线前准备与上线后运营 |

**环境检查**：对各 `bkm-bkte-*` MCP 发起一次只读探测；任一可用即记为部分就绪。

---

## 二、CLI 工具依赖

### 核心 CLI

| 工具 | 依赖方 | 必需 | 检查命令 | 说明 |
|------|--------|------|---------|------|
| `git` | tapd-iteration-runner, tapd-iteration-init, tapd-story-commit, tapd-story-pipeline, graph-engineering, issue-batch-analysis, speckit-git-*, code-review, bk-security-redlines | 是 | `git --version` | 版本控制 |

### Spec Kit / speckit（无独立 CLI）

| 依赖 | 依赖方 | 必需 | 说明 |
|------|--------|------|------|
| `{SKILL_ROOT}/speckit-*/SKILL.md` | speckit-execution-agent（经 `use_skill`） | 是 | specify / plan / tasks / analyze / implement / checklist / constitution |
| `.agents/commands/speckit.*.md` | 同上 | 是 | 命令入口（目标业务仓库） |
| `.specify/` 目录 | speckit-git-*、speckit-implement 等 | 是 | Spec Kit 项目结构（目标业务仓库） |

### 微服务开发 CLI

> 检测条件满足时才扫描；不满足则跳过整组。

| 工具 | 依赖方 | 必需 | 检测条件 | 检查命令 | 说明 |
|------|--------|------|---------|---------|------|
| `go` | go-micro-service, micro-service-project-init | 是 | `go.mod` 存在 | `go version` | Go 编译器 |
| `protoc` | go-micro-service, micro-service-project-init | 是 | `go.mod` 存在且含 go-micro 依赖 | `protoc --version` | Protocol Buffers 编译器（≥ v33.5） |
| `protoc-gen-go` | go-micro-service, micro-service-project-init | 是 | 同上 | `which protoc-gen-go` | protobuf Go 代码生成插件 |
| `protoc-gen-micro` | go-micro-service, micro-service-project-init | 是 | 同上 | `which protoc-gen-micro` | go-micro 代码生成插件 |
| `protoc-gen-grpc-gateway` | go-micro-service, micro-service-project-init | 是 | 同上 | `which protoc-gen-grpc-gateway` | grpc-gateway 代码生成插件 |
| `protoc-gen-openapiv2` | go-micro-service, micro-service-project-init | 是 | 同上 | `which protoc-gen-openapiv2` | OpenAPI 规范生成插件 |
| `make` | micro-service-project-init | 是 | 同上 | `make --version` | 构建工具 |
| `docker` | micro-service-project-init | 否 | — | — | 容器构建（可选，不主动扫描） |

### 通用 CLI

| 工具 | 依赖方 | 必需 | 检测条件 | 检查命令 | 说明 |
|------|--------|------|---------|---------|------|
| `bash` | tapd-iteration-report, 多个 skill 脚本 | 是 | 始终 | `bash --version` | Shell 脚本执行（Linux/macOS） |
| `jq` | tapd-iteration-report, report-scripts.md | 是* | 始终 | `jq --version` | 迭代报告 JSON 解析（*Linux/macOS；Windows 可编程替代） |
| `node` | brainstorming（server.cjs） | 否 | `package.json` 存在 | `node --version` | 可视化伴生服务 |
| `python3` | tapd-story-pipeline（log_usage.py）, skill-creator 脚本 | 否 | — | — | 成本采集 hook（可选，不主动扫描） |
| `gh` | finishing-a-development-branch | 否 | — | — | GitHub CLI（可选，不主动扫描） |
| `graphify` / `graphify` skill | harness-gardening（维度 8）、harness-generating（第三步-C） | 否 | `docs/dev-map/` 目录不存在或图谱需更新时 | 检查 `$SKILL_ROOT/graphify/SKILL.md` 与 `graphify --version` | 知识图谱 CLI + skill；不存在时跳过 dev-map 操作 |

---

## 三、配置文件依赖

> `{SKILL_ROOT}` 含义见「路径上下文说明」。

| 文件路径 | 依赖方 | 必需 | 说明 |
|---------|--------|------|------|
| `project.json` | tapd-product-discovery, tapd-iteration-runner, tapd-iteration-init, tapd-story-clarification, tapd-story-evaluation, tapd-iteration-plan, tapd-story-pipeline, graph-engineering | 是* | 含 `workspace_id`、可选 `owner`；*在**目标业务仓库**根目录，本治理仓可不包含 |
| `.specify/integration.json` | speckit-* | 条件 | Spec Kit 集成配置（目标业务仓库自身路径，不变） |
| `.specify/extensions.yml` | speckit-taskstoissues 等 | 条件 | 扩展 hook 注册（目标业务仓库自身路径，不变） |

---

## 四、按使用场景分组的检查清单

### 场景 A：TAPD 迭代研发流水线（v3 主线）

使用 `tapd-iteration-runner` / `tapd-story-pipeline` 时需要：

- [ ] MCP: tapd — TAPD MCP Server 已配置且可连接
- [ ] Agent: `tapd-story-agent`、`speckit-execution-agent` — `agents/<name>.md` 存在
- [ ] CLI: `git` — 已安装且可正常执行
- [ ] CLI: `jq` — JSON 解析（Linux/macOS，`tapd-iteration-report` 必需）
- [ ] CLI: `bash` — Shell 脚本执行（Linux/macOS）
- [ ] 配置: `project.json` — 包含 `workspace_id` 和 `owner`（目标业务仓库）

实验性 `graph-engineering` 以 `graph-steward` 为正式入口，并按 flow 声明专业 worker Agent；它不替代
上述稳定 Pipeline 依赖。历史「dispatcher」仅为兼容称呼，无独立 `agents/dispatcher.md`。

使用 `graph-engineering` / `graph-steward` 时额外需要：

- [ ] MCP: tapd — story flow 拉取/回写需求时（`mcps: ["tapd", "git"]`）
- [ ] CLI: `git` — 分支与 worktree 操作
- [ ] Agent: `graph-steward` — `agents/graph-steward.md` 存在
- [ ] Skill: `graph-engineering-clarify` 等内聚 worker — 按 flow YAML 声明加载（顶层兼容保留 `story-specify`）

### 场景 B：需求前期处理

使用 `tapd-product-discovery` / `tapd-story-clarification` / `tapd-story-review` / `tapd-story-evaluation` / `tapd-story-govern-pipeline` / `tapd-iteration-plan` 时需要：

- [ ] MCP: tapd — TAPD MCP Server 已配置且可连接
- [ ] 配置: `project.json` — 包含 `workspace_id`（目标业务仓库）

### 场景 C：微服务项目开发

使用 `go-micro-service` / `micro-service-project-init` 时需要（检测条件：`go.mod` 存在且含 go-micro 依赖）：

- [ ] CLI: `go` — Go 编译器（≥ 1.21）
- [ ] CLI: `protoc` — Protocol Buffers 编译器（≥ v33.5）
- [ ] CLI: `protoc-gen-go` — protobuf Go 插件
- [ ] CLI: `protoc-gen-micro` — go-micro 插件
- [ ] CLI: `protoc-gen-grpc-gateway` — grpc-gateway 插件
- [ ] CLI: `protoc-gen-openapiv2` — OpenAPI 规范插件
- [ ] CLI: `make` — 构建工具

### 场景 D：代码评审与安全检查

使用 `code-review` / `bk-security-redlines` 独立触发，或由 `tapd-story-validate` 经 subagent 加载时需要：

- [ ] CLI: `git` — 用于获取变更范围

### 场景 E：缺陷澄清与评估

使用 `tapd-bug-clarification` / `tapd-bug-evaluation` 时需要：

- [ ] MCP: tapd — TAPD MCP Server 已配置且可连接（需支持 `bugs_get`、`bugs_update`、`tapd_field_detail_get`）
- [ ] 配置: `project.json` — 包含 `workspace_id`（目标业务仓库）

### 场景 F：工蜂 Issue 前置

使用 `issue-feasibility-analysis` / `issue-batch-analysis` 时需要：

- [ ] MCP: gongfeng — 工蜂 MCP 已配置且可连接
- [ ] CLI: `git` — 分支 / worktree（batch 场景）
- [ ] Skill: 对应 `SKILL.md` 存在

### 场景 G：Harness 规范生成与巡检

使用 `harness-engineering` / `harness-generating` / `harness-gardening` 时需要：

- [ ] Skill: `harness-engineering` — `{SKILL_ROOT}/harness-engineering/SKILL.md` 存在
- [ ] 本文件 `tool-dependencies.md` — 作为环境检查唯一数据源（生成类会逐项探测）
- [ ] Skill: `graphify`（可选）— 知识图谱生成（检查 `$SKILL_ROOT/graphify/SKILL.md` 是否存在）；不存在时 dev-map 功能跳过

---

## 五、环境检查脚本参考

以下命令可批量验证当前环境的工具就绪状态。

```bash
# 自动识别项目类型（按优先级依次检测）
if [ -f "skills/harness-engineering/SKILL.md" ]; then
  SKILL_ROOT="skills"
elif [ -f ".codebuddy/skills/harness-engineering/SKILL.md" ]; then
  SKILL_ROOT=".codebuddy/skills"
elif [ -f ".claude/skills/harness-engineering/SKILL.md" ]; then
  SKILL_ROOT=".claude/skills"
elif [ -f ".cursor/skills/harness-engineering/SKILL.md" ]; then
  SKILL_ROOT=".cursor/skills"
else
  echo "[WARN] 无法自动识别项目类型，请手动设置 SKILL_ROOT"
  SKILL_ROOT="skills"
fi
echo "项目类型检测: SKILL_ROOT=${SKILL_ROOT}"

echo ""
echo "=== 通用 CLI 检查 ==="
for cmd in git bash jq; do
  if command -v "$cmd" &>/dev/null; then
    echo "[OK] $cmd: $(command -v $cmd)"
  else
    echo "[MISS] $cmd: 未安装"
  fi
done

echo ""
echo "=== Go 项目 CLI 检查（仅 go.mod 存在时适用）==="
if [ -f "go.mod" ]; then
  for cmd in go; do
    if command -v "$cmd" &>/dev/null; then
      echo "[OK] $cmd: $(command -v $cmd)"
    else
      echo "[MISS] $cmd: 未安装"
    fi
  done

  echo ""
  echo "=== go-micro 工具链检查（仅 go.mod 含 go-micro 依赖时适用）==="
  if grep -q "go-micro" go.mod 2>/dev/null; then
    for cmd in protoc protoc-gen-go protoc-gen-micro protoc-gen-grpc-gateway protoc-gen-openapiv2 make; do
      if command -v "$cmd" &>/dev/null; then
        echo "[OK] $cmd"
      else
        echo "[MISS] $cmd: 未安装"
      fi
    done
  else
    echo "[SKIP] 非 go-micro 项目，跳过工具链检查"
  fi
else
  echo "[SKIP] 非 Go 项目，跳过 Go CLI 检查"
fi

echo ""
echo "=== Node 项目 CLI 检查（仅 package.json 存在时适用）==="
if [ -f "package.json" ]; then
  for cmd in node; do
    if command -v "$cmd" &>/dev/null; then
      echo "[OK] $cmd: $(command -v $cmd)"
    else
      echo "[MISS] $cmd: 未安装"
    fi
  done
else
  echo "[SKIP] 非 Node 项目，跳过 Node CLI 检查"
fi
```

---

## 六、扫描更新流程

> **适用范围：本节仅适用于开发仓（ai-practice）的维护。**
> 接入仓通过 `harness-generating` 将扫描结果写入项目自身的 `docs/harness/tooling.md`，**不修改本文件**。

当开发仓新增/删除了 Skill 或 Agent，需要更新本文件时，按以下流程执行：

**触发条件**（任一满足时执行）：
- 用户显式要求："重新扫描工具依赖"、"更新 tool-dependencies"、"rescan tools"
- 开发仓新增了 Skill（`skills/`）或 Agent（`agents/`）定义文件
- 发现环境检查结果与预期不符（某 Skill 实际需要的工具未列出）

**扫描步骤**：

1. **扫描 Agent 定义**
   - 读取 `agents/*.md`（项目根，开发工作流 agent 入口）
   - 从 frontmatter 提取 `mcpServers`、`tools` 字段
   - 从正文提取 MCP 调用、CLI 命令

2. **扫描 Skill 定义**
   - 遍历 `skills/*/SKILL.md`
   - 从 `metadata.requires.mcps` 提取 MCP 依赖
   - 从 `metadata.requires.bins` 提取 CLI 依赖
   - 从正文中匹配 MCP 调用模式（`MCP`、`TAPD MCP`、`mcp__` 等平台关键词）
   - 从正文中匹配 CLI 调用模式（`git`、`go`、`protoc`、`make` 等）

3. **扫描 Hooks**
   - 读取 `.agents/hooks.yaml`（若存在）
   - 提取 hook 命令依赖

4. **合并更新**
   - 将扫描结果与本文件现有内容对比
   - 新增条目追加到对应章节
   - 已删除的 Skill 对应条目标注为废弃或移除
   - 更新 §四 场景分组检查清单
   - 更新 §五 环境检查脚本

**扫描产物示例**（中间格式，用于更新本文件）：

```yaml
- source: skills/tapd-story-pipeline/SKILL.md
  name: tapd-story-pipeline
  dependencies:
    - tool: tapd
      type: MCP
      interfaces: [stories_get]
      required: true
    - tool: git
      type: CLI
      required: true
```

---

## 维护说明

- **本文件仅在开发仓中维护**：接入仓部署本 skill 后不得修改本文件；接入仓的工具扫描结果由 `harness-generating` 写入项目自身的 `docs/harness/tooling.md`
- **本文件是 harness-generating 环境检查的唯一数据源**，不再每次运行时扫描项目目录
- 新增 Skill 时，在对应章节和场景分组中补充工具依赖（或触发 §六 扫描更新流程）
- 新增 MCP 依赖时，在「一、MCP 服务依赖」中添加条目
- 新增 CLI 工具时，同步更新「五、环境检查脚本参考」
- `harness-generating` 子 skill 在第一步直接读取本文件做环境检查，输出「环境工具缺口」报告
- 保持本文件与项目实际 Skill 同步是团队的日常维护职责
- **独立 skill**（如运维巡检类）若不在本清单维护范围内，由用户或 skill 自身前置条件说明，不强行写入本文件
