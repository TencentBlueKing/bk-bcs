### 描述

更新ClusterAutoScalingOption, 该接口为全量更新

更新ClusterAutoscalingOption，一个ClusterAutoScalingOption对应一个集群，索引信息为clusterID。适用于 Agent 需要执行写操作或修改资源的场景。需要提供必要的路径参数和请求体。返回结果可用于判断操作是否成功。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| body | `body` | ClusterManagerUpdateAutoScalingOptionBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerUpdateAutoScalingOptionOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerUpdateAutoScalingOptionDefaultBody

##### 内联模型

**ClusterManagerUpdateAutoScalingOptionBody**


> 更新AutoScalingOption请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bufferResourceCpuRatio | int64 (formatted integer)|  | bufferResourceCpuRatio,触发扩容资源阈值百分比，集群cpu资源使用率超过该值，则触发扩容，默认 100 |  |
| bufferResourceMemRatio | int64 (formatted integer)|  | bufferResourceMemRatio,触发扩容资源阈值百分比，集群mem资源使用率超过该值，则触发扩容，默认 100 |  |
| bufferResourceRatio | int64 (formatted integer)|  | buffer-resource-ratio,触发扩容资源阈值百分比，集群资源使用率超过该值，则触发扩容，默认 100 |  |
| clusterID | string| 是 | BCS集群ID，关联的集群ID |  |
| expander | string|  | expander,扩容算法，（默认）random, least-waste, most-pods, priority |  |
| expendablePodsPriorityCutoff | int32 (formatted integer)|  | 优先级低于此值的 pod，pending 不会触发扩容，缩容时会直接 kill, 默认值-10 |  |
| ignoreDaemonSetsUtilization | boolean (formatted boolean)|  | 忽略DaemonSet的资源计算 |  |
| isScaleDownEnable | boolean (formatted boolean)|  | scale-down-enabled,是否允许缩容节点 |  |
| maxEmptyBulkDelete | int64 (formatted integer)|  | max-empty-bulk-delete,单次缩容最大节点数 |  |
| maxGracefulTerminationSec | int64 (formatted integer)|  | max-graceful-termination-sec,缩容时，等待 Pod 退出最长时间 |  |
| maxNodeProvisionTime | int64 (formatted integer)|  | max-node-provision-time,等待节点提供最长时间，单位秒，默认 900 |  |
| maxTotalUnreadyPercentage | int64 (formatted integer)|  | max-total-unready-percentage,unready节点超过集群总节点的百分比，超过此阈值则停止自动扩容 |  |
| module | ClusterManagerUpdateAutoScalingOptionParamsBodyModule|  |  |  |
| newPodScaleUpDelay | int64 (formatted integer)|  | newPodScaleUpDelay，新pod建立多久之后才触发CA, 默认 0 |  |
| okTotalUnreadyCount | int64 (formatted integer)|  | ok-total-unready-count,允许 unready 节点的数量 |  |
| onlyUpdateInfo | boolean (formatted boolean)|  | 仅更新CA组件配置参数信息,不触发流程 |  |
| projectID | string| 是 | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string|  | 特性实现模式 |  |
| scaleDownDelay | int64 (formatted integer)|  | 集群扩容后多长时间开始判定缩容，默认10分钟 |  |
| scaleDownDelayAfterAdd | int64 (formatted integer)|  | scale-down-delay-after-add,扩容后判断缩容时间间隔，单位秒 |  |
| scaleDownDelayAfterDelete | int64 (formatted integer)|  | scale-down-delay-after-delete,连续两次缩容时间间隔，单位秒 |  |
| scaleDownDelayAfterFailure | int64 (formatted integer)|  | scale-down-delay-after-failure,缩容失败后重试时间间隔，单位秒 |  |
| scaleDownGpuUtilizationThreshold | int64 (formatted integer)|  | scale-down-gpu-utilization-threshold,触发缩容资源阈值(GPU)，节点资源使用量低于多少百分比判定为空闲，默认为50 |  |
| scaleDownUnneededTime | int64 (formatted integer)|  | scale-down-unneeded-time,执行缩容等待时间，节点连续空闲多少分钟后被缩容，默认10分钟 |  |
| scaleDownUnreadyTime | int64 (formatted integer)|  | scale-down-unready-time,unready节点缩容等待时间，单位秒 |  |
| scaleDownUtilizationThreahold | int64 (formatted integer)|  | scale-down-utilization-threshold,触发缩容资源阈值，节点资源使用量低于多少百分比判定为空闲，默认为50 |  |
| scaleUpFromZero | boolean (formatted boolean)|  | scale-up-from-zero,允许自动扩容（没有 ready 节点时） |  |
| scanInterval | int64 (formatted integer)|  | scan-interval,自动扩缩容检测时间间隔，单位为秒，默认为 10 |  |
| skipNodesWithLocalStorage | boolean (formatted boolean)|  | 忽略有本地存储的节点，默认为false |  |
| skipNodesWithSystemPods | boolean (formatted boolean)|  | 忽略kube-system NS下非DaemonSet管理的Pod的节点，默认false |  |
| unregisteredNodeRemovalTime | int64 (formatted integer)|  | cluster-autoscaler等待多长时间将删除未在kubernetes中注册的节点，默认30分钟 |  |
| updater | string| 是 | 更新人 |  |
| webhook | ClusterManagerUpdateAutoScalingOptionParamsBodyWebhook|  |  |  |



**ClusterManagerUpdateAutoScalingOptionDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerUpdateAutoScalingOptionDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerUpdateAutoScalingOptionDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerUpdateAutoScalingOptionOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | ClusterManagerUpdateAutoScalingOptionOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerUpdateAutoScalingOptionOKBodyWebAnnotations|  |  |  |



**ClusterManagerUpdateAutoScalingOptionOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| bufferResourceCpuRatio | int64 (formatted integer)|  | buffer-resource-ratio,触发扩容资源阈值百分比，集群资源使用率超过该值，则触发扩容，默认 100 |  |
| bufferResourceMemRatio | int64 (formatted integer)|  | buffer-resource-ratio,触发扩容资源阈值百分比，集群资源使用率超过该值，则触发扩容，默认 100 |  |
| bufferResourceRatio | int64 (formatted integer)|  | buffer-resource-ratio,触发扩容资源阈值百分比，集群资源使用率超过该值，则触发扩容，默认 100 |  |
| clusterID | string|  | BCS集群ID，关联的集群ID |  |
| createTime | string|  | 信息创建时间 |  |
| creator | string|  | 信息创建者 |  |
| devicePoolProvider | string|  | devicePoolProvider资源池提供者, 主要应用于自研云产品化, 默认 yunti |  |
| enableAutoscale | boolean (formatted boolean)|  | 是否开启自动扩缩容，开启后会在该集群部署 ClusterAutoScaling 组件，默认为 false |  |
| errorMessage | string|  | 错误信息 |  |
| expander | string|  | expander,扩容算法，（默认）random, least-waste, most-pods, priority |  |
| expendablePodsPriorityCutoff | int32 (formatted integer)|  | 优先级低于此值的 pod，pending 不会触发扩容，缩容时会直接 kill, CA默认值-10 |  |
| ignoreDaemonSetsUtilization | boolean (formatted boolean)|  | 忽略DaemonSet的资源计算 |  |
| isScaleDownEnable | boolean (formatted boolean)|  | scale-down-enabled,是否允许缩容节点 |  |
| maxEmptyBulkDelete | int64 (formatted integer)|  | max-empty-bulk-delete,单次缩容最大节点数 |  |
| maxGracefulTerminationSec | int64 (formatted integer)|  | max-graceful-termination-sec,缩容时，等待 Pod 退出最长时间 |  |
| maxNodeProvisionTime | int64 (formatted integer)|  | max-node-provision-time,等待节点提供最长时间，单位秒，默认 900 |  |
| maxTotalUnreadyPercentage | int64 (formatted integer)|  | max-total-unready-percentage,unready节点超过集群总节点的百分比，超过此阈值则停止自动扩容 |  |
| module | ClusterManagerUpdateAutoScalingOptionOKBodyDataModule|  |  |  |
| newPodScaleUpDelay | int64 (formatted integer)|  | newPodScaleUpDelay，新pod建立多久之后才触发CA, 默认 0 |  |
| okTotalUnreadyCount | int64 (formatted integer)|  | ok-total-unready-count,允许 unready 节点的数量 |  |
| projectID | string|  | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| provider | string|  | ClusterAutoScaling具体实现，用于覆盖所属cluster中定义的provider信息调整实现的细节。便于在一个云环境中扩展自定义的模板，使用不同 |  |
| scaleDownDelay | int64 (formatted integer)|  | 集群扩容后多长时间开始判定缩容，默认10分钟 |  |
| scaleDownDelayAfterAdd | int64 (formatted integer)|  | scale-down-delay-after-add,扩容后判断缩容时间间隔，单位秒 |  |
| scaleDownDelayAfterDelete | int64 (formatted integer)|  | scale-down-delay-after-delete,连续两次缩容时间间隔，单位秒 |  |
| scaleDownDelayAfterFailure | int64 (formatted integer)|  | scale-down-delay-after-failure,缩容失败后重试时间间隔，单位秒 |  |
| scaleDownGpuUtilizationThreshold | int64 (formatted integer)|  | scale-down-gpu-utilization-threshold,触发缩容资源阈值(GPU)，节点资源使用量低于多少百分比判定为空闲，默认为50 |  |
| scaleDownUnneededTime | int64 (formatted integer)|  | scale-down-unneeded-time,执行缩容等待时间，节点连续空闲多少分钟后被缩容，默认10分钟 |  |
| scaleDownUnreadyTime | int64 (formatted integer)|  | scale-down-unready-time,unready节点缩容等待时间，单位秒 |  |
| scaleDownUtilizationThreahold | int64 (formatted integer)|  | scale-down-utilization-threshold,触发缩容资源阈值，节点资源使用量低于多少百分比判定为空闲，默认为50 |  |
| scaleUpFromZero | boolean (formatted boolean)|  | scale-up-from-zero,允许自动扩容（没有 ready 节点时） |  |
| scanInterval | int64 (formatted integer)|  | scan-interval,自动扩缩容检测时间间隔，单位为秒，默认为 10 |  |
| skipNodesWithLocalStorage | boolean (formatted boolean)|  | 忽略有本地存储的节点，默认为false |  |
| skipNodesWithSystemPods | boolean (formatted boolean)|  | 忽略kube-system NS下非DaemonSet管理的Pod的节点，默认false |  |
| status | string|  | 状态, NORMAL:正常，UPDATING:更新中，UPDATE-FAILED:更新失败，STOPPED:已停止 |  |
| unregisteredNodeRemovalTime | int64 (formatted integer)|  | cluster-autoscaler等待多长时间将删除未在kubernetes中注册的节点，默认30分钟，暂时没使用 |  |
| updateTime | string|  | 信息更新时间 |  |
| updater | string|  | 信息更新者 |  |
| webhook | ClusterManagerUpdateAutoScalingOptionOKBodyDataWebhook|  |  |  |



**ClusterManagerUpdateAutoScalingOptionOKBodyDataModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerUpdateAutoScalingOptionOKBodyDataWebhook**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| mode | string|  | CA webhook模式(目前产品化仅支持Web模式) |  |
| server | string|  | CA webhook模式server配置信息 |  |
| token | string|  | CA webhook模式token |  |



**ClusterManagerUpdateAutoScalingOptionOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerUpdateAutoScalingOptionParamsBodyModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerUpdateAutoScalingOptionParamsBodyWebhook**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| mode | string|  | CA webhook模式(目前产品化仅支持Web模式) |  |
| server | string|  | CA webhook模式server配置信息 |  |
| token | string|  | CA webhook模式token |  |
