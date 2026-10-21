### 描述

查询任务列表

查询平台操作任务的列表，包含集群创建、节点添加等异步任务的状态信息。 适用于需要跟踪某个操作（如集群创建）的执行进度或查看历史任务的场景。 可通过 clusterID、taskType 等过滤条件筛选特定任务。 返回任务列表，每条包含 taskID、类型、状态和步骤执行详情。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `query` | string |  | 集群ID |
| creator | `query` | string |  | 创建者 |
| nodeGroupID | `query` | string |  | 节点组ID |
| nodeIP | `query` | string |  | 节点IP |
| projectID | `query` | string |  | 项目ID |
| status | `query` | string |  | 状态过滤条件 |
| taskType | `query` | string |  | 任务类型，如 CreateCluster、AddNodesToCluster 等 |
| updater | `query` | string |  | 更新者 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListTaskOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListTaskDefaultBody

##### 内联模型

**ClusterManagerListTaskDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []ClusterManagerListTaskDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**ClusterManagerListTaskDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**ClusterManagerListTaskOKBody**


> 查询任务列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []ClusterManagerListTaskOKBodyDataItems0|  | 任务列表 |  |
| latestTask | ClusterManagerListTaskOKBodyLatestTask|  |  |  |
| message | string|  | 返回错误信息 |  |
| result | boolean|  | 请求是否成功 | `false` |



**ClusterManagerListTaskOKBodyDataItems0**


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
| steps | map ofClusterManagerListTaskOKBodyDataItems0StepsAnon|  | 步骤详情，键为步骤名称 |  |
| taskID | string|  | 任务ID | `example-id-123` |
| taskName | string|  | 任务名称 |  |
| taskType | string|  | 任务类型，如 CreateCluster、AddNodesToCluster 等 |  |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerListTaskOKBodyDataItems0StepsAnon**


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



**ClusterManagerListTaskOKBodyLatestTask**


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
| steps | map ofClusterManagerListTaskOKBodyLatestTaskStepsAnon|  | 步骤详情，键为步骤名称 |  |
| taskID | string|  | 任务ID | `example-id-123` |
| taskName | string|  | 任务名称 |  |
| taskType | string|  | 任务类型，如 CreateCluster、AddNodesToCluster 等 |  |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**ClusterManagerListTaskOKBodyLatestTaskStepsAnon**


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
