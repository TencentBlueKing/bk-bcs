# AGENTS.md

> Agent 认知本项目的第一站——快速了解项目全貌、找到代码和规范、知道规矩。

## 项目概述

- **项目名称**：bk-bcs（蓝鲸容器调度平台 / BlueKing Container Service）
- **仓库地址**：https://github.com/Tencent/bk-bcs
- **定位**：连接云原生技术与复杂业务场景的容器编排与服务治理平台；纳管/创建 K8s 集群，提供游戏等复杂应用的一站式编排能力。
- **目标**：在本 monorepo 内安全、可验证地修改指定组件；局部约定优先于根。

## 目录结构

```
bcs-common/     — 共享库（日志、配置、加密、API 客户端等）
bcs-runtime/    — 运行时：K8s Operator / Mesos 组件
bcs-services/   — 控制面微服务（集群、项目、Helm、监控等）
bcs-ui/         — 控制台前端（Vue 2 + webpack）
bcs-scenarios/  — 场景化服务
bcs-ops/        — 运维脚本
docs/           — 产品/特性文档 + harness / standards
install/        — 部署配置与 Helm chart
scripts/        — 构建与工具脚本
```

## 关键规范

- Harness 规范（工具能力、架构约束等）→ [`docs/harness/README.md`](docs/harness/README.md)
- 技术开发规范 → [`docs/standards/README.md`](docs/standards/README.md)
- 词汇表 → [`docs/glossary.md`](docs/glossary.md)
- 架构总览 → [`docs/overview/README.md`](docs/overview/README.md)
- 贡献与提交 → [`CONTRIBUTING.md`](CONTRIBUTING.md)、[`docs/specification/commit-spec.md`](docs/specification/commit-spec.md)
- 开发地图：需安装 graphify 后生成本地图谱；当前未启用，见总结报告

## 局部入口（工作单元 AGENTS）

修改某路径前，阅读该路径向上最近的 `AGENTS.md`；**局部约定优先于根**。根不替代工作单元入口。

| 路径 | 角色 |
|------|------|
| [`bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/AGENTS.md`](bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/AGENTS.md) | Ingress Controller：网络扩展 CRD Operator |
| [`bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/AGENTS.md`](bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/AGENTS.md) | 容灾预案 Controller（Kubebuilder） |
| [`bcs-scenarios/bcs-terraform-bkprovider/AGENTS.md`](bcs-scenarios/bcs-terraform-bkprovider/AGENTS.md) | Terraform Provider：go-micro gRPC 代理 |

## 编码前必读（门闩）

写或改**业务代码**前（非纯文档/纯问答）：
1. 打开 `docs/standards/README.md`，确认「当前项目选用的规范」与「加载预算」。
2. 按预算表与「章节快速索引」**只 Read 本任务相关章节**（可用行号/偏移；**禁止**无差别灌入整份长规范）。
3. 未选用的端：不得假装存在规范；按 README「未覆盖的技术栈」处理或向用户确认。
4. 提交/宣称完成前：按相关节的检查清单自检，并运行**当前工作单元 / 仓库已有**的 `lint`/`test`/`build` 脚本（无脚本不得声称已通过）。
