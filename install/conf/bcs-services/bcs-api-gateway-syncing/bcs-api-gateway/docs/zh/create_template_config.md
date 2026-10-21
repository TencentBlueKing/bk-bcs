### 描述

创建模板配置

创建新的模板配置，需要提供完整的配置信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerCreateTemplateConfigBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCreateTemplateConfigOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCreateTemplateConfigDefaultBody

##### 内联模型

**ClusterManagerCreateTemplateConfigBody**


> 创建模板配置的请求参数



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| businessID | string| 是 | 业务ID |  |
| cloudTemplateConfig | ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfig|  |  |  |
| clusterID | string|  | 集群ID |  |
| configType | string| 是 | 模板配置类型，如：cloud/cluster/nodegroup |  |
| projectID | string| 是 | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string| 是 | 云服务商类型，如：aws/azure/gcp |  |



**ClusterManagerCreateTemplateConfigDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCreateTemplateConfigDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCreateTemplateConfigDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCreateTemplateConfigOKBody**


> 创建模板配置的响应结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回操作结果 |  |



**ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfig**


> 云模板配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cloudNetworkTemplateConfig | ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigCloudNetworkTemplateConfig|  |  |  |
| taskTimeTemplateConfig | ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigTaskTimeTemplateConfig|  |  |  |



**ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigCloudNetworkTemplateConfig**


> 云模板网络配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidrSteps | []ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigCloudNetworkTemplateConfigCidrStepsItems0|  | cidr自动扩容步长设置(随环境变量配置而变化) |  |
| perNodePodNum | []int64 (formatted integer)|  | 单节点pod数量配置 |  |
| serviceSteps | []int64 (formatted integer)|  | service数量扩容步长设置 |  |
| underlayAutoSteps | []int64 (formatted integer)|  | underlay网络模式下自动分配子网的步长 |  |
| underlaySteps | []int64 (formatted integer)|  | underlay网络模式下用户手动分配子网的步长 |  |



**ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigCloudNetworkTemplateConfigCidrStepsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| env | string|  |  |  |
| step | int64 (formatted integer)|  |  |  |



**ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigTaskTimeTemplateConfig**


> 任务时间配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| timeoutConfig | []ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigTaskTimeTemplateConfigTimeoutConfigItems0|  | 任务步骤超时时间配置 |  |



**ClusterManagerCreateTemplateConfigParamsBodyCloudTemplateConfigTaskTimeTemplateConfigTimeoutConfigItems0**


> 任务步骤超时时间配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| stepName | string|  | 步骤名 |  |
| timeout | int64 (formatted integer)|  | 超时时间，单位为分钟 |  |
