## 六、服务配置（trpc_go.yaml）


### 6.1 配置文件结构

```yaml
global:
  namespace: Production      # 环境隔离：Production / Development
  env_name: formal

server:
  app: {app}
  server: {server}
  service:
    - name: trpc.{app}.{server}.{Service}   # 四段式服务名
      network: tcp
      protocol: trpc
      ip: 0.0.0.0
      port: 8000
      timeout: 1000

client:
  timeout: 1000
  service:
    - name: trpc.{downstream-app}.{downstream-server}.{Service}
      target: polaris://trpc.{downstream-app}.{downstream-server}.{Service}
      protocol: trpc
      timeout: 800

log:
  - writer: console
    level: debug
  - writer: file
    level: info
    writer_config:
      log_path: /usr/local/trpc/log/

plugins:
  registry:
    polaris:
      address_list: "polaris-server:8090"
```

### 6.2 配置规则

| 规则 | 说明 |
|------|------|
| 服务名四段式 | `server.service[].name` 必须遵守 `trpc.{app}.{server}.{Service}` 格式 |
| 禁止硬编码 IP | `client.service[].target` 生产环境必须使用 `polaris://` 前缀 |
| 敏感信息 | 密码、Token 通过环境变量或 Secret 注入，不写入 yaml 文件 |
| 启动校验 | 启动时校验必填配置，缺失则 Fatal 退出 |

---
