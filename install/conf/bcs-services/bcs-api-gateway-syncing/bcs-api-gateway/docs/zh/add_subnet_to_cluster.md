### 描述

集群添加子网资源

集群添加子网资源, 腾讯云VPC-CNI模式下添加子网。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（如 BCS-K8S-12345）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的集群 ID |
| body | `body` | ClusterManagerAddSubnetToClusterBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerAddSubnetToClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerAddSubnetToClusterDefaultBody

##### 内联模型

**ClusterManagerAddSubnetToClusterBody**


> 集群添加子网请求



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string| 是 | 集群ID |  |
| operator | string|  | 操作人 |  |
| subnet | ClusterManagerAddSubnetToClusterParamsBodySubnet| 是 |  |  |



**ClusterManagerAddSubnetToClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerAddSubnetToClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerAddSubnetToClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerAddSubnetToClusterOKBody**


> 集群添加子网的返回



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerAddSubnetToClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerAddSubnetToClusterOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerAddSubnetToClusterParamsBodySubnet**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| existed | ClusterManagerAddSubnetToClusterParamsBodySubnetExisted|  |  |  |
| new | []ClusterManagerAddSubnetToClusterParamsBodySubnetNewItems0|  | VPC-CNI网络模式下申请subnet |  |



**ClusterManagerAddSubnetToClusterParamsBodySubnetExisted**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ids | []string|  | 直接使用已存在子网IDs |  |



**ClusterManagerAddSubnetToClusterParamsBodySubnetNewItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ipCnt | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的IP个数 |  |
| mask | int64 (formatted integer)|  | 申请vpc-cni网络subnet时的网络掩码大小 |  |
| zone | string|  | 申请vpc-cni网络subnet时region下的可用区 |  |
