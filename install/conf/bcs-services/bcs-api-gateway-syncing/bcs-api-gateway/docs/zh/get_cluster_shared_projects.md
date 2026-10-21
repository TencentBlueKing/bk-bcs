### 描述

查询集群共享项目信息

查询指定集群的共享配置信息，包括共享给哪些项目使用。 适用于需要了解某个集群的共享状态或查看有哪些项目在使用该集群的场景。 需要提供 clusterID 路径参数。 返回共享项目列表，包含项目 ID 和共享权限配置。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerGetClusterSharedProjectOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerGetClusterSharedProjectDefaultBody

##### 内联模型

**ClusterManagerGetClusterSharedProjectDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerGetClusterSharedProjectDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerGetClusterSharedProjectDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerGetClusterSharedProjectOKBody**


> 查询集群共享项目响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | ClusterManagerGetClusterSharedProjectOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerGetClusterSharedProjectOKBodyData**


> `ListValue` is a wrapper around a repeated field of values. The JSON representation for `ListValue` is JSON array.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| values | []interface{}|  | Repeated field of dynamically typed values. |  |
