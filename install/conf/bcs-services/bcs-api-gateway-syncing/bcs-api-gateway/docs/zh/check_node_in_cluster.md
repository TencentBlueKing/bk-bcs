### 描述

获取node信息是否存在bcs集群

检查节点是否存在bcs集群,防止误回收节点。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| body | `body` | ClusterManagerCheckNodeInClusterBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCheckNodeInClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCheckNodeInClusterDefaultBody

##### 内联模型

**ClusterManagerCheckNodeInClusterBody**


> 检查node是否存在bcs集群



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| innerIPs | []string| 是 | 节点内网IP列表 |  |



**ClusterManagerCheckNodeInClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCheckNodeInClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCheckNodeInClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCheckNodeInClusterOKBody**


> 返回节点状态



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | map ofClusterManagerCheckNodeInClusterOKBodyDataAnon|  |  |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |



**ClusterManagerCheckNodeInClusterOKBodyDataAnon**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 节点所在集群ID |  |
| clusterName | string|  | 节点所在集群名称 |  |
| isExist | boolean (formatted boolean)|  | 节点是否存在集群中 |  |
