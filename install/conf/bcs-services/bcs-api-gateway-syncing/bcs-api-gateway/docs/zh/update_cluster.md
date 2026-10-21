### 描述

更新Cluster

更新Cluster信息，该接口仅用于更新信息，不会触发任何集群管理的自动化动作。。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| body | `body` | ClusterManagerUpdateClusterBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateClusterDefaultBody

##### 内联模型

**ClusterManagerUpdateClusterBody**


> 更新集群信息请求，该接口仅用于信息更新和修正，不会触发任何行为动作。如果需要执行自动化相关流程，请根据需要使用对应的接口



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ExtraInfo | map of string|  |  |  |
| bcsAddons | map ofClusterManagerUpdateClusterParamsBodyBcsAddonsAnon|  |  |  |
| businessID | string|  | CMDB业务ID |  |
| cloudAccountID | string|  | 集群关联的云账号ID,默认为空。主要功能是获取云账号凭证。私有云场景下统一管理cloud账号，而公有云场景纳管不同云账号凭证 |  |
| clusterAdvanceSettings | ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettings|  |  |  |
| clusterBasicSettings | ClusterManagerUpdateClusterParamsBodyClusterBasicSettings|  |  |  |
| clusterCategory | string|  | 集群类别，主要用于区分该集群是否是自建、导入(builder/importer), 默认是自建 |  |
| clusterID | string| 是 | 集群ID，例如BCS-K8S-000000 |  |
| clusterName | string|  | 集群名称 |  |
| clusterType | string|  | 集群类型, 例如[federation, single], federation表示为联邦集群，single表示独立集群，默认为single |  |
| createTime | string|  | 更新集群创建时间,格式为(time.Now().Format(time.RFC3339)) |  |
| creator | string|  | 创建人员名称 |  |
| description | string|  | 集群简要描述 |  |
| engineType | string|  | 引擎类型，[k8s, mesos]，默认k8s |  |
| environment | string|  | 集群环境, 例如[prod, debug, stag] |  |
| extraAddons | map ofClusterManagerUpdateClusterParamsBodyExtraAddonsAnon|  |  |  |
| extraClusterID | string|  | 导入集群的额外集群ID标识信息,默认时空值 |  |
| federationClusterID | string|  | 如果该集群为联邦集群的一部分，用该ID记录联邦Host信息 |  |
| importCategory | string|  | 当clusterCategory参数为importer时, 主要用于区分该集群是kubeConfig/cloud导入,默认为空 |  |
| isCommonCluster | boolean (formatted boolean)|  | 是否为公共集群,默认false |  |
| isExclusive | boolean (formatted boolean)|  | 是否为业务独占集群 |  |
| isMixed | boolean (formatted boolean)|  | 标记该集群是否是混部集群 |  |
| is_shared | boolean (formatted boolean)|  | 是否为共享集群,默认false |  |
| labels | map of string|  |  |  |
| labels2 | ClusterManagerUpdateClusterParamsBodyLabels2|  |  |  |
| manageType | string|  | 集群管理类型，公有云时生效，MANAGED_CLUSTER(云上托管集群)，INDEPENDENT_CLUSTER(独立集群，自行维护) |  |
| master | []string|  | 集群master IP列表 |  |
| moduleID | string|  | 集群模块ID(CA添加节点时模块转移) |  |
| networkSettings | ClusterManagerUpdateClusterParamsBodyNetworkSettings|  |  |  |
| networkType | string|  | 集群网络类型(underlay/overlay) |  |
| nodeSettings | ClusterManagerUpdateClusterParamsBodyNodeSettings|  |  |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string|  | 集群提供者，云模板ID，影响具体云设施管理流程 |  |
| region | string|  | 集群所在地域 |  |
| sharedRanges | ClusterManagerUpdateClusterParamsBodySharedRanges|  |  |  |
| status | string|  | 集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED |  |
| systemID | string|  | kubernetes集群在各云平台上资源ID |  |
| updater | string|  | 操作人员名称 |  |
| vpcID | string|  | 集群私有网络id，云上集群填写 |  |



**ClusterManagerUpdateClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateClusterOKBody**


> 更新集群信息返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerUpdateClusterOKBodyData|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerUpdateClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerUpdateClusterOKBodyData**


> 记录集群信息，用于维护集群状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoGenerateMasterNodes | boolean (formatted boolean)|  | 创建集群是否使用已存在节点, 默认false, 即使用已经存在的节点, 从创建集群参数中获取 |  |
| bcsAddons | map ofClusterManagerUpdateClusterOKBodyDataBcsAddonsAnon|  |  |  |
| businessID | string|  | 项目绑定的业务ID, 集群所关联的业务ID，该ID来源于CMDB |  |
| cloudAccountID | string|  | 集群关联的云账号ID,默认为空。主要功能是获取云账号凭证。私有云场景下统一管理cloud账号，而公有云场景纳管不同云账号凭证 |  |
| clusterAdvanceSettings | ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettings|  |  |  |
| clusterBasicSettings | ClusterManagerUpdateClusterOKBodyDataClusterBasicSettings|  |  |  |
| clusterCategory | string|  | 集群类别，主要用于区分该集群是否是自建、导入(builder/importer), 默认是自建 |  |
| clusterID | string|  | BCS内部使用集群ID，格式为BCS-K8S-XXXXX |  |
| clusterIamRole | string|  | aws的iam role, 用于创建eks集群 |  |
| clusterName | string|  | 集群名称 |  |
| clusterType | string|  | 集群类型，federation、single |  |
| createTime | string|  |  |  |
| creator | string|  | 集群信息创建者 |  |
| description | string|  | 集群简要描述 |  |
| engineType | string|  | 集群编排引擎，k8s、mesos |  |
| environment | string|  | 集群所对应的环境信息，stag，debug，prod |  |
| extraAddons | map ofClusterManagerUpdateClusterOKBodyDataExtraAddonsAnon|  |  |  |
| extraClusterID | string|  | 导入集群的额外集群ID标识信息,默认时空值(废弃) |  |
| extraInfo | map of string|  |  |  |
| federationClusterID | string|  | 如果该集群为联邦集群的一部分，用该ID记录联邦Host信息 |  |
| importCategory | string|  | 当clusterCategory参数为importer时, 主要用于区分该集群是kubeConfig/cloud导入,默认为空 |  |
| isCommonCluster | boolean (formatted boolean)|  | 是否为公共集群,默认false(废弃) |  |
| isExclusive | boolean (formatted boolean)|  | 是否业务独占集群，默认为false |  |
| isMixed | boolean (formatted boolean)|  | 标记该集群是否是混部集群 |  |
| is_shared | boolean (formatted boolean)|  | 是否为共享集群,默认false |  |
| kubeConfig | string|  | 导入集群时若是kubeConfig导入则存储 |  |
| labels | map of string|  |  |  |
| manageType | string|  | 集群管理类型，公有云时生效，MANAGED_CLUSTER(云上托管集群)，INDEPENDENT_CLUSTER(独立集群，自行维护) |  |
| master | map ofClusterManagerUpdateClusterOKBodyDataMasterAnon|  |  |  |
| message | string|  | 集群状态异常时的失败原因展示 |  |
| moduleID | string|  | 集群模块ID(CA添加节点时模块转移 废弃) |  |
| networkSettings | ClusterManagerUpdateClusterOKBodyDataNetworkSettings|  |  |  |
| networkType | string|  | 集群网络类型，(underlay/overlay) |  |
| nodeSettings | ClusterManagerUpdateClusterOKBodyDataNodeSettings|  |  |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string|  | 集群云配置，为云类型ID |  |
| region | string|  | 集群区域信息 |  |
| sharedRanges | ClusterManagerUpdateClusterOKBodyDataSharedRanges|  |  |  |
| status | string|  | 集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION |  |
| systemID | string|  | kubernetes集群在各云平台上资源ID |  |
| template | []ClusterManagerUpdateClusterOKBodyDataTemplateItems0|  | autoGenerateMasterNodes为true时, 创建集群自动生成instance, 配置instance模板信息 |  |
| updateTime | string|  |  |  |
| updater | string|  | 集群信息更新者 |  |
| vpcID | string|  | 集群所在vpc，部分环境下非必填 |  |



**ClusterManagerUpdateClusterOKBodyDataBcsAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| IPVS | boolean (formatted boolean)|  | Kubernetes集群IPVS特性，默认为false |  |
| auditEnabled | boolean (formatted boolean)|  | 是否开启审计开关 |  |
| clusterConnectSetting | ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSetting|  |  |  |
| containerRuntime | string|  | 集群容器运行时实现，可选Docker、Containerd |  |
| deletionProtection | boolean (formatted boolean)|  | 是否启用集群删除保护 |  |
| enableHa | boolean (formatted boolean)|  | 自建集群是否开启高可用 |  |
| extraArgs | map of string|  |  |  |
| networkType | string|  | 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR.也支持CiliumOverlay模式（(flannel、calico、或者云上自研网络插件等)） |  |
| runtimeVersion | string|  | 容器运行时版本 |  |



**ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSetting**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| domain | string|  | 用户设置访问端点域名 |  |
| internet | ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSettingInternet|  |  |  |
| isExtranet | boolean (formatted boolean)|  | isExtranet集群内外网访问(true外网访问 false内网访问, 默认false) |  |
| securityGroup | string|  | 使用的安全组，只有外网访问需要传递（开启外网访问时必传） |  |
| subnetId | string|  | 集群端口所在的子网ID(开启内网访问所需) |  |



**ClusterManagerUpdateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSettingInternet**



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



**ClusterManagerUpdateClusterOKBodyDataClusterBasicSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| OS | string|  | 集群使用OS信息(镜像ID/公共镜像, 为空时使用默认镜像ID) |  |
| area | ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsArea|  |  |  |
| clusterLevel | string|  | 集群等级，针对托管集群生效 |  |
| clusterTags | map of string|  |  |  |
| isAutoUpgradeClusterLevel | boolean (formatted boolean)|  | 是否开启自动变配集群等级，针对托管集群生效 |  |
| module | ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsModule|  |  |  |
| subnetID | string|  | 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡 |  |
| upgradePolicy | ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsUpgradePolicy|  |  |  |
| version | string|  | kubernetes使用的版本 |  |
| versionName | string|  | kubernetes使用的版本名称 |  |



**ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| masterModuleID | string|  | CC模块ID,master节点转移模块 |  |
| masterModuleName | string|  | CC模块名称,master节点转移模块 |  |
| workerModuleID | string|  | CC模块ID,worker节点转移模块 |  |
| workerModuleName | string|  | CC模块名称,worker节点转移模块 |  |



**ClusterManagerUpdateClusterOKBodyDataClusterBasicSettingsUpgradePolicy**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| supportType | string|  | aws EKS升级策略类型(EXTENDED|STANDARD) |  |



**ClusterManagerUpdateClusterOKBodyDataExtraAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateClusterOKBodyDataMasterAnon**


> 记录节点信息，用于维护节点状态，被cluster与nodegroup使用



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 节点CPU信息 |  |
| GPU | int64 (formatted integer)|  | 节点GPU信息，可能为0 |  |
| VPC | string|  | 节点所属的VPC ID |  |
| chargeType | string|  | 机型计费模型，POSTPAID_BY_HOUR按小时后付费，PREPAID包年包月 |  |
| clusterID | string|  | 节点所属的ClusterID |  |
| dataDiskNum | int64 (formatted integer)|  | 节点数据盘数量 |  |
| deviceID | string|  | CA instance设备ID号 |  |
| failedReason | string|  | 上下架节点失败时原因 |  |
| innerIP | string|  | 节点内网主IP |  |
| innerIPv6 | string|  | 节点ipv6地址 |  |
| instanceType | string|  | 机型信息字符串，可能为空 |  |
| isGpuNode | boolean (formatted boolean)|  | 节点是否是GPU节点 |  |
| mem | int64 (formatted integer)|  | 节点内存信息 |  |
| nodeGroupID | string|  | 节点所属的NodeGroup ID, 可能为空(仅CA节点归属于nodeGroup) |  |
| nodeID | string|  | 节点云instanceID |  |
| nodeName | string|  | 节点名称, 存储各个云节点名称 |  |
| nodeTemplateID | string|  | 节点关联的节点模板ID, 可为空 |  |
| nodeType | string|  | 节点类型, 主要包含CVM/IDC, 可为空默认是CVM |  |
| passwd | string|  | 实例初始化密码, 需要进行解密 |  |
| region | string|  | 节点所属区域，例如广州，深圳，加州等 |  |
| status | string|  | 节点状态信息（根据弹性伸缩定义的），RUNNING, DELETEING, CREATING, FAILED, INITIALIZATION |  |
| taskID | string|  | 节点对应的任务ID, 仅在申请资源时追踪节点数据 |  |
| zone | int64 (formatted integer)|  | 节点可用区ID，可能为空; 该信息主要使用于CA |  |
| zoneID | string|  | 节点可用区，可能为空 |  |
| zoneName | string|  | 节点所在可用区名称 |  |



**ClusterManagerUpdateClusterOKBodyDataNetworkSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidrStep | int64 (formatted integer)|  | 自动分配cidr步长设置(默认是生产环境4096、其他环境2048) |  |
| claimExpiredSeconds | int64 (formatted integer)|  | VPC-CNI网络模式下,弹性网卡IP回收时间 |  |
| clusterIPv4CIDR | string|  | 集群总IPv4地址池 |  |
| clusterIPv6CIDR | string|  | 集群总IPv6地址池 |  |
| clusterIpType | string|  | 集群使用的网络类型(ipv4/ipv6/dual),目前默认是ipv4类型 |  |
| enableVPCCni | boolean (formatted boolean)|  | enableVPCCni是否开启VPC-CNI网络模式(使能vpc-cni模式时,新增流程开启) |  |
| eniSubnetIDs | []string|  | VPC-CNI网络模式下,弹性网卡的子网ID |  |
| isStaticIpMode | boolean (formatted boolean)|  | 集群VPC-CNI模式是否为非固定IP，默认: FALSE 非固定IP |  |
| maxNodePodNum | int64 (formatted integer)|  | 节点上最大Pod数量 |  |
| maxServiceNum | int64 (formatted integer)|  | 集群最大的Service数量 |  |
| multiClusterCIDR | []string|  | 集群扩展CIDR,不同cloud配置不同 |  |
| networkMode | string|  | 集群VpcCni网络模式(目前支持共享网卡模式tke-route-eni和独占网卡模式tke-direct-eni, 默认情况下是共享网卡模式) |  |
| serviceIPv4CIDR | string|  | 集群service使用的IPv4地址池 |  |
| serviceIPv6CIDR | string|  | 集群service使用的IPv6地址池 |  |
| status | string|  | 集群网络状态, 主要是变更网络组件过程中记录网络状态 |  |
| subnetSource | ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSource|  |  |  |



**ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSource**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| existed | ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSourceExisted|  |  |  |
| new | []ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSourceNewItems0|  | VPC-CNI网络模式下申请subnet |  |



**ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSourceExisted**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ids | []string|  | 直接使用已存在子网IDs |  |



**ClusterManagerUpdateClusterOKBodyDataNetworkSettingsSubnetSourceNewItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ipCnt | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的IP个数 |  |
| mask | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的网络掩码大小 |  |
| zone | string|  | 申请vpc-cni网络subnet时region下的可用区 |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| labels | map of string|  |  |  |
| masterLogin | ClusterManagerUpdateClusterOKBodyDataNodeSettingsMasterLogin|  |  |  |
| masterSecurityGroups | []string|  | master节点安全组设置 |  |
| mountTarget | string|  | 节点目录盘挂载点，BCS默认/data |  |
| taints | []ClusterManagerUpdateClusterOKBodyDataNodeSettingsTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| workerLogin | ClusterManagerUpdateClusterOKBodyDataNodeSettingsWorkerLogin|  |  |  |
| workerSecurityGroups | []string|  | worker节点安全组设置 |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettingsMasterLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerUpdateClusterOKBodyDataNodeSettingsMasterLoginKeyPair|  |  |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettingsMasterLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettingsTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettingsWorkerLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerUpdateClusterOKBodyDataNodeSettingsWorkerLoginKeyPair|  |  |  |



**ClusterManagerUpdateClusterOKBodyDataNodeSettingsWorkerLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateClusterOKBodyDataSharedRanges**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bizs | []string|  |  |  |
| projectIdOrCodes | []string|  |  |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| applyNum | int64 (formatted integer)|  | 申请实例数量 |  |
| charge | ClusterManagerUpdateClusterOKBodyDataTemplateItems0Charge|  |  |  |
| cloudDataDisks | []ClusterManagerUpdateClusterOKBodyDataTemplateItems0CloudDataDisksItems0|  | 节点数据盘信息(包含磁盘挂载信息以及格式化信息，用户若不需要可自行处理格式化挂载)，无默认设置则不购买 |  |
| dataDisks | []ClusterManagerUpdateClusterOKBodyDataTemplateItems0DataDisksItems0|  | 数据盘信息，无默认设置则不购买(已废弃掉，统一通过 CloudDataDisk字段兼容) |  |
| dockerGraphPath | string|  | 节点运行时容器目录，BCS默认设置/data/bcs/service/docker |  |
| imageInfo | ClusterManagerUpdateClusterOKBodyDataTemplateItems0ImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，tencentCloud包含POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。aws包含ON_DEMAND按需付费,SPOT竞价付费。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerUpdateClusterOKBodyDataTemplateItems0InternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 qcloud默认开启 |  |
| keyPair | ClusterManagerUpdateClusterOKBodyDataTemplateItems0KeyPair|  |  |  |
| nodeRole | string|  | 节点角色, 取值:MASTER_ETCD, WORKER。MASTER_ETCD只有在创建 INDEPENDENT_CLUSTER 独立集群时需要指定 |  |
| region | string|  | instance地域信息 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，使用系统默认的安全组信息 |  |
| subnetID | string|  | 子网ID, 和可用区域一一对应 |  |
| systemDisk | ClusterManagerUpdateClusterOKBodyDataTemplateItems0SystemDisk|  |  |  |
| vpcID | string|  | 私有网络ID |  |
| zone | string|  | 可用区域 |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0Charge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0CloudDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0DataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0ImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0InternetAccess**



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



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0KeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateClusterOKBodyDataTemplateItems0SystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerUpdateClusterOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerUpdateClusterParamsBodyBcsAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| IPVS | boolean (formatted boolean)|  | Kubernetes集群IPVS特性，默认为false |  |
| auditEnabled | boolean (formatted boolean)|  | 是否开启审计开关 |  |
| clusterConnectSetting | ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettingsClusterConnectSetting|  |  |  |
| containerRuntime | string|  | 集群容器运行时实现，可选Docker、Containerd |  |
| deletionProtection | boolean (formatted boolean)|  | 是否启用集群删除保护 |  |
| enableHa | boolean (formatted boolean)|  | 自建集群是否开启高可用 |  |
| extraArgs | map of string|  |  |  |
| networkType | string|  | 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR.也支持CiliumOverlay模式（(flannel、calico、或者云上自研网络插件等)） |  |
| runtimeVersion | string|  | 容器运行时版本 |  |



**ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettingsClusterConnectSetting**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| domain | string|  | 用户设置访问端点域名 |  |
| internet | ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettingsClusterConnectSettingInternet|  |  |  |
| isExtranet | boolean (formatted boolean)|  | isExtranet集群内外网访问(true外网访问 false内网访问, 默认false) |  |
| securityGroup | string|  | 使用的安全组，只有外网访问需要传递（开启外网访问时必传） |  |
| subnetId | string|  | 集群端口所在的子网ID(开启内网访问所需) |  |



**ClusterManagerUpdateClusterParamsBodyClusterAdvanceSettingsClusterConnectSettingInternet**



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



**ClusterManagerUpdateClusterParamsBodyClusterBasicSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| OS | string|  | 集群使用OS信息(镜像ID/公共镜像, 为空时使用默认镜像ID) |  |
| area | ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsArea|  |  |  |
| clusterLevel | string|  | 集群等级，针对托管集群生效 |  |
| clusterTags | map of string|  |  |  |
| isAutoUpgradeClusterLevel | boolean (formatted boolean)|  | 是否开启自动变配集群等级，针对托管集群生效 |  |
| module | ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsModule|  |  |  |
| subnetID | string|  | 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡 |  |
| upgradePolicy | ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsUpgradePolicy|  |  |  |
| version | string|  | kubernetes使用的版本 |  |
| versionName | string|  | kubernetes使用的版本名称 |  |



**ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| masterModuleID | string|  | CC模块ID,master节点转移模块 |  |
| masterModuleName | string|  | CC模块名称,master节点转移模块 |  |
| workerModuleID | string|  | CC模块ID,worker节点转移模块 |  |
| workerModuleName | string|  | CC模块名称,worker节点转移模块 |  |



**ClusterManagerUpdateClusterParamsBodyClusterBasicSettingsUpgradePolicy**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| supportType | string|  | aws EKS升级策略类型(EXTENDED|STANDARD) |  |



**ClusterManagerUpdateClusterParamsBodyExtraAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerUpdateClusterParamsBodyLabels2**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| values | map of string|  |  |  |



**ClusterManagerUpdateClusterParamsBodyNetworkSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cidrStep | int64 (formatted integer)|  | 自动分配cidr步长设置(默认是生产环境4096、其他环境2048) |  |
| claimExpiredSeconds | int64 (formatted integer)|  | VPC-CNI网络模式下,弹性网卡IP回收时间 |  |
| clusterIPv4CIDR | string|  | 集群总IPv4地址池 |  |
| clusterIPv6CIDR | string|  | 集群总IPv6地址池 |  |
| clusterIpType | string|  | 集群使用的网络类型(ipv4/ipv6/dual),目前默认是ipv4类型 |  |
| enableVPCCni | boolean (formatted boolean)|  | enableVPCCni是否开启VPC-CNI网络模式(使能vpc-cni模式时,新增流程开启) |  |
| eniSubnetIDs | []string|  | VPC-CNI网络模式下,弹性网卡的子网ID |  |
| isStaticIpMode | boolean (formatted boolean)|  | 集群VPC-CNI模式是否为非固定IP，默认: FALSE 非固定IP |  |
| maxNodePodNum | int64 (formatted integer)|  | 节点上最大Pod数量 |  |
| maxServiceNum | int64 (formatted integer)|  | 集群最大的Service数量 |  |
| multiClusterCIDR | []string|  | 集群扩展CIDR,不同cloud配置不同 |  |
| networkMode | string|  | 集群VpcCni网络模式(目前支持共享网卡模式tke-route-eni和独占网卡模式tke-direct-eni, 默认情况下是共享网卡模式) |  |
| serviceIPv4CIDR | string|  | 集群service使用的IPv4地址池 |  |
| serviceIPv6CIDR | string|  | 集群service使用的IPv6地址池 |  |
| status | string|  | 集群网络状态, 主要是变更网络组件过程中记录网络状态 |  |
| subnetSource | ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSource|  |  |  |



**ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSource**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| existed | ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSourceExisted|  |  |  |
| new | []ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSourceNewItems0|  | VPC-CNI网络模式下申请subnet |  |



**ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSourceExisted**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ids | []string|  | 直接使用已存在子网IDs |  |



**ClusterManagerUpdateClusterParamsBodyNetworkSettingsSubnetSourceNewItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ipCnt | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的IP个数 |  |
| mask | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的网络掩码大小 |  |
| zone | string|  | 申请vpc-cni网络subnet时region下的可用区 |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| labels | map of string|  |  |  |
| masterLogin | ClusterManagerUpdateClusterParamsBodyNodeSettingsMasterLogin|  |  |  |
| masterSecurityGroups | []string|  | master节点安全组设置 |  |
| mountTarget | string|  | 节点目录盘挂载点，BCS默认/data |  |
| taints | []ClusterManagerUpdateClusterParamsBodyNodeSettingsTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| workerLogin | ClusterManagerUpdateClusterParamsBodyNodeSettingsWorkerLogin|  |  |  |
| workerSecurityGroups | []string|  | worker节点安全组设置 |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettingsMasterLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerUpdateClusterParamsBodyNodeSettingsMasterLoginKeyPair|  |  |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettingsMasterLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettingsTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettingsWorkerLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerUpdateClusterParamsBodyNodeSettingsWorkerLoginKeyPair|  |  |  |



**ClusterManagerUpdateClusterParamsBodyNodeSettingsWorkerLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerUpdateClusterParamsBodySharedRanges**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bizs | []string|  |  |  |
| projectIdOrCodes | []string|  |  |  |
