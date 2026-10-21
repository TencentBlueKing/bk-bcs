### 描述

更新命名空间

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| namespace | `path` | string | 是 | [path_param参数] Kubernetes 命名空间名称（长度不超过 63 字符，仅小写字母、数字、连字符，以字母开头）。此参数在 URL 路径中传递，请替换为具体值。 |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 "k8s-demo"、"my-project"）。此参数在 URL 路径中传递，请替换为实际项目编码。 |
| body | `body` | NamespaceUpdateNamespaceBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

NamespaceUpdateNamespaceOKBody

#### 默认响应
An unexpected error response
##### SchemaNamespaceUpdateNamespaceDefaultBody

##### 内联模型

**NamespaceUpdateNamespaceBody**


> 更新命名空间配额请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | []NamespaceUpdateNamespaceParamsBodyAnnotationsItems0|  | 命名空间注解 |  |
| clusterID | string| 是 | 集群ID |  |
| labels | []NamespaceUpdateNamespaceParamsBodyLabelsItems0|  | 命名空间标签 |  |
| namespace | string| 是 | 命名空间名称, 长度不能超过63字符,只能包含小写字母、数字，以及 '-',必须以字母开头，不能以 '-' 结尾 |  |
| projectCode | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| quota | NamespaceUpdateNamespaceParamsBodyQuota|  |  |  |



**NamespaceUpdateNamespaceDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []NamespaceUpdateNamespaceDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**NamespaceUpdateNamespaceDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**NamespaceUpdateNamespaceOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID |  |
| web_annotations | NamespaceUpdateNamespaceOKBodyWebAnnotations|  |  |  |



**NamespaceUpdateNamespaceOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |



**NamespaceUpdateNamespaceParamsBodyAnnotationsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | key |  |
| value | string|  | value |  |



**NamespaceUpdateNamespaceParamsBodyLabelsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | key |  |
| value | string|  | value |  |



**NamespaceUpdateNamespaceParamsBodyQuota**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  | limits.cpu |  |
| cpuRequests | string|  | requests.cpu |  |
| memoryLimits | string|  | limits.cpu |  |
| memoryRequests | string|  | requests.memory |  |
