### 描述

更新 Deployment

### 路径参数

| 名称        | 位置   | 类型     | 必选 | 说明            |
|-----------|------|--------|----|---------------|
| projectID | path | string | 是  | 项目 ID         |
| clusterID | path | string | 是  | 集群 ID         |
| namespace | path | string | 是  | 命名空间          |
| name      | path | string | 是  | Deployment 名称 |

### Body 参数

```json
{
  "projectID": "string",
  "clusterID": "string",
  "namespace": "string",
  "name": "string",
  "rawData": {},
  "format": "string"
}
```

### 调用示例

```sh
curl -X PUT \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
-d '{}' \
'http://bcs-api.bkdomain/bcsapi/v4/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/deployments/{name}'
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "requestID": "",
  "data": {},
  "webAnnotations": {}
}
```