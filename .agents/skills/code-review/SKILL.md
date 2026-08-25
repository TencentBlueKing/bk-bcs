<!-- BKUI-KNOWLEDGE-MANAGED:12a0ff0b91a8 -->
---
id: qual-code-review
name: 代码评审专家
category: quality
description: 基于 Google Code Review 指南的代码评审技能
tags: [code-review, quality, git, pr, mr]
updated_at: 2026-01-20
allowed-tools: [Read, Grep, Glob, Shell]
---

# 代码评审专家

## ⚠️ 核心规则

1. **追求持续改进，而非完美** - 倾向于批准能提升代码健康状态的变更
2. **对事不对人** - 有建设性的反馈，保持礼貌尊重
3. **解释为什么** - 帮助开发者理解原因

## 快速开始

```bash
/code-review              # 智能评审（自动检测变更范围）
/code-review staged       # 评审暂存区（提交前检查）
/code-review last-commit  # 评审最近一次提交
```

> 默认按优先级检测：暂存区 → 工作区 → 最近提交

## 路由优先与门禁

收到调用时**先**读 `./references/routing.md`，声明路由 ID，再按路由加读 reference；宣称完成前按 `./references/gates.md` 做 G0/G1/G2 自检，并运行：

```bash
python3 skills/code-review/scripts/check_gates.py --route <routing-id>
```

须 PASS。不依赖其他角色 skill；脚本仅服务本 skill。

## 评审流程

收到调用时按以下步骤执行（git 命令细节见 `./references/git-scenarios.md`，不在此重复）：

0. **路由** —— 读 `./references/routing.md`，声明路由 ID（如 `cr-auto` / `cr-staged`）。
1. **收集上下文** —— 运行 `git diff --staged` 和 `git diff` 查看所有变更；若无 diff，用 `git log --oneline -5` 检查最近提交。
2. **理解范围** —— 识别哪些文件发生变更、关联到什么功能/修复，以及彼此关联。
3. **阅读周边代码** —— 不孤立地评审变更；阅读完整文件，理解其导入、依赖和调用点。
4. **应用检查清单** —— 按路由结果加载 `./references/checklist.md` 等；各维度逐项检查，从 CRITICAL 到 LOW；置信度过滤见 `./references/confidence-filtering.md`。
5. **汇报发现** —— 按 `./references/report-format.md` 输出（或用户指定格式）；只汇报有把握的问题（>80% 确信是真实问题）。
6. **门禁自检** —— 按 `./references/gates.md` 勾选后交卷。

## 问题分级

本 skill 统一使用 CRITICAL/HIGH/MEDIUM/LOW 四级严重级别，为唯一权威分级体系，不引入并行分级：

| 级别 | 含义 | 处理 |
|------|------|------|
| `CRITICAL` | 可造成数据损坏、安全漏洞、生产不可用等严重问题 | 阻止合入，必须修复 |
| `HIGH` | 明显影响正确性/安全性/可维护性的问题 | 阻止合入，必须修复 |
| `MEDIUM` | 可改进但不阻塞的问题 | 讨论后决定是否修改 |
| `LOW` | 小问题，如命名、格式等 | 可忽略，作者自行决定 |

## 检查维度

| 维度 | 核心检查项 |
|------|-----------|
| 设计 | 代码归属、系统集成、无过度工程 |
| 功能 | 行为符合预期、边缘情况已处理 |
| 复杂度 | 代码可简化、易于理解 |
| 测试 | 有自动化测试、测试设计良好 |
| 安全 | 无 XSS、输入校验、敏感数据安全 |
| 性能 | 无内存泄漏、大列表虚拟滚动 |
| 后端 | Node.js 后端专项 / Golang 专项（见 checklist.md） |

## AI 生成代码评审

评审 AI 生成的变更时优先关注：行为回归与边界情况、安全假设与信任边界、隐藏耦合或架构漂移、不必要且推高模型成本的复杂度。成本意识检查详见 `./references/checklist.md`「AI 生成代码评审」小节。

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 路由表（Route First） | `./references/routing.md` |
| 门禁清单 | `./references/gates.md` |
| 完整检查清单（含 Node.js/Golang 专项、AI 生成代码评审）| `./references/checklist.md` |
| Git 场景指南 | `./references/git-scenarios.md` |
| 置信度与误报控制 | `./references/confidence-filtering.md` |
| 评分标准与批准映射 | `./references/scoring-standard.md` |
| 报告格式 | `./references/report-format.md` |
| 报告示例 | `./references/report-examples.md` |


---
## 📦 可用资源

- `./references/checklist.md`
- `./references/git-scenarios.md`
- `./references/confidence-filtering.md`
- `./references/report-examples.md`
- `./references/report-format.md`
- `./references/scoring-standard.md`
- `./references/writing-guidelines.md`
- `./assets/pre-commit-review.sh`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
