### 描述

查询业务下集群列表

查询指定 CMDB 业务下所有集群的列表，包含集群 ID、名称和状态信息。 适用于需要根据业务 ID 查找对应集群或枚举业务管理的集群的场景。 需要提供 businessID（路径参数），可选填 operator 查询特定操作者的集群。 返回集群列表，每条包含 clusterID、名称、类型和当前状态。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| businessID | `path` | string | 是 | [path_param参数] 业务 ID，来源于 CMDB（如 123456）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的业务 ID |
| operator | `query` | string |  | 操作者 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListBusinessClusterOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListBusinessClusterDefaultBody

##### 内联模型

**ClusterManagerListBusinessClusterDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListBusinessClusterDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListBusinessClusterDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListBusinessClusterOKBody**


> 查询业务集群列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListBusinessClusterOKBodyDataItems0|  | 集群基本信息列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerListBusinessClusterOKBodyDataItems0**


> 集群基础信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bizMaintainer | string|  | 业务运维人员 |  |
| businessID | string|  | 业务ID | `example-id-123` |
| clusterCategory | date-time (formatted string)|  | 集群分类 | `2024-01-01T00:00:00Z` |
| clusterID | string|  | BCS 内部集群ID，格式 BCS-K8S-XXXXX | `BCS-K8S-00000` |
| clusterName | string|  | 集群名称 |  |
| clusterType | string|  | 集群类型 |  |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| description | string|  | 集群描述 |  |
| engineType | string|  | 引擎类型 |  |
| environment | string|  | 集群所处环境 |  |
| isCommonCluster | boolean|  | 是否为公共集群 | `false` |
| isShared | boolean|  | 是否为共享集群 | `false` |
| labels | map of string|  | 集群标签 |  |
| manageType | string|  | 管理方式 |  |
| moduleID | string|  | CMDB 模块ID | `example-id-123` |
| networkType | string|  | 网络类型 |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| provider | string|  | 集群云配置类型ID | `example-id-123` |
| region | string|  | 集群所在地域 |  |
| status | date-time (formatted string)|  | 集群状态 | `2024-01-01T00:00:00Z` |
| systemID | string|  | 系统ID | `example-id-123` |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |
| vpcID | string|  | VPC ID | `example-id-123` |
