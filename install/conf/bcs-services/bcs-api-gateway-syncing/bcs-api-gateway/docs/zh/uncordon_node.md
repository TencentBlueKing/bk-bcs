### 描述

节点可调度态

节点设置可调度状态。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerUnCordonNodeBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUnCordonNodeOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUnCordonNodeDefaultBody

##### 内联模型

**ClusterManagerUnCordonNodeBody**


> 设置节点可调度状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 更新节点所属的clusterID |  |
| innerIPs | []string|  | 节点内网IP列表 |  |
| nodes | []string|  | 节点名称列表，和节点内网IP列表二选一，优先节点名称列表 |  |



**ClusterManagerUnCordonNodeDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUnCordonNodeDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUnCordonNodeDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUnCordonNodeOKBody**


> 更新节点可调度状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUnCordonNodeOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerUnCordonNodeOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| fail | []ClusterManagerUnCordonNodeOKBodyDataFailItems0|  |  |  |
| success | []ClusterManagerUnCordonNodeOKBodyDataSuccessItems0|  |  |  |



**ClusterManagerUnCordonNodeOKBodyDataFailItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |



**ClusterManagerUnCordonNodeOKBodyDataSuccessItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeName | string|  |  |  |
