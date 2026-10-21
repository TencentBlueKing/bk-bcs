### 描述

更新配额

更新虚拟集群配额。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| body | `body` | ClusterManagerUpdateVirtualClusterQuotaBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateVirtualClusterQuotaOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateVirtualClusterQuotaDefaultBody

##### 内联模型

**ClusterManagerUpdateVirtualClusterQuotaBody**


> 更新虚拟集群配额



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string| 是 | 集群ID |  |
| quota | ClusterManagerUpdateVirtualClusterQuotaParamsBodyQuota| 是 |  |  |
| updater | string|  | 操作人员名称 |  |



**ClusterManagerUpdateVirtualClusterQuotaDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateVirtualClusterQuotaDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateVirtualClusterQuotaDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateVirtualClusterQuotaOKBody**


> 更新虚拟集群命名空间配额的返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUpdateVirtualClusterQuotaOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerUpdateVirtualClusterQuotaOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  |  |  |
| cpuRequests | string|  |  |  |
| memoryLimits | string|  |  |  |
| memoryRequests | string|  |  |  |
| serviceLimits | string|  |  |  |



**ClusterManagerUpdateVirtualClusterQuotaParamsBodyQuota**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  |  |  |
| cpuRequests | string|  |  |  |
| memoryLimits | string|  |  |  |
| memoryRequests | string|  |  |  |
| serviceLimits | string|  |  |  |
