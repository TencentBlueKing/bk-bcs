# iteration-state.json Schema

> 本文档定义 `iteration-state.json` 的完整字段结构。由 runner 主 SKILL.md §5 引用。

## 字段定义

```json
{
  "iteration_id": "<TAPD 迭代 ID>",
  "workspace_id": "<工作空间 ID>",
  "iteration_name": "<迭代名称>",
  "owner": "<当前用户名>",
  "project_path": "<git 仓库路径>",
  "iter_branch": "<迭代分支 v0.9.x>",
  "agent_tool": "<agent 工具名称，默认 agent>",
  "patch": 0,
  "started_at": "<ISO 8601>",
  "end_at": "",
  "status": "<initialized | analyzed | implementing | bugfix | completed | reported>",
  "all_parents": ["<所有父需求 ID>"],
  "selected_story": "<当前正在处理的需求 ID>",
  "sequence": ["<拓扑排序后的需求 ID 列表>"],
  "stories": {
    "<PARENT_ID>": {
      "name": "<父需求名称>",
      "children": {
        "<CHILD_ID>": {
          "name": "<子需求名称>",
          "phase": "<派生自 meta.yaml.phase>",
          "started_at": "<ISO 8601，runner 首次处理该需求时写入>",
          "stats": { 
            "total": 0, "add_code": 0, "delete_code": 0, "logic_code": 0, "test_code": 0, "docs": 0,"files": 0, "duration_sec": 0,
            "cost": { 
              "per_stage": {} 
            }
          }
        }
      }
    },
    "<INDEPENDENT_ID>": {
      "name": "<独立需求名称>",
      "phase": "<派生>",
      "stats": { },
      "children": null
    }
  }
}
```

## 字段语义

| 字段 | 决策场景 | 说明 |
|------|---------|------|
| `status` | 判断迭代所处阶段 | 决定加载哪个子 skill |
| `sequence` | loop 遍历顺序 | 拓扑排序后的可执行需求 ID |
| `selected_story` | 标识当前正在处理的需求 | loop 每轮设置 |
| `stories.*.phase` | 判断需求是否完成 | **派生缓存**——从 `meta.yaml.phase` 同步 |
| `stories.*.stats` | 汇总研发效能 | **派生缓存**——从 `meta.yaml.stats` 同步，由 report 阶段聚合 |
| `stories.*.stats.cost` | 需求成本数据 | 仅含 `per_stage`（执行层，各阶段 speckit 调用消耗）。需求总 credit 和 tokens 为各阶段求和 |
| `stories.*.stats.cost.per_stage` | 执行层各阶段消耗 | 由 pipeline 内部 `log_usage.py` hook 捕获的 speckit subagent 调用消耗，按 stage 分组；每个 stage 含 `credit`、`total_*_tokens`、`calls` 字段 |
| `stories.*.started_at` | 需求壁钟开始时间 | runner 在 loop 首次处理该需求时写入（ISO 8601），持久化以便恢复后仍可计算耗时 |
| `stories.*.stats.duration_sec` | 需求壁钟耗时（秒）| runner 在 phase 推进到 `committed` 时写入 = committed 时刻 − `started_at`（含等待用户决策的间隔）|
| `patch` | 构建版本号 | report 阶段自增，bugfix 模式恢复时也自增 |
| `iter_branch` | 传给 pipeline 的 commit 阶段 | 通过 §8.1 内联入参传递 |
| `agent_tool` | 传给 pipeline 的调度命令 | 通过 §8.1 内联入参传递 |

## 写入权限

- runner **完全拥有** `iteration-state.json`，因为迭代级状态只有一个写者，避免 runner 与 pipeline 并发修改导致数据冲突
- `stories.*.phase` 和 `stats` 是 `meta.yaml` 的派生缓存——runner 在 loop 中从各需求的 `meta.yaml` 同步过来
- runner **不直接修改** pipeline 的 `meta.yaml`——pipeline 对自己的 meta.yaml 有完整的校验逻辑，外部修改必须遵循 `state-mutation-guide.md` §7 速查表规则
