# 熵管理（Entropy Management）

> 目标：控制系统熵增速度，保持文档、代码与规范长期一致。

## 1. 文档园艺机制

### 1.1 自动一致性检测

| 检测项 | 频率 | 方式 | 责任人 |
|-------|------|------|-------|
| 文档引用路径是否存在 | 每次 PR / 按需 | harness-gardening 八维度扫描 | 自动化 |
| dev-map 与代码结构匹配 | 代码变更后 | harness-gardening 维度 8 | 自动化 |
| AGENTS.md 行数 ≤ 100 | 每次 harness 生成后 | 行数检查 | 自动化 |
| tooling.md 与 tool-dependencies 对齐 | 每次 harness 生成后 | 交叉验证 | 自动化 |
| Controller 注册完整性 | 新增 controller 时 | 对比 main.go SetupWithManager | 研发 |
| CRD 类型与 manifest 同步 | CRD 变更时 | 重新生成 deepcopy + manifest | 研发 |

### 1.2 园艺流程

1. 用户说「文档巡检」→ 触发 `harness-gardening`（mode=full）
2. P0 偏差（路径失效、必选文件缺失）→ 自动修复
3. P1 偏差（新增文件未入 dev-map）→ 列出方案请求确认
4. 修复后更新一致性报告

## 2. 架构违规检测

### 2.1 检测策略

| 检测类型 | 触发时机 | 工具 | 阻断级别 |
|---------|---------|------|---------|
| gofmt/goimports | PR 提交前 | `make test-ingress-controller` | 阻断合并 |
| 单元测试失败 | PR 提交前 | `go test` | 阻断合并 |
| 硬编码 Annotation | Code Review | 人工 + rg 搜索 | 警告 |
| 函数复杂度 > 10 | Code Review | 人工审查 | 警告 |
| 未注册 Controller | Code Review | 对比 main.go | 阻断合并 |

### 2.2 违规处理流程

- **阻断级别**：测试/编译失败，PR 不可合并
- **警告级别**：记入 Code Review 意见，当次修复或记录技术债
- **报告级别**：harness-gardening 输出，定期批量处理

## 3. 技术债追踪

### 3.1 追踪机制

| 债务类型 | 识别方式 | 记录位置 | 清理策略 |
|---------|---------|---------|---------|
| TODO/FIXME | `rg "TODO\|FIXME"` | Issue / 当次 PR | 每迭代 Review |
| 过时文档 | harness-gardening | 巡检报告 | 发现即修复 |
| 通用规范骨架 | standards Level 2 | `docs/standards/README.md` 待完善区 | 逐步填充或贡献预设库 |
| controller-runtime 版本老旧 | 技术评估 | specs/ 或 ADR | 大版本升级专项 |
| 历史拼写错误（ACESS_KEY） | 代码注释 | AGENTS.md 已记录 | 保持兼容，不修复 |

### 3.2 技术债预算

- 每个迭代新增 TODO/FIXME 不超过 5 条，超出需在 PR 说明原因
- 每迭代至少清理 2 条已有 TODO/FIXME
- harness-gardening P1 偏差超过 10 项时，优先文档修复再开发新功能

## 4. 熵增度量

### 4.1 度量指标

| 指标 | 计算方式 | 阈值 | 超标动作 |
|------|---------|------|---------|
| 文档一致性率 | harness-gardening 通过维度 / 8 | ≥ 87.5%（7/8） | 触发集中修复 |
| 测试通过率 | `make test-ingress-controller` | 100% | 阻断合并 |
| dev-map 覆盖率 | 源文件索引条目 / 实际 .go 文件数 | ≥ 90% | 更新 dev-map |
| 规范骨架比 | Level 2 规范数 / 总规范数 | ≤ 50% | 完善或贡献预设库 |

## 检查清单

- [x] 文档园艺检测机制已配置（harness-gardening）
- [x] 架构违规检测策略已定义
- [x] 技术债追踪机制已建立
- [x] 熵增度量指标已定义
- [x] PR 轻量检查脚本：`scripts/harness-gardening-pr.sh`
- [x] Cursor hooks：`.cursor/hooks.json`（`stop` 提醒 + `git commit` 后自动检查）
