# Resource Splitter Tool

这是一个用于将合并后的 JSON 资源文件拆分回独立 YAML 文件的命令行工具。

## 功能

- 将合并的 JSON 资源文件按类型拆分到不同目录
- 每个资源保存为独立的 YAML 文件
- 支持自定义输出目录
- 支持覆盖已存在的文件
- 智能文件名生成（基于资源名称、ID 等）

## 构建

```bash
# 构建拆分工具
make build-split
```

## 使用方法

### 基本用法

```bash
# 拆分 JSON 文件到默认目录 (./split-resources)
./bin/resource-splitter -input merged-resources.json

# 指定输出目录
./bin/resource-splitter -input merged-resources.json -output ./apisix-config

# 覆盖已存在的文件
./bin/resource-splitter -input merged-resources.json -output ./apisix-config -overwrite
```

### 参数说明

- `-input`: 输入的 JSON 文件路径（必需）
- `-output`: 输出目录路径（默认：./split-resources）
- `-overwrite`: 是否覆盖已存在的文件（默认：false）
- `-help`: 显示帮助信息

## 输出结构

工具会将 JSON 文件按以下结构拆分：

```
output-dir/
├── service/
│   ├── usermanager.yaml
│   ├── clustermanager-http.yaml
│   └── ...
├── route/
│   ├── clusterresources-grpc.yaml
│   ├── clusterresources-http.yaml
│   └── ...
├── plugin_config/
│   ├── bcs-http-config.yaml
│   ├── bcs-grpc-config.yaml
│   └── ...
├── ssl/
│   └── default-cert.yaml
└── upstream/
    ├── service-clusterresources-http.yaml
    └── ...
```

## 文件名生成规则

工具会按以下优先级生成文件名：

1. **name 字段**：如果资源有 `name` 字段，使用该值
2. **id 字段**：如果没有 `name`，使用 `id` 字段
3. **resource_id 字段**：如果都没有，使用 `resource_id` 字段
4. **索引**：如果都没有，使用 `resource-{index}` 格式

文件名会自动清理特殊字符（如 `/`、`\`、`:`、`*` 等），确保文件系统兼容性。

## 示例

### 输入 JSON 文件结构

```json
{
  "service": [
    {
      "id": "bk.s.gaakxUAAOQ",
      "name": "usermanager",
      "upstream": {
        "type": "roundrobin",
        "nodes": [
          {
            "host": "bcs-user-manager",
            "port": 8080,
            "weight": 1
          }
        ]
      }
    }
  ],
  "route": [
    {
      "resource_type": "route",
      "resource_id": "bk.r.ga6.P_AAOQ",
      "name": "clusterresources-grpc",
      "config": {
        "uris": ["/clusterresources.ClusterResources"],
        "methods": ["GET", "POST"]
      }
    }
  ]
}
```

### 输出文件

**service/usermanager.yaml**:
```yaml
id: bk.s.gaakxUAAOQ
name: usermanager
upstream:
  type: roundrobin
  nodes:
    - host: bcs-user-manager
      port: 8080
      weight: 1
```

**route/clusterresources-grpc.yaml**:
```yaml
resource_type: route
resource_id: bk.r.ga6.P_AAOQ
name: clusterresources-grpc
config:
  uris:
    - /clusterresources.ClusterResources
  methods:
    - GET
    - POST
```

## 错误处理

- 如果输入文件不存在，工具会报错并退出
- 如果输出目录无法创建，工具会报错并退出
- 如果 JSON 格式不正确，工具会报错并退出
- 默认情况下，如果文件已存在，工具会跳过（使用 `-overwrite` 强制覆盖）

## 与合并工具的配合使用

这个拆分工具与 `ResourceMerger` 工具配合使用，可以实现：

1. **合并**：将多个 YAML 文件合并为单个 JSON 文件
2. **拆分**：将合并的 JSON 文件拆分回独立的 YAML 文件

这样的工作流程特别适合：
- 配置文件的版本控制
- 配置的备份和恢复
- 配置的迁移和部署
- 配置的调试和分析
