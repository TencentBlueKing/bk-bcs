### 描述

节点上的 Pod 驱逐前风险检查

节点pod驱逐前风险检查。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerCheckDrainNodeBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCheckDrainNodeOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCheckDrainNodeDefaultBody

##### 内联模型

**ClusterManagerCheckDrainNodeBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | pod迁移节点所属的clusterID |  |
| deleteLocalData | boolean (formatted boolean)|  |  |  |
| disableEviction | boolean (formatted boolean)|  |  |  |
| force | boolean (formatted boolean)|  |  |  |
| gracePeriodSeconds | int32 (formatted integer)|  |  |  |
| ignoreAllDaemonSets | boolean (formatted boolean)|  |  |  |
| innerIPs | []string|  | 节点内网IP列表，和节点名称列表二选一，优先节点名称列表 |  |
| nodes | []string|  | 节点名称列表，和节点内网IP列表二选一，优先节点名称列表 |  |
| podSelector | string|  |  |  |
| selector | string|  |  |  |



**ClusterManagerCheckDrainNodeDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCheckDrainNodeDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCheckDrainNodeDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCheckDrainNodeOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | []ClusterManagerCheckDrainNodeOKBodyDataItems0|  | 返回数据 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerCheckDrainNodeOKBodyDataItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| evictionRisk | []ClusterManagerCheckDrainNodeOKBodyDataItems0EvictionRiskItems0|  |  |  |
| gracePeriodSeconds | int64 (formatted integer)|  |  |  |
| nameSpace | string|  |  |  |
| node | string|  |  |  |
| podName | string|  |  |  |
| podServiceAccount | string|  |  |  |
| podStatus | string|  |  |  |
| willBeEvicted | boolean (formatted boolean)|  |  |  |



**ClusterManagerCheckDrainNodeOKBodyDataItems0EvictionRiskItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| riskDescription | string|  |  |  |
| riskParameter | string|  |  |  |
