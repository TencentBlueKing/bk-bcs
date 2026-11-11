### 描述

查询用户的 BCS API token 列表。该接口需要用户认证，且只能查询当前认证用户自己的 token。

### 路径参数
| 参数名称 | 参数类型 | 必选 | 描述 |
| -------- | -------- | ---- | ---- |
| version  | string   | 是   | BCS API 版本，通常为 v4 |
| username | string   | 是   | 用户名，需与当前认证用户一致 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure 'https://bcs-api-gateway.apigw.com/prod/v4/usermanager/v1/users/admin/tokens'
```

### 响应示例
```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": [
        {
            "token": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
            "status": 1,
            "expired_at": "2027-01-01T00:00:00+08:00"
        }
    ]
}
```

### 响应参数说明
| 参数名称   | 参数类型 | 描述 |
| ---------- | -------- | ---- |
| token      | string   | BCS API token |
| tenant_id  | string   | 租户 ID，多租户模式下返回 |
| status     | int      | token 状态，0 为已过期，1 为有效 |
| expired_at | string   | 过期时间，为 null 表示永不过期 |
