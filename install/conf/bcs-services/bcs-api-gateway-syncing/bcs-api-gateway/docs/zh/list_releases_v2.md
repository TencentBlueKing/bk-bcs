### 描述

查询集群 Release 列表 V2

分页查询指定集群下所有 Helm Release 的概要信息（V2 版本），支持按命名空间和 Release 名称过滤。 功能与 V1 版本相同，但返回格式有所增强，建议优先使用此版本。 需要提供项目编码和集群 ID，可选填命名空间、名称、分页参数进行过滤。 返回分页的 Release 列表，每条包含 Release 名称、状态、Chart 版本和更新时间。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 k8s-demo、my-project）。 此参数在 URL 路径中传递，请替换为实际项目编码。 |
| name | `query` | string |  | Release 名称，用于过滤 |
| namespace | `query` | string |  | 命名空间，用于过滤指定命名空间下的 Release |
| page | `query` | uint32 (formatted integer) |  | 页码 |
| size | `query` | uint32 (formatted integer) |  | 每页数量 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

HelmManagerListReleaseV2OKBody

#### 默认响应
Default error response
##### SchemaHelmManagerListReleaseV2DefaultBody

##### 内联模型

**HelmManagerListReleaseV2DefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []HelmManagerListReleaseV2DefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**HelmManagerListReleaseV2DefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**HelmManagerListReleaseV2OKBody**


> 查询 Release 列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | HelmManagerListReleaseV2OKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求ID | `example-id-123` |
| result | boolean|  | 请求是否成功 | `false` |
| webAnnotations | HelmManagerListReleaseV2OKBodyWebAnnotations|  |  |  |



**HelmManagerListReleaseV2OKBodyData**


> Release 列表分页数据



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| data | []HelmManagerListReleaseV2OKBodyDataDataItems0|  | Release 列表 |  |
| page | uint32 (formatted integer)|  | 当前页码 |  |
| size | uint32 (formatted integer)|  | 每页数量 |  |
| total | uint32 (formatted integer)|  | 总数 |  |



**HelmManagerListReleaseV2OKBodyDataDataItems0**


> Helm Release 信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| appVersion | string|  | 应用版本号 |  |
| chart | string|  | Chart 名称 |  |
| chartVersion | string|  | Chart 版本号 |  |
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| createBy | date-time (formatted string)|  | 创建者 | `2024-01-01T00:00:00Z` |
| env | string|  | 部署环境 |  |
| iamNamespaceID | string|  | IAM 命名空间ID | `default` |
| message | string|  | 错误或状态消息 |  |
| name | string|  | Release 名称 | `example-name` |
| namespace | string|  | 所在命名空间 |  |
| projectCode | string|  | 项目编码 |  |
| repo | string|  | Chart 仓库名称 |  |
| revision | uint32 (formatted integer)|  | 当前版本号 |  |
| status | date-time (formatted string)|  | Release 当前状态 | `2024-01-01T00:00:00Z` |
| updateBy | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |



**HelmManagerListReleaseV2OKBodyWebAnnotations**


> Web 注解信息，主要包含权限数据



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
