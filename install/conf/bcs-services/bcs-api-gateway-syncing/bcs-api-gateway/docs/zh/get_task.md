### 描述

查询任务详情

查询指定操作任务的完整执行详情，包括每个步骤的状态、日志和错误信息。 适用于需要了解某个异步任务（如集群创建）当前执行到哪一步或排查任务失败原因的场景。 需要提供 taskID 路径参数。 返回任务的所有步骤列表，包含每步的名称、状态、开始/结束时间和错误消息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| taskID | `path` | string | 是 | [path_param参数] 任务ID。注意：此参数在 URL 路径中传递，请在请求路径中填充具体的任务ID值 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerGetTaskOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerGetTaskDefaultBody

##### 内联模型

**ClusterManagerGetTaskDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerGetTaskDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerGetTaskDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerGetTaskOKBody**


> 查询任务详情响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | ClusterManagerGetTaskOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerGetTaskOKBodyData**


> 异步任务信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 关联集群ID | `BCS-K8S-00000` |
| commonParams | map of string|  | 公共参数 |  |
| creator | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| currentStep | string|  | 当前执行的步骤名称 |  |
| end | string|  | 结束时间 |  |
| executionTime | uint32 (formatted integer)|  | 执行耗时（秒） |  |
| forceTerminate | boolean|  | 是否强制终止 | `false` |
| lastUpdate | date-time (formatted string)|  | 最后更新时间 | `2024-01-01T00:00:00Z` |
| message | string|  | 任务信息或错误信息 |  |
| nodeGroupID | string|  | 关联节点组ID | `example-id-123` |
| nodeIPList | []string|  | 关联节点IP列表 |  |
| projectID | string|  | 关联项目ID | `b37778ec757544868a01e1f01f07037f` |
| start | string|  | 开始时间 |  |
| status | date-time (formatted string)|  | 任务状态，可选值：INITIALIZING、RUNNING、SUCCESS、FAILURE、TIMEOUT、FORCETERM | `2024-01-01T00:00:00Z` |
| stepSequence | []string|  | 步骤执行顺序 |  |
| steps | map ofClusterManagerGetTaskOKBodyDataStepsAnon|  | 步骤详情，键为步骤名称 |  |
| taskID | string|  | 任务ID | `example-id-123` |
| taskName | string|  | 任务名称 |  |
| taskType | string|  | 任务类型，如 CreateCluster、AddNodesToCluster 等 |  |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerGetTaskOKBodyDataStepsAnon**


> 任务步骤信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkip | boolean|  | 是否允许跳过 | `false` |
| end | string|  | 结束时间 |  |
| executionTime | uint32 (formatted integer)|  | 执行耗时（秒） |  |
| lastUpdate | date-time (formatted string)|  | 最后更新时间 | `2024-01-01T00:00:00Z` |
| link | string|  | 标准运维任务链接 |  |
| maxRetry | uint32 (formatted integer)|  | 最大重试次数 |  |
| message | string|  | 步骤信息或错误信息 |  |
| name | string|  | 步骤名称 | `example-name` |
| params | map of string|  | 步骤参数 |  |
| retry | uint32 (formatted integer)|  | 已重试次数 |  |
| skipOnFailed | boolean|  | 失败时是否跳过 | `false` |
| start | string|  | 开始时间 |  |
| status | date-time (formatted string)|  | 步骤状态 | `2024-01-01T00:00:00Z` |
| system | string|  | 标准运维系统ID |  |
| taskMethod | string|  | 任务方法 |  |
| taskName | string|  | 任务名称 |  |
| translate | date-time (formatted string)|  | 国际化翻译标识 | `2024-01-01T00:00:00Z` |
