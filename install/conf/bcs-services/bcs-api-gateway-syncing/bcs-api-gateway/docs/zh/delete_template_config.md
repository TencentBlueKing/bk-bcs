### 描述

删除模板配置信息

根据配置ID删除特定的模板配置。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| templateConfigID | `path` | string | 是 | [path_param参数] 模板配置 ID。在 URL 路径中传递，请在请求路径中替换为具体值。 |
| businessID | `query` | string |  | businessID. 删除的模板配置所属的业务ID |
| projectID | `query` | string |  | BCS 项目唯一标识（32 位十六进制字符串，如 MD5）。删除的模板配置所属的项目；与 projectCode 区别：projectID 为内部标识，projectCode 为人类可读的项目英文编码。 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerDeleteTemplateConfigOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerDeleteTemplateConfigDefaultBody

##### 内联模型

**ClusterManagerDeleteTemplateConfigDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerDeleteTemplateConfigDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerDeleteTemplateConfigDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerDeleteTemplateConfigOKBody**


> 删除模板配置的响应结果



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回操作结果 |  |
