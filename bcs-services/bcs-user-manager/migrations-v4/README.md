# IAM V4 权限模型注册 Migration

对齐 IAM V3 `migrations/` 的 BCS 权限模型，按权限中心 V4（RBAC：System / ResourceType / Action / Role）注册。

V4 没有官方 upsert migrate API，由 `bcs-common/pkg/auth/iamv4.Client.Migrate` 读取本目录 JSON，
按 retrieve/list → create/update 做幂等注册。仅当 `iam_config.enable_v4=true` 时，
user-manager 启动会初始化 V4 client 并执行本目录 migration；默认关闭。现有 V3 migrate 保持不变。

`callback_url` 指向独立的 V4 Provider：`/bcsapi/v4/usermanager/v1/iamv4-provider/resources/`。
V3 仍使用 `/v1/iam-provider/resources/`，两套协议与 token 互不复用。

## 文件格式

`{version}_{name}.up.json`，version 从 `0` 开始。支持 go 模板，通过 `Migrate` 的 `templateVar` 渲染。

模板变量与 V3 一致：

| 变量 | 含义 |
|------|------|
| `BK_IAM_SYSTEM_ID` | 接入系统 ID |
| `APP_CODE` | 蓝鲸应用 ID（写入 clients） |
| `BCS_HOST` | BCS API 地址，用于 `callback_url` |

```json
{
  "system_id": "{{ .BK_IAM_SYSTEM_ID }}",
  "enabled": true,
  "operations": [
    {
      "operation": "upsert_system",
      "data": {
        "id": "{{ .BK_IAM_SYSTEM_ID }}",
        "name": "容器管理平台",
        "clients": ["{{ .APP_CODE }}", "bk_bcs_monitor", "bk_bcs", "bk_devops", "bk_harbor"],
        "callback_url": "{{ .BCS_HOST }}/bcsapi/v4/usermanager/v1/iamv4-provider/resources/"
      }
    }
  ]
}
```

支持的 operation：

- `upsert_system`
- `upsert_resource_type`
- `upsert_action`
- `upsert_role`（创建/更新角色，并同步 `actions`）

V3 的 instance_selection / action_groups / common_actions / resource_creator_actions 在 V4 不存在，对应物为 Role。

## 调用示例

```go
d := migrationsv4.MigrationFS
err := iamv4Client.Migrate(ctx, d, map[string]string{
    "BK_IAM_SYSTEM_ID": systemID,
    "APP_CODE":         appCode,
    "BCS_HOST":         bcsHost,
})
```
