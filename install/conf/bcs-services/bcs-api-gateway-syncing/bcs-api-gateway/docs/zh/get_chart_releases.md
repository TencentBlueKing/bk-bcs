### 描述

获取 chart 关联的 releases

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| name | `path` | string | 是 | [path_param参数] Chart 名称。此参数在 URL 路径中传递，请替换为具体值。 |
| projectCode | `path` | string | 是 | [path_param参数] 项目英文编码（非 32 位字符，如 "k8s-demo"、"my-project"）。⚠️ 注意与 projectId 的区别：projectCode 是人类可读的英文缩写（非 32 位），projectId 是 32 位十六进制唯一标识（如 MD5）。此参数在 URL 路径中传递，请替换为实际项目编码。 |
| repoName | `path` | string | 是 | [path_param参数] Helm 仓库名称。此参数在 URL 路径中传递，请替换为具体值。 |
| body | `body` | HelmManagerGetChartReleaseBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

HelmManagerGetChartReleaseOKBody

#### 默认响应
An unexpected error response
##### SchemaHelmManagerGetChartReleaseDefaultBody

##### 内联模型

**HelmManagerGetChartReleaseBody**


> 获取 chart 关联的 releases 请求参数



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | chart名称 |  |
| projectCode | string|  | 项目代码 |  |
| repoName | string|  | 仓库名称 |  |
| versions | []string|  | chart版本, 不传版本则获取所有版本的 release |  |



**HelmManagerGetChartReleaseDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []HelmManagerGetChartReleaseDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**HelmManagerGetChartReleaseDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**HelmManagerGetChartReleaseOKBody**


> 获取 chart 关联的 releases 的返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| data | []HelmManagerGetChartReleaseOKBodyDataItems0|  | release信息 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | requestID |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | HelmManagerGetChartReleaseOKBodyWebAnnotations|  |  |  |



**HelmManagerGetChartReleaseOKBodyDataItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| appVersion | string|  | 所处于的app version |  |
| chart | string|  | chart名称 |  |
| chartVersion | string|  | chart的版本 |  |
| clusterID | string|  | 集群 ID |  |
| createBy | string|  | 创建者 |  |
| env | string|  | 环境 |  |
| iamNamespaceID | string|  | iamNamespaceID |  |
| message | string|  | 报错消息 |  |
| name | string|  | chart release名称 |  |
| namespace | string|  | 所在的namespace |  |
| projectCode | string|  | 项目编码 |  |
| repo | string|  | chart 仓库 |  |
| revision | int64 (formatted integer)|  | 所处于的版本 |  |
| status | string|  | 当前的状态 |  |
| updateBy | string|  | 更新者 |  |
| updateTime | string|  | 更新时间 |  |



**HelmManagerGetChartReleaseOKBodyWebAnnotations**


> 权限信息返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
