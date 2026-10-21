### 描述

同步命名空间

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 "k8s-demo"、"my-project"）。此参数在 URL 路径中传递，请替换为实际项目编码。 |
| body | `body` | NamespaceSyncNamespaceBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

NamespaceSyncNamespaceOKBody

#### 默认响应
An unexpected error response
##### SchemaNamespaceSyncNamespaceDefaultBody

##### 内联模型

**NamespaceSyncNamespaceBody**


> 创建命名空间请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string| 是 | 集群ID |  |
| projectCode | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |



**NamespaceSyncNamespaceDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []NamespaceSyncNamespaceDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**NamespaceSyncNamespaceDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**NamespaceSyncNamespaceOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | string|  | 返回详细信息, 可以为空 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID |  |
