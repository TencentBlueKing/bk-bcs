### 描述

查询节点组内节点列表 V2

查询指定节点组（NodeGroup）内所有节点的详细信息（V2 版本，返回字段更丰富）。 适用于需要了解节点组内节点实例的详细配置和状态的场景。 需要提供 nodeGroupID 路径参数。 返回节点列表，包含 IP、状态、实例类型和资源配置等详细信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组ID。注意：此参数在 URL 路径中传递，请在请求路径中填充具体的节点组ID值 |
| output | `query` | string |  | 输出格式 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListNodesInGroupV2OKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListNodesInGroupV2DefaultBody

##### 内联模型

**ClusterManagerListNodesInGroupV2DefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListNodesInGroupV2DefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListNodesInGroupV2DefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListNodesInGroupV2OKBody**


> 查询节点组节点列表V2响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListNodesInGroupV2OKBodyDataItems0|  | 节点组节点列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |
| webAnnotations | ClusterManagerListNodesInGroupV2OKBodyWebAnnotations|  |  |  |



**ClusterManagerListNodesInGroupV2OKBodyDataItems0**


> 节点组节点信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | uint32 (formatted integer)|  | CPU 核数 |  |
| GPU | uint32 (formatted integer)|  | GPU 数量 |  |
| VPC | string|  | VPC ID |  |
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| deviceID | string|  | 设备ID | `example-id-123` |
| innerIP | string|  | 节点内网主IP |  |
| instanceRole | string|  | 实例角色 |  |
| instanceType | string|  | 机型 |  |
| mem | uint32 (formatted integer)|  | 内存大小（MB） |  |
| nodeGroupID | string|  | 节点组ID | `example-id-123` |
| nodeID | string|  | 节点云 Instance ID | `example-id-123` |
| nodeName | string|  | K8s 节点名称 |  |
| passwd | string|  | 登录密码 |  |
| region | string|  | 地域 |  |
| status | date-time (formatted string)|  | 节点状态 | `2024-01-01T00:00:00Z` |
| unSchedulable | uint32 (formatted integer)|  | 是否不可调度 |  |
| zone | uint32 (formatted integer)|  | 可用区编号 |  |
| zoneID | string|  | 可用区ID | `example-id-123` |



**ClusterManagerListNodesInGroupV2OKBodyWebAnnotations**


> 权限注解信息V2



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
