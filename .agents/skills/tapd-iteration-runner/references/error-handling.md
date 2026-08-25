# Runner 层错误处理规则

> 本文档约束 tapd-iteration-runner 在迭代调度过程中遇到的迭代级错误如何处理。
> 单需求层面的错误由 pipeline 自行管理，runner 仅通过 `meta.yaml.last_failure` 感知。

## 1. 错误分类

| 类型 | 触发场景 | 处理 |
|------|---------|------|
| **TAPD MCP 失败** | `iterations_get` / `stories_get` 等调用失败 | 重试 3 次（指数退避）；仍失败则停止 init / analysis，向用户报告错误，等待人工恢复 |
| **Git 操作失败** | 创建/切换迭代分支失败；推送失败 | 立即停止；展示错误码与命令；不自动重试（避免破坏分支状态） |
| **Git MCP 操作失败** | git MCP 调用异常 | 展示错误信息；输出手动操作命令供用户复制；不自动重试 |
| **iteration-state.json 解析失败** | JSON 格式损坏 / schema 不匹配 | 提示用户备份后人工修复；不自动覆盖；提示按 `recovery.md` 处理 |
| **PowerShell 编码乱码** | Windows 首次执行编码不正确 | 执行 `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8`；重新启动当前会话 |
| **pipeline 退出异常** | pipeline 退出但 `meta.yaml` 不存在 / `phase` 字段缺失 | 视为 `fail`；提示用户排查 `process.log` 与 pipeline 日志；不自动重试 |
| **pipeline 卡点决策** | pipeline 退出且 `meta.yaml.last_failure` 非空 | 按 `../tapd-story-pipeline/references/state-mutation-guide.md` §7 runner 视角速查表处理 |

## 2. 自动重试边界

| 错误类型 | 自动重试 |
|---------|---------|
| TAPD MCP 失败 | ✅ 重试 3 次（指数退避：1s / 4s / 16s）后停止 |
| Git 失败 | ❌（人工介入）|
| Git MCP 失败 | ❌（人工介入）|
| iteration-state.json 解析失败 | ❌（人工修复）|
| PowerShell 编码乱码 | ❌（提示用户在终端执行一次编码设置）|
| pipeline `last_failure.type=system` 且 `attempts < 2` | ✅（runner 清空 `last_failure` 后再次内联执行 pipeline）|
| pipeline `last_failure.type=semantic` | ❌（与用户对话后再决策）|
| pipeline `last_failure.type=mutation_invalid` | ❌（与用户对话，重新修复状态文件）|

## 3. 退出时机

runner 在以下时机**正常退出**：

- 全部需求 `phase=committed` 且 `tapd-iteration-report` 完成（最终 `status=reported`）
- 用户主动放弃当前迭代

**异常退出**：

- 任何 §1 列出的错误且无法恢复时，runner 落盘 `iteration-state.json`（保留现场）后退出
- 下次恢复时 `tapd-iteration-init` 按 `recovery.md` 流程接续

## 4. 与 pipeline 的错误边界

runner **不读** pipeline 内部产物（spec.md / plan.md / tasks.md / process.log 等）来诊断错误。
所有诊断信息必须由 pipeline 落到：

- `meta.yaml.last_failure.message`（结构化，≤200 字）
- `meta.yaml.last_failure.evidence`（可选，关键日志行号或文档路径）
- `process.log`（详细日志，供用户人工排查）

runner 若需要展示 `process.log` 给用户，**仅在用户明确请求时**通过文件路径告知；不主动读取并转述。

## 5. 用户交互的错误分支

部分错误场景需要 runner 与用户对话（pipeline 不直接问用户）：

| 场景 | runner 动作 |
|------|-----------|
| pipeline 返回 `last_failure.type=semantic` | 展示 `message` 摘要（≤200 字）+ 推荐处理步骤（修 req.md / context.md / spec.md / plan.md / tasks.md），询问用户：**原地修复重试** / **回退重入（指定 target phase）** / **挂起需求** |
| pipeline 返回 `last_failure.type=mutation_invalid` | 展示具体非法点，引导用户按 `state-mutation-guide.md` §4 修复 phase 与产物的对应关系 |
| pipeline 退出且 `pending_review` 非空 | 展示 `pending_review.artifacts` 文件路径，询问用户：**通过** / **回退（指定 target phase）** / **放弃需求** |
| pipeline 退出且 `questions.md` 有 `[open]` | 展示所有 `[open]` 条目，逐条获取用户答复，写回 `[answered]` |
| 回退重入 `attempts` 达上限（默认 3）| 展示累计失败摘要，询问用户：**继续回退**（无硬上限）/ **挂起需求**（从 `sequence` 临时移除）/ **终止迭代** |

> **注意**：`max_attempts` 默认值为 3，定义在 runner 的 loop 决策逻辑中（非 meta.yaml 字段）。这是一个软上限——runner 会展示累计失败摘要并询问用户，用户可选择继续（无硬上限）。

> 新架构下**没有"跳过需求"选项**——跳过会污染后续依赖该需求的故事。如确需跳过，用"挂起需求"路径（从 `iteration-state.json.sequence` 临时移除）。

## 6. 参考

- 迭代级断点恢复：`recovery.md`
- 外部修改 pipeline 状态契约：`../tapd-story-pipeline/references/state-mutation-guide.md`
- Pipeline 层错误规则（runner 不直接处理但需理解 last_failure 类型）：`../tapd-story-pipeline/references/error-handling.md`
