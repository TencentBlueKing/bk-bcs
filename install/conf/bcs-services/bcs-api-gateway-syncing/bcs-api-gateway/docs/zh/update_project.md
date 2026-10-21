### 描述

更新项目

更新项目信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectID | `path` | string | 是 | [path_param参数] 项目唯一 ID（32 位十六进制字符串，如 MD5 哈希）。此参数在 URL 路径中传递，请替换为实际项目 ID。 |
| body | `body` | BCSProjectUpdateProjectBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

BCSProjectUpdateProjectOKBody

#### 默认响应
An unexpected error response
##### SchemaBCSProjectUpdateProjectDefaultBody

##### 内联模型

**BCSProjectUpdateProjectBody**


> 更新项目请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| businessID | string|  | 项目绑定的蓝鲸CMDB中业务ID信息 |  |
| creator | string|  | 创建人 |  |
| description | string|  | 项目简要描述 |  |
| isOffline | boolean (formatted boolean)|  | 项目是否已经下线, 默认false |  |
| kind | string|  | 项目中集群类型, 可选k8s/mesos, 未来该字段可能废弃 |  |
| labels | map of string|  |  |  |
| managers | string|  | 项目管理员 |  |
| name | string|  | 项目中文名称, 长度不能超过64字符 |  |
| projectID | string|  | 项目ID |  |
| updater | string|  | 更新者 |  |
| useBKRes | boolean (formatted boolean)|  | 是否使用蓝鲸提供的资源池, 主要用于资源计费, 默认false |  |



**BCSProjectUpdateProjectDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []BCSProjectUpdateProjectDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**BCSProjectUpdateProjectDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**BCSProjectUpdateProjectOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | BCSProjectUpdateProjectOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID |  |
| web_annotations | BCSProjectUpdateProjectOKBodyWebAnnotations|  |  |  |



**BCSProjectUpdateProjectOKBodyData**


> 项目的基本信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| businessID | string|  | 项目绑定的蓝鲸CMDB中业务ID信息 |  |
| businessName | string|  | 项目绑定的蓝鲸CMDB中业务名称 |  |
| createTime | string|  | 创建时间 |  |
| creator | string|  | 项目创建者 |  |
| description | string|  | 项目描述, 尽量限制在100字符 |  |
| isOffline | boolean (formatted boolean)|  | 项目是否已经下线, 默认false |  |
| kind | string|  | 项目中集群类型, 可选k8s/mesos |  |
| labels | map of string|  |  |  |
| managers | string|  | 项目管理人员，默认为创建者+更新者 |  |
| name | string|  | 项目中文名称, 长度不能超过64字符 |  |
| projectCode | string|  | 项目编码(英文缩写), 全局唯一, 长度不能超过32字符 |  |
| projectID | string|  | 项目ID, 全局唯一, 长度为32位字符串, 自动生成 |  |
| updateTime | string|  | 更新时间 |  |
| updater | string|  | 项目更新者 |  |
| useBKRes | boolean (formatted boolean)|  | 是否使用蓝鲸提供的资源池, 主要用于资源计费, 默认false |  |



**BCSProjectUpdateProjectOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |
