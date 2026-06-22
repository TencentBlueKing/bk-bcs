# 执行与验证（Execution & Verification）

> 目标：通过执行循环和强制验证确保 Agent 任务被正确完成。

## 1. Agent Loop 执行循环

### 1.1 循环结构

```
while 任务未完成:
    1. 观察（Observe）— 读取 AGENTS.md、dev-map、相关代码
    2. 推理（Think）— 分析影响范围，规划变更步骤
    3. 行动（Act）— 编辑代码、运行测试
    4. 验证（Verify）— 编译、测试、lint 检查
    5. 更新（Update）— 同步文档，报告进度
```

### 1.2 循环保护

| 保护机制 | 配置 | 触发动作 |
|---------|------|---------|
| 最大循环次数 | 20 次 | 终止并报告 |
| 最大执行时间 | 30 分钟 | 终止并报告 |
| 无进展检测 | 连续 3 次状态无变化 | 暂停并请求人工介入 |
| 重复失败 | 同一错误连续 2 次 | 换思路或报告阻塞 |

## 2. 强制验证机制

### 2.1 预完成检查清单

Agent 在宣称任务完成前，必须逐项确认：

| 检查项 | 验证方式 | 跳过条件 |
|-------|---------|---------|
| 代码编译通过 | `cd .. && make ingress-controller` | 无代码变更 |
| 单元测试通过 | `cd .. && make test-ingress-controller` | 无测试变更 |
| gofmt/goimports 干净 | 格式化检查 | 无代码变更 |
| 新常量已入 constant.go | 检查 internal/constant/ | 无新 Annotation |
| 新 Controller 已注册 main.go | grep SetupWithManager | 无新 Controller |
| 新 metrics 已 init 注册 | 检查 internal/metrics/ | 无新指标 |
| 新 HTTP 路由已注册 | 检查 httpserver.go InitRouters | 无新路由 |
| 文档已同步 | 检查 AGENTS.md / dev-map | 无结构变更 |
| 安全红线 | 加载 bk-security-redlines skill | 无代码变更 |

### 2.2 验证失败处理

- 验证失败 → 回到执行循环修复
- 连续 3 次修复失败 → 暂停并请求人工介入
- 测试结果记录到执行日志

### 2.3 快速回归（Namespace Scope 等特性）

```bash
go test -count=1 -run 'TestParseExemptNamespaces|TestRuleConverter|TestMappingConverter|TestIsExempt|TestGetNsClient|TestReloadNsClient|TestNewNamespacedLB' \
  ./internal/cloud/namespacedlb/... ./internal/generator/... . 2>&1 | grep -E 'PASS|FAIL|ok'
```

## 3. 任务漂移检测

### 3.1 漂移信号

| 信号 | 含义 | 处理方式 |
|------|------|---------|
| 修改与任务无关的文件 | 范围蔓延 | 撤销变更并提醒 |
| 在 initClient 中添加业务逻辑 | 架构违规 | 回退，改用 initXxxClient |
| 使用 log/klog 替代 blog | 规范违规 | 立即修正 |
| 硬编码 Annotation Key | 规范违规 | 移入 constant.go |
| 假设 go.mod 在本目录 | 路径错误 | 从 bcs-network/ 目录执行 |

### 3.2 检查点机制

- 每完成一个子任务设置检查点（如「Controller 逻辑完成」「测试通过」「文档更新」）
- 检查点记录：已完成项、待完成项、已知风险

## 4. 结果可观测性

### 4.1 执行日志

使用 `bcs-common/common/blog` 记录运行时日志；开发阶段通过 `go test -v` 输出验证结果。

### 4.2 Prometheus 指标

| 子系统 | 文件 | namespace |
|--------|------|-----------|
| 全局注册 | `internal/metrics/metric.go` | `bkbcs_ingressctrl` |
| PortPool | `internal/metrics/portpool.go` | 同上 |
| HostNetPortPool | `internal/metrics/hostnetportpool.go` | 同上 |
| Listener | `internal/metrics/listener_controller.go` | 同上 |
| Webhook | `internal/metrics/webhook.go` | 同上 |
| Check | `internal/metrics/check.go` | 同上 |

### 4.3 关键指标

| 指标 | 计算方式 | 目标值 |
|------|---------|-------|
| 测试通过率 | make test-ingress-controller | 100% |
| 首次验证通过 | 无需返工的任务比例 | ≥ 90% |
| 文档同步率 | 代码变更 PR 含文档更新 | ≥ 80% |

## 检查清单

- [x] Agent Loop 执行循环已定义
- [x] 循环保护机制已配置
- [x] 预完成检查清单已制定（对齐 Pre-PR Checklist）
- [x] 任务漂移检测规则已明确
- [x] Prometheus 指标位置已文档化
