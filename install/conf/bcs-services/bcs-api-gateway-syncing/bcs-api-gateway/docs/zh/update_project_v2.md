### 描述

更新项目

更新项目信息（v2）。适用于需要执行写操作或修改项目资源的场景。需提供项目 ID 与请求体。返回结果可用于判断操作是否成功。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectID | `path` | string | 是 | [path_param参数] 项目唯一 ID（32 位十六进制字符串，如 MD5 哈希）。此参数在 URL 路径中传递，请替换为实际项目 ID。 |
| body | `body` | BCSProjectUpdateProjectV2Body | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

BCSProjectUpdateProjectV2OKBody

#### 默认响应
An unexpected error response
##### SchemaBCSProjectUpdateProjectV2DefaultBody

##### 内联模型

**BCSProjectUpdateProjectV2Body**


> 更新项目请求（v2）



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| businessID | string|  | 项目绑定的蓝鲸CMDB中业务ID信息 |  |
| description | string|  | 项目简要描述 |  |
| isOffline | boolean (formatted boolean)|  | 项目是否已经下线, 默认false |  |
| kind | string|  | 项目中集群类型, 可选k8s/mesos, 未来该字段可能废弃 |  |
| labels | map of string|  |  |  |
| managers | string|  | 项目管理员 |  |
| name | string|  | 项目中文名称, 长度不能超过64字符 |  |
| projectCode | string|  | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| projectID | string|  | 项目ID |  |
| quotaAttr | BCSProjectUpdateProjectV2ParamsBodyQuotaAttr|  |  |  |
| quotaSharedEnabled | boolean (formatted boolean)|  | 是否启用额度共享 |  |
| quotaSharedProjectList | []BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0|  | 额度共享配置 |  |
| useBKRes | boolean (formatted boolean)|  | 是否使用蓝鲸提供的资源池, 主要用于资源计费, 默认false |  |



**BCSProjectUpdateProjectV2DefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []BCSProjectUpdateProjectV2DefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**BCSProjectUpdateProjectV2DefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**BCSProjectUpdateProjectV2OKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | BCSProjectUpdateProjectV2OKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID |  |
| web_annotations | BCSProjectUpdateProjectV2OKBodyWebAnnotations|  |  |  |



**BCSProjectUpdateProjectV2OKBodyData**


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



**BCSProjectUpdateProjectV2OKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |



**BCSProjectUpdateProjectV2ParamsBodyQuotaAttr**


> 项目属性



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| SourceBkBizNames | string|  | 资源来源业务名称 |  |
| computeType | string|  | 计算类型 |  |
| endTime | string|  | 截止供应时间 UTC格式 |  |
| purchaseDurationSettings | string|  | 购买时长配置 |  |
| purchaseDurationType | string|  | 购买时长类型 |  |
| sourceBkBizIDs | string|  | 资源来源业务ID列表 |  |
| startTime | string|  | 开始供应时间 UTC格式 |  |



**BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0**


> 额度共享项目



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| projectCode | string|  | 项目Code |  |
| projectID | string|  | 项目ID |  |
| projectName | string|  | 项目名称 |  |
| shareStrategy | string|  | 共享策略 |  |
| usageLimit | BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0UsageLimit|  |  |  |
| usedAmount | BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0UsedAmount|  |  |  |



**BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0UsageLimit**


> 项目配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |



**BCSProjectUpdateProjectV2ParamsBodyQuotaSharedProjectListItems0UsedAmount**


> 项目配额信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |
