### 描述

更新minSize/maxSize信息

更新NodeGroup中autoscaling信息的最大最小限额，主要是协调多nodeGroup资源。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerUpdateGroupMinMaxSizeBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateGroupMinMaxSizeOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateGroupMinMaxSizeDefaultBody

##### 内联模型

**ClusterManagerUpdateGroupMinMaxSizeBody**


> 修改NodeGroup最小最大的扩容限额



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| maxSize | int64 (formatted integer)|  | 修改nodeGroup的autoscaling扩容上限 |  |
| minSize | int64 (formatted integer)|  | 修改nodeGroup的autoscaling扩容下限 |  |
| nodeGroupID | string| 是 | NodeGroupID |  |
| operator | string|  | 操作人员 |  |



**ClusterManagerUpdateGroupMinMaxSizeDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateGroupMinMaxSizeDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateGroupMinMaxSizeDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateGroupMinMaxSizeOKBody**


> 修改NodeGroup最小最大的扩容限额的请求结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerUpdateGroupMinMaxSizeOKBodyWebAnnotations|  |  |  |



**ClusterManagerUpdateGroupMinMaxSizeOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
