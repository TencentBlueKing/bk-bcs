# 评审报告示例

## Git 场景评审报告 (merge)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
代码评审报告 - Git 场景
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

项目: my-frontend
评审类型: merge review
评审时间: 2026-01-20 10:30:00

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Git 变更信息
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

评审类型: 分支合并评审 (merge)
源分支: feature/user-profile
目标分支: main
变更文件: 15 个
新增行数: +342
删除行数: -128
涉及 Commits: 5 个

Commits 列表:
  - a1b2c3d feat: add user profile component
  - e4f5g6h fix: fix profile image loading
  - i7j8k9l style: update profile styles
  - m0n1o2p test: add profile unit tests
  - q3r4s5t docs: update profile documentation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
变更文件列表
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📁 src/components/
   ├── UserProfile.vue        (+120, -30)
   ├── ProfileImage.vue       (+45, -10)
   └── ProfileSettings.vue    (+80, -25)

📁 src/utils/
   └── profile-helper.ts      (+35, -8)

📁 src/tests/
   └── UserProfile.spec.ts    (+62, -55)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Git 规范检查
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ GIT-001: Commit message 格式正确
✅ GIT-002: 变更范围合理 (342 行 < 400 行)
✅ GIT-003: 未发现潜在合并冲突
✅ GIT-004: 未发现敏感信息
⚠️ GIT-005: 发现 1 处调试代码
   文件: src/components/UserProfile.vue:85
   代码: console.log('debug:', userData);
✅ GIT-006: Commits 遵循单一职责原则

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
评审结论
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 结论: 需要小修改

🔴 CRITICAL: 0 个问题
🟠 HIGH: 0 个问题
🟡 MEDIUM: 2 个问题
⚪ LOW: 3 个问题
⚠️ Git 规范: 1 个警告

📊 代码质量评分: 82/100

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
下一步行动
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 移除调试代码 (console.log)
2. 处理 2 个 MEDIUM 问题
3. 可选：处理 3 个 LOW 优化建议
4. 修复完成后可安全合并到 main 分支
```

## 暂存区评审报告 (staged)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pre-commit 代码评审报告
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

项目: my-frontend
评审类型: staged (暂存区)
评审时间: 2026-01-20 15:45:00

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
暂存区变更
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

变更文件: 3 个
新增行数: +45
删除行数: -12

变更文件列表:
  M src/components/Header.vue    (+25, -8)
  M src/utils/format.ts          (+15, -4)
  A src/components/NewFeature.vue (+5, -0)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
提交前检查
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ 无敏感文件
✅ 无调试代码
✅ 变更范围合理
✅ 代码规范检查通过

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
评审结论: ✅ LGTM - 可以提交
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 严重级别汇总表 + 结论示例

> 严重级别统一使用 CRITICAL/HIGH/MEDIUM/LOW 四级（见 `./checklist.md`），下方为汇总表 + 结论样式；权威报告格式仍以 `./report-format.md` 为准，本表仅为一种补充呈现示例。

单条问题格式：

```
[CRITICAL] 源码中硬编码 API Key
文件: src/api/client.ts:42
问题: API Key "sk-abc..." 暴露在源码中，将被提交进 git 历史。
修复: 移至环境变量，并加入 .gitignore/.env.example

  const apiKey = "sk-abc123";           // 反例
  const apiKey = process.env.API_KEY;   // 正例
```

结尾汇总：

```
## 评审总结

| 严重级别 | 数量 | 状态 |
|----------|------|------|
| CRITICAL | 0    | pass |
| HIGH     | 2    | warn |
| MEDIUM   | 3    | info |
| LOW      | 1    | note |

结论: WARNING —— 2 个 HIGH 问题应在合并前解决。
```

> 结论三态（Approve / Warning / Block）与评分等级、一票否决项的映射见 `./scoring-standard.md`。
