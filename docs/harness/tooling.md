# 工具能力（Tooling）

> 目标：让 Agent "能行动"——封装标准化工具接口，保障执行稳定性。

本仓库当前 Skill 安装根：`.agents/skills`（探测结果）。

## 1. 工具清单

### 1.0 Skill 清单与触发（Harness 基线）

| Skill | 触发词 | 功能概要 |
|-------|--------|---------|
| harness-engineering | 生成 Harness 规范、文档园艺、文档巡检、生成开发地图 | 生成/巡检 Agent 运行环境规范 |
| tapd-product-discovery | 产品前置、PRD、想法建单、角色拆单 | 完整 PRD 之前的产品调研与拆单 |
| tapd-story-clarification | 需求澄清、澄清需求、完善需求描述 | 拉取规划中需求并回写规范化文档 |
| tapd-story-evaluation | 需求评估、需求拆分、size 评分、RICE | 拆分子需求并评分 |
| tapd-story-review | 需求评审、TAPD 评论处理 | 评审澄清/拆单结果并闭环评论 |
| tapd-iteration-plan | 迭代规划、排迭代、需求入迭代 | 将 approved 需求编入迭代 |
| tapd-iteration-runner | 迭代执行、批量需求实现、开发迭代 | 批量调度迭代内需求实现 |
| tapd-story-pipeline | 需求实现、开发需求、hotfix | 单需求从澄清到提交的流水线 |
| code-review | 代码评审、Code Review | 按 Google Code Review 指南评审 |
| bk-security-redlines | 安全红线、输入校验、鉴权、加密 | 蓝鲸三大安全红线检查 |

### 1.1 MCP 工具（Harness 基线）

| MCP 名称 | 所需接口 | 必需 | 检测方式 |
|---------|---------|------|---------|
| tapd | `get_stories_or_tasks` / `stories_get` 等价只读；规划/回写另需 update/create | 是 | 会话内 probe 一条需求；或 `harness-doctor` |
| gongfeng | `get_current_user` 等只读 | 条件（关联工蜂 Issue/MR 时） | 会话内 probe |
| bkm-bkte | metrics / logs 等只读 | 否 | 会话内对各 `bkm-bkte-*` 只读探测 |

### 1.2 CLI 工具（Harness 基线）

| 工具 | 必需 | 检测条件 | 检测命令 |
|------|------|---------|---------|
| `git` | 是 | 始终 | `command -v git` |
| `bash` | 是 | 始终 | `command -v bash` |
| `jq` | 是 | Linux/macOS 迭代报告 | `command -v jq` |
| `go` | 是 | 存在 `go.mod` | `go version` |
| `graphify` | 否 | 需生成/更新 `docs/dev-map/` 时 | `$SKILL_ROOT/graphify/SKILL.md` 与 `graphify --version` |

### 1.3 配置

| 文件 | 必需 | 说明 |
|------|------|------|
| `project.json` | 使用 TAPD 流水线时 | 含 `workspace_id`、可选 `owner`；本仓 gitignore，需本地创建 |
| `.specify/` | 使用 speckit TDD 时 | 本仓尚未初始化 |
| `.agents/commands/speckit.*.md` | 使用 speckit 命令入口时 | 本仓尚未安装 |

### 1.4 项目自有工具

| 名称 | 用途 |
|------|------|
| openspec-new-change | bcs-drplan-controller — 用 OpenSpec 实验工作流启动新变更 |
| openspec-continue-change | bcs-drplan-controller — 继续产出下一份 OpenSpec 工件 |
| openspec-apply-change | bcs-drplan-controller — 按 OpenSpec 变更实现任务 |
| openspec-verify-change | bcs-drplan-controller — 校验实现与变更工件一致 |
| openspec-archive-change | bcs-drplan-controller — 完成后归档 OpenSpec 变更 |
| openspec-bulk-archive-change | bcs-drplan-controller — 批量归档已完成变更 |
| openspec-ff-change | bcs-drplan-controller — 快速生成实现所需全部工件 |
| openspec-explore | bcs-drplan-controller — 变更前/中的探索与澄清 |
| openspec-onboard | bcs-drplan-controller — OpenSpec 工作流引导 |
| openspec-sync-specs | bcs-drplan-controller — 将 delta spec 同步到主 spec |
| api-standard | bcs-project-manager — 蓝鲸双协议兼容的 Axios 封装 |
| bk-monitor-dev-server | bcs-project-manager — 配置和启动本地开发服务器 |
| bk-monitor-security-audit | bcs-project-manager — 前端 XSS/CSRF 等安全审计 |
| bk-monitor-tapd-dev | bcs-project-manager — 根据 TAPD 单据分析开发方案 |
| bk-monitor-tapd-summary | bcs-project-manager — TAPD 待办归纳 |
| bk-monitor-weekly-report | bcs-project-manager — 基于 PR 生成周报 |
| bk-skill-creator | bcs-project-manager — 编写渐进式披露 Skill |
| bkui-builder | bcs-project-manager — 设计稿还原（先布局分析再套模版） |
| bkui-cheatsheet | bcs-project-manager — 蓝鲸前端高频属性避坑 |
| bkui-quick-start | bcs-project-manager — 蓝鲸前端知识库入口 |
| bundle-optimization | bcs-project-manager — 拆包与 Tree Shaking 优化首屏 |
| chat-x-contribute | bcs-project-manager — chat-x 组件库开发上手 |
| chat-x-custom-component | bcs-project-manager — chat-x 自定义 message 组件 |
| chat-x-unit-test | bcs-project-manager — Vue 组件单元测试 |
| js-security-check | bcs-project-manager — 前端 XSS/CSRF/原型污染检查 |
| nodejs-security-check | bcs-project-manager — Node RCE/SSRF/注入等检查 |
| permission-directive | bcs-project-manager — 蓝鲸 IAM 前端鉴权指令 |
| pinia-setup | bcs-project-manager — Pinia 全局状态规范 |
| unit-testing | bcs-project-manager — Vitest + Vue Test Utils 模版 |
| virtual-list | bcs-project-manager — 大数据列表虚拟滚动 |
| vite-migration | bcs-project-manager — Webpack 迁移 Vite 步骤 |
| vue-best-practices | bcs-project-manager — Vue 3 + TS + Vite 实践（仅该子树；仓级控制台仍为 Vue 2） |
| vue-composables | bcs-project-manager — useTable/useRequest 等 Hooks |
| web-security-guide | bcs-project-manager — OWASP 十大原理与修复 |

## 2. 工具接口规范

### 2.1 统一调用协议

所有自定义工具应遵循以下接口约定：

- **输入**：结构化参数（JSON 格式），包含必填和可选字段
- **输出**：结构化结果，包含成功数据或明确错误
- **错误处理**：返回可读错误信息，禁止把密钥写入日志

### 2.2 接口定义示例

```json
{
  "tool_name": "example_tool",
  "description": "工具功能描述",
  "parameters": {
    "required": ["param1"],
    "optional": ["param2"],
    "schema": {
      "param1": { "type": "string", "description": "参数说明" },
      "param2": { "type": "number", "default": 10, "description": "参数说明" }
    }
  },
  "returns": {
    "success": { "type": "boolean" },
    "data": { "type": "object" },
    "error": { "type": "string", "nullable": true }
  }
}
```

## 3. 稳定性保障

### 3.1 沙盒执行

| 执行环境 | 隔离方式 | 适用场景 |
|---------|---------|---------|
| 工作区 Shell | 仅仓库内路径；禁止扫家目录密钥 | 日常命令、测试、构建 |
| 集群命令 | 必须使用用户指定的 kubecontext；禁止默认打生产 | Operator 调试 |

### 3.2 容错策略

| 策略 | 配置 | 适用场景 |
|------|------|---------|
| 超时 | 120s（可按命令调整） | 外部 MCP / 网络请求 |
| 重试 | 最多 2 次，指数退避 | 瞬时网络失败 |
| 幂等 | 相同参数多次调用结果一致 | TAPD 回写、git 提交须先检查状态 |
| 降级 | 工具不可用时停止并报告缺口 | 缺 TAPD MCP 时不假装已回写 |

### 3.3 敏感操作防护

| 操作类型 | 防护措施 |
|---------|---------|
| 删除文件/目录 | 二次确认；禁止删用户未要求的路径 |
| 访问生产集群 | 严格禁止，除非用户明确授权 context |
| `git push --force` / 改 git config | 禁止 |
| 提交密钥、证书、真实用户信息 | 禁止；示例用占位符（ingress-controller 已强调） |

## 4. 工具扩展规范

### 4.1 新工具接入流程

1. 将 Skill 放入安装布局并 **git 跟踪** 后才会进入「项目自有工具」
2. 编写 `SKILL.md`（触发词、description、边界）
3. 在沙盒中验证
4. 更新本节表格（基线工具不改治理仓清单，只改本文件项目自有节）

### 4.2 扩展原则

- 新工具不应修改核心 Harness 逻辑
- 一个工具只做一件事
- 工具文档与工具代码同仓库版本控制

## 检查清单

- [x] Harness 基线表与权威清单场景 A/B/D/G 对齐
- [x] 项目自有节为 `名称 | 用途`
- [x] 契约表不含运行时探活列或探活单元格
- [x] 敏感操作有防护措施
