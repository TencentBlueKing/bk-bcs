### 描述

预览 releas

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| clusterID         | string       | 是     | 集群 ID     |
| namespace         | string       | 是     | 命名空间名称     |
| name         | string       | 是     | release 名称     |

### Body
```json
{
  "repository": "projecttest",
  "version": "0.0.1",
  "chart": "nginx",
  "values": [
    "replicas: 1"
  ],
  "args": [
    "--timeout=600s"
  ]
}
```


### 调用示例
```sh
curl -X POST -d 'your_body.json' -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/uat/helmmanager/v1/projects/projecttest/clusters/BCS-K8S-00000/namespaces/ns-test/releases/release-test/preview
```

### 响应示例
```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "data": {
    "newContents": {
      "additionalProp1": {
        "name": "string",
        "path": "string",
        "content": "string"
      },
      "additionalProp2": {
        "name": "string",
        "path": "string",
        "content": "string"
      },
      "additionalProp3": {
        "name": "string",
        "path": "string",
        "content": "string"
      }
    },
    "oldContents": {
      "additionalProp1": {
        "name": "string",
        "path": "string",
        "content": "string"
      },
      "additionalProp2": {
        "name": "string",
        "path": "string",
        "content": "string"
      },
      "additionalProp3": {
        "name": "string",
        "path": "string",
        "content": "string"
      }
    },
    "newContent": "string",
    "oldContent": "string"
  },
  "requestID": "42250a62-ae9c-4b1a-9dc7-b310cd96b5f9"
}
```