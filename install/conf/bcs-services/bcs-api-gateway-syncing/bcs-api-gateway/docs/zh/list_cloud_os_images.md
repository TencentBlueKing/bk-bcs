### 描述

根据云厂商、地域、云账号和镜像提供方等筛选条件查询节点可用的操作系统镜像列表。 适用于创建集群、扩容节点或配置节点模板前选择可用节点镜像的场景。 需要提供 cloudID 路径参数，可选填 region、accountID、provider、projectID、clusterID 等过滤条件。 返回镜像列表，包含镜像 ID、别名、CPU 架构、操作系统名称、状态和关联集群信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| cloudID | `path` | string | 是 | [path_param参数] 云厂商 ID（如 tencentPublicCloud、awsCloud）。注意：此参数在 URL 路径中传递，请在请求路径中替换为具体的云厂商 ID |
| accountID | `query` | string |  | 云凭证ID |
| clusterID | `query` | string |  | 集群ID，集群正在使用的镜像会添加额外信息 |
| projectID | `query` | string |  | 容器平台项目ID，用于查询使用某个镜像的集群信息列表 |
| provider | `query` | string |  | 镜像提供方，可选值：PUBLIC_IMAGE、PRIVATE_IMAGE、SHARED_IMAGE、MARKET_IMAGE、CLUSTER_IMAGE、ALL |
| region | `query` | string |  | 云地域信息 |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListCloudOsImageOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListCloudOsImageDefaultBody

##### 内联模型

**ClusterManagerListCloudOsImageDefaultBody**


> 通用错误响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 错误码 |  |
| details | []interface{}|  | 错误详情 |  |
| message | string|  | 错误信息 |  |



**ClusterManagerListCloudOsImageOKBody**


> 查询操作系统镜像列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | []ClusterManagerListCloudOsImageOKBodyDataItems0|  | 操作系统镜像列表 |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean|  | 返回结果 | `false` |



**ClusterManagerListCloudOsImageOKBodyDataItems0**


> 操作系统镜像



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| alias | string|  | 镜像别名 |  |
| arch | string|  | 镜像 CPU 架构 |  |
| clusters | []ClusterManagerListCloudOsImageOKBodyDataItems0ClustersItems0|  | 镜像关联的集群信息 |  |
| imageID | string|  | 镜像 ID |  |
| osCustomizeType | string|  | 定制系统类型 |  |
| osName | string|  | 操作系统名称 |  |
| provider | string|  | 镜像提供方，可选值：PUBLIC_IMAGE、PRIVATE_IMAGE、SHARED_IMAGE、MARKET_IMAGE |  |
| seriesName | string|  | 操作系统系列名称 |  |
| status | string|  | 镜像状态，如 NORMAL |  |



**ClusterManagerListCloudOsImageOKBodyDataItems0ClustersItems0**


> 镜像关联的集群信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| clusterID | string|  | 集群ID | `BCS-K8S-00000` |
| clusterName | string|  | 集群名称 |  |
