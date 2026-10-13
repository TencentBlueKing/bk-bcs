### 描述

获取模板文件变量

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |

### Body
```json
{
  "templateVersions": ["xxxx", "xxxx"],
  "clusterID": "BCS-K8S-12345",
  "namespace": "hito-test"
}
```


### 调用示例
```sh
curl -X POST -d 'your_body.json' -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/clusterresources/v1/projects/{project_code}/template/variables
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "9847f3f0-4b23-426a-bdbd-388672b0e17c",
  "data": {
    "vars": [
      {
        "key": "replicas",
        "readonly": false,
        "value": ""
      }
    ]
  },
  "webAnnotations": null
}
```