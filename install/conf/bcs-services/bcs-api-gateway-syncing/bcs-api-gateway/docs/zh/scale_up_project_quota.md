### 描述

扩容项目资源quota

扩容项目quota

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| quotaId | `path` | string | 是 | [Path参数] 配额Id。必须在 URL 路径中提供。 |
| body | `body` | BCSProjectQuotaScaleUpProjectQuotaBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

BCSProjectQuotaScaleUpProjectQuotaOKBody

#### 默认响应
An unexpected error response
##### SchemaBCSProjectQuotaScaleUpProjectQuotaDefaultBody

##### 内联模型

**BCSProjectQuotaScaleUpProjectQuotaBody**


> 根据不同的资源类型增量扩容项目维度的资源quota



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| quota | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuota|  |  |  |
| quotaId | string|  | 配额Id | `example-id-123` |
| updater | date-time (formatted string)|  | 操作人 | `2024-01-01T00:00:00Z` |



**BCSProjectQuotaScaleUpProjectQuotaDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []BCSProjectQuotaScaleUpProjectQuotaDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaDefaultBodyDetailsItems0**


> Any 类型，用于携带任意序列化的协议缓冲区消息及其类型 URL，便于在不同语言中打包/解包。



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | 标识序列化协议缓冲区消息类型的 URL 或资源名，须至少包含一个 "/"，路径最后一段为类型的完全限定名（如 path/google.protobuf.Duration），需为规范形式。 |  |
| value | byte (base64 string)|  | 上述类型对应的有效序列化协议缓冲区内容。 |  |



**BCSProjectQuotaScaleUpProjectQuotaOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求 ID | `example-id-123` |
| task | interface{}|  | 异步任务task详情信息 |  |
| web_annotations | BCSProjectQuotaScaleUpProjectQuotaOKBodyWebAnnotations|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 用户针对项目的权限 |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuota**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaCPU|  |  |  |
| gpu | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaGpu|  |  |  |
| mem | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaMem|  |  |  |
| zoneResources | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResources|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaCPU**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaGpu**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaMem**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| attributes | map of string|  |  |  |
| deviceQuota | string|  |  |  |
| deviceQuotaUsed | string|  |  |  |
| deviceType | string|  |  |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResources**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cpu | int64 (formatted integer)|  | 机型cpu信息，与机型instanceType互斥，比instanceType优先级高 |  |
| dataDisks | []BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResourcesDataDisksItems0|  | 数据盘信息，无默认设置则不购买 |  |
| gpu | int64 (formatted integer)|  | 机型gpu信息，与机型instanceType互斥，比instanceType优先级高，可以会为0 |  |
| instanceType | string|  | 机型 |  |
| mem | int64 (formatted integer)|  | 机型mem信息，与机型instanceType互斥，比instanceType优先级高 |  |
| quotaNum | int64 (formatted integer)|  | 机型申请的配额 |  |
| quotaUsed | int64 (formatted integer)|  | 机型已使用的配额 |  |
| region | string|  | 地域 |  |
| systemDisk | BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResourcesSystemDisk|  |  |  |
| zoneId | string|  | 可用区Id | `example-id-123` |
| zoneName | string|  | 可用区名称 |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResourcesDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |



**BCSProjectQuotaScaleUpProjectQuotaParamsBodyQuotaZoneResourcesSystemDisk**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
