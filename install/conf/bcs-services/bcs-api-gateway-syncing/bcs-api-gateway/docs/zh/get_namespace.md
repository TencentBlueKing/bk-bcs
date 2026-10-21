### 描述

查询单个命名空间详情

查询指定集群下某个命名空间的详细信息，包括资源配额、标签、注解和变量配置。 适用于需要确认命名空间当前配额使用情况或查看命名空间元数据的场景。 需要提供项目英文编码、集群 ID 和命名空间名称三个路径参数。 返回命名空间配额（quota/used）、CPU 和内存使用率以及标签注解信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| namespace | `path` | string | 是 | [path_param参数] Kubernetes 命名空间名称。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的命名空间名称 |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 k8s-demo、my-project）。 此参数在 URL 路径中传递，请替换为实际项目编码。 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

NamespaceGetNamespaceOKBody

#### 默认响应
Default error response
##### SchemaNamespaceGetNamespaceDefaultBody

##### 内联模型

**NamespaceGetNamespaceDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []NamespaceGetNamespaceDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**NamespaceGetNamespaceDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**NamespaceGetNamespaceOKBody**


> 查询单个命名空间响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | NamespaceGetNamespaceOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求ID | `example-id-123` |
| web_annotations | NamespaceGetNamespaceOKBodyWebAnnotations|  |  |  |



**NamespaceGetNamespaceOKBodyData**


> 命名空间详细信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | []NamespaceGetNamespaceOKBodyDataAnnotationsItems0|  | 命名空间注解列表 |  |
| cpuUseRate | float (formatted number)|  | CPU 使用率 |  |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| isSystem | boolean|  | 是否为系统命名空间，false 表示项目命名空间 | `false` |
| itsmTicketSN | string|  | ITSM 单号 |  |
| itsmTicketStatus | date-time (formatted string)|  | ITSM 单据状态 | `2024-01-01T00:00:00Z` |
| itsmTicketType | string|  | ITSM 单据类型，可选值：create、update、delete |  |
| itsmTicketURL | string|  | ITSM 单据链接 |  |
| labels | []NamespaceGetNamespaceOKBodyDataLabelsItems0|  | 命名空间标签列表 |  |
| managers | []string|  | 命名空间管理员列表 |  |
| memoryUseRate | float (formatted number)|  | 内存使用率 |  |
| name | string|  | 命名空间名称 | `example-name` |
| quota | NamespaceGetNamespaceOKBodyDataQuota|  |  |  |
| status | date-time (formatted string)|  | 命名空间状态 | `2024-01-01T00:00:00Z` |
| uid | string|  | 命名空间唯一标识 | `example-id-123` |
| used | NamespaceGetNamespaceOKBodyDataUsed|  |  |  |
| variables | []NamespaceGetNamespaceOKBodyDataVariablesItems0|  | 命名空间变量列表 |  |



**NamespaceGetNamespaceOKBodyDataAnnotationsItems0**


> 命名空间注解键值对



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | 注解键名 |  |
| value | string|  | 注解值 |  |



**NamespaceGetNamespaceOKBodyDataLabelsItems0**


> 命名空间标签键值对



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | 标签键名 |  |
| value | string|  | 标签值 |  |



**NamespaceGetNamespaceOKBodyDataQuota**


> 资源配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  | CPU 限制量 (limits.cpu) |  |
| cpuRequests | string|  | CPU 请求量 (requests.cpu) |  |
| memoryLimits | string|  | 内存限制量 (limits.memory) |  |
| memoryRequests | string|  | 内存请求量 (requests.memory) |  |



**NamespaceGetNamespaceOKBodyDataUsed**


> 资源配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  | CPU 限制量 (limits.cpu) |  |
| cpuRequests | string|  | CPU 请求量 (requests.cpu) |  |
| memoryLimits | string|  | 内存限制量 (limits.memory) |  |
| memoryRequests | string|  | 内存请求量 (requests.memory) |  |



**NamespaceGetNamespaceOKBodyDataVariablesItems0**


> 变量值信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| clusterName | string|  | 集群名称 |  |
| id | string|  | 变量ID | `example-id-123` |
| key | string|  | 变量键名 |  |
| name | string|  | 变量名称 | `example-name` |
| namespace | string|  | 命名空间 |  |
| scope | string|  | 变量作用域 |  |
| value | string|  | 变量的值 |  |



**NamespaceGetNamespaceOKBodyWebAnnotations**


> 权限信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
