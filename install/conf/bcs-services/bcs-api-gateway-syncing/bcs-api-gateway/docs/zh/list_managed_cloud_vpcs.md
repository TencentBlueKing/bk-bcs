### 描述

查询 Cloud VPC 列表

根据云厂商、地域、VPC ID、网络类型和业务 ID 过滤 cluster-manager 中登记的 Cloud VPC 列表，某项条件为空时忽略该条件，所有条件都为空时返回全部记录。 无路径参数，过滤条件均通过查询参数传入。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| businessID | `query` | string |  | 业务ID |
| cloudID | `query` | string |  | 云类型ID |
| networkType | `query` | string |  | 网络类型 |
| region | `query` | string |  | 地域信息 |
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

ClusterManagerListCloudVPCOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListCloudVPCDefaultBody

##### 内联模型

**ClusterManagerListCloudVPCDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListCloudVPCDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListCloudVPCDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListCloudVPCOKBody**


> 查询云 VPC 列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListCloudVPCOKBodyDataItems0|  | VPC 信息列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerListCloudVPCOKBodyDataItems0**


> VPC 响应信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| available | string|  | 可用状态 |  |
| availableIPNum | uint32 (formatted integer)|  | 可用 IP 数量 |  |
| businessID | string|  | 业务ID | `example-id-123` |
| cloudID | string|  | 云类型ID | `example-id-123` |
| extra | string|  | 额外信息 |  |
| networkType | string|  | 网络类型 |  |
| overlay | ClusterManagerListCloudVPCOKBodyDataItems0Overlay|  |  |  |
| region | string|  | 地域 |  |
| regionName | string|  | 地域名称 |  |
| reservedIPNum | uint32 (formatted integer)|  | 已保留 IP 数量 |  |
| underlay | ClusterManagerListCloudVPCOKBodyDataItems0Underlay|  |  |  |
| vpcID | string|  | VPC ID | `example-id-123` |
| vpcName | string|  | VPC 名称 |  |



**ClusterManagerListCloudVPCOKBodyDataItems0Overlay**


> CIDR 详细信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| availableIPNum | uint32 (formatted integer)|  | 可用 IP 数量 |  |
| cidrs | []ClusterManagerListCloudVPCOKBodyDataItems0OverlayCidrsItems0|  | CIDR 状态列表 |  |
| reservedIPNum | uint32 (formatted integer)|  | 已保留 IP 数量 |  |



**ClusterManagerListCloudVPCOKBodyDataItems0OverlayCidrsItems0**


> CIDR 状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| block | boolean|  | 是否被封锁 | `false` |
| cidr | string|  | CIDR 地址段 | `example-id-123` |



**ClusterManagerListCloudVPCOKBodyDataItems0Underlay**


> CIDR 详细信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| availableIPNum | uint32 (formatted integer)|  | 可用 IP 数量 |  |
| cidrs | []ClusterManagerListCloudVPCOKBodyDataItems0UnderlayCidrsItems0|  | CIDR 状态列表 |  |
| reservedIPNum | uint32 (formatted integer)|  | 已保留 IP 数量 |  |



**ClusterManagerListCloudVPCOKBodyDataItems0UnderlayCidrsItems0**


> CIDR 状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| block | boolean|  | 是否被封锁 | `false` |
| cidr | string|  | CIDR 地址段 | `example-id-123` |
