### 描述

回收节点组中的节点

指定IP地址从节点组中移除，也从集群中移除，并清理资源回收节点。该接口为异步接口，当该接口正常返回时，仅能确认清理任务正常创建，不能保证节点清理已经完成，如果需要确认进度，需要通过返回的Task...。适用于 Agent 需要执行写操作或修改资源的场景。需要提供必要的路径参数和请求体。返回结果可用于判断操作是否成功。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| nodeGroupID | `path` | string | 是 | [path_param参数] 节点组 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| body | `body` | ClusterManagerCleanNodesInGroupBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCleanNodesInGroupOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCleanNodesInGroupDefaultBody

##### 内联模型

**ClusterManagerCleanNodesInGroupBody**


> 从节点池中清理回收指定IP节点



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string| 是 | 集群ID |  |
| manual | boolean (formatted boolean)|  | 用户手动下架CA节点时，当失败时候支持用户跳过；自动下架时, 自动跳过失败步骤 |  |
| nodeGroupID | string| 是 | 关联的NodeGroupID |  |
| nodes | []string| 是 | 需要清理回收的IP列表 |  |
| operator | string| 是 | 操作者 |  |



**ClusterManagerCleanNodesInGroupDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCleanNodesInGroupDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCleanNodesInGroupDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCleanNodesInGroupOKBody**


> 清理节点返回，由于该请求为异步请求，该返回仅能确认清理任务是否正常创建



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | ClusterManagerCleanNodesInGroupOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerCleanNodesInGroupOKBodyWebAnnotations|  |  |  |



**ClusterManagerCleanNodesInGroupOKBodyData**


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
| steps | map ofClusterManagerCleanNodesInGroupOKBodyDataStepsAnon|  |  |  |
| taskID | string|  | 任务ID，用于追溯与确认任务状态 |  |
| taskName | string|  | 任务类型中文名称 |  |
| taskType | string|  | 任务类型，例如CreateCluster，AddNodesToCluster，DeleteNodesFromCluster等 |  |
| updater | string|  | 任务更新人 |  |



**ClusterManagerCleanNodesInGroupOKBodyDataStepsAnon**


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



**ClusterManagerCleanNodesInGroupOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
