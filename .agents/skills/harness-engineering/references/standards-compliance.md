# Standards 遵守与 IDE Rules 同步

> generating 与 gardening **共用**本文算法。禁止各写一套路径/裁剪逻辑。  
> 实现类 Skill（frontend/backend-developer 等）**不**耦合本流程。

## 三层模型

1. **短门闩**：AGENTS「编码前必读」+ IDE Rules（≤40 行/文件）  
2. **长规范**：`docs/standards/*`，按任务 **按节 Read**（先查 README「加载预算」+「章节快速索引」）  
3. **硬门禁**：仓库 lint/CI（规范「规范落地优先级」）

禁止把完整 frontend/backend 预设正文写入 Always Rules。  
禁止门闩/Rules 要求「无差别 Read 整份长规范」；P2 / 评分量表 / 陷阱章默认不预加载。

## Rules 落盘路径（对齐 `scripts/install-to-target.sh`）

候选 IDE：`cursor`、`codebuddy`、`claude`、`codex`。

1. 任一 `.$ide` 为指向 `.agents` 的软链 → 权威目录 **`.agents/rules/`**，同目录写 `.mdc`（Cursor）+ `.md`（CodeBuddy；frontmatter 含 Claude `paths` 以便 `.claude` 软链共用）  
2. `.cursor` 为实体目录 → `.cursor/rules/*.mdc`  
3. `.codebuddy` 为实体目录 → `.codebuddy/rules/*.md`  
4. `.claude` 为实体目录 → `.claude/rules/*.md`（Claude Code 格式：无 `paths` 始终加载；有 `paths` 按 glob 懒加载）  
5. **Codex**：不写 `.codex/rules/`（该路径为 execpolicy Starlark，非 agent 指令）；Standards / graphify 门闩以根目录 **`AGENTS.md`** 为准  
6. 以上皆无且存在 `.agents` → 同第 1 步双格式；否则 Skip Rules（保留 AGENTS 门闩）；**默认不**擅自创建软链  

可执行同步：`scripts/sync-standards-rules.sh <workspace>`（见脚本头注释）。

**graphify Rules** 使用**同一套落盘路径**，由 `scripts/harness-ide-setup.sh` 写入对应格式（模板在 `assets/ide-rules/`），并另行维护 CodeBuddy `settings.json` hook-guard。

参考：

- CodeBuddy 模块化规则：[官方文档](https://www.codebuddy.ai/docs/zh/cli/memory#%E4%BD%BF%E7%94%A8-codebuddy-rules-%E5%AE%9E%E7%8E%B0%E6%A8%A1%E5%9D%97%E5%8C%96%E8%A7%84%E5%88%99)
- Claude Code rules：[Memory / `.claude/rules/`](https://code.claude.com/docs/en/memory#organize-rules-with-claude/rules/)
- Codex 指令：[AGENTS.md](https://developers.openai.com/codex/guides/agents-md)；execpolicy：[Rules](https://developers.openai.com/codex/rules)

## 按选用裁剪

从 `docs/standards/README.md`「当前项目选用的规范」表解析已链接的 `*.md`：

| 选用文件名匹配 | 生成 rule 基名 |
|----------------|----------------|
| `frontend-*` | `standards-frontend` |
| `api-*` | `standards-api` |
| `backend-*` | `standards-backend` |
| `security-*` | `standards-security` |
| 上表任一命中 | 另写 `standards-gate` |
| 全无 | 不写任何 `standards-*` rule |

未选用的分类：删除已有多余 `standards-<cat>.*`。

## 写入时机

- **harness-generating** 交付前：默认执行同步  
- **harness-gardening**：维度「Standards Rules 一致性」P0 补写/删除  

## 模板

`assets/ide-rules/cursor/*.mdc.template`、`assets/ide-rules/codebuddy/*.md.template`、`assets/ide-rules/claude/*.md.template`；占位 `${STANDARD_PATH}` 换成 `docs/standards/<file>`。
