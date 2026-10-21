### 描述

更新节点污点

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerUpdateNodeTaintsBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateNodeTaintsOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateNodeTaintsDefaultBody

##### 内联模型

**ClusterManagerUpdateNodeTaintsBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 更新节点所属的clusterID |  |
| nodes | []ClusterManagerUpdateNodeTaintsParamsBodyNodesItems0|  | 节点列表 |  |



**ClusterManagerUpdateNodeTaintsDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateNodeTaintsDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateNodeTaintsDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateNodeTaintsOKBody**


> 更新节点taints



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUpdateNodeTaintsOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerUpdateNodeTaintsOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| fail | []ClusterManagerUpdateNodeTaintsOKBodyDataFailItems0|  |  |  |
| success | []ClusterManagerUpdateNodeTaintsOKBodyDataSuccessItems0|  |  |  |



**ClusterManagerUpdateNodeTaintsOKBodyDataFailItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeTaintsOKBodyDataSuccessItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUpdateNodeTaintsParamsBodyNodesItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| nodeName | string|  | 节点名称 |  |
| taints | []ClusterManagerUpdateNodeTaintsParamsBodyNodesItems0TaintsItems0|  | 节点污点信息 |  |



**ClusterManagerUpdateNodeTaintsParamsBodyNodesItems0TaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |
