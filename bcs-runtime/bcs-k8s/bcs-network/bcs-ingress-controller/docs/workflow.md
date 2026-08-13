# 迭代开发工作流

基于 TAPD 需求驱动的迭代开发流程，覆盖需求澄清、评估、规划到迭代执行的完整步骤。

## 基础配置

| 配置项 | 值 |
|-------|---|
| 需求来源 | tapd |
| 分支策略 | v{MAJOR}.{MINOR}.x |
| 工作目录 | specs/{iteration_name} |
| 重试次数 | 3 |

## 步骤

### 1. 需求澄清

- **skill**: tapd-story-clarification
- **输入**: workspace_id（project.json）, story_ids（用户输入）
- **输出**: docs/reqs/{需求名称}.md

提取 TAPD 需求，针对需求信息采用研发最佳实践进行澄清，明确交付内容。接受用户输入的需求 ID 列表，逐一拉取需求详情、整合背景知识进行澄清，输出规范化需求文档，并将最终文档同步回 TAPD（`v_status` 更新为"已澄清"）。

### 2. 需求评估

- **skill**: tapd-story-evaluation
- **输入**: {需求澄清.docs/reqs/*.md}, workspace_id
- **输出**: docs/reqs/{子需求名称}.md（拆分后的子需求文档）

基于规范需求文档进行逻辑分析，按接口定义、公共功能库、前端模块、后端模块进行子需求拆分，输出各子需求的规范文档，并使用 RICE 模型对每个子需求进行价值规模评分。拆分结果需用户确认后再继续。

### 3. 迭代规划

- **skill**: tapd-iteration-plan
- **输入**: workspace_id, iteration_name（用户输入）, 已评估需求列表
- **输出**: TAPD 迭代单据（iteration_id），各需求 iteration_id 已更新

针对需求池中所有已评估需求，分析依赖关系与规模评分，将合适的需求编排进入迭代（单迭代总分 500-1000）。可复用已有迭代（重入模式）或新建迭代。

### 4. 迭代开发

- **skill**: tapd-iteration-runner
- **输入**: workspace_id, iteration_id（来自迭代规划或用户输入）
- **输出**: 各需求代码变更 + specs/{iteration_name}/changelog.md + specs/{iteration_name}/summary.md

调用 `tapd-iteration-runner` skill，由其完整编排迭代的四个阶段：环境初始化与分支创建（tapd-iteration-init）→ 需求依赖分析与拓扑排序（tapd-iteration-analysis）→ 逐需求调度实现流水线（tapd-story-pipeline，六阶段：技术澄清 → 开发计划 → 任务拆分 → TDD 实现 → 架构/安全校验 → 代码提交）→ 迭代汇总报告（tapd-iteration-report）。

runner 自行管理 `iteration-state.json` 与各需求的 `meta.yaml` 状态文件；详细编排逻辑见 `tapd-iteration-runner/SKILL.md` 和 `tapd-story-pipeline/SKILL.md`，workflow 层不重复定义内部子步骤。

### 5. 文档巡检

- **skill**: harness-engineering（触发词：轻量检查）

调用 harness-engineering skill，执行 PR 轻量检查模式，维护 harness 文档与项目实际状态的一致性。
