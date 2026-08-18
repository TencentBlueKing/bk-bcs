# Pipeline 层错误处理规则

> 本文档约束 tapd-story-pipeline 及其内部子 skill（specify / plan / tasks / implement / validate / commit）
> 在执行过程中遇到的错误如何分类、落地到 `meta.yaml.last_failure`，以及调用方如何处理。
> 与可重入相关的决策（如 subagent 返回 fail / blocked_on_questions / needs_fix / *_insufficient）
> 详见 `reentry-protocol.md` 的决策矩阵。

## 1. 错误分类

| 类型 | `last_failure.type` | 触发场景 | 重试策略 |
|------|----------|---------|---------|
| 系统/网络 | `system` | speckit 命令超时；网络抖动；agent 工具进程异常退出；磁盘写失败；subagent 写入的产物文件未生成；subagent 回传 JSON 格式非法 | pipeline 第一次启动时若 `attempts==1` 自动清空 `last_failure` 并重试一次 |
| 需求实现 | `semantic` | speckit 返回结果不满足质量检查；validate 不通过；spec / plan / tasks 信息不充分；speckit.analyze 验证 CRITICAL/HIGH 违规；subagent 回传 `status=fail` 非系统类 | 必须由外部（用户或 runner 代用户）分析根因后修文件再调起 |
| 非法修改 | `mutation_invalid` | pipeline 启动校验时检测到外部对 `meta.yaml` / `phase` / 产物的非法修改（如 phase 跨链跳跃、前置产物缺失）；`questions.md` 中出现非法状态或格式破损 | 必须按 `state-mutation-guide.md` §4 修复后调起；不在 pipeline 内自动重试 |

## 2. last_failure 字段格式

```yaml
last_failure:
  type: system | semantic | mutation_invalid
  phase: <出错时所处的 phase>
  message: "<根因摘要，不超过 200 字>"
  occurred_at: "<ISO 8601 时间戳>"
  evidence:                # 可选，仅 semantic / mutation_invalid 用
    - "<process.log 关键行号或 banner>"
    - "<问题文档路径>"
```

## 3. blocked vs fail 的区分

子 skill 返回 `blocked` 与 `fail` 的判定准则：

| 现象 | 判定 |
|------|------|
| 子 skill 主动追加 `[open]` 条目到 `questions.md`，已生成部分产物 | `blocked`（不写 last_failure，调用方按 `state-mutation-guide.md` §2 卡点 2 处理）|
| 子 skill 命令失败但属临时/可重试 | `fail` + `type=system` |
| 子 skill 产出与输入显著矛盾，需要外部修正 | `fail` + `type=semantic` |
| 外部修改导致 phase 与产物不一致 / questions.md 破损 | `fail` + `type=mutation_invalid` |
| speckit.analyze 返回 CRITICAL/HIGH violations | `fail` + `type=semantic`（按 `reentry-protocol.md` 走回退重入）|
| speckit.analyze 返回 MEDIUM/LOW violations | round 重试中消除，不写 `last_failure` |

## 4. Subagent 返回状态决策

subagent 完成后返回的三态（`ok` / `blocked_on_questions` / `fail`）与 compliance verdict 决策
**统一由 `reentry-protocol.md` §2 决策矩阵定义**。本节仅提供速览：

| 现象 | 落地 | 进一步阅读 |
|------|------|----------|
| `status=ok` + `compliance.verdict=pass/LGTM` | 推进 phase | 决策矩阵行 A/E/L/Q/T |
| `status=ok` + `compliance.verdict=needs_fix` | round 重试 | `reentry-protocol.md` §3.2 |
| `status=ok` + `compliance.verdict=spec_insufficient` 等 | `fail` + `type=semantic`，回退重入 | `reentry-protocol.md` §3.3 |
| `status=blocked_on_questions` | 追加 questions.md，pipeline 退出（confirm 卡点之外） | `state-mutation-guide.md` §2 卡点 2 |
| `status=fail` | 写 `last_failure`（区分 system/semantic）退出 | 决策矩阵 + §4 |

## 5. 退出报告必含信息

pipeline 退出时若 `last_failure` 非空，主会话最后一条消息必须包含：

- 卡点类型：`fail`
- `last_failure.type`：`system` / `semantic` / `mutation_invalid`
- 失败 phase
- 根因摘要（取自 `last_failure.message`，≤200 字）
- 建议处理步骤（按 `state-mutation-guide.md` §2 卡点 3 或 §2 卡点 4）

不要在退出报告中粘贴 `process.log` 完整内容——`process.log` 本身是审计载体。

## 6. 日志摘取口径

从 `process.log` 中提取错误信息时，统一按以下口径：

- **当前模式**（use_skill 路径）：`tail -200 process.log` 的最近 N 行 banner 与
  关键事件（subagent 在阶段执行中追加的错误 / 警告 / 产物路径）
- **validate 阶段**：日志分布在独立文件中，按维度分别摘取：
  - 架构：`process-validate-arch.log`
  - 安全：`process-validate-security.log`
  - CodeReview：`process-validate-codereview.log`
  - 修复：`process-validate-fix.log`
  摘取口径与上述当前默认模式相同，只是目标文件不同。

提取到的信息仅用于填充 `last_failure.message`（≤200 字）；**完整日志不进主会话上下文**。

## 7. 参考

- 决策矩阵与动作序列：`reentry-protocol.md`
- 外部修改契约（含 last_failure 出卡点）：`state-mutation-guide.md` §2 §4
- 回传 JSON schema：`subagent-prompt-template.md`
- meta.yaml 字段与生命周期：`context-and-meta-template.md`
- questions.md 四态状态机：`questions-md-template.md`
