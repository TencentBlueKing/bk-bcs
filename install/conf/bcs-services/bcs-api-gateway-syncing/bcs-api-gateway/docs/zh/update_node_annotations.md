### 描述

更新节点注解

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerUpdateNodeAnnotationsBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateNodeAnnotationsOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateNodeAnnotationsDefaultBody

##### 内联模型

**ClusterManagerUpdateNodeAnnotationsBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 更新节点所属的clusterID |  |
| nodes | []ClusterManagerUpdateNodeAnnotationsParamsBodyNodesItems0|  | 节点列表 |  |



**ClusterManagerUpdateNodeAnnotationsDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateNodeAnnotationsDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateNodeAnnotationsDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateNodeAnnotationsOKBody**


> 更新节点annotation



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUpdateNodeAnnotationsOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerUpdateNodeAnnotationsOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| fail | []ClusterManagerUpdateNodeAnnotationsOKBodyDataFailItems0|  |  |  |
| success | []ClusterManagerUpdateNodeAnnotationsOKBodyDataSuccessItems0|  |  |  |



**ClusterManagerUpdateNodeAnnotationsOKBodyDataFailItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeAnnotationsOKBodyDataSuccessItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeAnnotationsParamsBodyNodesItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| nodeName | string|  | 节点名称 |  |
