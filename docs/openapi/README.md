# BCS OpenAPI 统一维护说明

本目录统一维护 BCS 各服务的 OpenAPI 3.0.1 规范文件，所有文件均输出到对应服务的子目录下。

## 目录结构

```
openapi/
├── README.md                                  # 本文件
├── generate.sh                                # 一键生成脚本
├── service_config.yaml                        # 服务配置（维护此处即可扩展）
├── scripts/
│   └── swagger2openapi_convert.py             # Swagger 2.0 → OpenAPI 3.0.1 转换工具
└── bcs-services/
    ├── bcs-cluster-manager/openapi.yaml       # proto 类，自动生成
    ├── bcs-helm-manager/openapi.yaml          # proto 类，自动生成
    ├── bcs-project-manager/openapi.yaml       # proto 类，自动生成
    ├── cluster-resources/openapi.yaml         # proto 类，自动生成
    ├── bcs-monitor/openapi.yaml               # swag 类，自动生成
    ├── bcs-user-manager/openapi.yaml          # restful 类，自动生成
    └── bcs-webconsole/openapi.yaml            # swag 类，原地手写维护
```

## 服务类型说明

| 类型 | 适用场景 | 维护方式 |
|------|----------|----------|
| `proto` | 服务有 `.proto` 文件且包含 `google.api.http` 注解 | 运行 `generate.sh` 自动生成 |
| `gorestful` | 服务使用 `go-restful` 框架，路由定义在 `router.go` 中 | 运行 `generate.sh` 自动生成 |
| `swag` | 服务使用 Go `swag` 注解（`@Summary`、`@Router` 等） | 运行 `generate.sh` 自动生成 |
| `manual` | 纯 gin/HTTP 路由，无法自动生成 | 直接编辑 `openapi.yaml` 文件 |

## 快速开始

### 生成所有服务

```bash
cd /root/code-fork/bk-bcs
./openapi/generate.sh
```

### 生成指定服务

```bash
./openapi/generate.sh --service bcs-cluster-manager
./openapi/generate.sh --service bcs-project-manager
```

### 按类型批量生成

```bash
# 仅生成 proto 类服务
./openapi/generate.sh --proto-only

# 仅生成 swag 类服务（需要安装 swag）
./openapi/generate.sh --swag-only

# 仅同步 manual 类服务
./openapi/generate.sh --manual-only
```

### 列出所有配置的服务

```bash
./openapi/generate.sh --list
```

## 前置依赖

| 工具 | 用途 | 安装命令 |
|------|------|----------|
| `protoc` | 编译 proto 文件 | 系统包管理器安装 |
| `protoc-gen-openapi` | 从 proto 生成 OpenAPI | `go install github.com/micro/micro/v3/cmd/protoc-gen-openapi@latest` |
| `swag` | 从 Go 注解生成 swagger | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `python3` + `pyyaml` | 脚本依赖 | `pip install pyyaml` |

## 新增服务

1. 在 `service_config.yaml` 中添加服务配置：

```yaml
services:
  my-new-service:
    type: proto          # 或 swag / gorestful / manual
    module: bcs-services
    source_dir: bcs-services/my-new-service
    proto_file: proto/my-service.proto
    proto_includes:
      - third_party
    title: "My New Service API"
    version: "0.0.1"

  my-gorestful-service:
    type: gorestful
    module: bcs-services
    source_dir: bcs-services/my-gorestful-service
    router_files:
      - app/v1http/router.go
      - app/v3http/router.go
    base_path: /myservice
    title: "My GoRestful Service API"
    version: "0.0.1"
```

2. 运行生成命令：

```bash
./openapi/generate.sh --service my-new-service
```

## manual 类服务维护

对于 `manual` 类服务（如 `bcs-webconsole`），直接编辑其 `openapi.yaml` 文件：

```
openapi/bcs-services/bcs-webconsole/openapi.yaml    ← 直接编辑此文件
```

对于从服务目录同步的 manual 类服务（如 `bcs-user-manager`），源文件在服务代码目录：

```
bcs-services/bcs-user-manager/openapi.yaml          ← 编辑此文件
openapi/bcs-services/bcs-user-manager/openapi.yaml  ← 运行 generate.sh --manual-only 同步
```

## 相关工具

- **描述注入**：`.cursor/skills/openapi-yaml-enhancer/scripts/inject_proto_descriptions.py`
  - 自动在 `generate.sh` 中调用，也可单独使用：
  ```bash
  python3 .cursor/skills/openapi-yaml-enhancer/scripts/inject_proto_descriptions.py \
    --config openapi/service_config.yaml --all
  ```

- **格式验证**：`.cursor/skills/openapi-yaml-enhancer/scripts/validate_openapi.py`
  ```bash
  python3 .cursor/skills/openapi-yaml-enhancer/scripts/validate_openapi.py \
    openapi/bcs-services/bcs-cluster-manager/openapi.yaml
  ```

- **swagger 转换**：`openapi/scripts/swagger2openapi_convert.py`
  ```bash
  python3 openapi/scripts/swagger2openapi_convert.py \
    bcs-services/bcs-monitor/docs/swagger.json \
    openapi/bcs-services/bcs-monitor/openapi.yaml
  ```
