# BCS Mesh Proxy

BCS Mesh Proxy 是一个用于在 Kubernetes 集群中代理 Istio 资源 API 请求的工具。它允许您在一个集群中通过 kubectl 访问另一个集群的 Istio 资源。

## 功能特性

- 代理 Istio 相关的 API 请求
- 支持 networking.istio.io、security.istio.io、telemetry.istio.io 等 API 组
- 支持 Bearer Token 和客户端证书认证
- **自动Token轮转**: 使用client-go内置的ServiceAccount token自动轮转机制
- 可配置的请求过滤和访问控制
- 优雅的启动和关闭

## 架构说明

```
控制面集群                   目标集群（有Istio资源）
┌─────────────────┐         ┌─────────────────┐
│                 │         │                 │
│  kubectl        │────────▶│  mesh-proxy     │
│                 │         │                 │
│  APIService     │         │  ┌─────────────┐│
│                 │         │  │  k8s API    ││
└─────────────────┘         │  └─────────────┘│
                            │                 │
                            └─────────────────┘
```
