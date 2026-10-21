### 描述

根据查询条件过滤指定项目下的节点模板列表，条件为空时默认返回该项目下所有节点模板。 适用于查找可复用的节点初始化配置，或根据节点模板 ID 定位模板信息的场景。 需要提供 projectID 路径参数，可选填 nodeTemplateID 精确过滤。 返回节点模板列表，包含模板名称、操作系统、标签、污点、脚本、运行时和镜像配置等信息。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectID | `path` | string | 是 | [path_param参数] 项目唯一 ID（32 位十六进制字符串，如 MD5 哈希）。 此参数在 URL 路径中传递，请替换为实际项目 ID。 |
| nodeTemplateID | `query` | string |  | 节点模板ID |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | OK |
| default | | Default error response |

### 响应

#### 200 - OK
Status: OK

##### Schema

ClusterManagerListNodeTemplateOKBody

#### 默认响应
Default error response
##### SchemaClusterManagerListNodeTemplateDefaultBody

##### 内联模型

**ClusterManagerListNodeTemplateDefaultBody**


> 通用错误响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 错误码 |  |
| details | []interface{}|  | 错误详情 |  |
| message | string|  | 错误信息 |  |



**ClusterManagerListNodeTemplateOKBody**


> 查询节点模板列表响应



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)| 是 | 返回错误码 |  |
| data | []ClusterManagerListNodeTemplateOKBodyDataItems0|  | 节点模板列表 |  |
| message | string| 是 | 返回错误信息 |  |
| result | boolean|  | 返回结果 | `false` |



**ClusterManagerListNodeTemplateOKBodyDataItems0**


> 节点模板，用于 Kubernetes 集群节点通用特性设置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipScaleInWhenFailed | boolean|  | 缩容失败时是否允许跳过流程 | `false` |
| allowSkipScaleOutWhenFailed | boolean|  | 扩容失败时是否允许跳过流程 | `false` |
| annotations | map of string|  | 节点注解 |  |
| bcsScaleInAddons | ClusterManagerListNodeTemplateOKBodyDataItems0BcsScaleInAddons|  |  |  |
| bcsScaleOutAddons | ClusterManagerListNodeTemplateOKBodyDataItems0BcsScaleOutAddons|  |  |  |
| createTime | string|  | 创建时间 |  |
| creator | string|  | 创建者 |  |
| dataDisks | []ClusterManagerListNodeTemplateOKBodyDataItems0DataDisksItems0|  | 数据盘配置列表 |  |
| desc | string|  | 节点模板描述信息 |  |
| dockerGraphPath | string|  | Docker 数据目录 |  |
| extraArgs | map of string|  | 额外参数 |  |
| extraInfo | map of string|  | 额外信息 |  |
| gpuArgs | ClusterManagerListNodeTemplateOKBodyDataItems0GpuArgs|  |  |  |
| image | ClusterManagerListNodeTemplateOKBodyDataItems0Image|  |  |  |
| labels | map of string|  | 节点标签 |  |
| maxPodsPerNode | int64 (formatted integer)|  | 每个节点最大 Pod 数量 |  |
| module | ClusterManagerListNodeTemplateOKBodyDataItems0Module|  |  |  |
| mountTarget | string|  | 数据盘挂载目录 |  |
| name | string|  | 节点模板名称 |  |
| nodeOS | string|  | 节点模板操作系统 |  |
| nodeTemplateID | string|  | 节点模板ID |  |
| preStartUserScript | string|  | 扩容节点前置脚本，base64 编码 |  |
| projectID | string|  | 节点模板关联项目ID | `b37778ec757544868a01e1f01f07037f` |
| runtime | ClusterManagerListNodeTemplateOKBodyDataItems0Runtime|  |  |  |
| scaleInExtraAddons | ClusterManagerListNodeTemplateOKBodyDataItems0ScaleInExtraAddons|  |  |  |
| scaleInPostScript | string|  | 缩容节点后置脚本，base64 编码 |  |
| scaleInPreScript | string|  | 缩容节点前置脚本，base64 编码 |  |
| scaleOutExtraAddons | ClusterManagerListNodeTemplateOKBodyDataItems0ScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean|  | 是否跳过系统初始化 | `false` |
| taints | []ClusterManagerListNodeTemplateOKBodyDataItems0TaintsItems0|  | 节点污点列表 |  |
| unSchedulable | int64 (formatted integer)|  | 加入节点是否参与调度 |  |
| updateTime | string|  | 更新时间 |  |
| updater | string|  | 更新者 |  |
| userScript | string|  | 扩容节点后置脚本，base64 编码 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0BcsScaleInAddons**


> 标准运维操作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0BcsScaleOutAddons**


> 标准运维操作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0DataDisksItems0**


> 云数据盘配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean|  | 是否自动格式化并挂载 | `false` |
| diskSize | string|  | 数据盘大小 |  |
| diskType | string|  | 数据盘类型 |  |
| fileSystem | string|  | 文件系统类型 |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0GpuArgs**


> GPU 参数



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cudaVersion | string|  | CUDA 版本 |  |
| driverVersion | string|  | 驱动版本 |  |
| gpuMachine | string|  | GPU 机型 |  |
| gpuType | string|  | GPU 类型 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0Image**


> 节点模板镜像信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID |  |
| imageName | string|  | 镜像名称 |  |
| imageOs | string|  | 镜像操作系统 |  |
| imageType | string|  | 镜像类型 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0Module**


> 业务模块信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInModuleID | string|  | 缩容模块ID |  |
| scaleInModuleName | string|  | 缩容模块名称 |  |
| scaleOutModuleID | string|  | 扩容模块ID |  |
| scaleOutModuleName | string|  | 扩容模块名称 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0Runtime**


> 容器运行时信息



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | string|  | 容器运行时 |  |
| runtimeVersion | string|  | 运行时版本 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0ScaleInExtraAddons**


> 标准运维操作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0ScaleOutExtraAddons**


> 标准运维操作配置



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| postActions | []string|  | 后置操作列表 |  |
| preActions | []string|  | 前置操作列表 |  |



**ClusterManagerListNodeTemplateOKBodyDataItems0TaintsItems0**


> 节点污点



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，可选值：NoSchedule、PreferNoSchedule、NoExecute |  |
| key | string|  | 污点键 |  |
| value | string|  | 污点值 |  |
