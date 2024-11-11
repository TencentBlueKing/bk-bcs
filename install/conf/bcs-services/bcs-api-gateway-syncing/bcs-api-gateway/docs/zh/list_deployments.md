### 描述

获取 Deployment 列表

### 路径参数

| 名称            | 位置    | 类型     | 必选 | 中文名 | 说明                           |
|---------------|-------|--------|----|-----|------------------------------|
| projectID     | path  | string | 是  |     | 项目 ID                        |
| clusterID     | path  | string | 是  |     | 集群 ID                        |
| namespace     | path  | string | 是  |     | 命名空间                         |
| labelSelector | query | string | 否  |     | 标签选择器                        |
| apiVersion    | query | string | 否  |     | apiVersion                   |
| ownerName     | query | string | 否  |     | 所属资源名称                       |
| ownerKind     | query | string | 否  |     | 所属资源类型                       |
| format        | query | string | 否  |     | 资源配置格式（manifest/selectItems） |
| scene         | query | string | 否  |     | 使用场景 仅 selectItems 格式下有效     |

### 调用示例

```sh
curl -X GET \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
'http://bcs-api.bkdomain/bcsapi/v4/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/deployments'
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