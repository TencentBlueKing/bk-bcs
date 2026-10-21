### 描述

更新NodeGroup

更新NodeGroup信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerUpdateNodeGroupBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateNodeGroupOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateNodeGroupDefaultBody

##### 内联模型

**ClusterManagerUpdateNodeGroupBody**


> 更新NodeGroup请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoScaling | ClusterManagerUpdateNodeGroupParamsBodyAutoScaling|  |  |  |
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| cloudAreaName | string|  | 云区域名称 |  |
| clusterID | string|  | 节点组所关联的集群，该ID为BCS集群ID |  |
| consumerID | string|  | consumerID对接resourceManager模块申请资源使用 |  |
| desc | string|  | 更新节点组描述信息 |  |
| enableAutoscale | boolean (formatted boolean)|  | 是否开启弹性伸缩，默认false |  |
| extraInfo | map of string|  |  |  |
| labels | map of string|  |  |  |
| launchTemplate | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplate|  |  |  |
| name | string|  | 节点组名称 |  |
| nodeGroupID | string|  | 节点组ID |  |
| nodeOS | string|  | 新实例选用的系统OS，参照云信息中配置的默认值 |  |
| nodeTemplate | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplate|  |  |  |
| onlyUpdateInfo | boolean (formatted boolean)|  | onlyUpdateInfo仅更新数据,不执行流程 |  |
| provider | string|  | NodeGroup具体依赖的云实现方案，用于方案组合与扩展。默认为空，如果要设置，必须是CloudID |  |
| region | string|  | 节点组所关联区域信息 |  |
| tags | map of string|  |  |  |
| taints | map of string|  |  |  |
| updater | string|  | 信息更新者 |  |



**ClusterManagerUpdateNodeGroupDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateNodeGroupDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateNodeGroupDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateNodeGroupOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | ClusterManagerUpdateNodeGroupOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerUpdateNodeGroupOKBodyWebAnnotations|  |  |  |



**ClusterManagerUpdateNodeGroupOKBodyData**


> 节点组，对蓝鲸、腾讯云、AWS弹性伸缩能力封装



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| area | ClusterManagerUpdateNodeGroupOKBodyDataArea|  |  |  |
| autoScaling | ClusterManagerUpdateNodeGroupOKBodyDataAutoScaling|  |  |  |
| cloudNodeGroupID | string|  | NodeGroup 在各个云上的 ID |  |
| clusterID | string|  | 节点组所关联的集群，该ID为BCS集群ID |  |
| consumerID | string|  | consumerID对接resourceManager模块申请资源使用 |  |
| createTime | string|  |  |  |
| creator | string|  |  |  |
| enableAutoscale | boolean (formatted boolean)|  | 是否开启弹性伸缩，默认false |  |
| extraInfo | map of string|  |  |  |
| labels | map of string|  |  |  |
| launchTemplate | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplate|  |  |  |
| name | string|  | 节点组名称 |  |
| nodeGroupID | string|  | 节点组ID |  |
| nodeGroupType | string|  | 节点池类型(normal/external)。普通弹性伸缩节点池和第三方弹性伸缩节点池, 即CVM节点或IDC节点 |  |
| nodeOS | string|  | 新实例选用的系统OS，参照云信息中配置的默认值 |  |
| nodeTemplate | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplate|  |  |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。需与集群所属项目一致。 |  |
| provider | string|  | NodeGroup具体实现，用于覆盖所属cluster中定义的provider信息调整实现的细节。便于在一个云环境中扩展自定义的模板，使用不同 |  |
| region | string|  | 节点组所关联区域信息 |  |
| status | string|  | 节点组状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED |  |
| tags | map of string|  |  |  |
| taints | map of string|  |  |  |
| updateTime | string|  |  |  |
| updater | string|  |  |  |



**ClusterManagerUpdateNodeGroupOKBodyDataArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataAutoScaling**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoScalingID | string|  | 弹性伸缩组ID，信息创建时忽略，该ID为自建ID或者云设施ID |  |
| autoScalingName | string|  | 弹性伸缩组名称 |  |
| autoUpgrade | boolean (formatted boolean)|  | 开启自动升级GKE版本功能,如果集群为发布版本,则强制开启 |  |
| defaultCooldown | int64 (formatted integer)|  | 弹性伸缩组默认扩缩容冷却周期，默认300秒 |  |
| desiredSize | int64 (formatted integer)|  | 弹性伸缩期望实例数 |  |
| maxSize | int64 (formatted integer)|  | 弹性伸缩组最大实例数 |  |
| minSize | int64 (formatted integer)|  | 伸缩组最小实例数 |  |
| multiZoneSubnetPolicy | string|  | 可选值：PRIORITY，EQUALITY。PRIORITY：默认在高优先级的子网与可用区创建实例。EQUALITY：所有可用区、子网机会均衡 |  |
| replaceUnhealthy | boolean (formatted boolean)|  | 开启监控不健康节点替换服务，默认false。连续1分钟ping不通则视为不健康！ |  |
| retryPolicy | string|  | 创建实例失败时的重试策略：IMMEDIATE_RETRY(最大5次)，INCREMENTAL_INTERVALS，NO_RETRY。默认值为IMMEDIATE_RETRY。扩容时默认会组合subnet和zone信息逐一尝试实例创建直到最大失败次数或者所有组合全失败。 |  |
| scalingMode | string|  | 扩容模式，CLASSIC_SCALING(默认)：扩容时创建新实例，缩容时销毁实例；WAKE_UP_STOPPED_SCALING：缩容时关机不销毁，扩容时优先唤醒关机实例。 |  |
| serviceRole | string|  | 创建eks节点池时指定的iam角色名称 |  |
| subnetIDs | []string|  | 弹性伸缩组扩容实例首选的的子网ID，多个子网ID按照先后顺序确认优先级，靠前优先级高 |  |
| timeRanges | []ClusterManagerUpdateNodeGroupOKBodyDataAutoScalingTimeRangesItems0|  | 定时自动扩容集群节点配置信息 |  |
| vpcID | string|  | 弹性伸缩组所属VPC ID |  |
| zones | []string|  | 弹性伸缩组扩容实例首选的可用区ID，多个可用区按照先后顺序确认优先级，靠前优先级高 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataAutoScalingTimeRangesItems0**


> autoScale定时扩缩配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desiredNum | int64 (formatted integer)|  | 期望节点数 |  |
| name | string|  | 定时扩缩job名称 |  |
| schedule | string|  | 定时cron时间 |  |
| zone | string|  | 定时cron时区,如果不执行时区则默认是服务所在时区 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplate**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| charge | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateCharge|  |  |  |
| dataDisks | []ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| imageInfo | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateInternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 yunti默认关闭 |  |
| keyPair | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateKeyPair|  |  |  |
| launchConfigurationID | string|  | 节点配置模板ID，自建ID或者云上ID |  |
| launchConfigureName | string|  | 配置模板名称 |  |
| networkTag | []string|  | 网络标记，标记代表应用于每个节点的防火墙规则 |  |
| projectID | string|  | 项目标识（该配置中可能指蓝鲸项目等，非 BCS 项目时含义以实际业务为准）。BCS 场景下为 32 位十六进制字符串；与 projectCode 区别：projectID 为内部标识，projectCode 为人类可读的项目英文编码。 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，BCS默认使用各云信息配置中预留的信息 |  |
| selector | map of string|  |  |  |
| systemDisk | ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateSystemDisk|  |  |  |
| userData | string|  | 经过 Base64 编码后的自定义数据，最大长度不超过16KB。 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateCharge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateInternetAccess**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bandwidthPackageId | string|  | 共享带宽包ID |  |
| internetChargeType | string|  | 网络计费类型。BANDWIDTH_PREPAID：预付费按带宽结算，TRAFFIC_POSTPAID_BY_HOUR：流量按小时后付费，BANDWIDTH_POSTPAID_BY_HOUR：带宽按小时后付费，BANDWIDTH_PACKAGE：带宽包用户 |  |
| internetMaxBandwidth | string|  | 公网出带宽上限，单位：Mbps。默认值：0Mbps。 |  |
| nodePublicIPPrefixID | string|  | 节点公网ip前缀ID |  |
| publicAccessCidrs | []string|  | 允许访问Kubernetes API Server的公网Ip cidr |  |
| publicIP | string|  | 公网ip地址 |  |
| publicIPAssigned | boolean (formatted boolean)|  | 是否分配公网IP。默认为false。 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateNodeGroupOKBodyDataLaunchTemplateSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplate**


> 节点模版, 主要用于对kubernetes集群节点的通用特性设置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipScaleInWhenFailed | boolean (formatted boolean)|  | 缩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| allowSkipScaleOutWhenFailed | boolean (formatted boolean)|  | 扩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| annotations | map of string|  |  |  |
| bcsScaleInAddons | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleInAddons|  |  |  |
| bcsScaleOutAddons | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleOutAddons|  |  |  |
| createTime | string|  |  |  |
| creator | string|  |  |  |
| dataDisks | []ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买. 多盘数据盘挂载信息：新建节点时请确保购买CVM的参数传递了购买多个数据盘的信息该参数主要是针对自动生产机器且设置多数据盘场景 |  |
| desc | string|  | 节点模板描述信息 |  |
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| extraInfo | map of string|  |  |  |
| gpuArgs | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgs|  |  |  |
| image | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateImage|  |  |  |
| labels | map of string|  |  |  |
| maxPodsPerNode | int64 (formatted integer)|  | 每个节点的最大Pod数量 |  |
| module | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateModule|  |  |  |
| mountTarget | string|  | 节点数据盘挂载点, 默认不挂载数据盘。已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载。无数据盘或有多块数据盘的云主机此设置不生效。BCS默认/data |  |
| name | string|  | 节点模版名称 |  |
| nodeOS | string|  | 节点模版OS |  |
| nodeTemplateID | string|  | 节点模版ID |  |
| preStartUserScript | string|  | 扩容节点前置脚本, base64 编码的用户脚本. 在初始化节点之前执行，注意目前只对添加已有节点生效 |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| runtime | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateRuntime|  |  |  |
| scaleInExtraAddons | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleInExtraAddons|  |  |  |
| scaleInPostScript | string|  | 缩容节点后置脚本, base64编码的用户脚本. |  |
| scaleInPreScript | string|  | 缩容节点前置脚本, base64编码的用户脚本. |  |
| scaleOutExtraAddons | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean (formatted boolean)|  | 是否跳过执行系统初始化, 默认执行 |  |
| taints | []ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| updateTime | string|  |  |  |
| updater | string|  |  |  |
| userScript | string|  | 扩容节点后置脚本, base64编码的用户脚本. k8s组件运行后执行，需要保证脚本的可重入和重试逻辑。如果要求节点需要在进行初始化完成后才可加入调度可配合 unschedulable 参数使用, 在 userScript 最后初始化完成后, 添加 kubectl uncordon nodename --kubeconfig=/root/.kube/config 命令使节点加入调度 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleInAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleInAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleInAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleOutAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleOutAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateBcsScaleOutAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgs**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cuda | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCuda|  |  |  |
| cudnn | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCudnn|  |  |  |
| customDriver | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCustomDriver|  |  |  |
| driver | ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsDriver|  |  |  |
| migEnable | boolean (formatted boolean)|  | 是否启用MIG特性 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCuda**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCudnn**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| devName | string|  | cuDNN的Dev名字 |  |
| docName | string|  | cuDNN的Doc名字 |  |
| name | string|  | cuDNN的名字 |  |
| version | string|  | cuDNN的版本 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsCustomDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| address | string|  | 自定义GPU驱动地址链接 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateGpuArgsDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateImage**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateRuntime**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | string|  | 容器运行时 |  |
| runtimeVersion | string|  | 运行时版本 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleInExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleInExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleInExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleOutExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleOutExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateScaleOutExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupOKBodyDataNodeTemplateTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerUpdateNodeGroupOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerUpdateNodeGroupParamsBodyAutoScaling**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoScalingID | string|  | 弹性伸缩组ID，信息创建时忽略，该ID为自建ID或者云设施ID |  |
| autoScalingName | string|  | 弹性伸缩组名称 |  |
| autoUpgrade | boolean (formatted boolean)|  | 开启自动升级GKE版本功能,如果集群为发布版本,则强制开启 |  |
| defaultCooldown | int64 (formatted integer)|  | 弹性伸缩组默认扩缩容冷却周期，默认300秒 |  |
| desiredSize | int64 (formatted integer)|  | 弹性伸缩期望实例数 |  |
| maxSize | int64 (formatted integer)|  | 弹性伸缩组最大实例数 |  |
| minSize | int64 (formatted integer)|  | 伸缩组最小实例数 |  |
| multiZoneSubnetPolicy | string|  | 可选值：PRIORITY，EQUALITY。PRIORITY：默认在高优先级的子网与可用区创建实例。EQUALITY：所有可用区、子网机会均衡 |  |
| replaceUnhealthy | boolean (formatted boolean)|  | 开启监控不健康节点替换服务，默认false。连续1分钟ping不通则视为不健康！ |  |
| retryPolicy | string|  | 创建实例失败时的重试策略：IMMEDIATE_RETRY(最大5次)，INCREMENTAL_INTERVALS，NO_RETRY。默认值为IMMEDIATE_RETRY。扩容时默认会组合subnet和zone信息逐一尝试实例创建直到最大失败次数或者所有组合全失败。 |  |
| scalingMode | string|  | 扩容模式，CLASSIC_SCALING(默认)：扩容时创建新实例，缩容时销毁实例；WAKE_UP_STOPPED_SCALING：缩容时关机不销毁，扩容时优先唤醒关机实例。 |  |
| serviceRole | string|  | 创建eks节点池时指定的iam角色名称 |  |
| subnetIDs | []string|  | 弹性伸缩组扩容实例首选的的子网ID，多个子网ID按照先后顺序确认优先级，靠前优先级高 |  |
| timeRanges | []ClusterManagerUpdateNodeGroupParamsBodyAutoScalingTimeRangesItems0|  | 定时自动扩容集群节点配置信息 |  |
| vpcID | string|  | 弹性伸缩组所属VPC ID |  |
| zones | []string|  | 弹性伸缩组扩容实例首选的可用区ID，多个可用区按照先后顺序确认优先级，靠前优先级高 |  |



**ClusterManagerUpdateNodeGroupParamsBodyAutoScalingTimeRangesItems0**


> autoScale定时扩缩配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desiredNum | int64 (formatted integer)|  | 期望节点数 |  |
| name | string|  | 定时扩缩job名称 |  |
| schedule | string|  | 定时cron时间 |  |
| zone | string|  | 定时cron时区,如果不执行时区则默认是服务所在时区 |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplate**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| charge | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateCharge|  |  |  |
| dataDisks | []ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| imageInfo | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateInternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 yunti默认关闭 |  |
| keyPair | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateKeyPair|  |  |  |
| launchConfigurationID | string|  | 节点配置模板ID，自建ID或者云上ID |  |
| launchConfigureName | string|  | 配置模板名称 |  |
| networkTag | []string|  | 网络标记，标记代表应用于每个节点的防火墙规则 |  |
| projectID | string|  | 项目标识（该配置中可能指蓝鲸项目等，非 BCS 项目时含义以实际业务为准）。BCS 场景下为 32 位十六进制字符串；与 projectCode 区别：projectID 为内部标识，projectCode 为人类可读的项目英文编码。 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，BCS默认使用各云信息配置中预留的信息 |  |
| selector | map of string|  |  |  |
| systemDisk | ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateSystemDisk|  |  |  |
| userData | string|  | 经过 Base64 编码后的自定义数据，最大长度不超过16KB。 |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateCharge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateInternetAccess**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bandwidthPackageId | string|  | 共享带宽包ID |  |
| internetChargeType | string|  | 网络计费类型。BANDWIDTH_PREPAID：预付费按带宽结算，TRAFFIC_POSTPAID_BY_HOUR：流量按小时后付费，BANDWIDTH_POSTPAID_BY_HOUR：带宽按小时后付费，BANDWIDTH_PACKAGE：带宽包用户 |  |
| internetMaxBandwidth | string|  | 公网出带宽上限，单位：Mbps。默认值：0Mbps。 |  |
| nodePublicIPPrefixID | string|  | 节点公网ip前缀ID |  |
| publicAccessCidrs | []string|  | 允许访问Kubernetes API Server的公网Ip cidr |  |
| publicIP | string|  | 公网ip地址 |  |
| publicIPAssigned | boolean (formatted boolean)|  | 是否分配公网IP。默认为false。 |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateNodeGroupParamsBodyLaunchTemplateSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplate**


> 节点模版, 主要用于对kubernetes集群节点的通用特性设置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipScaleInWhenFailed | boolean (formatted boolean)|  | 缩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| allowSkipScaleOutWhenFailed | boolean (formatted boolean)|  | 扩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| annotations | map of string|  |  |  |
| bcsScaleInAddons | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleInAddons|  |  |  |
| bcsScaleOutAddons | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleOutAddons|  |  |  |
| createTime | string|  |  |  |
| creator | string|  |  |  |
| dataDisks | []ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买. 多盘数据盘挂载信息：新建节点时请确保购买CVM的参数传递了购买多个数据盘的信息该参数主要是针对自动生产机器且设置多数据盘场景 |  |
| desc | string|  | 节点模板描述信息 |  |
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| extraInfo | map of string|  |  |  |
| gpuArgs | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgs|  |  |  |
| image | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateImage|  |  |  |
| labels | map of string|  |  |  |
| maxPodsPerNode | int64 (formatted integer)|  | 每个节点的最大Pod数量 |  |
| module | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateModule|  |  |  |
| mountTarget | string|  | 节点数据盘挂载点, 默认不挂载数据盘。已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载。无数据盘或有多块数据盘的云主机此设置不生效。BCS默认/data |  |
| name | string|  | 节点模版名称 |  |
| nodeOS | string|  | 节点模版OS |  |
| nodeTemplateID | string|  | 节点模版ID |  |
| preStartUserScript | string|  | 扩容节点前置脚本, base64 编码的用户脚本. 在初始化节点之前执行，注意目前只对添加已有节点生效 |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| runtime | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateRuntime|  |  |  |
| scaleInExtraAddons | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleInExtraAddons|  |  |  |
| scaleInPostScript | string|  | 缩容节点后置脚本, base64编码的用户脚本. |  |
| scaleInPreScript | string|  | 缩容节点前置脚本, base64编码的用户脚本. |  |
| scaleOutExtraAddons | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean (formatted boolean)|  | 是否跳过执行系统初始化, 默认执行 |  |
| taints | []ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| updateTime | string|  |  |  |
| updater | string|  |  |  |
| userScript | string|  | 扩容节点后置脚本, base64编码的用户脚本. k8s组件运行后执行，需要保证脚本的可重入和重试逻辑。如果要求节点需要在进行初始化完成后才可加入调度可配合 unschedulable 参数使用, 在 userScript 最后初始化完成后, 添加 kubectl uncordon nodename --kubeconfig=/root/.kube/config 命令使节点加入调度 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleInAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleInAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleInAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleOutAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleOutAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateBcsScaleOutAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgs**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cuda | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCuda|  |  |  |
| cudnn | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCudnn|  |  |  |
| customDriver | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCustomDriver|  |  |  |
| driver | ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsDriver|  |  |  |
| migEnable | boolean (formatted boolean)|  | 是否启用MIG特性 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCuda**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCudnn**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| devName | string|  | cuDNN的Dev名字 |  |
| docName | string|  | cuDNN的Doc名字 |  |
| name | string|  | cuDNN的名字 |  |
| version | string|  | cuDNN的版本 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsCustomDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| address | string|  | 自定义GPU驱动地址链接 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateGpuArgsDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateImage**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateRuntime**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | string|  | 容器运行时 |  |
| runtimeVersion | string|  | 运行时版本 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleInExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleInExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleInExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleOutExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleOutExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateScaleOutExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateNodeGroupParamsBodyNodeTemplateTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |
