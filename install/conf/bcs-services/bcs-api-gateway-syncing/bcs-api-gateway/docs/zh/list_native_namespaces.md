### 描述

创建命名空间

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectIDOrCode   | string       | 是     | 项目ID或项目英文名     |
| clusterID         | string       | 是     | 集群ID     |

1. 共享集群：projectIDOrCode 为 `-` 则列出集群下所有命名空间
2. 共享集群：projectIDOrCode 不为 `-` 则列出对应项目下的命名空间

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/testproject/clusters/BCS-K8S-12345/native/namespaces
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "uid": "c4eb240e-1234-1234-1234-6c41580839dc",
            "name": "ieg-testproject-test",
            "status": "Active",
            "createTime": "2022-09-01 02:58:19",
            "projectID": "xxxxxc8b5de5440a887a1c5126cxxxxx",
            "projectCode": "testproject"
        }
    ],
    "requestID": "9f5eb04c2cb1436cae9af75c4a43bd6f",
    "web_annotations": null
}
```