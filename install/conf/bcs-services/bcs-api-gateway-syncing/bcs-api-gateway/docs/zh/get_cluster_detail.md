### 描述

获取集群详情

### 路径参数
| 参数名称     | 参数类型 | 必选 | 描述   |
| ------------ | -------- | ---- | ------ |
| clusterID    | string   | 是   | 集群 ID |

### 查询参数
| 参数名称     | 参数类型 | 必选 | 描述 |
| ------------ | -------- | ---- | ---- |
| cloudInfo    | bool     | 否   | 是否返回云上扩展信息 |
| projectId    | string   | 否   | 项目 ID，用于权限或上下文校验 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure 'https://bcs-api-gateway.apigw.com/prod/clustermanager/v1/cluster/BCS-K8S-00000?cloudInfo=true&projectId=testproject'
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
        "clusterID": "BCS-K8S-00000",
        "clusterName": "test-cluster",
        "projectID": "1xxx3xxx5xxx4xxx8xxx1xxx2xxxexxx",
        "provider": "tencentCloud",
        "region": "ap-guangzhou",
        "status": "RUNNING"
    },
    "extra": {},
    "web_annotations": {
        "perms": null
    }
}
```
