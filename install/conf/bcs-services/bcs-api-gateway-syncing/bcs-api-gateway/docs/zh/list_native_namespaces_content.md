### 描述

查询原生 K8s 命名空间内容

查询指定集群的原生 Kubernetes 命名空间完整内容（Namespace 资源对象）。 适用于需要获取原始 Namespace YAML/JSON 配置或查看 K8s 原生元数据的场景。 需要提供项目 ID 或项目编码（projectIDOrCode）和集群 ID 两个路径参数。 返回原生 Kubernetes Namespace 对象列表，包含 metadata、spec 和 status 完整字段。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| projectIDOrCode | `path` | string | 是 | [path_param参数] 项目 ID（32 位十六进制，如 b37778ec757544868a01e1f01f07037f）或项目英文编码（非 32 位，如 k8s-demo）。 两者均可传入，系统自动识别。此参数在 URL 路径中传递，请替换为实际值。 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

any

#### 默认响应
Default error response
##### SchemaNamespaceListNativeNamespacesContentDefaultBody

##### 内联模型

**NamespaceListNativeNamespacesContentDefaultBody**


> 标准错误响应结构，包含错误码、错误信息和错误详情



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  | 状态码，对应 google.rpc.Code 枚举值 |  |
| details | []NamespaceListNativeNamespacesContentDefaultBodyDetailsItems0|  | 错误详情列表 |  |
| message | string|  | 错误信息描述 |  |



**NamespaceListNativeNamespacesContentDefaultBodyDetailsItems0**


> Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| @type | string|  | The type of the serialized message. |  |



**Additional Properties**

any
