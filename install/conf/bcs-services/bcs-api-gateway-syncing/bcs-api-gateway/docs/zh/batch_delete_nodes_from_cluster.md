### 描述

批量下架集群节点

批量下架集群节点(包括集群手动上下架节点、CA节点、第三方节点等)。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| clusterID | `path` | string | 是 | [path_param参数] 集群 ID（缩容目标集群，如 BCS-K8S-12345）。此参数在 URL 路径中传递，请替换为具体的集群 ID |
| deleteMode | `query` | string |  | deleteMode. 删除模式，retain(移除集群节点，但是保留主机)，terminate(只支持按量计费的机器)，默认是RETAIN |
| nodeIPs | `query` | string |  | nodeIPs. 需要缩容的IP列表,字符','隔开 |
| operator | `query` | string |  | operator. 操作者 |
| virtualNodeIDs | `query` | string |  | virtualNodeIDs. 需要缩容的虚拟节点列表,字符','隔开 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerBatchDeleteNodesFromClusterOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerBatchDeleteNodesFromClusterDefaultBody

##### 内联模型

**ClusterManagerBatchDeleteNodesFromClusterDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerBatchDeleteNodesFromClusterDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerBatchDeleteNodesFromClusterDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerBatchDeleteNodesFromClusterOKBody**


> 批量下架节点返回，主要是返回相应下架节点的任务，确认节点下架是否成功



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | []ClusterManagerBatchDeleteNodesFromClusterOKBodyDataItems0| 是 | 返回节点下架结果 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerBatchDeleteNodesFromClusterOKBodyWebAnnotations|  |  |  |



**ClusterManagerBatchDeleteNodesFromClusterOKBodyDataItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| message | string|  |  |  |
| nodeGroupID | string|  |  |  |
| nodeGroupType | string|  |  |  |
| nodeIPs | []string|  |  |  |
| success | boolean (formatted boolean)|  |  |  |
| taskID | string|  |  |  |



**ClusterManagerBatchDeleteNodesFromClusterOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |
