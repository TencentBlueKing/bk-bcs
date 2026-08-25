# 技术规范

> Agent 实现需求时的开发行为准则。根据需求涉及的端按需加载对应规范。

## 必选规范（横切）

| 分类 | 规范 | 文档 | 加载 |
|------|------|------|------|
| 安全 | 蓝鲸代码安全三大红线 | [security-bk-redlines.md](security-bk-redlines.md) | 按「加载预算」读**相关红线节**（非每次全文） |
| 质量 | 代码评审规范 | [quality-code-review.md](quality-code-review.md) | **按需**：仅 Code Review / 质量评分任务（见加载预算）；日常改代码不预加载 |

> 安全为横切红线；质量长文 **opt-in（按需）**，避免默认灌入。文件仍部署到本目录供按需 Read。

## 当前项目选用的规范

| 分类 | 规范 | 文档 | 技术栈 | 项目事实（命中根） |
|------|------|------|--------|-------------------|
| 安全 | 蓝鲸代码安全三大红线 | [security-bk-redlines.md](security-bk-redlines.md) | security | 横切（code-project 必选） |
| 前端 | Vue 2 + webpack | [frontend-vue2.md](frontend-vue2.md) | vue, webpack, options-api | `bcs-ui/frontend` |
| 接口 | Protobuf + gRPC-Gateway | [api-grpc-gateway.md](api-grpc-gateway.md) | protobuf, grpc, grpc-gateway | `bcs-common`（primary）；同仓另有 Gin+Swagger 候选 |
| 后端 | Go + Gin REST | [backend-gin.md](backend-gin.md) | go, gin, rest | `bcs-common`（primary）；大量微服务为 go-micro，改这些服务时以该服务 `go.mod` 与局部 AGENTS 为准，勿硬套 Gin 分层 |

## Agent 加载步骤（强制）

1. 解析本文件「当前项目选用的规范」表，得到与本任务相关的规范**文件路径**。
2. 查「加载预算」：确定本任务要读哪些**章节**（不是整份文件）。
3. 用「章节快速索引」定位后，**按节 Read**（不要凭记忆复述；**禁止**默认全文灌入）。
4. 实现过程中若偏离规范，先说明冲突再改代码或请用户裁断。
5. IDE Rules（`standards-*.mdc` / `standards-*.md`）仅督促按节加载；完整条款以 `docs/standards/` 为准。

## 加载预算

| 任务类型 | 应 Read | 默认不预加载 |
|---------|---------|--------------|
| 文案/样式/单行无逻辑改动 | 无，或前端「陷阱」相关小节 | 安全全文、质量评分、后端/接口全文 |
| 改前端组件/状态/路由/请求 | 前端规范中对应章节（见索引） | 前端全文、质量评分量表、其它端全文 |
| 改接口契约/鉴权/联调 | 安全相关红线节 + 接口规范相关分册 | 前端全文、质量全文 |
| 改后端业务逻辑 | 后端规范相关节 + 安全相关红线节；go-micro 服务改读该服务代码与局部 AGENTS | 前端全文、质量评分量表 |
| 新建敏感接口/鉴权/加密逻辑 | 安全规范对应红线整节 + 检查清单 | 无关前端/质量长文 |
| Code Review / 质量评分 | 质量入口决策表 + 相关**分册**（`./quality-code-review/`） | 日常改代码、质量评分量表（非评分任务）、前端最佳实践全书 |
| 命中业务 tags/scenarios | `docs/business-standards/` 匹配条目 | 全量 business-standards |

## Agent 加载策略

| 需求类型 | 应加载的规范 |
|---------|------------|
| 任何涉及业务代码的需求 | 按「加载预算」读已选用安全规范的**相关节**（若表中有）；非每次全文 |
| 涉及前端页面 | 按节加载 `frontend-vue2.md` |
| 涉及接口定义/联调 | 按节加载 `api-grpc-gateway.md` 对应分册 |
| 涉及后端逻辑 | 按节加载 `backend-gin.md`；若目标服务为 go-micro / Operator，以局部 AGENTS 与该服务布局为准 |
| 命中业务领域/标签 | 按 tags/scenarios 选择性加载 `docs/business-standards/` |

## 规范约束力

- 标注"禁止"/"必须"的条目：**强制**遵守，违反需明确说明原因
- 标注"推荐"/"优先"的条目：**优先**遵守，有合理理由可偏离
- 常见场景参考：**参考**实现，可根据具体情况调整

## 章节快速索引

### 安全 [security-bk-redlines.md](security-bk-redlines.md)

- 本次必读决策表
- 一、红线总览
- 二、红线 1：外部输入未校验
- 三、红线 2：敏感接口未鉴权
- 四、红线 3：敏感数据未加密
- 五、代码评审检查清单
- 规范落地优先级

### 质量 [quality-code-review.md](quality-code-review.md)（按需）

- 入口：加载说明、本次必读决策表、分册目录
- [01-核心原则.md](./quality-code-review/01-核心原则.md)
- [02-问题分级.md](./quality-code-review/02-问题分级.md)
- [03-检查维度与规则清单.md](./quality-code-review/03-检查维度与规则清单.md)
- [04-代码质量评分标准.md](./quality-code-review/04-代码质量评分标准.md)
- [05-评审意见撰写规范.md](./quality-code-review/05-评审意见撰写规范.md)
- [06-评审报告格式.md](./quality-code-review/06-评审报告格式.md)

### 前端 [frontend-vue2.md](frontend-vue2.md)

- 本次必读决策表
- 一、技术栈要求
- 二、项目结构
- 三、编码规范（Options API）
- 四、状态管理（Vuex）
- 五、网络请求规范
- 六、路由（vue-router 3）
- 七、构建与工程（webpack / Vue CLI）
- 八、UI 组件使用规范
- 九、质量保证
- 十、常见陷阱与避坑
- 规范落地优先级

### 接口 [api-grpc-gateway.md](api-grpc-gateway.md)

- [01-架构概述.md](./api-grpc-gateway/01-架构概述.md)
- [02-接口定义规范-proto.md](./api-grpc-gateway/02-接口定义规范-proto.md)
- [03-数据类型映射.md](./api-grpc-gateway/03-数据类型映射.md)
- [04-响应格式规范.md](./api-grpc-gateway/04-响应格式规范.md)
- [05-前端接入规范.md](./api-grpc-gateway/05-前端接入规范.md)
- [06-核心契约模式.md](./api-grpc-gateway/06-核心契约模式.md)
- [07-联调标准流程.md](./api-grpc-gateway/07-联调标准流程.md)
- [08-api-网关配置规范.md](./api-grpc-gateway/08-api-网关配置规范.md)
- [09-常见陷阱与解决方案.md](./api-grpc-gateway/09-常见陷阱与解决方案.md)
- [10-附录.md](./api-grpc-gateway/10-附录.md)

### 后端 [backend-gin.md](backend-gin.md)

- 本次必读决策表
- 一、技术栈要求
- 二、项目结构
- 三、编码规范
- 四、错误处理与日志
- 五、配置、超时与安全
- 六、测试（可选）
- 七、构建与质量门禁
- 八、常见陷阱
- 规范落地优先级
