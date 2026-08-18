# 技术规范预设管理

本文档定义 harness-generating 管理 `assets/standards/` 预设库的规范。

## 命名规范

```
{category}-{stack}.md        # 特定技术栈预设入口（唯一 index.yaml file 字段）
{category}-{stack}/          # 可选：大预设的分册目录（与入口 {stem} 同名）
  01-xxx.md                  # 分册正文（按章节拆分）
```

- `category`：`frontend` / `api` / `backend`（固定三类）
- `stack`：技术栈标识，kebab-case（如 `vue3`、`trpc-go`、`grpc-gateway`）
- **不再使用** `{category}-generic.md`：未匹配栈时**不设**该分类规范（策略 `skip_unmatched`）

## 薄入口 + 分册（>400 行长预设）

大预设拆为**薄入口** + **分册目录**，降低 Agent 默认加载成本：

|  artifact | 路径 | 约束 |
|-----------|------|------|
| 入口 | `assets/standards/{name}.md` | `index.yaml` 的 `file` 仍指向此文件；**≤120 行** |
| 分册 | `assets/standards/{name}/*.md` | 原文章节正文；入口「分册目录」表链接 `./{name}/xx.md` |

**部署（generating / gardening 维度 4 覆写）**：

1. 复制入口 `{name}.md` → `docs/standards/{name}.md`
2. 若预设库存在 `assets/standards/{name}/`，**递归复制**到 `docs/standards/{name}/`（与入口成对，不可只复制入口）

入口须含：标题与短引言、「本次必读决策表」（若有）、「分册目录」、可选「章节快速索引」指针、「规范落地优先级」（可留入口或末册，优先入口）。

接入仓 `docs/standards/README.md` 的「章节快速索引」应汇总**入口 + 各分册**标题，供按节 Read。

## 新增预设流程

扩展新技术栈只需两步，无需修改 SKILL.md：

1. 在 `assets/standards/` 下新建规范文件（参考已有预设的章节结构）
2. 在 `index.yaml` 的对应 category 中添加 preset 条目（含 id、name、file、detect、tags）

## index.yaml 格式与 detect 语义

```yaml
categories:
  {category}:
    description: "分类描述"
    presets:
      - id: {stack-id}
        name: "展示名称"
        file: {category}-{stack}.md
        detect:
          match: all          # 默认 AND；多条 rules 必须全部满足才 Level-1
          rules:
            - file: "package.json"
              contains_dep: ["vue"]
              version_gte: "3.0.0"
            - any_of_files: ["vite.config.ts", "vite.config.js"]
            - file: "go.mod"
              contains_require: ["trpc.group/trpc-go"]
            - require_dirs: ["proto", "stub"]
        tags: [tag1, tag2]

fallback:
  strategy: skip_unmatched   # 未匹配 / planned：不写入该分类规范文件
  contribute_hint: "贡献引导文案"
```

**禁止**：对 `package.json` / `go.mod` 做全文 `contains: ["vue"]` 子串匹配。

**禁止**：未匹配时复制或生成 `*-generic.md` 到接入仓。

**预设正文**：不得写死业务仓目录或强制不存在的测试命令；目录/命令以仓库为准或生成时填「项目事实」。

## Monorepo / 子项目根探测（F9 / R2）

权威实现：`scripts/detect-standards.sh`（`--json` 可供 Agent 解析）。

1. **发现根**
   - frontend 候选：含 `package.json` 的目录（深度≤4；跳过 `node_modules`/`e2e`/`tests`/隐藏目录）
   - backend/api 候选：含 `go.mod` 的目录（同样跳过规则）
2. **规则求值**：`file` / `any_of_files` / `contains_dep` / `contains_require` / `require_dirs` 相对**候选根**（及仓库根）解析；`match: all` 时全部 rule 命中才 Level-1。
3. **直接依赖**：`contains_require` 忽略 `// indirect` 行。
4. **同一 `contains_require` 数组**：多项为 OR。
5. **每分类 primary**：第一条 Level-1 作为部署选用；其余命中写入报告「候选」。
6. **项目事实**：generating 写入 `docs/standards/README.md`（如「前端根：`apps/ui`」「后端根：`apps/bkms-server`」）。

## 规范落地优先级（写入预设文末）

长规范靠 Agent 记忆无法稳定执法。每个完整预设**文首**还应有 ≤15 行的「本次必读决策表」（禁止/必须/验证 + 章节指针），与文末优先级表交叉引用：决策表指向「详见文末规范落地优先级」，文末优先级节首行指回文首决策表。

每个完整预设文末应有「规范落地优先级」短表，把条款分成：

| 优先级 | 含义 | 典型落地 |
|--------|------|----------|
| P0 机器门禁 | 可脚本/CI 判定 | lint、format、类型检查、禁止密钥扫描、契约/codegen 校验 |
| P1 Agent 必读 | 写代码前 Read 本文件对应节 | 分层/目录约定、错误码、状态机、鉴权流程 |
| P2 参考 | 有理由可偏离 | 命名偏好、布局惯例、非强制性能建议 |

Harness 只生成短 IDE Rules 督促 Read；**不**把整本预设塞进 Always Rules。机器可检项应沉到仓库已有 scripts / CI，而不是散文「必须遵守」。

范例见 `assets/standards/frontend-vue2.md` / `frontend-vue3.md` / `frontend-vue3-webpack.md` 文末。
