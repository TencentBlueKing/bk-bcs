### 描述

查询集群命名空间列表

查询指定项目下某个集群内所有已创建的命名空间，包含配额使用情况和标签信息。 适用于需要枚举集群内所有可用命名空间或查找特定命名空间的场景。 需要提供项目英文编码（projectCode）和集群 ID（clusterID），两者均为路径参数。 返回命名空间列表，每条包含名称、状态、资源配额和使用率等详情。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
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

NamespaceListNamespacesOKBody

#### 默认响应
Default error response
##### SchemaNamespaceListNamespacesDefaultBody

##### 内联模型

**NamespaceListNamespacesDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []NamespaceListNamespacesDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**NamespaceListNamespacesDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**NamespaceListNamespacesOKBody**


> 查询命名空间列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | []NamespaceListNamespacesOKBodyDataItems0|  | 命名空间列表 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求ID | `example-id-123` |
| web_annotations | NamespaceListNamespacesOKBodyWebAnnotations|  |  |  |



**NamespaceListNamespacesOKBodyDataItems0**


> 命名空间详细信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | []NamespaceListNamespacesOKBodyDataItems0AnnotationsItems0|  | 命名空间注解列表 |  |
| cpuUseRate | float (formatted number)|  | CPU 使用率 |  |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| isSystem | boolean|  | 是否为系统命名空间，false 表示项目命名空间 | `false` |
| itsmTicketSN | string|  | ITSM 单号 |  |
| itsmTicketStatus | date-time (formatted string)|  | ITSM 单据状态 | `2024-01-01T00:00:00Z` |
| itsmTicketType | string|  | ITSM 单据类型，可选值：create、update、delete |  |
| itsmTicketURL | string|  | ITSM 单据链接 |  |
| labels | []NamespaceListNamespacesOKBodyDataItems0LabelsItems0|  | 命名空间标签列表 |  |
| managers | []string|  | 命名空间管理员列表 |  |
| memoryUseRate | float (formatted number)|  | 内存使用率 |  |
| name | string|  | 命名空间名称 | `example-name` |
| quota | NamespaceListNamespacesOKBodyDataItems0Quota|  |  |  |
| status | date-time (formatted string)|  | 命名空间状态 | `2024-01-01T00:00:00Z` |
| uid | string|  | 命名空间唯一标识 | `example-id-123` |
| used | NamespaceListNamespacesOKBodyDataItems0Used|  |  |  |
| variables | []NamespaceListNamespacesOKBodyDataItems0VariablesItems0|  | 命名空间变量列表 |  |



**NamespaceListNamespacesOKBodyDataItems0AnnotationsItems0**


> 命名空间注解键值对



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | 注解键名 |  |
| value | string|  | 注解值 |  |



**NamespaceListNamespacesOKBodyDataItems0LabelsItems0**


> 命名空间标签键值对



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| key | string|  | 标签键名 |  |
| value | string|  | 标签值 |  |



**NamespaceListNamespacesOKBodyDataItems0Quota**


> 资源配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  | CPU 限制量 (limits.cpu) |  |
| cpuRequests | string|  | CPU 请求量 (requests.cpu) |  |
| memoryLimits | string|  | 内存限制量 (limits.memory) |  |
| memoryRequests | string|  | 内存请求量 (requests.memory) |  |



**NamespaceListNamespacesOKBodyDataItems0Used**


> 资源配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpuLimits | string|  | CPU 限制量 (limits.cpu) |  |
| cpuRequests | string|  | CPU 请求量 (requests.cpu) |  |
| memoryLimits | string|  | 内存限制量 (limits.memory) |  |
| memoryRequests | string|  | 内存请求量 (requests.memory) |  |



**NamespaceListNamespacesOKBodyDataItems0VariablesItems0**


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



**NamespaceListNamespacesOKBodyWebAnnotations**


> 权限信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
