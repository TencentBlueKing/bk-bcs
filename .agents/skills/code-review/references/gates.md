# Code Review — 门禁清单（Checklist + Local Script）

> G0/G1/G2 清单由 Agent 自检；**G2 结案前必须再跑本 skill 本地脚本**（无共享包）。未通过不得宣称评审完成。

## 可执行结案命令

```bash
python3 skills/code-review/scripts/check_gates.py --route <routing-id>
```

校验本 skill 的 `routing.md` / `gates.md` / `checklist.md` 存在且路由 ID 合法。

## G0 入口门禁

| # | 检查项 | 通过标准 | 未通过时 |
|---|--------|----------|----------|
| G0.1 | 路由已声明 | 已读 `routing.md` 并输出路由 ID | 先路由 |
| G0.2 | 变更范围已采集 | staged/unstaged/commit 之一有实质 diff，或已说明无变更 | 按 git-scenarios 再采集 |
| G0.3 | 项目背景策略已执行 | 信任调用方背景或完成只读感知 | 补感知（只读） |
| G0.4 | 只读约束 | 未修改被评审代码或项目背景文件（除非用户明确要求改） | 停止改写，回到评审 |

## G1 阶段门禁

| 从 → 到 | 必须已满足 | 未通过时 |
|---------|------------|----------|
| 收集 → 下结论 | 已阅读周边代码，非孤立看 hunk | 补读调用点/导入 |
| 出 CRITICAL/HIGH | 置信度 >80% 且可指出位置与理由 | 降级或删除该条 |
| 使用分级 | 仅用 CRITICAL/HIGH/MEDIUM/LOW | 去掉并行分级体系 |
| AI 代码专项 | 已覆盖行为/安全/耦合/复杂度中的相关项 | 补查 |

## G2 结案门禁

| # | 检查项 | 通过标准 |
|---|--------|----------|
| G2.0 | 脚本门禁 | `check_gates.py --route …` 输出 PASS |
| G2.1 | 报告结构 | 符合 `report-format` 或用户指定格式 |
| G2.2 | 建设性 | 问题含「为什么」；对事不对人 |
| G2.3 | 批准逻辑 | 若给合并建议，与 `scoring-standard` 一致 |
| G2.4 | 无变更场景 | 明确说明无 diff/无可评范围，而非虚构问题 |

## 自检记录格式

```text
[code-review gates] G0=pass G1=pass G2=pass routing=cr-auto script=PASS
```
