# 开发地图文档模板

本文件包含 docs/dev-map/README.md 的模板，由 `harness-generating` 第三步-C 读取并写入目标项目。

配套忽略清单见同目录 `dev-map.gitignore`（复制为 `docs/dev-map/.gitignore`）。

---

## 模板：docs/dev-map/README.md

以下为 README.md 的目标内容，harness-generating 第三步-C 执行时将此内容写入项目的 `docs/dev-map/README.md`：

# 开发地图（Dev Map）

> 基于 graphify 知识图谱，帮助 Agent 和开发者理解项目结构与概念关联。
> 相比旧式 Markdown 索引，token 消耗降低约 70x。

## 使用方式

调用 graphify skill 查询图谱（无需指定 `--graph` 路径参数，skill 已配置图谱位于 `docs/dev-map/graph.json`）：

| 调用方式 | 用途 |
|---------|------|
| `/graphify query "<问题>"` | 自然语言查询 |
| `/graphify path "<A>" "<B>"` | 路径查询——找两个概念之间的最短连接路径 |
| `/graphify explain "<概念>"` | 概念解释——展开某节点的定义、关联节点和所在 community |

## 仓库中提交什么

原则：**提交让 graphify 跑起来的配置与说明，不提交跑出来的结果。**

| 入库 | 不入库（见本目录 `.gitignore` 白名单） |
|------|----------------------------------------|
| 本 `README.md`（用法与约定） | **除下列白名单外的全部文件**（含 `graph.json`、`GRAPH_REPORT.md`、`cache/`、`wiki/`、可视化导出等） |
| 本目录 `.gitignore` | （策略：`*` + `!README.md` + `!.gitignore`） |

**禁止**在仓库根（或其它路径）把整个 `docs/dev-map/` 写进 gitignore——会丢掉约定落点与本说明。本目录内用白名单忽略即可。  
IDE Rules（`graphify.mdc` / `graphify.md`）由 harness 同步，属于「让图跑起来」的规则侧配置。

## 维护

由 harness-gardening 在代码变更后**本地**增量更新图谱（AST，无 API 费用）；**不**把 `graph.json` 纳入提交。  
图谱文件不存在时自动全量生成到本地。  
手动触发词："更新 dev map" 或 "生成开发地图"。

## 首次生成（克隆后）

若本地尚无 `graph.json`：

```bash
export GRAPHIFY_OUT=docs/dev-map GRAPHIFY_NO_BACKUP=1
graphify update . --no-cluster   # AST-only，无 LLM；或按 skill 跑全量 /graphify .
```

亦可触发 harness：「更新开发地图」/「生成开发地图」。
