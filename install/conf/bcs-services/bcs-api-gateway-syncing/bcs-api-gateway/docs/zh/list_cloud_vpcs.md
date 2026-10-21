### 描述

查询云厂商 VPC 列表（V1）

查询指定云厂商地域下所有可用 VPC 的列表，包含 VPC ID 和网络配置。 适用于需要为新建集群选择 VPC 或列举可用网络环境的场景。 需要提供 cloudID 路径参数，可选填地域 ID 和账号参数。 返回 VPC 列表，包含 vpcID、名称、CIDR 和可用区信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| cloudID | `path` | string | 是 | [path_param参数] 云厂商 ID（如 tencentPublicCloud、awsCloud）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的云厂商 ID |
| accountID | `query` | string |  | 云账号ID |
| region | `query` | string |  | 地域信息 |
| resourceGroupName | `query` | string |  | 资源组名称 |
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

ClusterManagerListCloudVpcsOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListCloudVpcsDefaultBody

##### 内联模型

**ClusterManagerListCloudVpcsDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListCloudVpcsDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListCloudVpcsDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListCloudVpcsOKBody**


> 查询云 VPC 列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListCloudVpcsOKBodyDataItems0|  | VPC 列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerListCloudVpcsOKBodyDataItems0**


> 云 VPC 信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allocateIpNum | uint32 (formatted integer)|  | 已分配 IP 数量 |  |
| cidrs | []ClusterManagerListCloudVpcsOKBodyDataItems0CidrsItems0|  | 辅助 CIDR 列表 |  |
| ipv4Cidr | string|  | IPv4 CIDR 地址段 | `example-id-123` |
| ipv6Cidr | string|  | IPv6 CIDR 地址段 | `example-id-123` |
| name | string|  | VPC 名称 | `example-name` |
| vpcId | string|  | VPC ID | `example-id-123` |



**ClusterManagerListCloudVpcsOKBodyDataItems0CidrsItems0**


> 辅助 CIDR 信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidr | string|  | CIDR 地址段 | `example-id-123` |
| cidrType | int32 (formatted integer)|  | 辅助CIDR类型（0：普通辅助CIDR，1：容器辅助CIDR），默认为0 |  |
