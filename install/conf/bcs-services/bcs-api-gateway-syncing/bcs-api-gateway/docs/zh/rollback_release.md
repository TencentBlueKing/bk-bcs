### 描述

回滚 release，该接口是异步接口，需通过 release 列表接口或详情接口获取 release 状态

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
  "revision": 1
}
```


### 调用示例
```sh
curl -X PUT -d 'your_body.json' -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/BCS-K8S-00000/namespaces/ns-test/releases/release-test/rollback
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "requestID": "a2a46e9a-9c4d-4547-8b40-82373ce0b9ff"
}
```