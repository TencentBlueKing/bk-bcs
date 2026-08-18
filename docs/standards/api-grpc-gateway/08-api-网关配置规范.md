## 八、API 网关配置规范

### 8.1 配置文件结构

```yaml
# bk_apigw_resources.yaml
- path: /api/v1/users
  method: GET
  operationId: list_users
  description: 获取用户列表
  labels:
    - 用户管理
  backend:
    method: GET
    path: /api/v1/users
  auth:
    required: true

- path: /api/v1/users/{user_id}
  method: GET
  operationId: get_user
  description: 获取用户详情
  labels:
    - 用户管理
  backend:
    method: GET
    path: /api/v1/users/{user_id}
  auth:
    required: true
```

### 8.2 关键规则

| 规则 | 说明 |
|------|------|
| operationId 用 snake_case | `list_users` 而不是 `listUsers` |
| backend 路径与前端一致 | 前后端使用相同 URL |
| 路径参数直接使用 proto 字段名 | `{user_id}` 对应 message 中的字段 |
| 文档文件名 = operationId | `list_users.md` |

### 8.3 API 文档模板

```markdown
# {操作描述}

## 请求参数

### 路径参数

| 参数 | 类型 | 必选 | 描述 |
|------|------|------|------|
| user_id | string | 是 | 用户 ID |

### 查询参数

| 参数 | 类型 | 必选 | 描述 |
|------|------|------|------|
| page | int64 | 否 | 页码，默认 1 |
| page_size | int64 | 否 | 每页数量，默认 20 |

## 响应示例

```json
{
  "data": {
    "items": [...],
    "total": 100
  }
}
```
```
