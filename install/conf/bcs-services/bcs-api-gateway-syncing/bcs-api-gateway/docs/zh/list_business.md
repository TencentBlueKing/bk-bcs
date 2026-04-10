### 描述

获取业务列表

### 查询参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| useBCS       | bool         | 否     | 是否仅返回已启用 BCS 的业务 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure 'https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/business?useBCS=true'
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "businessID": "100000",
            "name": "test-biz",
            "maintainer": ["testuser"]
        }
    ],
    "requestID": "894f7249f43c4e5xxxx9207058045e8e",
    "web_annotations": {
        "perms": null
    }
}
```
