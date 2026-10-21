### 描述

下架集群节点

指定IP地址缩容集群node, 支持云和第三方节点。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（缩容目标集群，如 BCS-K8S-12345）。此参数在 URL 路径中传递，请替换为具体的集群 ID |
| deleteMode | `query` | string |  | deleteMode. 删除模式，RETAIN(移除集群，但是保留主机)，TERMINATE(只支持按量计费的机器)，默认是RETAIN |
| isExternalNode | `query` | boolean (formatted boolean) |  | 是否是第三方节点(IDC节点). 下架第三方节点需置为true, 且必须关联nodeGroupID |
| isForce | `query` | boolean (formatted boolean) |  | isForce. 不管节点处于任何状态都强制删除，例如可能刚初始化，NotReady等 |
| nodeGroupID | `query` | string |  | nodeGroupID. 下架第三方节点需关联节点池ID |
| nodeTemplateID | `query` | string |  | nodeTemplateID. 关联的节点模板ID |
| nodes | `query` | string |  | nodes. 需要缩容的IP列表,字符','隔开 |
| onlyDeleteInfo | `query` | boolean (formatted boolean) |  | onlyDeleteInfo. 默认为false。设置为true时，仅删除cluster-manager所记录的信息，不会触发任何自动化流程. |
| operator | `query` | string |  | operator. 操作者 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerDeleteNodesFromClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerDeleteNodesFromClusterDefaultBody

##### 内联模型

**ClusterManagerDeleteNodesFromClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerDeleteNodesFromClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerDeleteNodesFromClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerDeleteNodesFromClusterOKBody**


> 下架节点返回，主要是返回下架节点的任务，确认节点下架是否成功



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | ClusterManagerDeleteNodesFromClusterOKBodyData| 是 |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerDeleteNodesFromClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerDeleteNodesFromClusterOKBodyData**


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
| steps | map ofClusterManagerDeleteNodesFromClusterOKBodyDataStepsAnon|  |  |  |
| taskID | string|  | 任务ID，用于追溯与确认任务状态 |  |
| taskName | string|  | 任务类型中文名称 |  |
| taskType | string|  | 任务类型，例如CreateCluster，AddNodesToCluster，DeleteNodesFromCluster等 |  |
| updater | string|  | 任务更新人 |  |



**ClusterManagerDeleteNodesFromClusterOKBodyDataStepsAnon**


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



**ClusterManagerDeleteNodesFromClusterOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
