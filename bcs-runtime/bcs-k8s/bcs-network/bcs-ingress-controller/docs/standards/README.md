# 技术规范

> Agent 实现需求时的开发行为准则。根据需求涉及的端按需加载对应规范。

## 必选规范（横切）

| 分类 | 规范 | 文档 | 加载 |
|------|------|------|------|
| 安全 | 蓝鲸代码安全三大红线 | [security-bk-redlines.md](security-bk-redlines.md) | 按「加载预算」读**相关红线节**（非每次全文） |
| 质量 | 代码评审规范（Google Code Review 指南） | [quality-code-review.md](quality-code-review.md) | **按需**：仅 Code Review / 质量评分任务（见加载预算）；日常改代码不预加载 |

> 安全为横切红线；质量长文 **opt-in（按需）**，避免默认灌入。文件仍部署到本目录供按需 Read。

## 当前项目选用的规范

| 分类 | 规范 | 文档 | 技术栈 | 项目事实（命中根） |
|------|------|------|--------|-------------------|
| 后端 | K8s Operator 开发规范 | [backend-k8s-operator.md](backend-k8s-operator.md) | Go + controller-runtime v0.6.3 | 本组件（无对应预设，项目定制） |
| 接口 | go-restful HTTP API 规范 | [api-go-restful.md](api-go-restful.md) | go-restful | 本组件（无对应预设，项目定制） |

## Agent 加载步骤（强制）

1. 解析本文件「当前项目选用的规范」表，得到与本任务相关的规范**文件路径**。
2. 查「加载预算」：确定本任务要读哪些**章节**（不是整份文件）。
3. 用「章节快速索引」定位后，**按节 Read**（不要凭记忆复述；**禁止**默认全文灌入）。
4. 若存在 `docs/business-standards/README.md`：读「规范索引」，按任务 tags/scenarios **只 Read 命中文件**（无匹配不加载；未登记视为未生效）。
5. 实现过程中若偏离规范，先说明冲突再改代码或请用户裁断。
6. IDE Rules（`standards-*.mdc` / `standards-*.md`）仅督促按节加载；完整条款以 `docs/standards/` 为准。禁止为业务规范逐文件生成 Always Rule。

## 加载预算

| 任务类型 | 应 Read | 默认不预加载 |
|---------|---------|--------------|
| 文案/样式/单行无逻辑改动 | 无 | 安全全文、质量评分、后端/接口全文 |
| 改接口契约/鉴权/联调 | 安全相关红线节 + [api-go-restful.md](api-go-restful.md) 相关节 | 质量全文、后端全文 |
| 改后端业务逻辑 / Controller | [backend-k8s-operator.md](backend-k8s-operator.md) 相关节 + 安全相关红线节 | 质量评分量表、接口全文 |
| 新增/修改 HTTP API | 后端相关节 + 接口相关节 + 安全相关红线节 | 质量评分量表 |
| Webhook / 云适配器 | 后端相关节 + 安全相关红线节 + 相关 ADR | 质量全文 |
| 新建敏感接口/鉴权/加密逻辑 | 安全规范对应红线整节 + 检查清单 | 无关质量长文 |
| Code Review / 质量评分 | 质量入口决策表 + 相关**分册**（`./quality-code-review/`） | 日常改代码、质量评分量表（非评分任务） |

## Agent 加载策略

| 需求类型 | 应加载的规范 |
|---------|------------|
| 任何涉及业务代码的需求 | 按「加载预算」读已选用安全规范的**相关节**；非每次全文 |
| 新增/修改 Controller | 后端规范相关节 |
| 新增/修改 HTTP API | 后端规范相关节 + 接口规范相关节 |
| Webhook 逻辑 | 后端规范相关节 + 安全相关红线节 |
| 云适配器修改 | 后端规范相关节 + 安全相关红线节 + 相关 ADR |
| Code Review / MR 评审 | 质量入口 + 对应分册 |

## 规范约束力

- 标注"禁止"/"必须"的条目：**强制**遵守
- 标注"推荐"/"优先"的条目：**优先**遵守，有合理理由可偏离
- 常见场景参考：**参考**实现，可根据具体情况调整

## 章节快速索引

### security-bk-redlines.md

- 本次必读决策表、一、红线总览
- 二、红线 1：外部输入未校验
- 三、红线 2：敏感接口未鉴权
- 四、红线 3：敏感数据未加密
- 五、代码评审检查清单、规范落地优先级

### quality-code-review.md（按需；分册目录 `./quality-code-review/`）

- 加载说明、本次必读决策表
- [01-核心原则.md](./quality-code-review/01-核心原则.md)
- [02-问题分级.md](./quality-code-review/02-问题分级.md)
- [03-检查维度与规则清单.md](./quality-code-review/03-检查维度与规则清单.md)
- [04-代码质量评分标准.md](./quality-code-review/04-代码质量评分标准.md)
- [05-评审意见撰写规范.md](./quality-code-review/05-评审意见撰写规范.md)
- [06-评审报告格式.md](./quality-code-review/06-评审报告格式.md)

### backend-k8s-operator.md

- 一、技术栈要求；二、项目结构；三、Controller 开发规范
- 四、分层架构；五、Webhook；六、Metrics；七、缓存
- 八、编码规范；九、测试规范；十、配置管理；十一、构建与部署；十二、新增功能检查清单

### api-go-restful.md

- 一、架构概述；二、路由规范；三、响应格式规范
- 四、Handler 编写规范；五、Metrics；六、数据类型
- 七、安全规范；八、新增 API 检查清单；九、接口变更管理
