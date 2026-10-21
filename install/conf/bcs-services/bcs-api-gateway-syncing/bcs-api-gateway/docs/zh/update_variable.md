### 描述

更新变量

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectCode | `path` | string | 是 | [Path参数] 项目编码(英文缩写), 全局唯一, 长度不能超过64字符。必须在 URL 路径中提供。 |
| variableID | `path` | string | 是 | [Path参数] 项目编码(英文缩写), 全局唯一, 长度不能超过64字符。必须在 URL 路径中提供。 |
| body | `body` | VariableUpdateVariableBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

VariableUpdateVariableOKBody

#### 默认响应
An unexpected error response
##### SchemaVariableUpdateVariableDefaultBody

##### 内联模型

**VariableUpdateVariableBody**


> 更新变量请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| default | string|  | 变量默认值 |  |
| desc | string|  | 变量说明与描述, 限制在100字符以内 |  |
| key | string|  | 变量 key, 项目内唯一, 长度不能超过64字符 |  |
| name | string|  | 变量名称, 长度不能超过32字符 | `example-name` |
| projectCode | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| scope | string|  | 变量作用域, 取值范围: global, cluster, namespace |  |
| variableID | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 | `example-id-123` |



**VariableUpdateVariableDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []VariableUpdateVariableDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**VariableUpdateVariableDefaultBodyDetailsItems0**


> Any 类型，用于携带任意序列化的协议缓冲区消息及其类型 URL，便于在不同语言中打包/解包。



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | 标识序列化协议缓冲区消息类型的 URL 或资源名，须至少包含一个 "/"，路径最后一段为类型的完全限定名（如 path/google.protobuf.Duration），需为规范形式。 |  |
| value | byte (base64 string)|  | 上述类型对应的有效序列化协议缓冲区内容。 |  |



**VariableUpdateVariableOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | VariableUpdateVariableOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID | `example-id-123` |



**VariableUpdateVariableOKBodyData**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| category | date-time (formatted string)|  | 变量类型, 取值范围: sys, custom | `2024-01-01T00:00:00Z` |
| default | string|  | 变量默认值 |  |
| desc | string|  | 变量说明与描述, 限制在100字符以内 |  |
| id | string|  | 变量 id | `example-id-123` |
| key | string|  | 变量 key, 项目内唯一, 长度不能超过64字符 |  |
| name | string|  | 变量名称, 长度不能超过32字符 | `example-name` |
| projectCode | string|  | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| scope | string|  | 变量作用域, 取值范围: global, cluster, namespace |  |
