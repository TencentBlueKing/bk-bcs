### 描述

更新节点标签

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerUpdateNodeLabelsBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateNodeLabelsOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateNodeLabelsDefaultBody

##### 内联模型

**ClusterManagerUpdateNodeLabelsBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 更新节点所属的clusterID |  |
| nodes | []ClusterManagerUpdateNodeLabelsParamsBodyNodesItems0|  | 节点列表 |  |



**ClusterManagerUpdateNodeLabelsDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateNodeLabelsDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateNodeLabelsDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateNodeLabelsOKBody**


> 更新节点label



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUpdateNodeLabelsOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerUpdateNodeLabelsOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| fail | []ClusterManagerUpdateNodeLabelsOKBodyDataFailItems0|  |  |  |
| success | []ClusterManagerUpdateNodeLabelsOKBodyDataSuccessItems0|  |  |  |



**ClusterManagerUpdateNodeLabelsOKBodyDataFailItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeLabelsOKBodyDataSuccessItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeLabelsParamsBodyNodesItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| labels | map of string|  |  |  |
| nodeName | string|  | 节点名称 |  |
