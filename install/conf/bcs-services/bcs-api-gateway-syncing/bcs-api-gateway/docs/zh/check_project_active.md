### 描述

检测项目活跃度

### 路径参数
| 参数名称         | 参数类型 | 必选 | 描述           |
| ---------------- | -------- | ---- | -------------- |
| project_code     | string   | 是   | 项目 ID 或项目英文名 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/testproject/active
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "isActive": true
    },
    "requestID": "894f7249f43c4e5xxxx9207058045e8e"
}
```
