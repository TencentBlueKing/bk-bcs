# 执行与验证（Execution & Verification）

> 目标：让 Agent "做对事"——通过执行循环和强制验证确保任务被正确完成。

## 1. Agent Loop 执行循环

### 1.1 循环结构

```
while 任务未完成:
    1. 观察（Observe）— 获取当前环境状态（最近 AGENTS、相关测试命令、git status）
    2. 推理（Think）— 分析状态，规划下一步（范围收敛到目标组件）
    3. 行动（Act）— 调用工具执行操作
    4. 验证（Verify）— 在正确组件根运行已有 lint/test/build
    5. 更新（Update）— 更新任务状态
```

### 1.2 循环保护

| 保护机制 | 配置 | 触发动作 |
|---------|------|---------|
| 最大循环次数 | 20 次 | 终止并报告已做/未做 |
| 最大执行时间 | 60 分钟 | 终止并报告 |
| 无进展检测 | 连续 3 次状态无变化 | 暂停并请求人工介入 |
| 范围蔓延 | 修改与任务无关的组件 | 撤销无关变更 |

## 2. 强制验证机制

### 2.1 预完成检查清单

Agent 在宣称任务完成前，必须逐项确认。**验证命令与工作目录以根 / 最近工作单元 `AGENTS.md` 及仓库脚本为准**；按组件根执行，禁止虚构测试栈。无单测或未运行测试时**不得声称测试已通过**。

| 检查项 | 验证方式 | 跳过条件 |
|-------|---------|---------|
| 构建/编译 | 见下方「组件命令」；未知组件则读该目录 Makefile/README | 无代码变更 |
| 测试 | 仅运行入口层已写明或该组件 Makefile 已有的测试 | 入口允许跳过 / 无相关变更 / **无测试则不得声称通过** |
| Lint | 组件 `.golangci.yml` 或前端 `npm run lint` | 无代码变更 |
| 对照任务 Spec | 逐条检查需求点 | 无 |
| 文档已同步更新 | 特性进 `docs/features/`；组件约定同步局部 AGENTS | 无文档影响 |
| 安全红线 | 按 `docs/standards/security-bk-redlines.md` 相关节自检 | 纯文档 |

### 2.2 仓库与工作单元已有命令（禁止编造）

**仓库根**

| 目的 | 命令 | 说明 |
|------|------|------|
| 构建（按目标） | `make` / `make bcs-runtime` / `make bcs-services` 等 | 根 Makefile 按组件 target；不要假设一次编过全仓 |
| 测试 | `make test` | **当前主要依赖 `test-user-manager` 链路，不是全仓测试** |

**bcs-ui/frontend**

| 目的 | 命令（在 `bcs-ui/frontend`） |
|------|------------------------------|
| Lint | `npm run lint` |
| 构建 | `npm run build` |
| 开发 | `npm run dev` |

**bcs-ingress-controller**（在 `bcs-network/` 目录）

```bash
make ingress-controller
make test-ingress-controller
```

详见该组件 `AGENTS.md`。

**bcs-drplan-controller**

```bash
make lint
make test
make manifests generate   # 改 CRD types 之后
```

e2e 必须使用独立 Kind 集群，禁止打真实集群。

**bcs-terraform-bkprovider**

```bash
make test
make build
make proto    # 改 proto 之后
```

其它 `bcs-services/*`：进入该服务目录，使用其 `Makefile` / README；没有测试目标则只做编译与人工说明，**勿声称全量单测通过**。

### 2.3 验证失败处理

- 验证失败 → 自动回到执行循环修复
- 连续 3 次修复失败 → 暂停并请求人工介入
- 验证结果记录到会话说明中（命令、工作目录、exit code）

## 3. 任务漂移检测

### 3.1 漂移信号

| 信号 | 含义 | 处理方式 |
|------|------|---------|
| Agent 开始处理 Spec 以外的任务 | 任务漂移 | 回退到最近的检查点 |
| 修改与任务无关的文件/组件 | 范围蔓延 | 撤销变更并提醒 |
| 自主决定改变技术方案 | 决策越权 | 暂停并请求确认 |
| 循环执行相同操作 | 陷入死循环 | 终止并报告 |
| 用根 `make test` 宣称全仓通过 | 验证不诚实 | 按组件根重跑真实目标 |

### 3.2 检查点机制

- 每完成一个子任务设置检查点（git 可提交的中间状态或明确文件列表）
- 检查点记录：任务状态、已完成项、待完成项
- 漂移发生时可回退到最近检查点重新执行

## 4. 结果可观测性

### 4.1 执行日志

每次任务执行应记录以下信息：

| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | string | 任务唯一标识（需求 ID 或简述） |
| timestamp | ISO 8601 | 操作时间 |
| action | string | 执行的操作（如 read_file、run_test） |
| input | object | 工作目录与命令 |
| output | object | exit code / 关键错误 |
| status | string | success / failure / skipped |

### 4.2 Trace 分析

| 分析维度 | 工具 | 用途 |
|---------|------|---------|
| 执行路径 | 会话记录 | 定位异常操作 |
| 工具调用频次 | 会话记录 | 发现冗余调用 |
| 组件测试输出 | Makefile / `go test` | 确认验证发生在正确根 |

### 4.3 指标看板

| 指标 | 计算方式 | 目标值 |
|------|---------|--------|
| 任务完成率 | 成功任务 / 总任务 | ≥ 95% |
| 验证诚实率 | 实际执行的测试命令 / 宣称通过的测试 | 100% |

## 检查清单

- [x] Agent Loop 已定义
- [x] 预完成检查命令来自仓内 / 局部 AGENTS
- [x] 写明根 `make test` 并非全仓测试
- [x] 无单测不得声称通过
