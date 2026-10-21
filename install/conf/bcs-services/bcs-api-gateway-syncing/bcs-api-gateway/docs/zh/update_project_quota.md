### 描述

更新项目资源quota

更新项目quota信息

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| quotaId | `path` | string | 是 | [Path参数] 资源quotaId。必须在 URL 路径中提供。 |
| body | `body` | BCSProjectQuotaUpdateProjectQuotaBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

BCSProjectQuotaUpdateProjectQuotaOKBody

#### 默认响应
An unexpected error response
##### SchemaBCSProjectQuotaUpdateProjectQuotaDefaultBody

##### 内联模型

**BCSProjectQuotaUpdateProjectQuotaBody**


> 更新项目quota请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| labels | map of string|  |  |  |
| name | string|  | 项目中文名称, 长度不能超过64字符 | `example-name` |
| quota | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuota|  |  |  |
| quotaAttr | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaAttr|  |  |  |
| quotaId | string|  | 资源quotaId | `example-id-123` |
| quotaSharedEnabled | boolean (formatted boolean)|  | 是否启用额度共享 | `false` |
| quotaSharedProjectList | []BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0|  | 额度共享配置 |  |
| updater | date-time (formatted string)|  | 更新者 | `2024-01-01T00:00:00Z` |



**BCSProjectQuotaUpdateProjectQuotaDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []BCSProjectQuotaUpdateProjectQuotaDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaDefaultBodyDetailsItems0**


> Any 类型，用于携带任意序列化的协议缓冲区消息及其类型 URL，便于在不同语言中打包/解包。



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | 标识序列化协议缓冲区消息类型的 URL 或资源名，须至少包含一个 "/"，路径最后一段为类型的完全限定名（如 path/google.protobuf.Duration），需为规范形式。 |  |
| value | byte (base64 string)|  | 上述类型对应的有效序列化协议缓冲区内容。 |  |



**BCSProjectQuotaUpdateProjectQuotaOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | BCSProjectQuotaUpdateProjectQuotaOKBodyData|  |  |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID | `example-id-123` |
| task | interface{}|  | 异步任务task详情信息 |  |
| web_annotations | BCSProjectQuotaUpdateProjectQuotaOKBodyWebAnnotations|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyData**


> 项目配额信息(支持不同类型)



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| annotations | map of string|  |  |  |
| businessID | string|  | 项目绑定的蓝鲸CMDB中业务ID信息 | `example-id-123` |
| businessName | string|  | 项目绑定的蓝鲸CMDB中业务名称 |  |
| clusterId | string|  | 集群ID | `BCS-K8S-00000` |
| clusterName | string|  | 集群名称 |  |
| createTime | date-time (formatted string)|  | 创建时间 | `2024-01-01T00:00:00Z` |
| creator | date-time (formatted string)|  | 项目创建者 | `2024-01-01T00:00:00Z` |
| description | string|  | 项目配额描述 |  |
| isDeleted | boolean (formatted boolean)|  | 项目下该资源配额是否已下线(默认软删),同类型同种类资源仅不允许重复申请,可更改配额 | `false` |
| labels | map of string|  |  |  |
| message | string|  | 不同状态中的具体原因(审批不通过、资源不足等原因) |  |
| nameSpace | string|  | 集群名称空间维度的额度管理 |  |
| nodeGroups | []BCSProjectQuotaUpdateProjectQuotaOKBodyDataNodeGroupsItems0|  | 相关nodeGroups |  |
| projectCode | string|  | 项目编码(英文缩写), 全局唯一, 长度不能超过32字符 |  |
| projectID | string|  | 项目ID, 全局唯一, 长度为32位字符串, 自动生成 | `b37778ec757544868a01e1f01f07037f` |
| provider | string|  | 云底层资源提供方 | `example-id-123` |
| quota | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuota|  |  |  |
| quotaAttr | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaAttr|  |  |  |
| quotaId | string|  | 资源配额ID | `example-id-123` |
| quotaName | string|  | 配额中文名称, 长度不能超过64字符 |  |
| quotaSharedEnabled | boolean (formatted boolean)|  | 是否启用额度共享 | `false` |
| quotaSharedProjectList | []BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0|  | 额度共享配置 |  |
| quotaType | date-time (formatted string)|  | 资源类型(目前支持CA整机资源、共享集群维度资源) | `2024-01-01T00:00:00Z` |
| status | date-time (formatted string)|  | 资源quota的状态(申请中、审批中、已申请) | `2024-01-01T00:00:00Z` |
| updateTime | date-time (formatted string)|  | 更新时间 | `2024-01-01T00:00:00Z` |
| updater | date-time (formatted string)|  | 项目更新者 | `2024-01-01T00:00:00Z` |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataNodeGroupsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterId | string|  | 集群ID | `BCS-K8S-00000` |
| nodeGroupId | string|  | 节点池ID | `example-id-123` |
| quotaNum | int64 (formatted integer)|  | 机型申请的配额 |  |
| quotaUsed | int64 (formatted integer)|  | 机型已使用的配额 |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuota**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaCPU|  |  |  |
| gpu | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaGpu|  |  |  |
| mem | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaMem|  |  |  |
| zoneResources | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResources|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaAttr**


> 项目属性



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| SourceBkBizNames | string|  | 资源来源业务名称 |  |
| computeType | string|  | 计算类型，计算类型 [General 通算, Intelligent智算] |  |
| endTime | date-time (formatted string)|  | 截止供应时间 UTC格式：2024-01-31 23:59:59 | `2024-01-01T00:00:00Z` |
| purchaseDurationSettings | date-time (formatted string)|  | 购买时长配置, 可能存在不限制时长 | `2024-01-01T00:00:00Z` |
| purchaseDurationType | date-time (formatted string)|  | 购买时长类型，【once一次性、periodic周期性】 目前对于整机额度来说仅支持一次性购买 | `2024-01-01T00:00:00Z` |
| sourceBkBizIDs | string|  | 资源来源业务ID列表，资源来源 [all]不限业务; [100148]可使用100148业务 [1068, 100148]资源可来源于多个业务 | `example-id-123` |
| startTime | date-time (formatted string)|  | 开始供应时间 UTC格式：2024-01-25 23:59:59 | `2024-01-01T00:00:00Z` |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaCPU**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaGpu**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaMem**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0**


> 被共享的项目配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| projectCode | string|  | 项目Code |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| projectName | string|  | 项目名称 |  |
| shareEndTime | date-time (formatted string)|  | 共享结束时间 | `2024-01-01T00:00:00Z` |
| shareStartTime | date-time (formatted string)|  | 共享开始时间 | `2024-01-01T00:00:00Z` |
| shareStrategy | date-time (formatted string)|  | 共享策略 [elastic弹性共享模式, rigid刚性共享模式] | `2024-01-01T00:00:00Z` |
| status | date-time (formatted string)|  | 共享状态，[active生效中, expired已过期, suspended已暂停] | `2024-01-01T00:00:00Z` |
| usageLimit | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0UsageLimit|  |  |  |
| usedAmount | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0UsedAmount|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0UsageLimit**


> 项目配额信息(支持不同类型)



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaSharedProjectListItems0UsedAmount**


> 项目配额信息(支持不同类型)



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResources**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | int64 (formatted integer)|  | 机型cpu信息，与机型instanceType互斥，比instanceType优先级高 |  |
| dataDisks | []BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResourcesDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| gpu | int64 (formatted integer)|  | 机型gpu信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| instanceType | string|  | 机型 |  |
| mem | int64 (formatted integer)|  | 机型mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| quotaNum | int64 (formatted integer)|  | 机型申请的配额 |  |
| quotaUsed | int64 (formatted integer)|  | 机型已使用的配额 |  |
| region | string|  | 地域 |  |
| systemDisk | BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResourcesSystemDisk|  |  |  |
| zoneId | string|  | 可用区Id | `example-id-123` |
| zoneName | string|  | 可用区名称 |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResourcesDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyDataQuotaZoneResourcesSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**BCSProjectQuotaUpdateProjectQuotaOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuota**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaCPU|  |  |  |
| gpu | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaGpu|  |  |  |
| mem | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaMem|  |  |  |
| zoneResources | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResources|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaAttr**


> 项目属性



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| SourceBkBizNames | string|  | 资源来源业务名称 |  |
| computeType | string|  | 计算类型，计算类型 [General 通算, Intelligent智算] |  |
| endTime | date-time (formatted string)|  | 截止供应时间 UTC格式：2024-01-31 23:59:59 | `2024-01-01T00:00:00Z` |
| purchaseDurationSettings | date-time (formatted string)|  | 购买时长配置, 可能存在不限制时长 | `2024-01-01T00:00:00Z` |
| purchaseDurationType | date-time (formatted string)|  | 购买时长类型，【once一次性、periodic周期性】 目前对于整机额度来说仅支持一次性购买 | `2024-01-01T00:00:00Z` |
| sourceBkBizIDs | string|  | 资源来源业务ID列表，资源来源 [all]不限业务; [100148]可使用100148业务 [1068, 100148]资源可来源于多个业务 | `example-id-123` |
| startTime | date-time (formatted string)|  | 开始供应时间 UTC格式：2024-01-25 23:59:59 | `2024-01-01T00:00:00Z` |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaCPU**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaGpu**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaMem**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0**


> 被共享的项目配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| projectCode | string|  | 项目Code |  |
| projectID | string|  | 项目ID | `b37778ec757544868a01e1f01f07037f` |
| projectName | string|  | 项目名称 |  |
| shareEndTime | date-time (formatted string)|  | 共享结束时间 | `2024-01-01T00:00:00Z` |
| shareStartTime | date-time (formatted string)|  | 共享开始时间 | `2024-01-01T00:00:00Z` |
| shareStrategy | date-time (formatted string)|  | 共享策略 [elastic弹性共享模式, rigid刚性共享模式] | `2024-01-01T00:00:00Z` |
| status | date-time (formatted string)|  | 共享状态，[active生效中, expired已过期, suspended已暂停] | `2024-01-01T00:00:00Z` |
| usageLimit | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0UsageLimit|  |  |  |
| usedAmount | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0UsedAmount|  |  |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0UsageLimit**


> 项目配额信息(支持不同类型)



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaSharedProjectListItems0UsedAmount**


> 项目配额信息(支持不同类型)



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quotaNum | int64 (formatted string)|  | 配额数量（整机数量） |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResources**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | int64 (formatted integer)|  | 机型cpu信息，与机型instanceType互斥，比instanceType优先级高 |  |
| dataDisks | []BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResourcesDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| gpu | int64 (formatted integer)|  | 机型gpu信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| instanceType | string|  | 机型 |  |
| mem | int64 (formatted integer)|  | 机型mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| quotaNum | int64 (formatted integer)|  | 机型申请的配额 |  |
| quotaUsed | int64 (formatted integer)|  | 机型已使用的配额 |  |
| region | string|  | 地域 |  |
| systemDisk | BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResourcesSystemDisk|  |  |  |
| zoneId | string|  | 可用区Id | `example-id-123` |
| zoneName | string|  | 可用区名称 |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResourcesDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**BCSProjectQuotaUpdateProjectQuotaParamsBodyQuotaZoneResourcesSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
