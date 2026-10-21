### 描述

查询集群节点组列表

查询指定集群下所有节点组（NodeGroup）的列表，包含每个节点组的配置和当前副本数。 适用于需要了解集群节点自动扩缩容配置或查找特定节点组的场景。 需要提供 clusterID 路径参数。 返回节点组列表，包含名称、最小/最大节点数、当前节点数和实例类型。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| enableFilter | `query` | boolean |  | 是否启用过滤 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListClusterNodeGroupOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListClusterNodeGroupDefaultBody

##### 内联模型

**ClusterManagerListClusterNodeGroupDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListClusterNodeGroupDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListClusterNodeGroupDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListClusterNodeGroupOKBody**


> 查询集群节点组列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListClusterNodeGroupOKBodyDataItems0|  | 节点组列表 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |
| webAnnotations | ClusterManagerListClusterNodeGroupOKBodyWebAnnotations|  |  |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0**


> 节点组信息，用于集群弹性伸缩



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| area | ClusterManagerListClusterNodeGroupOKBodyDataItems0Area|  |  |  |
| autoScaling | ClusterManagerListClusterNodeGroupOKBodyDataItems0AutoScaling|  |  |  |
| cloudNodeGroupID | string|  | 云节点组ID | `example-id-123` |
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| consumerID | string|  | 消费者ID | `example-id-123` |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| enableAutoscale | boolean|  | 是否开启弹性伸缩 | `false` |
| extraInfo | map of string|  | 额外信息 |  |
| labels | map of string|  | 节点标签 |  |
| launchTemplate | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplate|  |  |  |
| name | string|  | 节点组名称 | `example-name` |
| nodeGroupID | string|  | 节点组ID | `example-id-123` |
| nodeGroupType | string|  | 节点组类型 |  |
| nodeOS | string|  | 节点操作系统 |  |
| nodeTemplate | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplate|  |  |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| provider | string|  | 云服务商类型 | `example-id-123` |
| region | string|  | 地域 |  |
| status | date-time (formatted string)|  | 节点组状态 | `2024-01-01T00:00:00Z` |
| tags | map of string|  | 标签 |  |
| taints | map of string|  | 节点污点 |  |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0Area**


> 云区域信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | uint32 (formatted integer)|  | 蓝鲸云区域ID |  |
| bkCloudName | string|  | 蓝鲸云区域名称 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0AutoScaling**


> 弹性伸缩组配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoScalingID | string|  | 弹性伸缩组ID | `example-id-123` |
| autoScalingName | string|  | 弹性伸缩组名称 |  |
| autoUpgrade | boolean|  | 是否自动升级 | `false` |
| defaultCooldown | uint32 (formatted integer)|  | 默认冷却时间（秒） |  |
| desiredSize | uint32 (formatted integer)|  | 期望节点数 |  |
| maxSize | uint32 (formatted integer)|  | 最大节点数 |  |
| minSize | uint32 (formatted integer)|  | 最小节点数 |  |
| multiZoneSubnetPolicy | string|  | 多可用区子网策略 |  |
| replaceUnhealthy | boolean|  | 是否替换不健康节点 | `false` |
| retryPolicy | string|  | 重试策略 |  |
| scalingMode | string|  | 伸缩模式 |  |
| serviceRole | string|  | 服务角色 |  |
| subnetIDs | []string|  | 子网ID列表 |  |
| timeRanges | []ClusterManagerListClusterNodeGroupOKBodyDataItems0AutoScalingTimeRangesItems0|  | 定时伸缩时间范围 |  |
| vpcID | string|  | VPC ID | `example-id-123` |
| zones | []string|  | 可用区列表 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0AutoScalingTimeRangesItems0**


> 定时伸缩时间范围



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desiredNum | uint32 (formatted integer)|  | 期望节点数 |  |
| name | string|  | 时间范围名称 | `example-name` |
| schedule | string|  | Cron 表达式 |  |
| zone | string|  | 可用区 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplate**


> 节点组启动配置模板



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | uint32 (formatted integer)|  | CPU 核数 |  |
| GPU | uint32 (formatted integer)|  | GPU 数量 |  |
| Mem | uint32 (formatted integer)|  | 内存大小（MB） |  |
| charge | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateCharge|  |  |  |
| dataDisks | []ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateDataDisksItems0|  | 数据盘配置列表 |  |
| imageInfo | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateImageInfo|  |  |  |
| initLoginPassword | string|  | 初始登录密码 |  |
| initLoginUsername | string|  | 初始登录用户名 |  |
| instanceChargeType | string|  | 计费类型 |  |
| instanceType | string|  | 机型 |  |
| internetAccess | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateInternetAccess|  |  |  |
| isMonitorService | boolean|  | 是否开启监控服务 | `false` |
| isSecurityService | boolean|  | 是否开启安全服务 | `false` |
| keyPair | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateKeyPair|  |  |  |
| launchConfigurationID | date-time (formatted string)|  | 启动配置ID | `2024-01-01T00:00:00Z` |
| launchConfigureName | string|  | 启动配置名称 |  |
| networkTag | []string|  | 网络标签列表 |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| securityGroupIDs | []string|  | 安全组ID列表 |  |
| selector | map of string|  | 标签选择器 |  |
| systemDisk | ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateSystemDisk|  |  |  |
| userData | date-time (formatted string)|  | 用户自定义数据脚本 | `2024-01-01T00:00:00Z` |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateCharge**


> 包年包月计费配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | uint32 (formatted integer)|  | 包月时长（月） |  |
| renewFlag | string|  | 续费标志 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateDataDisksItems0**


> 数据盘定义



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 磁盘大小 |  |
| diskType | string|  | 磁盘类型 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateImageInfo**


> 镜像信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID | `example-id-123` |
| imageName | string|  | 镜像名称 |  |
| imageOs | string|  | 镜像操作系统 |  |
| imageType | string|  | 镜像类型 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateInternetAccess**


> 公网带宽设置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bandwidthPackageId | string|  | 带宽包ID | `example-id-123` |
| internetChargeType | string|  | 网络计费类型 |  |
| internetMaxBandwidth | string|  | 最大出带宽 | `example-id-123` |
| nodePublicIPPrefixID | string|  | 节点公网IP前缀ID | `example-id-123` |
| publicAccessCidrs | []string|  | 公网访问 CIDR 列表 |  |
| publicIP | string|  | 公网IP |  |
| publicIPAssigned | boolean|  | 是否分配公网IP | `false` |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateKeyPair**


> 密钥信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID | `example-id-123` |
| keyPublic | string|  | 公钥内容 |  |
| keySecret | string|  | 密钥内容 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0LaunchTemplateSystemDisk**


> 数据盘定义



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 磁盘大小 |  |
| diskType | string|  | 磁盘类型 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplate**


> 节点模板，用于节点通用配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipScaleInWhenFailed | boolean|  | 缩容失败时是否允许跳过 | `false` |
| allowSkipScaleOutWhenFailed | boolean|  | 扩容失败时是否允许跳过 | `false` |
| annotations | map of string|  | 节点注解 |  |
| bcsScaleInAddons | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleInAddons|  |  |  |
| bcsScaleOutAddons | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleOutAddons|  |  |  |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| dataDisks | []ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateDataDisksItems0|  | 数据盘配置列表 |  |
| desc | string|  | 模板描述 |  |
| dockerGraphPath | date-time (formatted string)|  | Docker 数据目录 | `2024-01-01T00:00:00Z` |
| extraArgs | map of string|  | 额外参数 |  |
| extraInfo | map of string|  | 额外信息 |  |
| gpuArgs | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgs|  |  |  |
| image | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateImage|  |  |  |
| labels | map of string|  | 节点标签 |  |
| maxPodsPerNode | uint32 (formatted integer)|  | 单节点最大 Pod 数 |  |
| module | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateModule|  |  |  |
| mountTarget | string|  | 数据盘挂载目录 |  |
| name | string|  | 模板名称 | `example-name` |
| nodeOS | string|  | 节点操作系统 |  |
| nodeTemplateID | date-time (formatted string)|  | 节点模板ID | `2024-01-01T00:00:00Z` |
| preStartUserScript | string|  | 预启动用户脚本 |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| runtime | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateRuntime|  |  |  |
| scaleInExtraAddons | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleInExtraAddons|  |  |  |
| scaleInPostScript | string|  | 缩容后置脚本 |  |
| scaleInPreScript | string|  | 缩容前置脚本 |  |
| scaleOutExtraAddons | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean|  | 是否跳过系统初始化 | `false` |
| taints | []ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateTaintsItems0|  | 节点污点列表 |  |
| unSchedulable | uint32 (formatted integer)|  | 是否不可调度 |  |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |
| userScript | string|  | 用户自定义脚本 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleInAddons**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleInAddonsPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleInAddonsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleOutAddons**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleOutAddonsPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateBcsScaleOutAddonsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateDataDisksItems0**


> 云数据盘配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean|  | 是否自动格式化并挂载 | `false` |
| diskPartition | string|  | 磁盘分区 |  |
| diskSize | string|  | 磁盘大小 |  |
| diskType | string|  | 磁盘类型 |  |
| fileSystem | string|  | 文件系统类型 |  |
| mountTarget | string|  | 挂载目标路径 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgs**


> GPU 参数配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cuda | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCuda|  |  |  |
| cudnn | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCudnn|  |  |  |
| customDriver | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCustomDriver|  |  |  |
| driver | ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsDriver|  |  |  |
| migEnable | boolean|  | 是否启用 MIG（多实例GPU） | `false` |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCuda**


> 驱动版本信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 驱动名称 | `example-name` |
| version | string|  | 版本号 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCudnn**


> CUDNN 版本信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| devName | string|  | CUDNN 开发包名称 |  |
| docName | string|  | CUDNN 文档名称 |  |
| name | string|  | CUDNN 名称 | `example-name` |
| version | string|  | CUDNN 版本号 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsCustomDriver**


> 自定义驱动配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| address | string|  | 自定义驱动地址 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateGpuArgsDriver**


> 驱动版本信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 驱动名称 | `example-name` |
| version | string|  | 版本号 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateImage**


> 镜像信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID | `example-id-123` |
| imageName | string|  | 镜像名称 |  |
| imageOs | string|  | 镜像操作系统 |  |
| imageType | string|  | 镜像类型 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateModule**


> 业务模块信息，用于节点模块转移



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容业务ID | `example-id-123` |
| scaleInModuleID | string|  | 缩容模块ID | `example-id-123` |
| scaleInModuleName | string|  | 缩容模块名称 |  |
| scaleOutBizID | string|  | 扩容业务ID | `example-id-123` |
| scaleOutModuleID | string|  | 扩容模块ID | `example-id-123` |
| scaleOutModuleName | string|  | 扩容模块名称 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateRuntime**


> 容器运行时信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | date-time (formatted string)|  | 容器运行时类型，如 docker、containerd | `2024-01-01T00:00:00Z` |
| runtimeVersion | date-time (formatted string)|  | 运行时版本 | `2024-01-01T00:00:00Z` |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleInExtraAddons**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleInExtraAddonsPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleInExtraAddonsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleOutExtraAddons**


> 操作动作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleOutExtraAddonsPluginsAnon|  | 插件配置列表 |  |
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateScaleOutExtraAddonsPluginsAnon**


> 标准运维插件配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean|  | 失败时是否允许跳过 | `false` |
| link | string|  | 标准运维任务链接 |  |
| params | map of string|  | 任务参数 |  |
| system | string|  | 标准运维系统ID |  |



**ClusterManagerListClusterNodeGroupOKBodyDataItems0NodeTemplateTaintsItems0**


> 节点污点



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，可选值：NoSchedule、PreferNoSchedule、NoExecute |  |
| key | string|  | 污点键 |  |
| value | string|  | 污点值 |  |



**ClusterManagerListClusterNodeGroupOKBodyWebAnnotations**


> 权限注解信息V2



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
