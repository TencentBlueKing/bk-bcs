# 技术规范

> Agent 实现需求时的开发行为准则。根据需求涉及的端按需加载对应规范。

## 必选规范（所有项目强制）

| 分类 | 规范 | 文档 |
|------|------|------|
| 安全 | 蓝鲸代码安全三大红线 | [security-bk-redlines.md](security-bk-redlines.md) |
| 质量 | 代码评审规范（Google Code Review 指南） | [quality-code-review.md](quality-code-review.md) |

> 安全规范为横切关注点，**无论需求类型、技术栈如何，每次 Code Review 均须核查**。

## 当前项目选用的规范

| 分类 | 规范 | 文档 | 技术栈 |
|------|------|------|--------|
| 后端 | K8s Operator 开发规范 | [backend-k8s-operator.md](backend-k8s-operator.md) | Go + controller-runtime v0.6.3 |
| 接口 | go-restful HTTP API 规范 | [api-go-restful.md](api-go-restful.md) | go-restful |

## Agent 加载策略

| 需求类型 | 应加载的规范 |
|---------|------------|
| 任何需求 | 安全规范 + 质量规范（必选） |
| 新增/修改 Controller | 后端规范 |
| 新增/修改 HTTP API | 后端规范 + 接口规范 |
| Webhook 逻辑 | 后端规范 + 安全规范 |
| 云适配器修改 | 后端规范 + 安全规范 + 相关 ADR |

## 规范约束力

- 标注"禁止"/"必须"的条目：**强制**遵守
- 标注"推荐"/"优先"的条目：**优先**遵守，有合理理由可偏离
- 常见场景参考：**参考**实现，可根据具体情况调整

## 章节快速索引

### security-bk-redlines.md
- 输入校验红线、鉴权红线、数据加密红线

### quality-code-review.md
- 设计、功能、复杂度、测试、命名、注释、风格、文档

### backend-k8s-operator.md
- 技术栈、Controller/Reconcile 模式、分层架构、Webhook、Metrics、缓存、测试、构建

### api-go-restful.md
- 路由清单、响应格式、Handler 规范、Metrics、新增 API 检查清单
