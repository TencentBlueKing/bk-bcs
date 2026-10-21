### 描述

导入用户Cluster任务

导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerImportClusterBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerImportClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerImportClusterDefaultBody

##### 内联模型

**ClusterManagerImportClusterBody**


> 用户已存在的集群导入BCS进行纳管



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| accountID | string|  | 导入集群关联的cloud凭证信息, 可以为空; 当为空时, 可以通过cloud获取cloud凭证信息 |  |
| area | ClusterManagerImportClusterParamsBodyArea|  |  |  |
| businessID | string| 是 | CMDB业务ID |  |
| cloudMode | ClusterManagerImportClusterParamsBodyCloudMode|  |  |  |
| clusterCategory | string|  | 集群类别，主要用于区分该集群是否是自建、导入(builder/importer), 默认是自建 |  |
| clusterID | string|  | 集群ID，例如BCS-K8S-000000(手动录入信息时直接指定clusterID,自动导入会默认生成集群ID) |  |
| clusterName | string| 是 | 集群名称 |  |
| clusterType | string| 是 | 集群类型, 例如[federation, single], federation表示为联邦集群，single表示独立集群，默认为single |  |
| creator | string| 是 | 创建人 |  |
| description | string|  | 集群简要描述 |  |
| engineType | string| 是 | 引擎类型，默认k8s |  |
| environment | string| 是 | 集群环境, 例如[prod, debug, stag] |  |
| extraClusterID | string|  | 导入集群的额外集群ID标识信息,默认时空值 |  |
| extraInfo | map of string|  |  |  |
| isExclusive | boolean (formatted boolean)| 是 | 是否为业务独占集群,默认为true |  |
| is_shared | boolean (formatted boolean)|  | 是否为共享集群,默认false |  |
| labels | map of string|  |  |  |
| manageType | string|  | 集群管理类型，公有云时生效，MANAGED_CLUSTER(云上托管集群)，默认是 INDEPENDENT_CLUSTER(独立集群，自行维护) |  |
| networkType | string|  | 集群网络类型(underlay/overlay),默认是overlay |  |
| projectID | string| 是 | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string| 是 | 云模版ID, 集群所属云进行流程管理 |  |
| region | string|  | 集群所在地域 |  |
| version | string|  | 导入集群版本信息 |  |



**ClusterManagerImportClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerImportClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerImportClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerImportClusterOKBody**


> 导入集群的返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | interface{}| 是 | 返回导入集群和任务 |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerImportClusterParamsBodyArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerImportClusterParamsBodyCloudMode**


> 云provider导入集群模式



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cloudID | string|  | 云集群ID(例如: 腾讯云: cls-xxx) |  |
| inter | boolean (formatted boolean)|  | 默认外网方式导入, 为true时候以内网方式导入. 该字段仅对云导入生效 |  |
| kubeConfig | string|  | 集群kubeConfig文件 |  |
| nodeIps | []string|  | 集群Master或者节点IP列表, 需要提前纳管至蓝鲸平台。通过反向代理实现集群纳管 |  |
| resourceGroup | string|  | 集群所在的资源组(azure云) |  |
