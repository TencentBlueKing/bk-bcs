# Helm Hook Delete Policy 对齐设计

本文档记录 `helm.sh/hook-delete-policy` 从早期方案到当前实现的对齐过程。

> 状态说明（2026-04-01）：当前实现已采用 `hookCleanup.beforeCreate/onSuccess/onFailure`，
> hook action 的 `subscription.operation` 为 `Create`，不再使用 `Replace` 近似 delete policy。

## 背景

历史版本生成器曾将所有 hook 统一映射为：

- `type: Subscription`
- `operation: Replace`

这只能近似表达 `before-hook-creation`，但无法表达：

- `hook-succeeded`
- `hook-failed`
- 多个 delete policy 组合

因此历史实现与 Helm 的真实 hook 生命周期存在差距。

## Helm 原始语义

Helm 的 `hook-delete-policy` 本质上描述的是“在什么时机删除 hook 资源”。

常见值如下：

| Helm policy | 语义 |
|---|---|
| `before-hook-creation` | 本次执行前，如果旧 hook 资源存在，则先删除 |
| `hook-succeeded` | 本次 hook 执行成功后删除 |
| `hook-failed` | 本次 hook 执行失败后删除 |

这些值可以组合使用，例如：

```yaml
helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
```

## 历史问题

历史方案中，`operation: Replace` 的行为是：

1. 执行前删除旧资源
2. 创建新资源

这只能覆盖 `before-hook-creation` 的一部分语义，不能表达：

- 成功后自动删除
- 失败后自动删除
- 根据执行结果决定是否清理

因此不应继续把 delete policy 压缩为单个 `operation` 字段。

## 设计目标

目标是把 Helm delete policy 建模为“前置清理 + 结果驱动清理”，而不是继续硬编码成 `Replace`。

设计要求：

- 对齐 Helm 的三个核心 delete 时机
- 支持多个 policy 组合
- 不污染 rollback 语义
- 兼容当前 `Subscription` hook 执行模型
- 对 `PerCluster` 场景保持一致行为

## 建议数据模型

建议在 hook action 上增加结构化清理策略，例如：

```yaml
hookCleanup:
  beforeCreate: true
  onSuccess: false
  onFailure: false
```

字段含义：

| 字段 | 说明 |
|---|---|
| `beforeCreate` | 执行前删除旧 hook 资源 |
| `onSuccess` | hook 成功后删除 |
| `onFailure` | hook 失败后删除 |

对应 Helm 映射如下：

| Helm policy | DRPlan hookCleanup |
|---|---|
| `before-hook-creation` | `beforeCreate: true` |
| `hook-succeeded` | `onSuccess: true` |
| `hook-failed` | `onFailure: true` |

例如：

```yaml
helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
```

可映射为：

```yaml
hookCleanup:
  beforeCreate: true
  onSuccess: true
  onFailure: false
```

## 生成器改动建议

生成器侧需要做两件事：

1. 解析 `helm.sh/hook-delete-policy`
2. 将其转为结构化 `hookCleanup`

建议：

- `operationForHook()` 不再根据 delete policy 返回 `Replace`
- hook 的创建操作统一回归显式 create/apply 语义
- delete policy 只通过 `hookCleanup` 控制

这能避免把“生命周期清理策略”和“资源创建方式”混为一谈。

## 执行器改动建议

执行器侧按三个时机处理：

### 1. 执行前清理

当 `beforeCreate=true` 时：

- 删除旧 parent Subscription
- 再创建新的 parent Subscription

这对应 Helm 的 `before-hook-creation`。

### 2. 执行成功后清理

当 action 最终 `phase=Succeeded` 且 `onSuccess=true` 时：

- 删除 parent Subscription

这对应 Helm 的 `hook-succeeded`。

### 3. 执行失败后清理

当 action 最终 `phase=Failed` 且 `onFailure=true` 时：

- 删除 parent Subscription

这对应 Helm 的 `hook-failed`。

## PerCluster 场景

当前 hook 采用：

- parent Subscription
- child Subscription
- child 由 OwnerReference 依附 parent

因此在清理时：

- 只删除 parent Subscription 即可
- child Subscription 应通过级联删除自动清理

这意味着 `hookCleanup` 只需要作用在 parent 这一层，不需要为 child 单独设计 delete policy。

## 为什么不要复用 rollback

`hook-delete-policy` 和 rollback 不是一回事。

区别如下：

| 能力 | 触发时机 | 目的 |
|---|---|---|
| hook delete policy | 单次 hook 执行生命周期内 | 清理 hook 资源 |
| rollback | 用户触发 Revert 时 | 撤销已执行动作 |

因此不能把 `hook-succeeded` 或 `hook-failed` 实现成 rollback。

## 推荐落地顺序

建议分三步实现：

1. 生成器增加 `hookCleanup` 建模
2. 执行器增加 pre-create / post-success / post-failure 清理
3. 回归测试覆盖以下场景

建议测试矩阵：

- `before-hook-creation`
- `hook-succeeded`
- `hook-failed`
- `before-hook-creation,hook-succeeded`
- `PerCluster` hook 清理 parent 后 child 级联删除

## 结论

要真正对齐 Helm，`hook-delete-policy` 应建模为：

- 执行前是否清理
- 成功后是否清理
- 失败后是否清理

而不是继续将所有 hook 统一压缩为 `operation: Replace`。

一句话总结：

`hook-delete-policy` 是生命周期清理策略，不是资源创建操作类型。
