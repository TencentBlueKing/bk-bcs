# IDE Rules 模板

| 目录 | 格式 | 落盘 |
|------|------|------|
| `cursor/*.mdc.template` | Cursor Project Rules（`alwaysApply` / `globs`） | `.cursor/rules/*.mdc` |
| `codebuddy/*.md.template` | CodeBuddy 模块化规则（`enabled` / `alwaysApply`）+ 可选 `paths`（fallback 时兼容 Claude） | `.codebuddy/rules/*.md` |
| `claude/*.md.template` | [Claude Code `.claude/rules/`](https://code.claude.com/docs/en/memory#organize-rules-with-claude/rules/)（无 `paths`=会话始终加载；有 `paths`=按 glob 懒加载） | `.claude/rules/*.md` |

**Codex**：无模块化 agent instruction rules。官方 `.codex/rules/*.rules` 是 **execpolicy**（Starlark `prefix_rule`），不是编码规范。Standards / graphify 门闩写入根目录 `AGENTS.md`（Codex 原生读取），勿往 `.codex/rules/` 写 markdown。

| 模板 | 写入脚本 | 说明 |
|------|----------|------|
| `standards-*.template` | `../../scripts/sync-standards-rules.sh` | Standards 加载门闩；算法见 `../../references/standards-compliance.md` |
| `graphify.*template` | `../../scripts/harness-ide-setup.sh` | graphify 探索门闩；按布局写对应格式；另写 CodeBuddy `settings.json` hook-guard |

落盘路径算法：

1. 任一 `.$ide` → `.agents` 软链（fallback）→ `.agents/rules/` 同时写 `.mdc` + `.md`（`.md` 含 `paths`，供 Claude 经 `.claude` 软链读取）
2. 实体 `.cursor` → `.mdc`；实体 `.codebuddy` → CodeBuddy `.md`；实体 `.claude` → Claude `.md`
3. 皆无且存在 `.agents` → 同 fallback 双格式
4. 仅有 `.codex` → **不写** IDE rules（依赖 `AGENTS.md`）
