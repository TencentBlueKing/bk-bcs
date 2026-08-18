# Commit 与代码提交规范

由 tapd-story-commit 阶段使用。

## Commit Message 格式

采用 Conventional Commits 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### type 取值

| type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 重构（既非新功能也非修复） |
| `docs` | 文档变更 |
| `test` | 测试相关 |
| `chore` | 构建过程或辅助工具变更 |
| `perf` | 性能优化 |
| `style` | 代码格式（不影响逻辑） |

### scope

变更涉及的模块名，例如 `auth`、`order`、`user` 等。

### subject

一句话概括变更目的，使用祈使语气（如"add"而非"added"）。

### body

详细说明（必须使用中文）：
- 变更了什么
- 为什么需要变更
- 关联的需求 ID（格式：`--story=${STORY_ID}`）

### footer

关联信息：
- `--story=${STORY_ID}` — 关联的 TAPD 需求 ID

## Commit Message 示例

```
feat(user): 实现登录超时检测

增加会话过期检测和自动重定向到登录页面。
处理并发会话失效的边缘情况。

--story=1001001
```

## 提交前检查

1. 确认所有测试通过
2. 确认代码已通过 lint 检查
3. 确认没有遗留 debug 代码或 TODO
4. 确认变更范围与需求一致（不包含无关改动）
