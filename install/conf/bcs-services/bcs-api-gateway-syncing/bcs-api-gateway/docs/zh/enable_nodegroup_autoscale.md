### 描述

开启节点池节点自动扩缩容

开启节点组的自动伸缩功能。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerEnableNodeGroupAutoScaleBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerEnableNodeGroupAutoScaleOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerEnableNodeGroupAutoScaleDefaultBody

##### 内联模型

**ClusterManagerEnableNodeGroupAutoScaleBody**


> 开启NodeGroup自动扩缩容



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| nodeGroupID | string| 是 | NodeGroupID |  |



**ClusterManagerEnableNodeGroupAutoScaleDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerEnableNodeGroupAutoScaleDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerEnableNodeGroupAutoScaleDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerEnableNodeGroupAutoScaleOKBody**


> 开启NodeGroup自动扩缩容返回结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerEnableNodeGroupAutoScaleOKBodyWebAnnotations|  |  |  |



**ClusterManagerEnableNodeGroupAutoScaleOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
