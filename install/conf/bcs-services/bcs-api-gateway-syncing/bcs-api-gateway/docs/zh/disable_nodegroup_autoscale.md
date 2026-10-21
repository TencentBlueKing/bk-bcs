### 描述

关闭节点池节点自动扩缩容

关闭节点组的自动伸缩功能。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerDisableNodeGroupAutoScaleBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerDisableNodeGroupAutoScaleOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerDisableNodeGroupAutoScaleDefaultBody

##### 内联模型

**ClusterManagerDisableNodeGroupAutoScaleBody**


> 关闭NodeGroup自动扩缩容



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| nodeGroupID | string| 是 | NodeGroupID |  |



**ClusterManagerDisableNodeGroupAutoScaleDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerDisableNodeGroupAutoScaleDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerDisableNodeGroupAutoScaleDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerDisableNodeGroupAutoScaleOKBody**


> 关闭NodeGroup自动扩缩容返回结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerDisableNodeGroupAutoScaleOKBodyWebAnnotations|  |  |  |



**ClusterManagerDisableNodeGroupAutoScaleOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
