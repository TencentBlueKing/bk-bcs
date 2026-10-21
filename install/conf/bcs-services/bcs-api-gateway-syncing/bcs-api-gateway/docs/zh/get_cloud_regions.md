### 描述

查询云厂商地域列表

查询指定云厂商下所有可用地域（Region）的列表。 适用于需要了解某个云厂商支持哪些地域或选择集群部署地域的场景。 需要提供 cloudID 路径参数，可选填认证账号相关参数。 返回地域列表，包含地域 ID、名称和支持的可用区信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| cloudID | `path` | string | 是 | [path_param参数] 云厂商 ID（如 tencentPublicCloud、awsCloud）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的云厂商 ID |
| accountID | `query` | string |  | 云账号ID |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerGetCloudRegionsOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerGetCloudRegionsDefaultBody

##### 内联模型

**ClusterManagerGetCloudRegionsDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerGetCloudRegionsDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerGetCloudRegionsDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerGetCloudRegionsOKBody**


> 查询云区域列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerGetCloudRegionsOKBodyDataItems0|  | 地域信息列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerGetCloudRegionsOKBodyDataItems0**


> 地域信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| region | string|  | 地域ID |  |
| regionName | string|  | 地域名称 |  |
| regionState | date-time (formatted string)|  | 地域状态 | `2024-01-01T00:00:00Z` |
