# 总结报告格式模板

第五步生成的报告遵循以下格式：

```markdown
## Harness Engineering 规范生成报告

### 已完成的组件
- [x] AGENTS.md — 项目入口（仓库根目录）
- [x] 词汇表 — docs/glossary.md

### AGENTS 理解（若改写前已有根 AGENTS）

> 规程：`references/agents-merge.md`。先理解再裁决；禁止描点作为充分条件。

- 意图摘要：${AGENTS_INTENT_SUMMARY}
- 四件套（地图 / 边界与禁令 / 做法与参照 / 诚实验证）：${AGENTS_FOUR_PACK}
- 去留表：`RETAIN-ENTRY` / `MERGE-HEADER` / `MOVE-TOOLING` / `PROMOTE-SUMMARY` / `DROP` — ${AGENTS_DISPOSITION}
- DROP 理由：${AGENTS_DROP_REASONS}
- 行数 WARN（若有）：${AGENTS_LENGTH_WARN}

### 工作单元 AGENTS

> 规程：`references/agents-work-units.md`。默认只索引不覆写局部正文。

- 扫描命令：`git ls-files -- 'AGENTS.md' '**/AGENTS.md'`
- 非根路径清单：${WORK_UNIT_AGENTS_PATHS}（无则写「无」）
- 根短头已索引：是 / 否 / 不适用（清单为空）
- nearest 优先句已写入根：是 / 否 / 不适用

- [x] 上下文工程 — docs/harness/context-engineering.md
- [x] 架构约束 — docs/harness/architectural-constraints.md
- [x] 熵管理 — docs/harness/entropy-management.md
- [x] 工具能力 — docs/harness/tooling.md
- [x] 执行与验证 — docs/harness/execution-verification.md
- [x] harness 导航 — docs/harness/README.md

### 待补充的内容
- {组件} > {内容}：{说明}
...

### Dev Map git 策略
- 入库：`docs/dev-map/README.md`、`docs/dev-map/.gitignore`（白名单：仅此二者；**非**根上整目录 ignore）
- 本地生成、默认不入库：本目录其余文件（含 `graph.json` 等）
- 克隆后首次：`GRAPHIFY_OUT=docs/dev-map graphify update .`（AST-only）

### Skill 布局
- Skill 安装根（探测值）：`${SKILL_INSTALL_ROOT}`
- 是否 IDE→`.agents` 软链：是/否
- 布局 WARN（若安装根为 `.<ide>/skills` 实体且非 `.agents/skills`）：建议收敛到 `.agents/skills` + IDE 整目录软链；见 `references/skill-install-root.md` / `docs/agent-install.md`

### 技术栈探测（detect-standards）
- 命令：`bash …/detect-standards.sh --json .`
- primary：frontend=`${FE_PRESET@root}`；api=`${API_PRESET@root}`；backend=`${BE_PRESET@root}`
- 未覆盖分类：`${UNMATCHED}`

### Standards Rules
- Rules 根：`${RULES_ROOT}`（如 `.agents/rules/` / `.cursor/rules/` / Skip）
- 写入/删除：`${RULES_ACTIONS}`（sync-standards-rules.sh stdout 摘要）
- Skip 原因（若有）：无 IDE 目录 / 用户拒绝 / 无选用规范

### 环境工具缺口

> 来源：`harness-doctor` stdout + 所选场景 MCP probe。结果**不**写回 tooling.md。

#### 场景：${SCENARIO_NAME}
- [ ] [${TOOL_TYPE}] ${TOOL_NAME} — 被 ${SKILL_OR_AGENT} 依赖，当前环境不可用
  - 安装/配置方式：${INSTALL_HINT}

> ⚠️ 以上工具缺口会导致对应 Skill 无法正常运行，请优先补齐。

若无缺口，输出：环境工具探测完成，所选场景依赖均 present（仍以会话内 MCP 为准）

若 `docs/dev-map/` 存在但无 `graph.json`：提示  
`GRAPHIFY_OUT=docs/dev-map GRAPHIFY_NO_BACKUP=1 graphify update .`

### 后续建议
1. 完善标注为"待补充"的内容
2. **优先按 harness-doctor / 上表补齐 MCP/CLI**，否则相关 Agent/Skill 无法正常运行
3. 项目自研 Skill/MCP 纳入 git 后会出现在 tooling「项目自有工具」节；Harness 基线变更仍改治理仓 `tool-dependencies.md`
4. 新增术语后更新 docs/glossary.md
5. 定期运行文档巡检（含 tooling 基线与权威清单、项目自有与 git 对账）
6. 需要时再跑 `harness-doctor` 确认本机缺口
7. **根 README 口径**（若有冲突）：探测栈与根 README 主推关键词不一致时，请人工修订 README（harness **不**自动改）；冲突信号：`${README_STACK_CONFLICTS}`
```
