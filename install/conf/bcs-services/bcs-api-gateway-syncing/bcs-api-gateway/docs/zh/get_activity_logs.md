### 描述

获取 BCS 操作记录

### 路径参数
| 参数名称     | 参数类型 | 必选 | 描述     |
| ------------ | -------- | ---- | -------- |
| project_code | string   | 是   | 项目英文名 |

### 查询参数
| 参数名称     | 参数类型 | 必选 | 描述 |
| ------------ | -------- | ---- | ---- |
| limit        | int      | 否   | 返回记录数量上限 |
| offset       | int      | 否   | 分页偏移量 |
| resourceType | string   | 否   | 资源类型过滤 |
| action       | string   | 否   | 操作类型过滤 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure 'https://bcs-api-gateway.apigw.com/prod/v4/usermanager/v3/projects/testproject/activity_logs?limit=20&offset=0'
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "count": 1,
        "results": [
            {
                "id": "1",
                "username": "admin",
                "project_code": "testproject",
                "resourceType": "cluster",
                "resourceName": "test-cluster",
                "activity_type": "update",
                "action": "create",
                "source_ip": "1.1.1.1",
                "status": "success",
                "created_at": "2024-01-01T00:00:00Z"
            }
        ]
    },
    "requestID": "894f7249f43c4e5xxxx9207058045e8e"
}
```
