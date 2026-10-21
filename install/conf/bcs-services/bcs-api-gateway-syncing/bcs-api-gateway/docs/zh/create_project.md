### 描述

创建项目

创建项目信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | BCSProjectCreateProjectBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

BCSProjectCreateProjectOKBody

#### 默认响应
An unexpected error response
##### SchemaBCSProjectCreateProjectDefaultBody

##### 内联模型

**BCSProjectCreateProjectBody**


> 创建项目



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| BGID | string|  | 事业群ID, 保留字段, 默认为0 |  |
| BGName | string|  | 事业群名称, 保留字段, 默认为空 |  |
| annotations | map of string|  |  |  |
| businessID | string|  | 项目绑定的蓝鲸CMDB中业务ID信息 |  |
| centerID | string|  | 中心ID, 保留字段, 默认为0 |  |
| centerName | string|  | 中心名称, 保留字段, 默认为空 |  |
| createTime | string|  | 创建时间 |  |
| creator | string|  | 项目创建者 |  |
| deployType | int64 (formatted integer)|  | 业务部署类型, 保留字段, 1:物理机部署, 2:容器部署 |  |
| deptID | string|  | 部门ID, 保留字段, 默认为0 |  |
| deptName | string|  | 部门名称, 保留字段, 默认为空 |  |
| description | string|  | 项目描述, 尽量限制在100字符 |  |
| isOffline | boolean (formatted boolean)|  | 项目是否已经下线, 默认false |  |
| isSecret | boolean (formatted boolean)|  | 是否为保密项目, 默认为false |  |
| kind | string|  | 项目中集群类型, 可选k8s/mesos |  |
| labels | map of string|  |  |  |
| name | string| 是 | 项目中文名称, 长度不能超过64字符 |  |
| projectCode | string| 是 | 项目编码(英文缩写), 全局唯一, 长度不能超过64字符 |  |
| projectID | string|  | 项目ID, 全局唯一, 长度为32位字符串, 自动生成 |  |
| projectType | int64 (formatted integer)|  | 项目类型, 保留字段, 默认为0, 可选 1:手游, 2:端游, 3:页游, 4:平台产品, 5:支撑产品 |  |
| useBKRes | boolean (formatted boolean)|  | 是否使用蓝鲸提供的资源池, 主要用于资源计费, 默认false |  |



**BCSProjectCreateProjectDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []BCSProjectCreateProjectDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**BCSProjectCreateProjectDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**BCSProjectCreateProjectOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | BCSProjectCreateProjectOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID |  |
| web_annotations | BCSProjectCreateProjectOKBodyWebAnnotations|  |  |  |



**BCSProjectCreateProjectOKBodyData**


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



**BCSProjectCreateProjectOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |
