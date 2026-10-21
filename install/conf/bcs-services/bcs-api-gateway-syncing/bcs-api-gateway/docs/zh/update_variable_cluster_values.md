### 描述

更新所有集群的变量值

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectCode | `path` | string | 是 | [Path参数] 项目编码(英文缩写), 全局唯一, 长度不能超过64字符。必须在 URL 路径中提供。 |
| variableID | `path` | string | 是 | [Path参数] 变量 id。必须在 URL 路径中提供。 |
| body | `body` | VariableUpdateClustersVariablesBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

VariableUpdateClustersVariablesOKBody

#### 默认响应
An unexpected error response
##### SchemaVariableUpdateClustersVariablesDefaultBody

##### 内联模型

**VariableUpdateClustersVariablesBody**


> 更新项目下集群变量值请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| data | []VariableUpdateClustersVariablesParamsBodyDataItems0|  | 集群变量值 |  |
| projectCode | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| variableID | string| 是 | 变量 id | `example-id-123` |



**VariableUpdateClustersVariablesDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []VariableUpdateClustersVariablesDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**VariableUpdateClustersVariablesDefaultBodyDetailsItems0**


> Any 类型，用于携带任意序列化的协议缓冲区消息及其类型 URL，便于在不同语言中打包/解包。



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | 标识序列化协议缓冲区消息类型的 URL 或资源名，须至少包含一个 "/"，路径最后一段为类型的完全限定名（如 path/google.protobuf.Duration），需为规范形式。 |  |
| value | byte (base64 string)|  | 上述类型对应的有效序列化协议缓冲区内容。 |  |



**VariableUpdateClustersVariablesOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID | `example-id-123` |



**VariableUpdateClustersVariablesParamsBodyDataItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| clusterName | string|  | 集群名称 |  |
| id | string|  | 变量ID | `example-id-123` |
| key | string|  | 变量key |  |
| name | string|  | 变量名称 | `example-name` |
| namespace | string|  | 命名空间 |  |
| scope | string|  | 变量作用域 |  |
| value | string|  | 变量的值 |  |
