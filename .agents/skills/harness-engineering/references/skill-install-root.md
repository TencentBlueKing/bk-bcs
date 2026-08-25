# Skill 安装根（运行时探测）

> `${SKILL_INSTALL_ROOT}` / `$SKILL_ROOT` 是 Agent **执行期**变量：扫描项目级 Skill 的根目录。  
> **推荐布局 ≠ 唯一合法路径**；每人/每 IDE 可能不同，文档禁止写死「权威路径=.cursor/skills」。

路径约定对齐治理仓 `scripts/install-to-target.sh` 与 `docs/agent-install.md`（fallback：`.agents/skills` + IDE 整目录软链到 `.agents`）。

## 探测优先级

按顺序命中第一个即赋值 `${SKILL_INSTALL_ROOT}`：

| 顺序 | 条件 | 安装根 |
|------|------|--------|
| 1 | 存在 `.agents/skills/*/SKILL.md`（或至少 harness-engineering） | `.agents/skills`（推荐） |
| 2 | 存在 `skills/*/SKILL.md`（治理/开发源） | `skills` |
| 3 | `.cursor` / `.codebuddy` / `.claude` / `.codex` 为指向 `.agents` 的软链 | 仍报 `.agents/skills` |
| 4 | 上述 IDE 下 `skills/` 为实体目录且含 `SKILL.md` | `.<ide>/skills`；总结报告 WARN：建议收敛到 `.agents/skills` + 软链 |

## 推荐布局（对人）

```bash
mkdir -p .agents/skills
# 将 skill 安装到 .agents/skills/<name>/
ln -sfn .agents .cursor       # 整目录软链（install-to-target fallback）
# 或仅 skills 子树：
# ln -sfn ../.agents/skills .cursor/skills
```

`install-to-target.sh` 对**尚无 git 跟踪内容**的 `.agents/`、`.cursor/`、`.claude/` 等写入目标仓 `.gitignore`（纯安装产物通常不进 git）。  
若目标仓已跟踪该路径下文件（项目自有 Skill/Rules），脚本**不会** ignore，并会移除此前误加的冲突行。

## 文档表述

- 对 Agent：写「扫描 `${SKILL_INSTALL_ROOT}/*/SKILL.md`」，并注明变量由本探测步骤赋值。  
- 生成物可写「本仓库当前 Skill 安装根：`.agents/skills`（探测结果）」——相对仓库事实，非本机家目录。  
- 禁止：「权威路径必须为 `.codebuddy/skills`」类唯一字面量断言。
