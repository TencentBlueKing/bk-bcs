### 描述

查询项目关联业务信息

查询指定项目关联的 CMDB 业务信息，包括业务 ID、业务名称和运维人员列表。 适用于需要了解项目对应的 CMDB 业务归属或联系运维人员的场景。 需要提供项目英文编码（projectCode），该参数为非 32 位字符的英文编码，与 projectID 不同。 返回业务 ID（businessID）、业务名称和运维负责人列表。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
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

BusinessGetBusinessOKBody

#### 默认响应
Default error response
##### SchemaBusinessGetBusinessDefaultBody

##### 内联模型

**BusinessGetBusinessDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []BusinessGetBusinessDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**BusinessGetBusinessDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**BusinessGetBusinessOKBody**


> 查询项目业务信息响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | BusinessGetBusinessOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求ID | `example-id-123` |
| web_annotations | BusinessGetBusinessOKBodyWebAnnotations|  |  |  |



**BusinessGetBusinessOKBodyData**


> 业务信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| businessID | string|  | CMDB 业务ID | `example-id-123` |
| maintainer | []string|  | 业务运维人员列表 |  |
| name | string|  | 业务名称 | `example-name` |



**BusinessGetBusinessOKBodyWebAnnotations**


> 权限信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
