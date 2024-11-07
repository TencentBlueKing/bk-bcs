### 描述

创建 Deployment

### 路径参数

| 名称        | 位置   | 类型     | 必选 | 说明    |
|-----------|------|--------|----|-------|
| projectID | path | string | 是  | 项目 ID |
| clusterID | path | string | 是  | 集群 ID |

### Body 请求参数

```json
{
  "rawData": {},
  "format": "string"
}
```

| 名称      | 类型     | 必选 | 说明                        |
|---------|--------|----|---------------------------|
| rawData | body   | 是  | 资源配置信息                    |
| format  | string | 是  | 资源配置格式（manifest/formData） |

### 调用示例

```sh
curl -X POST \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
-d '{ "rawData": {}, "format": "" }' \
'http://bcs-api.bkdomain/bcsapi/v4/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/workloads/deployments'
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