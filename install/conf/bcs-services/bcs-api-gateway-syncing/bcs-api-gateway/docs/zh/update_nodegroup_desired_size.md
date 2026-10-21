### 描述

更新DesiredSize信息

更新NodeGroup中autoscaling信息的DesiredSize节点数，主要是修正数据不一致情况。适用于 Agent 需要执行写操作或修改资源的场景。需要提供必要的路径参数和请求体。返回结果可用于判断操作是否成功。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerUpdateGroupDesiredSizeBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateGroupDesiredSizeOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateGroupDesiredSizeDefaultBody

##### 内联模型

**ClusterManagerUpdateGroupDesiredSizeBody**


> 修改NodeGroup期望的节点数,扩容前保证数据一致性



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desiredSize | int64 (formatted integer)| 是 | 修改nodeGroup |  |
| nodeGroupID | string| 是 | NodeGroupID |  |
| operator | string|  | 操作人员 |  |



**ClusterManagerUpdateGroupDesiredSizeDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateGroupDesiredSizeDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateGroupDesiredSizeDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateGroupDesiredSizeOKBody**


> 更新NodeGroup期望节点数的请求结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerUpdateGroupDesiredSizeOKBodyWebAnnotations|  |  |  |



**ClusterManagerUpdateGroupDesiredSizeOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
