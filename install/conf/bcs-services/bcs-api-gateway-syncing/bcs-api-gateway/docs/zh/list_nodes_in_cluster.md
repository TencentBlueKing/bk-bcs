### 描述

查询集群所有节点列表

查询指定集群内所有节点的基本信息列表，包含节点 IP 和状态。 适用于需要快速枚举集群节点 IP 列表或统计节点数量的场景。 需要提供 clusterID 路径参数，可选填额外过滤条件。 返回所有节点的 IP、状态和所在可用区信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| instanceType | `query` | string |  | 机型信息 |
| limit | `query` | uint32 (formatted integer) |  | 每页数量 |
| nodeGroupID | `query` | string |  | 节点组ID |
| offset | `query` | uint32 (formatted integer) |  | 分页偏移量 |
| region | `query` | string |  | 地域信息 |
| showPwd | `query` | boolean |  | 是否显示密码 |
| status | `query` | string |  | 状态过滤条件 |
| vpcID | `query` | string |  | VPC ID |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListNodesInClusterOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListNodesInClusterDefaultBody

##### 内联模型

**ClusterManagerListNodesInClusterDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListNodesInClusterDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListNodesInClusterDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListNodesInClusterOKBody**


> 查询集群节点列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListNodesInClusterOKBodyDataItems0|  | 集群节点列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |
| webAnnotations | ClusterManagerListNodesInClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerListNodesInClusterOKBodyDataItems0**


> 集群节点信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | uint32 (formatted integer)|  | CPU 核数 |  |
| GPU | uint32 (formatted integer)|  | GPU 数量 |  |
| VPC | string|  | VPC ID |  |
| agent | uint32 (formatted integer)|  | Agent 状态 |  |
| annotations | map of string|  | 节点注解 |  |
| bkCloudID | uint32 (formatted integer)|  | 蓝鲸云区域ID |  |
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| deviceClass | string|  | 设备类型 |  |
| deviceID | string|  | 设备ID | `example-id-123` |
| failedReason | string|  | 失败原因 |  |
| idc | string|  | IDC 机房 | `example-id-123` |
| innerIP | string|  | 节点内网主IP |  |
| innerIPv6 | string|  | 节点内网 IPv6 |  |
| instanceType | string|  | 机型 |  |
| labels | map of string|  | 节点标签 |  |
| mem | uint32 (formatted integer)|  | 内存大小（MB） |  |
| nodeGroupID | string|  | 节点组ID | `example-id-123` |
| nodeGroupName | string|  | 节点组名称 |  |
| nodeID | string|  | 节点云 Instance ID | `example-id-123` |
| nodeName | string|  | K8s 节点名称 |  |
| passwd | string|  | 登录密码 |  |
| rack | string|  | 机架 |  |
| region | string|  | 地域 |  |
| status | date-time (formatted string)|  | 节点状态，可选值：RUNNING、DELETING、CREATING、FAILED、INITIALIZATION | `2024-01-01T00:00:00Z` |
| taints | []ClusterManagerListNodesInClusterOKBodyDataItems0TaintsItems0|  | 节点污点列表 |  |
| taskID | string|  | 关联任务ID | `example-id-123` |
| unSchedulable | uint32 (formatted integer)|  | 是否不可调度，1 表示不可调度 |  |
| zone | uint32 (formatted integer)|  | 可用区编号 |  |
| zoneID | string|  | 可用区ID | `example-id-123` |
| zoneName | string|  | 可用区名称 |  |



**ClusterManagerListNodesInClusterOKBodyDataItems0TaintsItems0**


> 节点污点



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，可选值：NoSchedule、PreferNoSchedule、NoExecute |  |
| key | string|  | 污点键 |  |
| value | string|  | 污点值 |  |



**ClusterManagerListNodesInClusterOKBodyWebAnnotations**


> 权限注解信息V2



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
