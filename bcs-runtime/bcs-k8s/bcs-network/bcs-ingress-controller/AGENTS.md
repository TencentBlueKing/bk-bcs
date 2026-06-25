# AGENTS.md — bcs-ingress-controller

> Agent 认知本项目的第一站。详细规范见 `docs/harness/`，代码索引见 `docs/dev-map/`。

## 项目概述

- **名称**：BCS Ingress Controller
- **定位**：Kubernetes Operator，管理网络扩展 CRD（Ingress、Listener、PortPool、PortBinding、HostNetPortPool）
- **技术栈**：Go 1.20+、controller-runtime v0.6.3、go-restful、Prometheus
- **模块路径**：go.mod 在 `../`（bcs-network/）；CRD 类型在 `../../kubernetes/apis/networkextension/v1/`

## 目录结构

```
bcs-ingress-controller/
├── main.go                     # 入口：注册 Controller / HTTP / Webhook
├── {name}controller/           # 每个 CRD 一个 Reconcile 控制器
├── internal/
│   ├── constant/               # Annotation Key 与共享常量
│   ├── option/                 # CLI 参数
│   ├── generator/              # Ingress → Listener 转换
│   ├── cloud/                  # 多云 LB 适配（aws/azure/gcp/tencentcloud）
│   ├── httpsvr/                # REST 管理 API
│   ├── webhookserver/          # Admission Webhook
│   ├── portpoolcache/          # PortPool 内存缓存
│   ├── hostnetportpoolcache/   # HostNetPortPool 内存缓存
│   ├── metrics/                # Prometheus 指标（namespace bkbcs_ingressctrl）
│   └── check/                  # 周期性一致性检查
├── docs/
│   ├── harness/                # AI Agent 运行环境规范（五大组件）
│   ├── standards/              # 技术开发规范（安全/质量/后端/接口）
│   ├── dev-map/                # 开发地图（源文件/模块/依赖索引）
│   ├── adr/                    # 架构决策记录（ADR）
│   ├── reqs/                   # 需求文档（TAPD 澄清产出）
│   └── glossary.md             # 词汇表
├── specs/                      # 功能设计文档
├── cli-util/                   # 独立 CLI 工具（如 validate-listener-name）
├── project.json                # TAPD workspace_id / owner 配置
└── bcs-ingress-inspector/      # 独立诊断二进制（非主 Controller）
```

## 关键规范

| 类型 | 入口 |
|------|------|
| Harness 规范（架构约束、工具能力、执行验证） | [docs/harness/README.md](docs/harness/README.md) |
| 技术开发规范（安全红线、代码评审、后端/API） | [docs/standards/README.md](docs/standards/README.md) |
| 开发地图（模块索引与依赖） | [docs/dev-map/README.md](docs/dev-map/README.md) |
| 架构决策记录（ADR） | [docs/adr/README.md](docs/adr/README.md) |
| 词汇表 | [docs/glossary.md](docs/glossary.md) |

## 构建与测试

```bash
cd .. && make ingress-controller          # 构建
cd .. && make test-ingress-controller     # 全量测试 + 覆盖率
go test -v -run TestReconcile ./hostnetportcontroller/...  # 单包测试
```

部署：`kubectl rollout restart -n bcs-system deployment/bcsingresscontroller`

## 核心约定（速查）

- 日志用 `bcs-common/common/blog`，禁止 log/klog
- Annotation Key 放 `internal/constant/constant.go`，禁止硬编码
- Controller Reconcile 必须幂等；新 Controller 须在 `main.go` 注册
- 导出函数须有 GoDoc 注释（英文）
- 函数名 ≤ 35 字符（含测试函数）；圈复杂度 > 10 须拆分
- `initClient` 保持纯分发，云初始化放 `initXxxClient`
- 测试：表驱动，colocated `*_test.go`，fake client

## 全流程覆盖规划

| 环节 | Skill / 入口 | 状态 |
|------|-------------|------|
| Harness 规范生成 | harness-engineering → harness-generating | ✅ 已建设 |
| 文档一致性巡检 | harness-engineering → harness-gardening | ✅ 已建设 |
| 需求澄清/评估 | tapd-story-clarification / tapd-story-evaluation | ✅ 已建设（TAPD workspace 70046748） |
| 迭代研发 | tapd-iteration-runner / tapd-story-pipeline | ✅ 已建设（workspace 70046748，owner adelaidahe） |
| Spec Kit TDD | speckit-specify → plan → tasks → implement | ✅ 已建设 |
| 代码评审 | code-review | ✅ 已建设 |
| 安全检查 | bk-security-redlines | ✅ 已建设 |
| 工作汇总 | work-summary | ✅ 已建设 |

> Skill 清单、MCP 配置、环境检查结果详见 [docs/harness/tooling.md](docs/harness/tooling.md)。

## 已完成特性

- HostNetPortPool hostNetwork 端口动态分配（`specs/001-hostnet-port-allocation/`）
- Namespace Scope Exemption：白名单 NS 可跨 NS 引用 Service 并使用全局云凭证（详见 harness 架构约束文档）
- SSL 证书过期 Prometheus 指标：CertificateChecker 周期性查询腾讯云证书剩余天数（`specs/stories/1070046748135050873/`，默认关闭，需 CLI 开关启用）
