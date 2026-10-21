### 描述

创建Cluster

创建Cluster，集群类型兼容联邦集群和独立集群。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerCreateClusterBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCreateClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCreateClusterDefaultBody

##### 内联模型

**ClusterManagerCreateClusterBody**


> 创建集群请求，如果集群已经存在，直接报错。更新集群信息请使用Update接口



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoGenerateMasterNodes | boolean (formatted boolean)|  | 创建集群Master节点来源, 默认false表示使用已有节点; 为true时, 需要生成master instance |  |
| bcsAddons | map ofClusterManagerCreateClusterParamsBodyBcsAddonsAnon|  |  |  |
| businessID | string| 是 | CMDB业务ID |  |
| cloudAccountID | string|  | 集群关联的云账号ID,默认为空。主要功能是获取云账号凭证。私有云场景下统一管理cloud账号，而公有云场景纳管不同云账号凭证 |  |
| cloudID | string|  | kubernetes集群上资源ID |  |
| clusterAdvanceSettings | ClusterManagerCreateClusterParamsBodyClusterAdvanceSettings|  |  |  |
| clusterBasicSettings | ClusterManagerCreateClusterParamsBodyClusterBasicSettings|  |  |  |
| clusterCategory | string|  | 集群类别，主要用于区分该集群是否是自建、导入(builder/importer), 默认是自建 |  |
| clusterID | string| 是 | 集群ID，例如BCS-K8S-000000(写入信息时直接指定clusterID，新创建时需平台生成) |  |
| clusterIamRole | string|  | aws的iam role, 用于创建eks集群 |  |
| clusterName | string| 是 | 集群名称 |  |
| clusterType | string| 是 | 集群类型, 例如[federation, single], federation表示为联邦集群，single表示独立集群，默认为single |  |
| creator | string| 是 | 创建人 |  |
| description | string|  | 集群简要描述 |  |
| engineType | string| 是 | 引擎类型，默认k8s |  |
| environment | string| 是 | 集群环境, 例如[prod, debug, stag] |  |
| extraAddons | map ofClusterManagerCreateClusterParamsBodyExtraAddonsAnon|  |  |  |
| extraClusterID | string|  | 导入集群的额外集群ID标识信息,默认时空值(废弃) |  |
| extraInfo | map of string|  |  |  |
| federationClusterID | string|  | 如果该集群为联邦集群的一部分，用该ID记录联邦Host信息 |  |
| initLoginPassword | string|  | 重装master节点的系统时，初始化password，机器被托管情况下有效 |  |
| instances | []ClusterManagerCreateClusterParamsBodyInstancesItems0|  | 创建集群时 autoGenerateMasterNodes 为true, 系统自动生成master节点, 需要指定instances生成的配置信息,支持不同可用区实例 |  |
| isCommonCluster | boolean (formatted boolean)|  | 是否为公共集群,默认false(废弃) |  |
| isExclusive | boolean (formatted boolean)| 是 | 是否为业务独占集群 |  |
| isMixed | boolean (formatted boolean)|  | 标记该集群是否是混部集群, 默认是false |  |
| is_shared | boolean (formatted boolean)|  | 是否为共享集群,默认false |  |
| labels | map of string|  |  |  |
| manageType | string|  | 集群管理类型，公有云时生效，MANAGED_CLUSTER(云上托管集群)，默认是 INDEPENDENT_CLUSTER(独立集群，自行维护) |  |
| master | []string|  | 集群master IP列表 |  |
| moduleID | string|  | 集群模块ID(CA添加节点时模块转移)(废弃) |  |
| networkSettings | ClusterManagerCreateClusterParamsBodyNetworkSettings|  |  |  |
| networkType | string|  | 集群网络类型(underlay/overlay) |  |
| nodeGroups | []ClusterManagerCreateClusterParamsBodyNodeGroupsItems0|  | 创建集群时的节点池信息 |  |
| nodeSettings | ClusterManagerCreateClusterParamsBodyNodeSettings|  |  |  |
| nodeTemplateID | string|  | 创建集群上架节点时的节点模板 |  |
| nodes | []string|  | 集群node IP列表 |  |
| onlyCreateInfo | boolean (formatted boolean)|  | 仅创建集群信息，不进行真实集群创建，默认为false |  |
| projectID | string| 是 | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string| 是 | 集群提供者，云模板ID，影响具体云设施管理流程 |  |
| region | string| 是 | 集群所在地域 |  |
| sharedRanges | ClusterManagerCreateClusterParamsBodySharedRanges|  |  |  |
| status | string|  | 创建集群导入集群数据时设置集群状态,默认情况下设置空即可,仅onlyCreateInfo为true时起作用 |  |
| systemReinstall | boolean (formatted boolean)|  | 是否重装master节点的系统，机器被托管情况下有效 |  |
| vpcID | string|  | 集群私有网络id，云上集群填写 |  |



**ClusterManagerCreateClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCreateClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCreateClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCreateClusterOKBody**


> 创建集群的返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerCreateClusterOKBodyData| 是 |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| task | ClusterManagerCreateClusterOKBodyTask|  |  |  |
| web_annotations | ClusterManagerCreateClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerCreateClusterOKBodyData**


> 记录集群信息，用于维护集群状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoGenerateMasterNodes | boolean (formatted boolean)|  | 创建集群是否使用已存在节点, 默认false, 即使用已经存在的节点, 从创建集群参数中获取 |  |
| bcsAddons | map ofClusterManagerCreateClusterOKBodyDataBcsAddonsAnon|  |  |  |
| businessID | string|  | 项目绑定的业务ID, 集群所关联的业务ID，该ID来源于CMDB |  |
| cloudAccountID | string|  | 集群关联的云账号ID,默认为空。主要功能是获取云账号凭证。私有云场景下统一管理cloud账号，而公有云场景纳管不同云账号凭证 |  |
| clusterAdvanceSettings | ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettings|  |  |  |
| clusterBasicSettings | ClusterManagerCreateClusterOKBodyDataClusterBasicSettings|  |  |  |
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
| extraAddons | map ofClusterManagerCreateClusterOKBodyDataExtraAddonsAnon|  |  |  |
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
| master | map ofClusterManagerCreateClusterOKBodyDataMasterAnon|  |  |  |
| message | string|  | 集群状态异常时的失败原因展示 |  |
| moduleID | string|  | 集群模块ID(CA添加节点时模块转移 废弃) |  |
| networkSettings | ClusterManagerCreateClusterOKBodyDataNetworkSettings|  |  |  |
| networkType | string|  | 集群网络类型，(underlay/overlay) |  |
| nodeSettings | ClusterManagerCreateClusterOKBodyDataNodeSettings|  |  |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string|  | 集群云配置，为云类型ID |  |
| region | string|  | 集群区域信息 |  |
| sharedRanges | ClusterManagerCreateClusterOKBodyDataSharedRanges|  |  |  |
| status | string|  | 集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION |  |
| systemID | string|  | kubernetes集群在各云平台上资源ID |  |
| template | []ClusterManagerCreateClusterOKBodyDataTemplateItems0|  | autoGenerateMasterNodes为true时, 创建集群自动生成instance, 配置instance模板信息 |  |
| updateTime | string|  |  |  |
| updater | string|  | 集群信息更新者 |  |
| vpcID | string|  | 集群所在vpc，部分环境下非必填 |  |



**ClusterManagerCreateClusterOKBodyDataBcsAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| IPVS | boolean (formatted boolean)|  | Kubernetes集群IPVS特性，默认为false |  |
| auditEnabled | boolean (formatted boolean)|  | 是否开启审计开关 |  |
| clusterConnectSetting | ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSetting|  |  |  |
| containerRuntime | string|  | 集群容器运行时实现，可选Docker、Containerd |  |
| deletionProtection | boolean (formatted boolean)|  | 是否启用集群删除保护 |  |
| enableHa | boolean (formatted boolean)|  | 自建集群是否开启高可用 |  |
| extraArgs | map of string|  |  |  |
| networkType | string|  | 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR.也支持CiliumOverlay模式（(flannel、calico、或者云上自研网络插件等)） |  |
| runtimeVersion | string|  | 容器运行时版本 |  |



**ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSetting**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| domain | string|  | 用户设置访问端点域名 |  |
| internet | ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSettingInternet|  |  |  |
| isExtranet | boolean (formatted boolean)|  | isExtranet集群内外网访问(true外网访问 false内网访问, 默认false) |  |
| securityGroup | string|  | 使用的安全组，只有外网访问需要传递（开启外网访问时必传） |  |
| subnetId | string|  | 集群端口所在的子网ID(开启内网访问所需) |  |



**ClusterManagerCreateClusterOKBodyDataClusterAdvanceSettingsClusterConnectSettingInternet**



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



**ClusterManagerCreateClusterOKBodyDataClusterBasicSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| OS | string|  | 集群使用OS信息(镜像ID/公共镜像, 为空时使用默认镜像ID) |  |
| area | ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsArea|  |  |  |
| clusterLevel | string|  | 集群等级，针对托管集群生效 |  |
| clusterTags | map of string|  |  |  |
| isAutoUpgradeClusterLevel | boolean (formatted boolean)|  | 是否开启自动变配集群等级，针对托管集群生效 |  |
| module | ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsModule|  |  |  |
| subnetID | string|  | 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡 |  |
| upgradePolicy | ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsUpgradePolicy|  |  |  |
| version | string|  | kubernetes使用的版本 |  |
| versionName | string|  | kubernetes使用的版本名称 |  |



**ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| masterModuleID | string|  | CC模块ID,master节点转移模块 |  |
| masterModuleName | string|  | CC模块名称,master节点转移模块 |  |
| workerModuleID | string|  | CC模块ID,worker节点转移模块 |  |
| workerModuleName | string|  | CC模块名称,worker节点转移模块 |  |



**ClusterManagerCreateClusterOKBodyDataClusterBasicSettingsUpgradePolicy**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| supportType | string|  | aws EKS升级策略类型(EXTENDED|STANDARD) |  |



**ClusterManagerCreateClusterOKBodyDataExtraAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterOKBodyDataMasterAnon**


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



**ClusterManagerCreateClusterOKBodyDataNetworkSettings**



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
| subnetSource | ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSource|  |  |  |



**ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSource**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| existed | ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSourceExisted|  |  |  |
| new | []ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSourceNewItems0|  | VPC-CNI网络模式下申请subnet |  |



**ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSourceExisted**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ids | []string|  | 直接使用已存在子网IDs |  |



**ClusterManagerCreateClusterOKBodyDataNetworkSettingsSubnetSourceNewItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ipCnt | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的IP个数 |  |
| mask | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的网络掩码大小 |  |
| zone | string|  | 申请vpc-cni网络subnet时region下的可用区 |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| labels | map of string|  |  |  |
| masterLogin | ClusterManagerCreateClusterOKBodyDataNodeSettingsMasterLogin|  |  |  |
| masterSecurityGroups | []string|  | master节点安全组设置 |  |
| mountTarget | string|  | 节点目录盘挂载点，BCS默认/data |  |
| taints | []ClusterManagerCreateClusterOKBodyDataNodeSettingsTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| workerLogin | ClusterManagerCreateClusterOKBodyDataNodeSettingsWorkerLogin|  |  |  |
| workerSecurityGroups | []string|  | worker节点安全组设置 |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettingsMasterLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerCreateClusterOKBodyDataNodeSettingsMasterLoginKeyPair|  |  |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettingsMasterLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettingsTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettingsWorkerLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerCreateClusterOKBodyDataNodeSettingsWorkerLoginKeyPair|  |  |  |



**ClusterManagerCreateClusterOKBodyDataNodeSettingsWorkerLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterOKBodyDataSharedRanges**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bizs | []string|  |  |  |
| projectIdOrCodes | []string|  |  |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| applyNum | int64 (formatted integer)|  | 申请实例数量 |  |
| charge | ClusterManagerCreateClusterOKBodyDataTemplateItems0Charge|  |  |  |
| cloudDataDisks | []ClusterManagerCreateClusterOKBodyDataTemplateItems0CloudDataDisksItems0|  | 节点数据盘信息(包含磁盘挂载信息以及格式化信息，用户若不需要可自行处理格式化挂载)，无默认设置则不购买 |  |
| dataDisks | []ClusterManagerCreateClusterOKBodyDataTemplateItems0DataDisksItems0|  | 数据盘信息，无默认设置则不购买(已废弃掉，统一通过 CloudDataDisk字段兼容) |  |
| dockerGraphPath | string|  | 节点运行时容器目录，BCS默认设置/data/bcs/service/docker |  |
| imageInfo | ClusterManagerCreateClusterOKBodyDataTemplateItems0ImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，tencentCloud包含POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。aws包含ON_DEMAND按需付费,SPOT竞价付费。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerCreateClusterOKBodyDataTemplateItems0InternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 qcloud默认开启 |  |
| keyPair | ClusterManagerCreateClusterOKBodyDataTemplateItems0KeyPair|  |  |  |
| nodeRole | string|  | 节点角色, 取值:MASTER_ETCD, WORKER。MASTER_ETCD只有在创建 INDEPENDENT_CLUSTER 独立集群时需要指定 |  |
| region | string|  | instance地域信息 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，使用系统默认的安全组信息 |  |
| subnetID | string|  | 子网ID, 和可用区域一一对应 |  |
| systemDisk | ClusterManagerCreateClusterOKBodyDataTemplateItems0SystemDisk|  |  |  |
| vpcID | string|  | 私有网络ID |  |
| zone | string|  | 可用区域 |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0Charge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0CloudDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0DataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0ImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0InternetAccess**



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



**ClusterManagerCreateClusterOKBodyDataTemplateItems0KeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterOKBodyDataTemplateItems0SystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterOKBodyTask**


> 集群管理，节点管理背后所映射任务信息，主要是关联的自动化步骤（整合云api与蓝鲸标准运维）



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 任务所属集群 |  |
| commonParams | map of string|  |  |  |
| creator | string|  | 任务创建人 |  |
| currentStep | string|  | 任务正在执行的步骤名称 |  |
| end | string|  | 任务结束时间 |  |
| executionTime | int64 (formatted integer)|  | 任务执行时间 |  |
| forceTerminate | boolean (formatted boolean)|  | 任务设置为强制终止，用于流程控制 |  |
| lastUpdate | string|  | 上次更新时间 |  |
| message | string|  | 任务执行信息，报错时为错误提示信息。 |  |
| nodeGroupID | string|  | 记录CA扩缩容时任务对应的nodeGroupID |  |
| nodeIPList | []string|  | 当前任务操作的IP列表 |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| start | string|  | 任务启动时间 |  |
| status | string|  | 任务状态 |  |
| stepSequence | []string|  | 任务执行步骤 |  |
| steps | map ofClusterManagerCreateClusterOKBodyTaskStepsAnon|  |  |  |
| taskID | string|  | 任务ID，用于追溯与确认任务状态 |  |
| taskName | string|  | 任务类型中文名称 |  |
| taskType | string|  | 任务类型，例如CreateCluster，AddNodesToCluster，DeleteNodesFromCluster等 |  |
| updater | string|  | 任务更新人 |  |



**ClusterManagerCreateClusterOKBodyTaskStepsAnon**


> 任务执行的动作步骤，用于定义长时间运行的异步动作，例如机器重新初始化，插件安装等。



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkip | boolean (formatted boolean)|  | 标识该步骤失败时候是否允许用户手动跳过。当为true失败时候允许用户手动跳过, 为false失败时用户仅能通过重试解决 |  |
| end | string|  | 结束时间 |  |
| executionTime | int64 (formatted integer)|  | 执行时间 |  |
| lastUpdate | string|  | 任务上次更新时间 |  |
| link | string|  | 补充system信息所用 |  |
| maxRetry | int64 (formatted integer)|  | 该步骤最大重试次数 |  |
| message | string|  | 步骤执行信息，主要用于提示当前错误 |  |
| name | string|  | 步骤名称 |  |
| params | map of string|  |  |  |
| retry | int64 (formatted integer)|  | 该步骤重试次数 |  |
| skipOnFailed | boolean (formatted boolean)|  | 失败时自动跳过某步骤,由用户设置。当设置失败跳过时，系统自动跳过当前失败步骤，执行后续流程 |  |
| start | string|  | 启动时间 |  |
| status | string|  | 步骤状态，初始默认为NOTSTARTED |  |
| system | string|  | 所关联的系统，例如bksops/api |  |
| taskMethod | string|  | task对应的代码函数 |  |
| taskName | string|  | task对应的中文名称 |  |
| translate | string|  | 解决当使用同一taskMethod时,是否转换翻译展示文本 |  |



**ClusterManagerCreateClusterOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerCreateClusterParamsBodyBcsAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyClusterAdvanceSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| IPVS | boolean (formatted boolean)|  | Kubernetes集群IPVS特性，默认为false |  |
| auditEnabled | boolean (formatted boolean)|  | 是否开启审计开关 |  |
| clusterConnectSetting | ClusterManagerCreateClusterParamsBodyClusterAdvanceSettingsClusterConnectSetting|  |  |  |
| containerRuntime | string|  | 集群容器运行时实现，可选Docker、Containerd |  |
| deletionProtection | boolean (formatted boolean)|  | 是否启用集群删除保护 |  |
| enableHa | boolean (formatted boolean)|  | 自建集群是否开启高可用 |  |
| extraArgs | map of string|  |  |  |
| networkType | string|  | 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR.也支持CiliumOverlay模式（(flannel、calico、或者云上自研网络插件等)） |  |
| runtimeVersion | string|  | 容器运行时版本 |  |



**ClusterManagerCreateClusterParamsBodyClusterAdvanceSettingsClusterConnectSetting**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| domain | string|  | 用户设置访问端点域名 |  |
| internet | ClusterManagerCreateClusterParamsBodyClusterAdvanceSettingsClusterConnectSettingInternet|  |  |  |
| isExtranet | boolean (formatted boolean)|  | isExtranet集群内外网访问(true外网访问 false内网访问, 默认false) |  |
| securityGroup | string|  | 使用的安全组，只有外网访问需要传递（开启外网访问时必传） |  |
| subnetId | string|  | 集群端口所在的子网ID(开启内网访问所需) |  |



**ClusterManagerCreateClusterParamsBodyClusterAdvanceSettingsClusterConnectSettingInternet**



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



**ClusterManagerCreateClusterParamsBodyClusterBasicSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| OS | string|  | 集群使用OS信息(镜像ID/公共镜像, 为空时使用默认镜像ID) |  |
| area | ClusterManagerCreateClusterParamsBodyClusterBasicSettingsArea|  |  |  |
| clusterLevel | string|  | 集群等级，针对托管集群生效 |  |
| clusterTags | map of string|  |  |  |
| isAutoUpgradeClusterLevel | boolean (formatted boolean)|  | 是否开启自动变配集群等级，针对托管集群生效 |  |
| module | ClusterManagerCreateClusterParamsBodyClusterBasicSettingsModule|  |  |  |
| subnetID | string|  | 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡 |  |
| upgradePolicy | ClusterManagerCreateClusterParamsBodyClusterBasicSettingsUpgradePolicy|  |  |  |
| version | string|  | kubernetes使用的版本 |  |
| versionName | string|  | kubernetes使用的版本名称 |  |



**ClusterManagerCreateClusterParamsBodyClusterBasicSettingsArea**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerCreateClusterParamsBodyClusterBasicSettingsModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| masterModuleID | string|  | CC模块ID,master节点转移模块 |  |
| masterModuleName | string|  | CC模块名称,master节点转移模块 |  |
| workerModuleID | string|  | CC模块ID,worker节点转移模块 |  |
| workerModuleName | string|  | CC模块名称,worker节点转移模块 |  |



**ClusterManagerCreateClusterParamsBodyClusterBasicSettingsUpgradePolicy**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| supportType | string|  | aws EKS升级策略类型(EXTENDED|STANDARD) |  |



**ClusterManagerCreateClusterParamsBodyExtraAddonsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| applyNum | int64 (formatted integer)|  | 申请实例数量 |  |
| charge | ClusterManagerCreateClusterParamsBodyInstancesItems0Charge|  |  |  |
| cloudDataDisks | []ClusterManagerCreateClusterParamsBodyInstancesItems0CloudDataDisksItems0|  | 节点数据盘信息(包含磁盘挂载信息以及格式化信息，用户若不需要可自行处理格式化挂载)，无默认设置则不购买 |  |
| dataDisks | []ClusterManagerCreateClusterParamsBodyInstancesItems0DataDisksItems0|  | 数据盘信息，无默认设置则不购买(已废弃掉，统一通过 CloudDataDisk字段兼容) |  |
| dockerGraphPath | string|  | 节点运行时容器目录，BCS默认设置/data/bcs/service/docker |  |
| imageInfo | ClusterManagerCreateClusterParamsBodyInstancesItems0ImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，tencentCloud包含POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。aws包含ON_DEMAND按需付费,SPOT竞价付费。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerCreateClusterParamsBodyInstancesItems0InternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 qcloud默认开启 |  |
| keyPair | ClusterManagerCreateClusterParamsBodyInstancesItems0KeyPair|  |  |  |
| nodeRole | string|  | 节点角色, 取值:MASTER_ETCD, WORKER。MASTER_ETCD只有在创建 INDEPENDENT_CLUSTER 独立集群时需要指定 |  |
| region | string|  | instance地域信息 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，使用系统默认的安全组信息 |  |
| subnetID | string|  | 子网ID, 和可用区域一一对应 |  |
| systemDisk | ClusterManagerCreateClusterParamsBodyInstancesItems0SystemDisk|  |  |  |
| vpcID | string|  | 私有网络ID |  |
| zone | string|  | 可用区域 |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0Charge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0CloudDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0DataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0ImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0InternetAccess**



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



**ClusterManagerCreateClusterParamsBodyInstancesItems0KeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterParamsBodyInstancesItems0SystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterParamsBodyNetworkSettings**



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
| subnetSource | ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSource|  |  |  |



**ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSource**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| existed | ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSourceExisted|  |  |  |
| new | []ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSourceNewItems0|  | VPC-CNI网络模式下申请subnet |  |



**ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSourceExisted**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ids | []string|  | 直接使用已存在子网IDs |  |



**ClusterManagerCreateClusterParamsBodyNetworkSettingsSubnetSourceNewItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ipCnt | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的IP个数 |  |
| mask | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的网络掩码大小 |  |
| zone | string|  | 申请vpc-cni网络subnet时region下的可用区 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0**


> 节点组，对蓝鲸、腾讯云、AWS弹性伸缩能力封装



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| area | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0Area|  |  |  |
| autoScaling | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0AutoScaling|  |  |  |
| cloudNodeGroupID | string|  | NodeGroup 在各个云上的 ID |  |
| clusterID | string|  | 节点组所关联的集群，该ID为BCS集群ID |  |
| consumerID | string|  | consumerID对接resourceManager模块申请资源使用 |  |
| createTime | string|  |  |  |
| creator | string|  |  |  |
| enableAutoscale | boolean (formatted boolean)|  | 是否开启弹性伸缩，默认false |  |
| extraInfo | map of string|  |  |  |
| labels | map of string|  |  |  |
| launchTemplate | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplate|  |  |  |
| name | string|  | 节点组名称 |  |
| nodeGroupID | string|  | 节点组ID |  |
| nodeGroupType | string|  | 节点池类型(normal/external)。普通弹性伸缩节点池和第三方弹性伸缩节点池, 即CVM节点或IDC节点 |  |
| nodeOS | string|  | 新实例选用的系统OS，参照云信息中配置的默认值 |  |
| nodeTemplate | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplate|  |  |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。需与集群所属项目一致。 |  |
| provider | string|  | NodeGroup具体实现，用于覆盖所属cluster中定义的provider信息调整实现的细节。便于在一个云环境中扩展自定义的模板，使用不同 |  |
| region | string|  | 节点组所关联区域信息 |  |
| status | string|  | 节点组状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED |  |
| tags | map of string|  |  |  |
| taints | map of string|  |  |  |
| updateTime | string|  |  |  |
| updater | string|  |  |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0Area**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bkCloudID | int64 (formatted integer)|  | 云区域 ID |  |
| bkCloudName | string|  | 云区域名称 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0AutoScaling**



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
| timeRanges | []ClusterManagerCreateClusterParamsBodyNodeGroupsItems0AutoScalingTimeRangesItems0|  | 定时自动扩容集群节点配置信息 |  |
| vpcID | string|  | 弹性伸缩组所属VPC ID |  |
| zones | []string|  | 弹性伸缩组扩容实例首选的可用区ID，多个可用区按照先后顺序确认优先级，靠前优先级高 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0AutoScalingTimeRangesItems0**


> autoScale定时扩缩配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| desiredNum | int64 (formatted integer)|  | 期望节点数 |  |
| name | string|  | 定时扩缩job名称 |  |
| schedule | string|  | 定时cron时间 |  |
| zone | string|  | 定时cron时区,如果不执行时区则默认是服务所在时区 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplate**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| CPU | int64 (formatted integer)|  | 机型CPU信息，与机型instanceType互斥，比instanceType优先级高 |  |
| GPU | int64 (formatted integer)|  | 机型GPU信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| Mem | int64 (formatted integer)|  | 机型Mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| charge | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateCharge|  |  |  |
| dataDisks | []ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| imageInfo | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateImageInfo|  |  |  |
| initLoginPassword | string|  | 新实例初始设置password |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| instanceChargeType | string|  | 机型计费模型，POSTPAID_BY_HOUR按小时后付费，SPOTPAID竞价付费，PREPAID包年包月。如不设置，BCS暂时默认提供包年包月的类型 |  |
| instanceType | string|  | 机型信息字符串，若配置CPU和内存，该值不强制，默认自行向上就近匹配机型信息 |  |
| internetAccess | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateInternetAccess|  |  |  |
| isMonitorService | boolean (formatted boolean)|  | 新实例启动时的是否开启云监控, 选填 默认开启 |  |
| isSecurityService | boolean (formatted boolean)|  | 新实例启动时的是否开启云安全, 选填 yunti默认关闭 |  |
| keyPair | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateKeyPair|  |  |  |
| launchConfigurationID | string|  | 节点配置模板ID，自建ID或者云上ID |  |
| launchConfigureName | string|  | 配置模板名称 |  |
| networkTag | []string|  | 网络标记，标记代表应用于每个节点的防火墙规则 |  |
| projectID | string|  | 项目标识（该配置中可能指蓝鲸项目等，非 BCS 项目时含义以实际业务为准）。BCS 场景下为 32 位十六进制字符串；与 projectCode 区别：projectID 为内部标识，projectCode 为人类可读的项目英文编码。 |  |
| securityGroupIDs | []string|  | 新实例安全组关联的ID列表，如果不填写，BCS默认使用各云信息配置中预留的信息 |  |
| selector | map of string|  |  |  |
| systemDisk | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateSystemDisk|  |  |  |
| userData | string|  | 经过 Base64 编码后的自定义数据，最大长度不超过16KB。 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateCharge**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| period | int64 (formatted integer)|  | 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48, 60。 |  |
| renewFlag | string|  | 自动续费标识。取值范围：NOTIFY_AND_AUTO_RENEW：通知过期且自动续费; NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费;DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费. 默认取值：NOTIFY_AND_MANUAL_RENEW。若该参数指定为NOTIFY_AND_AUTO_RENEW，在账户余额充足的情况下，实例到期后将按月自动续费。 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateInternetAccess**



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



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0LaunchTemplateSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplate**


> 节点模版, 主要用于对kubernetes集群节点的通用特性设置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipScaleInWhenFailed | boolean (formatted boolean)|  | 缩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| allowSkipScaleOutWhenFailed | boolean (formatted boolean)|  | 扩容失败时是否允许跳过流程, 默认不允许跳过 |  |
| annotations | map of string|  |  |  |
| bcsScaleInAddons | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleInAddons|  |  |  |
| bcsScaleOutAddons | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleOutAddons|  |  |  |
| createTime | string|  |  |  |
| creator | string|  |  |  |
| dataDisks | []ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateDataDisksItems0|  | 数据盘信息，无默认设置则不购买. 多盘数据盘挂载信息：新建节点时请确保购买CVM的参数传递了购买多个数据盘的信息该参数主要是针对自动生产机器且设置多数据盘场景 |  |
| desc | string|  | 节点模板描述信息 |  |
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| extraInfo | map of string|  |  |  |
| gpuArgs | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgs|  |  |  |
| image | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateImage|  |  |  |
| labels | map of string|  |  |  |
| maxPodsPerNode | int64 (formatted integer)|  | 每个节点的最大Pod数量 |  |
| module | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateModule|  |  |  |
| mountTarget | string|  | 节点数据盘挂载点, 默认不挂载数据盘。已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载。无数据盘或有多块数据盘的云主机此设置不生效。BCS默认/data |  |
| name | string|  | 节点模版名称 |  |
| nodeOS | string|  | 节点模版OS |  |
| nodeTemplateID | string|  | 节点模版ID |  |
| preStartUserScript | string|  | 扩容节点前置脚本, base64 编码的用户脚本. 在初始化节点之前执行，注意目前只对添加已有节点生效 |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| runtime | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateRuntime|  |  |  |
| scaleInExtraAddons | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleInExtraAddons|  |  |  |
| scaleInPostScript | string|  | 缩容节点后置脚本, base64编码的用户脚本. |  |
| scaleInPreScript | string|  | 缩容节点前置脚本, base64编码的用户脚本. |  |
| scaleOutExtraAddons | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean (formatted boolean)|  | 是否跳过执行系统初始化, 默认执行 |  |
| taints | []ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| updateTime | string|  |  |  |
| updater | string|  |  |  |
| userScript | string|  | 扩容节点后置脚本, base64编码的用户脚本. k8s组件运行后执行，需要保证脚本的可重入和重试逻辑。如果要求节点需要在进行初始化完成后才可加入调度可配合 unschedulable 参数使用, 在 userScript 最后初始化完成后, 添加 kubectl uncordon nodename --kubeconfig=/root/.kube/config 命令使节点加入调度 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleInAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleInAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleInAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleOutAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleOutAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateBcsScaleOutAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgs**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cuda | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCuda|  |  |  |
| cudnn | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCudnn|  |  |  |
| customDriver | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCustomDriver|  |  |  |
| driver | ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsDriver|  |  |  |
| migEnable | boolean (formatted boolean)|  | 是否启用MIG特性 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCuda**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCudnn**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| devName | string|  | cuDNN的Dev名字 |  |
| docName | string|  | cuDNN的Doc名字 |  |
| name | string|  | cuDNN的名字 |  |
| version | string|  | cuDNN的版本 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsCustomDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| address | string|  | 自定义GPU驱动地址链接 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateGpuArgsDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateImage**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateRuntime**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | string|  | 容器运行时 |  |
| runtimeVersion | string|  | 运行时版本 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleInExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleInExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleInExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleOutExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleOutExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateScaleOutExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateClusterParamsBodyNodeGroupsItems0NodeTemplateTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerCreateClusterParamsBodyNodeSettings**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| labels | map of string|  |  |  |
| masterLogin | ClusterManagerCreateClusterParamsBodyNodeSettingsMasterLogin|  |  |  |
| masterSecurityGroups | []string|  | master节点安全组设置 |  |
| mountTarget | string|  | 节点目录盘挂载点，BCS默认/data |  |
| taints | []ClusterManagerCreateClusterParamsBodyNodeSettingsTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| workerLogin | ClusterManagerCreateClusterParamsBodyNodeSettingsWorkerLogin|  |  |  |
| workerSecurityGroups | []string|  | worker节点安全组设置 |  |



**ClusterManagerCreateClusterParamsBodyNodeSettingsMasterLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerCreateClusterParamsBodyNodeSettingsMasterLoginKeyPair|  |  |  |



**ClusterManagerCreateClusterParamsBodyNodeSettingsMasterLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterParamsBodyNodeSettingsTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |



**ClusterManagerCreateClusterParamsBodyNodeSettingsWorkerLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerCreateClusterParamsBodyNodeSettingsWorkerLoginKeyPair|  |  |  |



**ClusterManagerCreateClusterParamsBodyNodeSettingsWorkerLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |



**ClusterManagerCreateClusterParamsBodySharedRanges**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bizs | []string|  |  |  |
| projectIdOrCodes | []string|  |  |  |
