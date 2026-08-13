# 工具能力（Tooling）

> 目标：封装标准化工具接口，保障 Agent 执行稳定性。权威清单：`{SKILL_ROOT}/harness-engineering/references/tool-dependencies.md`（当前 `$SKILL_ROOT=.codebuddy/skills`）

## 1. 工具清单

### 1.0 Skill 清单与触发

> 仅列出 `tool-dependencies.md` 中登记的 Skill。扫描路径：`{SKILL_ROOT}/*/SKILL.md`（顶层；嵌套子 skill 触发词合并到父 skill）。

| Skill | 触发词（摘要） | 功能概要 | 环境状态 |
|-------|--------------|---------|---------|
| harness-engineering | Harness Engineering、驾驭工程、文档巡检、生成 Harness 规范 | Harness 规范生成与九维度巡检编排器（含 harness-generating / harness-gardening） | ✅ 已就绪 |
| tapd-story-clarification | 需求澄清、clarify story | TAPD 需求规范化与回写 | ✅ 已就绪 |
| tapd-story-evaluation | 需求评估、RICE 评分、需求拆分 | 需求评估与子需求创建 | ✅ 已就绪 |
| tapd-story-review | 需求评审、review 需求 | TAPD 需求评审与评论汇总 | ✅ 已就绪 |
| tapd-story-govern-pipeline | 需求整理流水线、refinement pipeline | 澄清→评审→评估前置编排 | ✅ 已就绪 |
| tapd-product-discovery | 产品前置、产品调研、PRD | 产品父单/角色拆单与调研 | ✅ 已就绪 |
| tapd-bug-clarification | 缺陷澄清、bug clarify | TAPD 缺陷根因澄清与回写 | ✅ 已就绪 |
| tapd-bug-evaluation | 缺陷评估、bug evaluate | TAPD 缺陷工时/规模评估 | ✅ 已就绪 |
| tapd-iteration-plan | 迭代规划、排迭代 | TAPD 迭代编排 | ✅ 已就绪 |
| tapd-iteration-runner | 迭代执行、开发迭代、批量需求实现 | 迭代批量开发调度器 | ✅ 已就绪 |
| tapd-story-pipeline | 需求实现、开发需求、story pipeline | 单需求 TDD 流水线 | ✅ 已就绪 |
| graph-engineering | graph-engineering、按流程推进 | 可恢复多阶段流水线编排 | ✅ 已就绪 |
| story-specify | story specify（由 graph-engineering 加载） | graph-engineering 技术澄清 worker | ✅ 已就绪 |
| issue-feasibility-analysis | Issue 可行性分析 | 工蜂 Issue 可行性分析 | ✅ 已就绪 |
| issue-batch-analysis | Issue 批量分析 | 多 Issue 批量前置与并行交接 | ✅ 已就绪 |
| code-review | 代码评审 | Google Code Review 指南评审 | ✅ 已就绪 |
| bk-security-redlines | 安全红线检查 | 蓝鲸代码安全三大红线 | ✅ 已就绪 |
| sre-engineer | SRE、可观测性、上线准备 | SRE 上线前/后运营（bkm-bkte） | ✅ 已就绪 |
| go-micro-service | go-micro、grpc 服务开发 | go-micro 微服务开发指南 | ✅ 已就绪（非本项目主栈） |
| micro-service-project-init | 微服务项目初始化 | go-micro 项目脚手架生成 | ✅ 已就绪（非本项目主栈） |
| graphify | 代码图谱、graphify query | 知识图谱生成与查询（dev-map） | ✅ 已就绪 |
| speckit-specify | speckit specify | 功能规格生成 | ✅ 已就绪 |
| speckit-plan | speckit plan | 实现计划生成 | ✅ 已就绪 |
| speckit-tasks | speckit tasks | 任务拆分 | ✅ 已就绪 |
| speckit-implement | speckit implement | TDD 实现 | ✅ 已就绪 |
| speckit-analyze | speckit analyze | 产物一致性分析 | ✅ 已就绪 |
| speckit-checklist | speckit checklist | 检查清单生成 | ✅ 已就绪 |
| speckit-clarify | speckit clarify | 规格澄清 | ✅ 已就绪 |
| speckit-constitution | speckit constitution | 项目宪法 | ✅ 已就绪 |
| speckit-git-commit | speckit git commit | 自动提交 | ✅ 已就绪 |
| speckit-git-feature | speckit git feature | 功能分支创建 | ✅ 已就绪 |
| speckit-git-initialize | speckit git initialize | Git 初始化 | ✅ 已就绪 |
| speckit-git-remote | speckit git remote | 远程检测 | ✅ 已就绪 |
| speckit-git-validate | speckit git validate | 分支命名校验 | ✅ 已就绪 |
| speckit-taskstoissues | speckit tasks to issues | 任务转 Issue | ✅ 已就绪 |

**接入仓额外安装（未在 tool-dependencies 登记，不参与一致性强制检查）：**

| Skill | 路径 | 说明 |
|-------|------|------|
| work-summary | `.cursor/skills/work-summary/` | 工作汇总与 TAPD 同步 |
| bcs-cluster-checklist 等 | `{SKILL_ROOT}/` | 运维巡检类，按需使用 |

### 1.1 Agent 清单

> 定义目录：`.codebuddy/agents/`（Cursor 侧同步于 `.cursor/agents/`）。

| Agent | 定位 | 说明 | 环境状态 |
|-------|------|------|---------|
| `ai-engineer` | AI 工程开发 | Skill/Agent/Eval 驱动交付（定义：`.codebuddy/agents/ai-engineer.md`） | ✅ 已就绪 |
| `backend-developer` | 后端开发 | API-First / TDD 后端交付（定义：`.codebuddy/agents/backend-developer.md`） | ✅ 已就绪 |
| `business-analyst` | 商业分析 | 产品价值与服务梳理（定义：`.codebuddy/agents/business-analyst.md`） | ✅ 已就绪 |
| `code-reviewer` | 代码评审 | 质量/安全/可维护性评审（定义：`.codebuddy/agents/code-reviewer.md`） | ✅ 已就绪 |
| `flow-steward` | Graph Engineering 入口 | 单需求 flow 编排与裁决（定义：`.codebuddy/agents/flow-steward.md`） | ✅ 已就绪 |
| `frontend-developer` | 前端开发 | 组件驱动前端交付（定义：`.codebuddy/agents/frontend-developer.md`） | ✅ 已就绪 |
| `product-manager` | 产品经理 | PRD / 路线图（定义：`.codebuddy/agents/product-manager.md`） | ✅ 已就绪 |
| `qa-engineer` | QA | 对抗性测试与质量报告（定义：`.codebuddy/agents/qa-engineer.md`） | ✅ 已就绪 |
| `speckit-execution-agent` | Spec Kit 执行器 | 隔离执行 speckit-* skill（定义：`.codebuddy/agents/speckit-execution-agent.md`） | ✅ 已就绪 |
| `sre-engineer` | SRE | 上线前准备与运营（定义：`.codebuddy/agents/sre-engineer.md`） | ✅ 已就绪 |
| `tapd-story-agent` | 单需求流水线 | 加载 tapd-story-pipeline 状态机（定义：`.codebuddy/agents/tapd-story-agent.md`） | ✅ 已就绪 |
| `tech-lead` | 技术负责人 | 技术方案与任务拆解（定义：`.codebuddy/agents/tech-lead.md`） | ✅ 已就绪 |
| `ux-designer` | UX 设计 | 场景与产品设计文档（定义：`.codebuddy/agents/ux-designer.md`） | ✅ 已就绪 |
| `workflow-agent` | 工作流执行 | 按 docs/workflow.md 驱动步骤（定义：`.codebuddy/agents/workflow-agent.md`） | ✅ 已就绪 |

### 1.2 MCP 工具

> 仅列出 `tool-dependencies.md` §一 登记的 MCP。

| MCP 名称 | 所需接口 | 必需 | 环境状态 |
|---------|---------|------|---------|
| tapd | stories_get, stories_update, iterations_get, bugs_* 等 | 是（TAPD 流水线） | ✅ 已就绪（user-tapd，workspace 70046748） |
| gongfeng | Issue / MR / 提交查询 | 条件（工蜂 Issue / workflow） | ✅ 已就绪（user-gongfeng） |
| bkm-bkte | metrics / logs / dashboards 等 | 条件（sre-engineer） | 待重新检测 |

**已配置但未列入权威清单的 MCP（不纳入 Harness 规范强制项）：**

| MCP | 用途 |
|-----|------|
| user-bcs-api-gateway-mcp-cluster | BCS 集群操作 |
| user-bcs-api-gateway-mcp-resource | BCS 资源操作 |
| user-iWiki | 内部文档 |
| cursor-ide-browser | 浏览器自动化 |

### 1.3 CLI 工具

| 工具 | 必需 | 检测条件 | 环境状态 |
|------|------|---------|---------|
| `git` | 是 | 始终 | ✅ 已就绪 |
| `bash` | 是 | 始终 | ✅ 已就绪 |
| `jq` | 是 | 始终（迭代报告） | ✅ 已就绪 |
| `go` | 是 | go.mod 存在（`../go.mod`） | ✅ 已就绪 |
| `graphify` | 否 | docs/dev-map 需更新时 | ✅ v0.9.20 |

**go-micro 工具链（检测条件：go.mod 含 go-micro 直接依赖）：** 跳过——本项目主栈为 K8s Operator，go-micro 仅为间接依赖。

**可选工具（按需安装，不主动扫描）：** `docker`、`gh`、`python3`、`kubectl`、`uv`

### 1.4 配置文件

| 文件路径 | 必需 | 环境状态 |
|---------|------|---------|
| `project.json`（含 workspace_id、owner） | TAPD 流水线必需 | ⚠️ 需本地创建（已 gitignore，不入库） |
| `.specify/` 目录 | Spec Kit 必需 | ✅ 已就绪 |
| `{SKILL_ROOT}/work-summary/meta.json`（若使用 work-summary） | 条件 | ✅ `.cursor/skills/work-summary/` |

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

### 场景 A：TAPD 迭代研发流水线

- [x] MCP: tapd — 已接入（user-tapd）
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
- [x] tool-dependencies.md — 已作为数据源使用

### Graph Engineering / 工作流

- [x] Agent: flow-steward、workflow-agent
- [x] Skill: graph-engineering、story-specify
- [x] 文档: docs/workflow.md

## 检查清单

- [x] Skill 清单已与 tool-dependencies 交叉验证
- [x] Agent 清单已与 `.codebuddy/agents/` 对齐
- [x] MCP 清单含权威登记条目（tapd / gongfeng / bkm-bkte）
- [x] CLI 环境状态已探测
- [x] 环境缺口已记录（project.json 本地配置、bkm-bkte 待检测）
