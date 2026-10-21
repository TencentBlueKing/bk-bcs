### 描述

查询 Release 扩展资源信息

查询指定 Helm Release 的扩展资源信息，包括 Ingress、Service、Secret 等关联资源。 适用于需要了解 Release 对外暴露的服务入口或敏感配置的场景。 需要提供项目编码、集群 ID、命名空间和 Release 名称。 返回扩展资源的详细配置，可用于网络访问排查。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| name | `path` | string | 是 | [path_param参数] Helm Release 名称。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的 Release 名称 |
| namespace | `path` | string | 是 | [path_param参数] Kubernetes 命名空间名称。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的命名空间名称 |
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

HelmManagerGetReleaseDetailExtendOKBody

#### 默认响应
Default error response
##### SchemaHelmManagerGetReleaseDetailExtendDefaultBody

##### 内联模型

**HelmManagerGetReleaseDetailExtendDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []HelmManagerGetReleaseDetailExtendDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**HelmManagerGetReleaseDetailExtendDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any

**HelmManagerGetReleaseDetailExtendOKBody**


> 通用响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | uint32 (formatted integer)|  | 返回错误码，0表示成功 |  |
| data | interface{}|  | 返回数据 |  |
| message | string|  | 返回错误信息 |  |
| requestID | string|  | 请求ID | `example-id-123` |
| result | boolean|  | 请求是否成功 | `false` |
| web_annotations | HelmManagerGetReleaseDetailExtendOKBodyWebAnnotations|  |  |  |



**HelmManagerGetReleaseDetailExtendOKBodyWebAnnotations**


> Web 注解信息，主要包含权限数据



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限详情 |  |
