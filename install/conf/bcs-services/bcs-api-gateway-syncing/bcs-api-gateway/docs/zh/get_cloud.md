### 描述

查询云厂商详情

查询指定云厂商（Cloud）的完整配置信息，包括认证配置和区域信息。 适用于需要了解某个云厂商的具体参数配置或验证云厂商是否正确配置的场景。 需要提供 cloudID 路径参数。 返回云厂商的完整配置，包含 provider、region、认证相关字段。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| cloudID | `path` | string | 是 | [path_param参数] 云厂商 ID（如 tencentPublicCloud、awsCloud）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的云厂商 ID |
| businessID | `query` | string |  | 业务ID |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerGetCloudOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerGetCloudDefaultBody

##### 内联模型

**ClusterManagerGetCloudDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerGetCloudDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerGetCloudDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerGetCloudOKBody**


> 查询云信息响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | ClusterManagerGetCloudOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerGetCloudOKBodyData**


> 云配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cloudCredential | ClusterManagerGetCloudOKBodyDataCloudCredential|  |  |  |
| cloudID | string|  | 云类型ID | `example-id-123` |
| cloudProvider | string|  | 云服务商实现类型，如 qcloud/aws/blueking | `example-id-123` |
| clusterManagement | ClusterManagerGetCloudOKBodyDataClusterManagement|  |  |  |
| confInfo | ClusterManagerGetCloudOKBodyDataConfInfo|  |  |  |
| config | string|  | 云配置信息 |  |
| creatTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| description | string|  | 云描述 |  |
| editable | boolean|  | 是否可编辑 | `false` |
| enable | string|  | 是否启用 |  |
| engineType | string|  | 引擎类型 |  |
| extraPlugins | map ofClusterManagerGetCloudOKBodyDataExtraPluginsAnon|  | 额外操作插件 |  |
| name | string|  | 云名称 | `example-name` |
| networkInfo | ClusterManagerGetCloudOKBodyDataNetworkInfo|  |  |  |
| nodeGroupManagement | ClusterManagerGetCloudOKBodyDataNodeGroupManagement|  |  |  |
| opsPlugins | map ofClusterManagerGetCloudOKBodyDataOpsPluginsAnon|  | 标准运维操作插件 |  |
| osManagement | ClusterManagerGetCloudOKBodyDataOsManagement|  |  |  |
| platformInfo | map of string|  | 平台信息 |  |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerGetCloudOKBodyDataCloudCredential**


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



**ClusterManagerGetCloudOKBodyDataClusterManagement**


> 集群管理配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| addNodesToCluster | ClusterManagerGetCloudOKBodyDataClusterManagementAddNodesToCluster|  |  |  |
| availableVersion | []string|  | 可用版本列表 |  |
| checkExternalNodeEmptyAction | ClusterManagerGetCloudOKBodyDataClusterManagementCheckExternalNodeEmptyAction|  |  |  |
| commonMixedAction | ClusterManagerGetCloudOKBodyDataClusterManagementCommonMixedAction|  |  |  |
| createCluster | ClusterManagerGetCloudOKBodyDataClusterManagementCreateCluster|  |  |  |
| deleteCluster | ClusterManagerGetCloudOKBodyDataClusterManagementDeleteCluster|  |  |  |
| deleteNodesFromCluster | ClusterManagerGetCloudOKBodyDataClusterManagementDeleteNodesFromCluster|  |  |  |
| importCluster | ClusterManagerGetCloudOKBodyDataClusterManagementImportCluster|  |  |  |
| managedConfig | []ClusterManagerGetCloudOKBodyDataClusterManagementManagedConfigItems0|  | 托管配置列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementAddNodesToCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementAddNodesToClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementAddNodesToClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCheckExternalNodeEmptyAction**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementCheckExternalNodeEmptyActionPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCheckExternalNodeEmptyActionPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCommonMixedAction**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementCommonMixedActionPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCommonMixedActionPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCreateCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementCreateClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementCreateClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementDeleteCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementDeleteClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementDeleteClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementDeleteNodesFromCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementDeleteNodesFromClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementDeleteNodesFromClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementImportCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataClusterManagementImportClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementImportClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataClusterManagementManagedConfigItems0**


> 托管配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desc | string|  | 描述 |  |
| spec | string|  | 规格 |  |



**ClusterManagerGetCloudOKBodyDataConfInfo**


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



**ClusterManagerGetCloudOKBodyDataExtraPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNetworkInfo**


> 云网络配置信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidrStep | []uint32 (formatted integer)|  | CIDR 步长列表 |  |
| cidrSteps | []ClusterManagerGetCloudOKBodyDataNetworkInfoCidrStepsItems0|  | 环境CIDR步长配置 |  |
| perNodePodNum | []uint32 (formatted integer)|  | 单节点Pod数量选项 |  |
| serviceSteps | []uint32 (formatted integer)|  | Service CIDR 步长列表 |  |
| underlayAutoSteps | []uint32 (formatted integer)|  | Underlay 自动网络步长列表 |  |
| underlayRatio | uint32 (formatted integer)|  | Underlay 比例 |  |
| underlaySteps | []uint32 (formatted integer)|  | Underlay 网络步长列表 |  |
| vpcCniModes | []ClusterManagerGetCloudOKBodyDataNetworkInfoVpcCniModesItems0|  | VPC-CNI 网络模式列表 |  |



**ClusterManagerGetCloudOKBodyDataNetworkInfoCidrStepsItems0**


> 环境 CIDR 步长配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| env | string|  | 环境标识 |  |
| step | uint32 (formatted integer)|  | 步长值 |  |



**ClusterManagerGetCloudOKBodyDataNetworkInfoVpcCniModesItems0**


> 网络模式



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| default | boolean|  | 是否默认模式 | `false` |
| mode | string|  | 模式标识 |  |
| name | string|  | 模式名称 | `example-name` |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagement**


> 节点组管理配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| addExternalNodesToCluster | ClusterManagerGetCloudOKBodyDataNodeGroupManagementAddExternalNodesToCluster|  |  |  |
| cleanNodesInGroup | ClusterManagerGetCloudOKBodyDataNodeGroupManagementCleanNodesInGroup|  |  |  |
| createNodeGroup | ClusterManagerGetCloudOKBodyDataNodeGroupManagementCreateNodeGroup|  |  |  |
| deleteExternalNodesFromCluster | ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteExternalNodesFromCluster|  |  |  |
| deleteNodeGroup | ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteNodeGroup|  |  |  |
| moveNodesToGroup | ClusterManagerGetCloudOKBodyDataNodeGroupManagementMoveNodesToGroup|  |  |  |
| removeNodesFromGroup | ClusterManagerGetCloudOKBodyDataNodeGroupManagementRemoveNodesFromGroup|  |  |  |
| updateDesiredNodes | ClusterManagerGetCloudOKBodyDataNodeGroupManagementUpdateDesiredNodes|  |  |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementAddExternalNodesToCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementAddExternalNodesToClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementAddExternalNodesToClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementCleanNodesInGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementCleanNodesInGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementCleanNodesInGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementCreateNodeGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementCreateNodeGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementCreateNodeGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteExternalNodesFromCluster**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteExternalNodesFromClusterPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteExternalNodesFromClusterPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteNodeGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteNodeGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementDeleteNodeGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementMoveNodesToGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementMoveNodesToGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementMoveNodesToGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementRemoveNodesFromGroup**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementRemoveNodesFromGroupPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementRemoveNodesFromGroupPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementUpdateDesiredNodes**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerGetCloudOKBodyDataNodeGroupManagementUpdateDesiredNodesPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerGetCloudOKBodyDataNodeGroupManagementUpdateDesiredNodesPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataOpsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerGetCloudOKBodyDataOsManagement**


> 操作系统管理信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| availableVersion | []string|  | 可用版本列表 |  |
| defaultSecurityGroups | []string|  | 默认安全组列表 |  |
| instanceType | []string|  | 实例类型列表 |  |
| regions | map of string|  | 地域映射 |  |
