### 描述

上架集群节点

指定IP，给集群增加节点(云CVM节点和IDC第三方节点)，会触发相关自动化流程。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（扩容目标集群，如 BCS-K8S-12345）。此参数在 URL 路径中传递，请替换为具体的集群 ID |
| body | `body` | ClusterManagerAddNodesToClusterV2Body | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerAddNodesToClusterV2OKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerAddNodesToClusterV2DefaultBody

##### 内联模型

**ClusterManagerAddNodesToClusterV2Body**


> 给指定的集群进行扩容，！注意！扩容的机器如果是腾讯云cvm默认会重装系统，不能使用竞价计费类型的机器



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| advance | ClusterManagerAddNodesToClusterV2ParamsBodyAdvance|  |  |  |
| clusterID | string| 是 | 扩容的集群ID |  |
| initLoginPassword | string|  | 重装时的初始密码, 可为空 |  |
| isExternalNode | boolean (formatted boolean)|  | 上架第三方节点需置为true, 且必须关联nodeGroupID |  |
| login | ClusterManagerAddNodesToClusterV2ParamsBodyLogin|  |  |  |
| nodeGroupID | string|  | 扩容节点如果需要加入NodeGroup，需要填写关联ID |  |
| nodeTemplateID | string|  | 关联的节点模板ID |  |
| nodes | []string| 是 | 扩容的集群节点IP地址 |  |
| onlyCreateInfo | boolean (formatted boolean)|  | 仅写入节点信息，默认是false。仅写入节点信息时，会通过cluster所属cloudprovider查询节点信息，并设置为RUNNING。 |  |
| operator | string|  | 操作者 |  |



**ClusterManagerAddNodesToClusterV2DefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerAddNodesToClusterV2DefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerAddNodesToClusterV2DefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerAddNodesToClusterV2OKBody**


> 增加集群节点返回，主要是反馈相关自动化过程的任务信息，便于查阅状态，V2版本返回任务列表



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | []ClusterManagerAddNodesToClusterV2OKBodyDataItems0| 是 | 节点添加任务信息 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerAddNodesToClusterV2OKBodyWebAnnotations|  |  |  |



**ClusterManagerAddNodesToClusterV2OKBodyDataItems0**


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
| steps | map ofClusterManagerAddNodesToClusterV2OKBodyDataItems0StepsAnon|  |  |  |
| taskID | string|  | 任务ID，用于追溯与确认任务状态 |  |
| taskName | string|  | 任务类型中文名称 |  |
| taskType | string|  | 任务类型，例如CreateCluster，AddNodesToCluster，DeleteNodesFromCluster等 |  |
| updater | string|  | 任务更新人 |  |



**ClusterManagerAddNodesToClusterV2OKBodyDataItems0StepsAnon**


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



**ClusterManagerAddNodesToClusterV2OKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerAddNodesToClusterV2ParamsBodyAdvance**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| isGPUNode | boolean (formatted boolean)|  | 是否GPU节点 |  |
| nodeOs | string|  | 上架节点时可支持设置自定义节点镜像,不设置时复用集群镜像 |  |



**ClusterManagerAddNodesToClusterV2ParamsBodyLogin**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| initLoginPassword | string|  | 新实例初始设置password,如果不设置系统自动生成 |  |
| initLoginUsername | string|  | 新实例初始设置username |  |
| keyPair | ClusterManagerAddNodesToClusterV2ParamsBodyLoginKeyPair|  |  |  |



**ClusterManagerAddNodesToClusterV2ParamsBodyLoginKeyPair**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| keyID | string|  | 密钥ID，必填字段 |  |
| keyPublic | string|  | key公钥信息(base64存储) |  |
| keySecret | string|  | key私钥信息(base64存储) |  |
