# 开发地图（Dev Map）

> 本目录提供 BCS Ingress Controller 代码结构的导航索引，帮助 Agent 和开发者评估变更影响范围、定位具体实现文件。

## 文件说明

<!-- dev-map:auto -->
| 文件 | 用途 |
|------|------|
| [source-index.md](source-index.md) | 源文件目录——按目录分组列出所有 .go 文件的路径和职责描述 |
| [module-index.md](module-index.md) | 模块索引——每个模块的职责和关联文件，用于评估模块内部影响 |
| [module-dependencies.md](module-dependencies.md) | 模块依赖索引——模块间 import 依赖关系与 mermaid 图 |
<!-- /dev-map:auto -->

## 使用方式

评估某次变更的影响时，按以下顺序查阅：

1. **module-index** — 找到所属模块，确认模块内所有关联文件
2. **module-dependencies** — 查该模块被哪些模块依赖，评估跨模块影响
3. **source-index** — 查看具体文件的职责描述，确定修改点

## 维护规则

<!-- dev-map:auto -->
| 变更类型 | 应更新的文档 | 优先级 |
|---------|------------|-------|
| 新增源文件 | source-index.md、module-index.md | 建议 |
| 删除源文件 | source-index.md、module-index.md | 必须 |
| 文件移动/重命名 | source-index.md、module-index.md | 必须 |
| 新增模块间 import | module-dependencies.md | 建议 |
| 新增 Controller/子系统 | module-index.md、module-dependencies.md | 必须 |
<!-- /dev-map:auto -->

**自动维护**：对我说「文档巡检」触发 harness-gardening 维度 8 检测偏差。

**手动更新**：触发词「更新 dev map」或「生成开发地图」可全量更新。
