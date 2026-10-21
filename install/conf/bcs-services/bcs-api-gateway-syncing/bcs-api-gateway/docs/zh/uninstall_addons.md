### 描述

卸载集群组件

删除组件所有内容。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| name | `path` | string | 是 | [path_param参数] 组件名称。此参数在 URL 路径中传递，请替换为具体值。 |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 "k8s-demo"、"my-project"）。此参数在 URL 路径中传递，请替换为实际项目编码。 |
| body | `body` | ClusterAddonsUninstallAddonsBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterAddonsUninstallAddonsOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterAddonsUninstallAddonsDefaultBody

##### 内联模型

**ClusterAddonsUninstallAddonsBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 所在的集群ID |  |
| name | string|  | 组件名称 |  |
| projectCode | string|  | 项目代码 |  |



**ClusterAddonsUninstallAddonsDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterAddonsUninstallAddonsDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterAddonsUninstallAddonsDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterAddonsUninstallAddonsOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | requestID |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterAddonsUninstallAddonsOKBodyWebAnnotations|  |  |  |



**ClusterAddonsUninstallAddonsOKBodyWebAnnotations**


> 权限信息返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
