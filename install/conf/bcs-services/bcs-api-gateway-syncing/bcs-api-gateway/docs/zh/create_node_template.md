### 描述

创建集群节点模版

创建集群节点模版(项目维度)。

### 输入参数

| 参数名称 | 参数位置 | 类型 | 必选 | 描述 |
|------|--------|------| :------: |-------------|
| projectID | `path` | string | 是 | [path_param参数] 项目唯一 ID（32 位十六进制字符串，如 MD5 哈希）。此参数在 URL 路径中传递，请替换为实际项目 ID。 |
| body | `body` | ClusterManagerCreateNodeTemplateBody | 是 |  |

### 所有响应
| 状态码 | 状态 | 描述 |
|------|--------|-------------|
| 200 | OK | A successful response. |
| default | | An unexpected error response |

### 响应

#### 200 - A successful response.
Status: OK

##### Schema

ClusterManagerCreateNodeTemplateOKBody

#### 默认响应
An unexpected error response
##### SchemaClusterManagerCreateNodeTemplateDefaultBody

##### 内联模型

**ClusterManagerCreateNodeTemplateBody**


> 创建节点模板



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| ScaleInPostScript | string|  | 缩容节点后置脚本, base64编码的用户脚本. |  |
| ScaleInPreScript | string|  | 缩容节点后置脚本, base64编码的用户脚本. |  |
| annotations | ClusterManagerCreateNodeTemplateParamsBodyAnnotations|  |  |  |
| creator | string| 是 | cloud账号创建者 |  |
| dataDisks | []ClusterManagerCreateNodeTemplateParamsBodyDataDisksItems0|  | 数据盘信息，无默认设置则不购买. 多盘数据盘挂载信息：新建节点时请确保购买CVM的参数传递了购买多个数据盘的信息该参数主要是针对自动生产机器且设置多数据盘场景 |  |
| desc | string|  | 节点模板描述信息 |  |
| dockerGraphPath | string|  | 节点运行时为docker时设置信息数据目录，BCS默认设置/data/bcs/service/docker |  |
| extraArgs | map of string|  |  |  |
| extraInfo | ClusterManagerCreateNodeTemplateParamsBodyExtraInfo|  |  |  |
| gpuArgs | ClusterManagerCreateNodeTemplateParamsBodyGpuArgs|  |  |  |
| imageInfo | ClusterManagerCreateNodeTemplateParamsBodyImageInfo|  |  |  |
| labels | map of string|  |  |  |
| module | ClusterManagerCreateNodeTemplateParamsBodyModule|  |  |  |
| mountTarget | string|  | 节点数据盘挂载点, 默认不挂载数据盘。已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载。无数据盘或有多块数据盘的云主机此设置不生效。BCS默认/data |  |
| name | string| 是 | 节点模板名称 |  |
| nodeOS | string|  | 节点模版OS |  |
| preStartUserScript | string|  | 上架节点前置脚本，base64 编码的用户脚本，在初始化节点之前执行，注意目前只对添加已有节点生效 |  |
| projectID | string| 是 | BCS 项目唯一标识，32 位十六进制字符串（如 MD5）。与 projectCode 区别：projectID 为 API 内部标识，projectCode 为人类可读的项目英文编码。 |  |
| runtime | ClusterManagerCreateNodeTemplateParamsBodyRuntime|  |  |  |
| scaleInExtraAddons | ClusterManagerCreateNodeTemplateParamsBodyScaleInExtraAddons|  |  |  |
| scaleOutExtraAddons | ClusterManagerCreateNodeTemplateParamsBodyScaleOutExtraAddons|  |  |  |
| skipSystemInit | boolean (formatted boolean)|  | 是否跳过执行系统初始化, 默认执行 |  |
| taints | []ClusterManagerCreateNodeTemplateParamsBodyTaintsItems0|  | 新实例初始化taints信息 |  |
| unSchedulable | int64 (formatted integer)|  | 设置加入的节点是否参与调度，默认值0表示参与调度，非0表示不参与调度(加入节点时可覆盖参数) |  |
| userScript | string|  | 上架节点后置脚本, base64编码的用户脚本，k8s组件运行后执行，需要保证脚本的可重入和重试逻辑。如果要求节点需要在进行初始化完成后才可加入调度 |  |



**ClusterManagerCreateNodeTemplateDefaultBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int32 (formatted integer)|  |  |  |
| details | []ClusterManagerCreateNodeTemplateDefaultBodyDetailsItems0|  |  |  |
| error | string|  |  |  |
| message | string|  |  |  |



**ClusterManagerCreateNodeTemplateDefaultBodyDetailsItems0**


> Arbitrary serialized protobuf message with a type URL (google.protobuf.Any).



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| type_url | string|  | Type URL that identifies the serialized protobuf message type. |  |
| value | byte (base64 string)|  | Serialized protocol buffer bytes of the type specified by type_url. |  |



**ClusterManagerCreateNodeTemplateOKBody**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| code | int64 (formatted integer)|  | 返回错误码 |  |
| message | string|  | 返回错误信息 |  |
| result | boolean (formatted boolean)|  | 返回结果 |  |
| web_annotations | ClusterManagerCreateNodeTemplateOKBodyWebAnnotations|  |  |  |



**ClusterManagerCreateNodeTemplateOKBodyWebAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| perms | interface{}|  | 权限信息 |  |



**ClusterManagerCreateNodeTemplateParamsBodyAnnotations**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| values | map of string|  |  |  |



**ClusterManagerCreateNodeTemplateParamsBodyDataDisksItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| autoFormatAndMount | boolean (formatted boolean)|  | 是否自动化格式盘并挂载 |  |
| diskPartition | string|  | 挂载设备名或分区名，当且仅当添加已有节点时需要 |  |
| diskSize | string|  | 数据盘大小，10G起跳，默认为0，为0时不购买 |  |
| diskType | string|  | 数据盘类型，LOCAL_BASIC（默认）,LOCAL_SSD,CLOUD_BASE,CLOUD_SSD,CLOUD_PREMIUM(高性能云硬盘) |  |
| fileSystem | string|  | 文件系统(ext3/ext4/xfs) |  |
| mountTarget | string|  | 挂载目录 |  |



**ClusterManagerCreateNodeTemplateParamsBodyExtraInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| values | map of string|  |  |  |



**ClusterManagerCreateNodeTemplateParamsBodyGpuArgs**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| cuda | ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCuda|  |  |  |
| cudnn | ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCudnn|  |  |  |
| customDriver | ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCustomDriver|  |  |  |
| driver | ClusterManagerCreateNodeTemplateParamsBodyGpuArgsDriver|  |  |  |
| migEnable | boolean (formatted boolean)|  | 是否启用MIG特性 |  |



**ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCuda**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCudnn**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| devName | string|  | cuDNN的Dev名字 |  |
| docName | string|  | cuDNN的Doc名字 |  |
| name | string|  | cuDNN的名字 |  |
| version | string|  | cuDNN的版本 |  |



**ClusterManagerCreateNodeTemplateParamsBodyGpuArgsCustomDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| address | string|  | 自定义GPU驱动地址链接 |  |



**ClusterManagerCreateNodeTemplateParamsBodyGpuArgsDriver**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| name | string|  | 节点模版镜像信息 |  |
| version | string|  | 节点模版镜像信息 |  |



**ClusterManagerCreateNodeTemplateParamsBodyImageInfo**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| imageID | string|  | 镜像ID，必填字段 |  |
| imageName | string|  | 镜像名 |  |
| imageOs | string|  | 镜像系统 |  |
| imageType | string|  | 镜像类型(PUBLIC_IMAGE/PRIVATE_IMAGE/MARKET_IMAGE/BCS_IMAGE) |  |



**ClusterManagerCreateNodeTemplateParamsBodyModule**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| scaleInBizID | string|  | 缩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleInModuleID | string|  | CC模块ID,CA缩容节点转移模块 |  |
| scaleInModuleName | string|  | CC模块名称,CA缩容节点转移模块 |  |
| scaleOutBizID | string|  | 扩容节点CC业务ID, 默认是集群所属业务ID |  |
| scaleOutModuleID | string|  | CC模块ID,CA扩容节点转移模块 |  |
| scaleOutModuleName | string|  | CC模块名称,CA扩容节点转移模块 |  |



**ClusterManagerCreateNodeTemplateParamsBodyRuntime**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| containerRuntime | string|  | 容器运行时 |  |
| runtimeVersion | string|  | 运行时版本 |  |



**ClusterManagerCreateNodeTemplateParamsBodyScaleInExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateNodeTemplateParamsBodyScaleInExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateNodeTemplateParamsBodyScaleInExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateNodeTemplateParamsBodyScaleOutExtraAddons**


> 记录各模块功能自动化行为模板, 对接job系统, 完成自定义操作; 针对不同云的不同操作模版参数需自定义针对大部分操作，仅需要前置动作或后置动作即可. 可自由扩展



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| plugins | map ofClusterManagerCreateNodeTemplateParamsBodyScaleOutExtraAddonsPluginsAnon|  |  |  |
| postActions | []string|  | 行为后置自动化模板列表，action在操作行为后置执行 |  |
| preActions | []string|  | 行为前置自动化模板列表，action在操作行为前置执行 |  |



**ClusterManagerCreateNodeTemplateParamsBodyScaleOutExtraAddonsPluginsAnon**


> 标准运维模板信息记录，用于自动化工作流程记录



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| allowSkipWhenFailed | boolean (formatted boolean)|  | allowSkipWhenFailed插件执行失败时是否允许跳过, 默认不允许跳过 |  |
| link | string|  | bksops模板链接，用于发起标准运维事务远程调用 |  |
| params | map of string|  |  |  |
| system | string|  | 插件系统，本插件为bksops |  |



**ClusterManagerCreateNodeTemplateParamsBodyTaintsItems0**



**Properties**

| 名称 | 类型 | 必选 | 描述 | 示例 |
|------|------|:--------:|-------------|---------|
| effect | string|  | 污点效果，(NoSchedule, PreferNoSchedule, NoExecute) |  |
| key | string|  | 污点 Key |  |
| value | string|  | 污点 Value |  |
