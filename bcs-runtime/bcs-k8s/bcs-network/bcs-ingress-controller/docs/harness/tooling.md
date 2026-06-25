# 工具能力（Tooling）

> 目标：封装标准化工具接口，保障 Agent 执行稳定性。权威清单：`.cursor/skills/harness-engineering/references/tool-dependencies.md`

## 1. 工具清单

### 1.0 Skill 清单与触发

> 仅列出 `tool-dependencies.md` 中登记的 Skill。扫描路径：`.cursor/skills/*/SKILL.md`（顶层）。

| Skill | 触发词（摘要） | 功能概要 | 环境状态 |
|-------|--------------|---------|---------|
| harness-engineering | Harness Engineering、驾驭工程、文档巡检 | Harness 规范生成与巡检编排器 | ✅ 已就绪 |
| harness-generating | 生成 Harness 规范、开发地图 | 规范文档生成子 skill | ✅ 已就绪 |
| harness-gardening | 文档园艺、文档巡检、扫描文档 | 八维度文档一致性巡检 | ✅ 已就绪 |
| tapd-story-clarification | 需求澄清、clarify story | TAPD 需求规范化与回写 | ✅ 已就绪 |
| tapd-story-evaluation | 需求评估、RICE 评分、需求拆分 | 需求评估与子需求创建 | ✅ 已就绪 |
| tapd-iteration-plan | 迭代规划、排迭代 | TAPD 迭代编排 | ✅ 已就绪 |
| tapd-iteration-runner | 迭代执行、开发迭代、批量需求实现 | 迭代批量开发调度器 | ✅ 已就绪 |
| tapd-story-pipeline | 需求实现、开发需求、story pipeline | 单需求 TDD 流水线 | ✅ 已就绪 |
| work-summary | 总结工作内容、统计 AI 工时 | 工作汇总与 TAPD 同步 | ✅ 已就绪 |
| code-review | 代码评审 | Google Code Review 指南评审 | ✅ 已就绪 |
| bk-security-redlines | 安全红线检查 | 蓝鲸代码安全三大红线 | ✅ 已就绪 |
| go-micro-service | go-micro、grpc 服务开发 | go-micro 微服务开发指南 | ✅ 已就绪（非本项目主栈） |
| micro-service-project-init | 微服务项目初始化 | go-micro 项目脚手架生成 | ✅ 已就绪（非本项目主栈） |
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

**Agent 定义（接入仓 `.cursor/agents/`，非 Skill 目录）：**

| Agent | 路径 | 依赖方 | 环境状态 |
|-------|------|--------|---------|
| tapd-story-agent | `.cursor/agents/tapd-story-agent.md` | tapd-iteration-runner | ✅ 已就绪 |
| speckit-executor-agent | `.cursor/agents/speckit-executor-agent.md` | tapd-story-pipeline | ✅ 已就绪 |

### 1.1 MCP 工具

> 仅列出 `tool-dependencies.md` §一 登记的 MCP。

| MCP 名称 | 所需接口 | 必需 | 环境状态 |
|---------|---------|------|---------|
| tapd | stories_get, stories_update, iterations_get 等 | 是（TAPD 流水线） | ✅ 已就绪（user-tapd，workspace 70046748 已验证） |

**已配置但未列入权威清单的 MCP（不纳入 Harness 规范）：**

| MCP | 用途 |
|-----|------|
| user-codegraph | 代码符号搜索与调用链分析 |
| user-gongfeng | 工蜂 Git/MR 操作 |
| user-bcs-api-gateway-mcp-cluster | BCS 集群操作 |
| user-bcs-api-gateway-mcp-resource | BCS 资源操作 |
| 内部文档 MCP | 内部文档操作 |
| cursor-ide-browser | 浏览器自动化 |

### 1.2 CLI 工具

| 工具 | 必需 | 检测条件 | 环境状态 |
|------|------|---------|---------|
| `git` | 是 | 始终 | ✅ v2.43.7 |
| `bash` | 是 | 始终 | ✅ 5.2.15 |
| `jq` | 是 | 始终（迭代报告） | ✅ 1.8.1 |
| `go` | 是 | go.mod 存在（`../go.mod`） | ✅ go@1.23 |

**go-micro 工具链（检测条件：go.mod 含 go-micro 直接依赖）：** 跳过——本项目主栈为 K8s Operator，go-micro 仅为间接依赖。

**可选工具（按需安装，不主动扫描）：** `docker`、`gh`、`python3`、`kubectl`

### 1.3 配置文件

| 文件路径 | 必需 | 环境状态 |
|---------|------|---------|
| `project.json`（含 workspace_id、owner） | TAPD 流水线必需 | ✅ workspace_id=70046748，owner=adelaidahe |
| `.cursor/skills/work-summary/meta.json` | work-summary 必需 | ✅ 已就绪 |
| `.cursor/skills/work-summary/references/user-config.json` | work-summary 必需 | ✅ 已就绪 |
| `.specify/` 目录 | Spec Kit 必需 | ✅ 已就绪 |

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
| force push main/master | 严格禁止 |

## 4. 按场景的环境就绪状态

### 场景 A：TAPD 迭代研发流水线

- [x] MCP: tapd — 已接入（user-tapd，stories_get 验证通过）
- [x] Agent: tapd-story-agent — `.cursor/agents/tapd-story-agent.md`
- [x] Agent: speckit-executor-agent — `.cursor/agents/speckit-executor-agent.md`
- [x] Skill: speckit-* 系列 — 已就绪
- [x] CLI: git, bash, jq — 已就绪
- [x] 配置: project.json — workspace_id=70046748，owner=adelaidahe

### 场景 D：代码评审与安全检查

- [x] CLI: git — 已就绪
- [x] Skill: code-review — 已就绪
- [x] Skill: bk-security-redlines — 已就绪

### 场景 F：Harness 规范生成与巡检

- [x] Skill: harness-engineering — 已就绪
- [x] tool-dependencies.md — 已作为数据源使用

## 检查清单

- [x] Skill 清单已与 tool-dependencies 交叉验证
- [x] MCP 清单仅含权威登记条目
- [x] CLI 环境状态已探测
- [x] 环境缺口已记录
- [x] project.json 已配置（workspace_id + owner）
