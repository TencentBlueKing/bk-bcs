### 描述

查询云厂商列表

查询平台中所有已配置的云厂商（Cloud）列表，支持按云厂商 ID、名称等条件过滤。 适用于需要了解平台支持哪些云厂商或查找特定云厂商配置的场景。 各查询条件可单独使用或组合；所有条件为空时返回所有云厂商。 返回云厂商列表，包含 cloudID、名称、提供商类型和配置信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| businessID | `query` | string |  | 业务ID |
| cloudID | `query` | string |  | 云类型ID |
| cloudProvider | `query` | string |  | 云服务商实现类型，如 qcloud/aws/blueking |
| creator | `query` | string |  | 创建者 |
| editable | `query` | boolean |  | 是否可编辑 |
| name | `query` | string |  | 名称，用于过滤 |
| updater | `query` | string |  | 更新者 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListCloudOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListCloudDefaultBody

##### 内联模型

**ClusterManagerListCloudDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListCloudDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListCloudDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListCloudOKBody**


> 查询云列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListCloudOKBodyDataItems0|  | 云配置信息列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerListCloudOKBodyDataItems0**


> 云配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cloudCredential | ClusterManagerListCloudOKBodyDataItems0CloudCredential|  |  |  |
| cloudID | string|  | 云类型ID | `example-id-123` |
| cloudProvider | string|  | 云服务商实现类型，如 qcloud/aws/blueking | `example-id-123` |
| clusterManagement | ClusterManagerListCloudOKBodyDataItems0ClusterManagement|  |  |  |
| confInfo | ClusterManagerListCloudOKBodyDataItems0ConfInfo|  |  |  |
| config | string|  | 云配置信息 |  |
| creatTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| description | string|  | 云描述 |  |
| editable | boolean|  | 是否可编辑 | `false` |
| enable | string|  | 是否启用 |  |
| engineType | string|  | 引擎类型 |  |
| extraPlugins | map ofClusterManagerListCloudOKBodyDataItems0ExtraPluginsAnon|  | 额外操作插件 |  |
| name | string|  | 云名称 | `example-name` |
| networkInfo | ClusterManagerListCloudOKBodyDataItems0NetworkInfo|  |  |  |
| nodeGroupManagement | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagement|  |  |  |
| opsPlugins | map ofClusterManagerListCloudOKBodyDataItems0OpsPluginsAnon|  | 标准运维操作插件 |  |
| osManagement | ClusterManagerListCloudOKBodyDataItems0OsManagement|  |  |  |
| platformInfo | map of string|  | 平台信息 |  |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerListCloudOKBodyDataItems0CloudCredential**


> 云凭证信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clientID | string|  | Azure 客户端ID | `example-id-123` |
| clientSecret | string|  | Azure 客户端密钥 |  |
| gkeProjectID | string|  | GKE 项目ID | `b37778ec757544868a01e1f01f07037f` |
| key | string|  | 凭证密钥ID |  |
| resourceGroupName | string|  | Azure 资源组名称 |  |
| secret | string|  | 凭证密钥 |  |
| serviceAccountSecret | string|  | GKE 服务账号密钥 |  |
| subscriptionID | string|  | Azure 订阅ID | `example-id-123` |
| tenantID | string|  | Azure 租户ID | `example-id-123` |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagement**


> 集群管理配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| addNodesToCluster | ClusterManagerListCloudOKBodyDataItems0ClusterManagementAddNodesToCluster|  |  |  |
| availableVersion | []string|  | 可用版本列表 |  |
| checkExternalNodeEmptyAction | ClusterManagerListCloudOKBodyDataItems0ClusterManagementCheckExternalNodeEmptyAction|  |  |  |
| commonMixedAction | ClusterManagerListCloudOKBodyDataItems0ClusterManagementCommonMixedAction|  |  |  |
| createCluster | ClusterManagerListCloudOKBodyDataItems0ClusterManagementCreateCluster|  |  |  |
| deleteCluster | ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteCluster|  |  |  |
| deleteNodesFromCluster | ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteNodesFromCluster|  |  |  |
| importCluster | ClusterManagerListCloudOKBodyDataItems0ClusterManagementImportCluster|  |  |  |
| managedConfig | []ClusterManagerListCloudOKBodyDataItems0ClusterManagementManagedConfigItems0|  | 托管配置列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementAddNodesToCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementAddNodesToClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementAddNodesToClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCheckExternalNodeEmptyAction**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementCheckExternalNodeEmptyActionPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCheckExternalNodeEmptyActionPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCommonMixedAction**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementCommonMixedActionPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCommonMixedActionPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCreateCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementCreateClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementCreateClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteNodesFromCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteNodesFromClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementDeleteNodesFromClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementImportCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0ClusterManagementImportClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementImportClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0ClusterManagementManagedConfigItems0**


> 托管配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desc | string|  | 描述 |  |
| spec | string|  | 规格 |  |



**ClusterManagerListCloudOKBodyDataItems0ConfInfo**


> 云特殊配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cloudDomain | string|  | 云域名 |  |
| cloudInternalEnable | boolean|  | 是否启用内部云 | `false` |
| createClusterNodesLimit | uint32 (formatted integer)|  | 创建集群时节点数量限制 |  |
| disableCheckGroupResource | boolean|  | 是否禁止检查节点组资源 | `false` |
| disableCreateCluster | boolean|  | 是否禁止创建集群 | `false` |
| disableImportCluster | boolean|  | 是否禁止导入集群 | `false` |
| disableNodeGroup | boolean|  | 是否禁止节点组 | `false` |
| machineDomain | string|  | 机器域名 |  |
| maxNodeCount | uint32 (formatted integer)|  | 最大节点数量 |  |
| regionHaWhiteList | []string|  | 地域高可用白名单 |  |
| vpcDomain | string|  | VPC 域名 |  |
| whiteImageOsName | []string|  | 白名单镜像系统名称 |  |



**ClusterManagerListCloudOKBodyDataItems0ExtraPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NetworkInfo**


> 云网络配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidrStep | []uint32 (formatted integer)|  | CIDR 步长列表 |  |
| cidrSteps | []ClusterManagerListCloudOKBodyDataItems0NetworkInfoCidrStepsItems0|  | 环境CIDR步长配置 |  |
| perNodePodNum | []uint32 (formatted integer)|  | 单节点Pod数量选项 |  |
| serviceSteps | []uint32 (formatted integer)|  | Service CIDR 步长列表 |  |
| underlayAutoSteps | []uint32 (formatted integer)|  | Underlay 自动网络步长列表 |  |
| underlayRatio | uint32 (formatted integer)|  | Underlay 比例 |  |
| underlaySteps | []uint32 (formatted integer)|  | Underlay 网络步长列表 |  |
| vpcCniModes | []ClusterManagerListCloudOKBodyDataItems0NetworkInfoVpcCniModesItems0|  | VPC-CNI 网络模式列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NetworkInfoCidrStepsItems0**


> 环境 CIDR 步长配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| env | string|  | 环境标识 |  |
| step | uint32 (formatted integer)|  | 步长值 |  |



**ClusterManagerListCloudOKBodyDataItems0NetworkInfoVpcCniModesItems0**


> 网络模式



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| default | boolean|  | 是否默认模式 | `false` |
| mode | string|  | 模式标识 |  |
| name | string|  | 模式名称 | `example-name` |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagement**


> 节点组管理配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| addExternalNodesToCluster | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementAddExternalNodesToCluster|  |  |  |
| cleanNodesInGroup | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCleanNodesInGroup|  |  |  |
| createNodeGroup | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCreateNodeGroup|  |  |  |
| deleteExternalNodesFromCluster | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteExternalNodesFromCluster|  |  |  |
| deleteNodeGroup | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteNodeGroup|  |  |  |
| moveNodesToGroup | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementMoveNodesToGroup|  |  |  |
| removeNodesFromGroup | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementRemoveNodesFromGroup|  |  |  |
| updateDesiredNodes | ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementUpdateDesiredNodes|  |  |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementAddExternalNodesToCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementAddExternalNodesToClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementAddExternalNodesToClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCleanNodesInGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCleanNodesInGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCleanNodesInGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCreateNodeGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCreateNodeGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementCreateNodeGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteExternalNodesFromCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteExternalNodesFromClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteExternalNodesFromClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteNodeGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteNodeGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementDeleteNodeGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementMoveNodesToGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementMoveNodesToGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementMoveNodesToGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementRemoveNodesFromGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementRemoveNodesFromGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementRemoveNodesFromGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementUpdateDesiredNodes**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListCloudOKBodyDataItems0NodeGroupManagementUpdateDesiredNodesPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListCloudOKBodyDataItems0NodeGroupManagementUpdateDesiredNodesPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0OpsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListCloudOKBodyDataItems0OsManagement**


> 操作系统管理信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| availableVersion | []string|  | 可用版本列表 |  |
| defaultSecurityGroups | []string|  | 默认安全组列表 |  |
| instanceType | []string|  | 实例类型列表 |  |
| regions | map of string|  | 地域映射 |  |
