## 六、服务配置（trpc_go.yaml）


```yaml
global:
  namespace: Production      # 环境隔离：Production / Development
  env_name: formal

server:
  app: {app}
  server: {server}
  service:
    - name: trpc.{app}.{server}.{Service}
      network: tcp
      protocol: trpc
      ip: 0.0.0.0
      port: 8000
      timeout: 1000

client:
  service:
    # LLM 服务：timeout 必须设为 0，由 Agent 框架层控制超时
    - name: hunyuan-pro
      target: polaris://hunyuan-pro
      protocol: http
      timeout: 0                         # ⚠️ LLM 客户端 timeout 必须为 0

    # 普通下游服务：正常设置超时
    - name: trpc.{downstream-app}.{downstream-server}.{Service}
      target: polaris://trpc.{downstream-app}.{downstream-server}.{Service}
      protocol: trpc
      timeout: 800

plugins:
  registry:
    polaris:
      address_list: "polaris-server:8090"
```

### 配置规则

| 规则 | 说明 |
|------|------|
| LLM timeout 为 0 | LLM 客户端的 `timeout` 必须设为 `0`，超时由 `llmagent.WithTimeout` 控制 |
| API Key 禁止入库 | 凭证必须通过环境变量注入，禁止写入 yaml 文件或代码 |
| 四段式服务名 | `server.service[].name` 遵守 `trpc.{app}.{server}.{Service}` 格式 |

---
