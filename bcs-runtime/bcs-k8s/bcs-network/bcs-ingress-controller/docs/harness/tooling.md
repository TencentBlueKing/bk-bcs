# 工具能力（Tooling）

> 目标：封装标准化工具接口，保障 Agent 执行稳定性。仅列出 Harness 基线契约；环境是否就绪由 `harness-doctor` 探测，不写入本文件。

## 1. 工具清单

### 1.0 Skill 清单与触发（Harness 基线）

> 仅列出已在 `tool-dependencies.md` 登记的 Skill。扫描路径：`{SKILL_ROOT}/*/SKILL.md`（顶层；嵌套子 skill 触发词合并到父 skill）。当前 Cursor 会话 `$SKILL_ROOT=.cursor/skills`。

| Skill | 触发词（摘要） | 功能概要 |
|-------|--------------|---------|
| harness-engineering | Harness Engineering、驾驭工程、文档巡检、生成 Harness 规范 | Harness 规范生成与十二维度巡检编排器（含 harness-generating / harness-gardening） |
| tapd-story-clarification | 需求澄清、clarify story | TAPD 需求规范化与回写 |
| tapd-story-evaluation | 需求评估、RICE 评分、需求拆分 | 需求评估与子需求创建 |
| tapd-story-review | 需求评审、review 需求 | TAPD 需求评审与评论汇总 |
| tapd-story-govern-pipeline | 需求整理流水线、refinement pipeline | 澄清→评审→评估前置编排 |
| tapd-product-discovery | 产品前置、产品调研、PRD | 产品父单/角色拆单与调研 |
| tapd-bug-clarification | 缺陷澄清、bug clarify | TAPD 缺陷根因澄清与回写 |
| tapd-bug-evaluation | 缺陷评估、bug evaluate | TAPD 缺陷工时/规模评估 |
| tapd-iteration-plan | 迭代规划、排迭代 | TAPD 迭代编排 |
| tapd-iteration-runner | 迭代执行、开发迭代、批量需求实现 | 迭代批量开发调度器 |
| tapd-story-pipeline | 需求实现、开发需求、story pipeline | 单需求 TDD 流水线 |
| graph-engineering | graph-engineering、按流程推进 | 可恢复多阶段流水线编排 |
| story-specify | story specify（由 graph-engineering 加载） | graph-engineering 技术澄清 worker |
| issue-feasibility-analysis | Issue 可行性分析 | 工蜂 Issue 可行性分析 |
| issue-batch-analysis | Issue 批量分析 | 多 Issue 批量前置与并行交接 |
| code-review | 代码评审 | Google Code Review 指南评审 |
| bk-security-redlines | 安全红线检查 | 蓝鲸代码安全三大红线 |
| sre-engineer | SRE、可观测性、上线准备 | SRE 上线前/后运营（bkm-bkte） |
| go-micro-service | go-micro、grpc 服务开发 | go-micro 微服务开发指南（非本项目主栈） |
| micro-service-project-init | 微服务项目初始化 | go-micro 项目脚手架生成（非本项目主栈） |
| graphify | 代码图谱、graphify query | 知识图谱生成与查询（dev-map） |
| speckit-specify | speckit specify | 功能规格生成 |
| speckit-plan | speckit plan | 实现计划生成 |
| speckit-tasks | speckit tasks | 任务拆分 |
| speckit-implement | speckit implement | TDD 实现 |
| speckit-analyze | speckit analyze | 产物一致性分析 |
| speckit-checklist | speckit checklist | 检查清单生成 |
| speckit-clarify | speckit clarify | 规格澄清 |
| speckit-constitution | speckit constitution | 项目宪法 |
| speckit-git-commit | speckit git commit | 自动提交 |
| speckit-git-feature | speckit git feature | 功能分支创建 |
| speckit-git-initialize | speckit git initialize | Git 初始化 |
| speckit-git-remote | speckit git remote | 远程检测 |
| speckit-git-validate | speckit git validate | 分支命名校验 |
| speckit-taskstoissues | speckit tasks to issues | 任务转 Issue |

**接入仓额外安装（未在 tool-dependencies 登记，不参与一致性强制检查）：**

| Skill | 路径 | 说明 |
|-------|------|------|
| work-summary | `.cursor/skills/work-summary/` | 工作汇总与 TAPD 同步 |
| bcs-cluster-checklist 等 | `{SKILL_ROOT}/` | 运维巡检类，按需使用 |

### 1.1 Agent 清单

> 定义目录：`.cursor/agents/`（CodeBuddy 侧同步于 `.codebuddy/agents/`）。

| Agent | 定位 | 说明 |
|-------|------|------|
| `ai-engineer` | AI 工程开发 | Skill/Agent/Eval 驱动交付（定义：`.cursor/agents/ai-engineer.md`） |
| `backend-developer` | 后端开发 | API-First / TDD 后端交付（定义：`.cursor/agents/backend-developer.md`） |
| `business-analyst` | 商业分析 | 产品价值与服务梳理（定义：`.cursor/agents/business-analyst.md`） |
| `code-reviewer` | 代码评审 | 质量/安全/可维护性评审（定义：`.cursor/agents/code-reviewer.md`） |
| `flow-steward` | Graph Engineering 入口 | 单需求 flow 编排与裁决（定义：`.cursor/agents/flow-steward.md`） |
| `frontend-developer` | 前端开发 | 组件驱动前端交付（定义：`.cursor/agents/frontend-developer.md`） |
| `graph-steward` | Graph Engineering 调度 | graph-engineering 正式入口（定义：`.cursor/agents/graph-steward.md`） |
| `product-manager` | 产品经理 | PRD / 路线图（定义：`.cursor/agents/product-manager.md`） |
| `qa-engineer` | QA | 对抗性测试与质量报告（定义：`.cursor/agents/qa-engineer.md`） |
| `speckit-execution-agent` | Spec Kit 执行器 | 隔离执行 speckit-* skill（定义：`.cursor/agents/speckit-execution-agent.md`） |
| `sre-engineer` | SRE | 上线前准备与运营（定义：`.cursor/agents/sre-engineer.md`） |
| `tapd-story-agent` | 单需求流水线 | 加载 tapd-story-pipeline 状态机（定义：`.cursor/agents/tapd-story-agent.md`） |
| `tech-lead` | 技术负责人 | 技术方案与任务拆解（定义：`.cursor/agents/tech-lead.md`） |
| `ux-designer` | UX 设计 | 场景与产品设计文档（定义：`.cursor/agents/ux-designer.md`） |

`workflow-agent` 已废弃：Harness 不再接入 workflow；可选编排请用 `flow-steward` / `graph-engineering`。`docs/workflow.md` 仅作人类可读参考，不强制执行。

### 1.2 MCP 工具（Harness 基线）

| MCP 名称 | 所需接口 | 必需 | 检测方式 |
|---------|---------|------|---------|
| tapd | stories_get, stories_update, iterations_get, bugs_* 等 | 是（TAPD 流水线） | 会话内 probe `stories_get` |
| gongfeng | Issue / MR / 提交查询 | 条件（工蜂 Issue） | 会话内 probe `get_current_user` |
| bkm-bkte | metrics / logs / dashboards 等 | 条件（sre-engineer） | 对各 bkm-bkte MCP 只读探测 |

**已配置但未列入权威清单的 MCP（不纳入 Harness 规范强制项）：**

| MCP | 用途 |
|-----|------|
| user-bcs-api-gateway-mcp-cluster | BCS 集群操作 |
| user-bcs-api-gateway-mcp-resource | BCS 资源操作 |
| user-iWiki | 内部文档 |
| cursor-ide-browser | 浏览器自动化 |

### 1.3 CLI 工具（Harness 基线）

| 工具 | 必需 | 检测条件 | 检测命令 |
|------|------|---------|---------|
| `git` | 是 | 始终 | `command -v git` |
| `bash` | 是 | 始终 | `command -v bash` |
| `jq` | 是 | 始终（迭代报告） | `command -v jq` |
| `go` | 是 | go.mod 存在（`../go.mod`） | `command -v go` |
| `graphify` | 否 | docs/dev-map 需更新时 | `$SKILL_ROOT/graphify/SKILL.md` 与 `graphify --version` |

**go-micro 工具链（检测条件：go.mod 含 go-micro 直接依赖）：** 跳过——本项目主栈为 K8s Operator，go-micro 仅为间接依赖。

**可选工具（按需安装，不主动扫描）：** `docker`、`gh`、`python3`、`kubectl`、`uv`

### 1.4 配置文件

| 文件路径 | 必需 | 说明 |
|---------|------|------|
| `project.json`（含 workspace_id、owner） | TAPD 流水线必需 | 需本地创建（已 gitignore，不入库） |
| `.specify/` 目录 | Spec Kit 必需 | Spec Kit 项目结构 |
| `{SKILL_ROOT}/work-summary/meta.json`（若使用 work-summary） | 条件 | `.cursor/skills/work-summary/` |

## 2. 工具接口规范

### 2.1 统一调用协议

- **输入**：结构化参数（JSON），区分必填和可选
- **输出**：`{success, data, error}` 结构
- **错误处理**：明确错误码 + 可读错误信息

### 2.2 Controller 开发工具约定

| 操作 | 命令 | 工作目录 |
|------|------|---------|
| 构建 | `cd .. && make ingress-controller` | bcs-network/ |
| 全量测试 | `cd .. && make test-ingress-controller` | bcs-network/ |
| 单包测试 | `go test -v -run TestXxx ./bcs-ingress-controller/{pkg}/...` | bcs-network/ |
| 格式化 | `gofmt` / `goimports` | — |
| 部署重启 | `kubectl rollout restart -n bcs-system deployment/bcsingresscontroller` | — |

## 3. 稳定性保障

### 3.1 沙盒执行

| 执行环境 | 隔离方式 | 适用场景 |
|---------|---------|---------|
| Shell 沙盒 | 文件系统 + 网络限制 | 日常命令执行 |
| K8s 集群 | RBAC 权限边界 | Controller 验证 |

### 3.2 容错策略

| 策略 | 配置 | 适用场景 |
|------|------|---------|
| 超时 | 30s（默认）/ 300s（构建） | 所有外部调用 |
| 重试 | 最多 3 次，指数退避 | 网络请求、K8s API |
| 幂等 | Reconcile 必须幂等 | Controller 写操作 |

### 3.3 敏感操作防护

| 操作类型 | 防护措施 |
|---------|---------|
| 删除文件/目录 | 二次确认 |
| 修改生产集群 | 需 KUBECONFIG 授权，禁止未确认操作 |
| 云凭证操作 | 禁止提交 Secret 内容到 git |
| 用户业务信息 | 禁止将用户信息（域名、证书名称/ID、账号、业务标识等）写入代码、注释或文档；示例与测试数据须使用占位符（如 `example.com`、`cert-xxx`） |
| force push main/master | 严格禁止 |

## 4. 按场景的环境就绪状态

> 勾选表示契约已登记；实际连通性请运行 harness-doctor，不要把本机探测结果写回本文件。

### 场景 A：TAPD 迭代研发流水线

- [x] MCP: tapd
- [x] Agent: tapd-story-agent、speckit-execution-agent
- [x] Skill: tapd-iteration-runner / tapd-story-pipeline / speckit-*
- [x] CLI: git, bash, jq
- [ ] 配置: project.json — 需本地创建（gitignore）

### 场景 B：需求前期处理

- [x] MCP: tapd
- [x] Skill: tapd-product-discovery / tapd-story-clarification / tapd-story-review / tapd-story-evaluation / tapd-story-govern-pipeline / tapd-iteration-plan
- [ ] 配置: project.json — 需本地创建

### 场景 D：代码评审与安全检查

- [x] CLI: git
- [x] Skill: code-review、bk-security-redlines
- [x] Agent: code-reviewer

### 场景 E：缺陷澄清与评估

- [x] MCP: tapd（bugs_*）
- [x] Skill: tapd-bug-clarification、tapd-bug-evaluation
- [ ] 配置: project.json — 需本地创建

### 场景 F：工蜂 Issue 前置

- [x] MCP: gongfeng
- [x] CLI: git
- [x] Skill: issue-feasibility-analysis、issue-batch-analysis

### 场景 G：Harness 规范生成与巡检

- [x] Skill: harness-engineering
- [x] Skill: graphify（可选，dev-map）

### Graph Engineering

- [x] Agent: flow-steward、graph-steward
- [x] Skill: graph-engineering、story-specify

## 检查清单

- [x] Skill 清单已与 tool-dependencies 交叉验证
- [x] Agent 清单已与 `.cursor/agents/` 对齐（workflow-agent 已标废弃）
- [x] MCP 清单含权威登记条目（tapd / gongfeng / bkm-bkte）
- [x] CLI 表含检测命令、不含环境状态列
