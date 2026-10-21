### 描述

创建命名空间

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| clusterID         | string       | 是     | 集群ID     |

### Body
```json
{
    "name": "testnamespace",
    "quota": {
        "cpuRequests": "20",
        "memoryRequests": "20Gi",
        "cpuLimits": "20",
        "memoryLimits": "20Gi"
    },
    "labels": [
        {
            "key": "string",
            "value": "string"
        }
    ],
    "annotations": [
        {
            "key": "string",
            "value": "string"
        }
    ]
}
```
1. quota 为空则默认不创建 ResourceQuota
2. 共享集群 quota 为必填项
3. 共享集群不允许设置 labels 和 annotations

### 调用示例
```sh
curl -X POST -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/testproject/clusters/BCS-K8S-12345/namespaces
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": null,
    "requestID": "226ab8605c534cbc88ed021c36d5e256",
    "web_annotations": null
}
```