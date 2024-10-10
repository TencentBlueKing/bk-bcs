/* eslint-disable @typescript-eslint/no-empty-interface */
// 自动生成的, 请勿手动编辑!!!
declare namespace ClusterResource {
  export interface EchoReq {
    str: string // 待回显字符串，长度在 2-30 之间，仅包含大小写字母及数字
  }
  export interface EchoResp {
    ret: string // 回显字符串
  }
  export interface PingReq {
  }
  export interface PingResp {
    ret: string // Pong
  }
  export interface HealthzReq {
    raiseErr: boolean // 存在依赖服务异常的情况，是否返回错误（默认只在响应体中标记）
    token: string // Healthz API Token
  }
  export interface HealthzResp {
    callTime: string // API 请求时间
    status: string // 服务状态
    redis: string // Redis 状态
  }
  export interface VersionReq {
  }
  export interface VersionResp {
    version: string // 服务版本
    gitCommit: string // 最新 Commit ID
    buildTime: string // 构建时间
    goVersion: string // Go 版本
    runMode: string // 运行模式
    callTime: string // API 请求时间
  }
  export interface ResListReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    labelSelector: string // 标签选择器
    apiVersion: string // apiVersion
    ownerName: string // 所属资源名称
    ownerKind: string // 所属资源类型
    format: string // 资源配置格式（manifest/selectItems）
    scene: string // 仅 selectItems 格式下有效
  }
  export interface ResGetReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    apiVersion: string // apiVersion
    format: string // 资源配置格式（manifest/formData）
  }
  export interface ResCreateReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    rawData: any // 资源配置信息
    format: string // 资源配置格式（manifest/formData）
  }
  export interface ResUpdateReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    rawData: any // 资源配置信息
    format: string // 资源配置格式（manifest/formData）
  }
  export interface ResRestartReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
  }
  export interface ResPauseOrResumeReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    $value: boolean // 暂停或者恢复(true暂停，false恢复)
  }
  export interface ResScaleReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    replicas: number // 副本数量（0-8192）
  }
  export interface ResDeleteReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
  }
  export interface GetResHistoryReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
  }
  export interface RolloutRevisionReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    $revision: number // revision 版本
  }
  export interface ResBatchRescheduleReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $name: string // 资源名称
    labelSelector: string // 标签选择器
    podNames: string[] // 待重新调度 Pod 名称列表
  }
  export interface ListPoByNodeReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $nodeName: string // 节点名称
  }
  export interface ContainerListReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $podName: string // Pod 名称
  }
  export interface ContainerGetReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    $namespace: string // 命名空间
    $podName: string // Pod 名称
    $containerName: string // 容器名称
  }
  export interface GetK8SResTemplateReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    kind: string // 资源类型
    namespace: string // 命名空间
  }
  export interface CObjListReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    namespace: string // 命名空间
    format: string // 资源配置格式（manifest/selectItems）
    scene: string // 仅 selectItems 格式下有效
  }
  export interface CObjGetReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
    format: string // 资源配置格式（manifest/formData）
  }
  export interface CObjHistoryReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
  }
  export interface CObjRestartReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
  }
  export interface CObjRolloutReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
    $revision: number // revision 版本
  }
  export interface CObjCreateReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    rawData: any // 资源配置信息
    format: string // 资源配置格式（manifest/formData）
  }
  export interface CObjUpdateReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
    rawData: any // 资源配置信息
    format: string // 资源配置格式（manifest/formData）
  }
  export interface CObjScaleReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    namespace: string // 命名空间
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    replicas: number // 副本数量（0-8192）
  }
  export interface CObjDeleteReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    namespace: string // 命名空间
  }
  export interface CObjBatchRescheduleReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    namespace: string // 命名空间
    cRDName: string // CRD 名称
    $cobjName: string // 自定义资源名称
    labelSelector: string // 标签选择器
    podNames: string[] // 待重新调度 Pod 名称列表
  }
  export interface CommonResp {
    code: number // 返回错误码
    message: string // 返回错误信息
    requestID: string // 请求 ID
    data: any // 资源信息
    webAnnotations: any // web 注解
  }
  export interface CommonListResp {
    code: number // 返回错误码
    message: string // 返回错误信息
    requestID: string // 请求 ID
    data: any[] // 资源信息
    webAnnotations: any // web 注解
  }
  export interface SubscribeReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    resourceVersion: string // 资源版本号
    kind: string // 资源类型
    cRDName: string // CRD 名称
    apiVersion: string // API 版本
    namespace: string // 命名空间
  }
  export interface SubscribeResp {
    code: number // 返回错误码
    message: string // 返回错误信息
    kind: string // 资源类型
    type: string // 操作类型
    uid: string // 唯一标识
    manifest: any // 资源配置信息
    manifestExt: any // 资源扩展信息
  }
  export interface InvalidateDiscoveryCacheReq {
    projectID: string // 项目 ID
    clusterID: string // 集群 ID
    authToken: string // AuthToken
  }
  export interface FormRenderPreviewReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    kind: string // 资源类型
    formData: any // 表单化数据
  }
  export interface FormData {
    apiVersion: string // api版本
    kind: string // 资源类型
    formData: any // 表单化数据
  }
  export interface FormToYAMLReq {
    $projectCode?: string // 项目编码
    resources: FormData[] // 资源列表
  }
  export interface YAMLToFormReq {
    $projectCode?: string // 项目编码
    yaml: string // YAML 数据
  }
  export interface FormResourceType {
    apiVersion: string // 资源版本
    kind: string // 资源类型
  }
  export interface GetMultiResFormSchemaReq {
    $projectCode?: string // 项目编码
    resourceTypes: FormResourceType[] // 资源类型
  }
  export interface GetResFormSchemaReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    kind: string // 资源类型
    namespace: string // 命名空间
    action: string // 模板使用场景（如表单创建，表单更新等）
  }
  export interface GetFormSupportedApiVersionsReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    kind: string // 资源类型
  }
  export interface GetResSelectItemsReq {
    $projectId?: string // 项目 ID
    $clusterId: string // 集群 ID
    kind: string // 资源类型
    // NOTE 目前使用该 API 的请求资源都有命名空间，因此先加上 min_len == 1 的限制
    namespace: string // 命名空间
    scene: string // 使用场景
  }
  export interface ListViewConfigsReq {
    $projectCode?: string // 项目编码
  }
  export interface GetViewConfigReq {
    $id: string // 视图 ID
    $projectCode?: string // 项目编码
  }
  export interface ViewFilter {
    name: string
    creator: string[]
    labelSelector: LabelSelector[]
  }
  export interface CreateViewConfigReq {
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    name: string // 视图名称
    filter: ViewFilter // 筛选条件
    saveAs: boolean // 另存为
  }
  export interface UpdateViewConfigReq {
    $id: string // 视图 ID
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    name: string // 视图名称
    filter: ViewFilter // 筛选条件
  }
  export interface RenameViewConfigReq {
    $id: string // 视图 ID
    $projectCode?: string // 项目编码
    name: string // 视图名称
  }
  export interface DeleteViewConfigReq {
    $id: string // 视图 ID
    $projectCode?: string // 项目编码
  }
  export interface ViewSuggestReq {
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    label: string // 标签
  }
  export interface ClusterNamespaces {
    clusterID: string // 集群 ID
    namespaces: string[]
  }
  export interface LabelSelector {
    key: string // key
    op: string // op
    values: string[] // values
  }
  export interface GetTemplateSpaceReq {
    $id: string // 文件夹 ID
    $projectCode?: string // 项目编码
  }
  export interface ListTemplateSpaceReq {
    $projectCode?: string // 项目编码
    name?: string // 文件夹名称
  }
  export interface CreateTemplateSpaceReq {
    $projectCode?: string // 项目编码
    name: string // 文件夹名称
    description: string // 文件夹描述
  }
  export interface UpdateTemplateSpaceReq {
    $id: string // 文件夹 ID
    $projectCode?: string // 项目编码
    name: string // 文件夹名称
    description: string // 文件夹描述
  }
  export interface DeleteTemplateSpaceReq {
    $id: string // 文件夹 ID
    $projectCode?: string // 项目编码
  }
  export interface GetTemplateMetadataReq {
    $id: string // 模板元数据 ID
    $projectCode?: string // 项目编码
  }
  export interface ListTemplateMetadataReq {
    $projectCode?: string // 项目编码
    $templateSpaceID: string // 模板文件文件夹
  }
  export interface CreateTemplateMetadataReq {
    $projectCode?: string // 项目编码
    name: string // 模板文件元数据名称
    description: string // 模板文件元数据描述
    $templateSpaceID: string // 模板文件文件夹
    tags: string[] // 模板文件元数据标签
    versionDescription: string // 模板文件版本描述
    version: string // 模板文件版本
    content: string // 模板文件版本Content
    isDraft: boolean // 是否草稿
    draftVersion: string // 基于模板文件版本
    draftContent: string // 模板文件草稿内容
    draftEditFormat: 'form' | 'yaml' | undefined // 草稿态编辑格式
  }
  export interface UpdateTemplateMetadataReq {
    $id: string // 模板文件元数据 ID
    $projectCode?: string // 项目编码
    name: string // 模板文件元数据名称
    description: string // 模板文件元数据描述
    tags: string[] // 模板文件元数据标签
    version: string // 模板文件版本
    versionMode: any // 模板文件版本模式
    isDraft: boolean // 是否草稿
    draftVersion: string // 基于模板文件版本
    draftContent: string // 模板文件草稿内容
    draftEditFormat?: 'form' | 'yaml' | undefined // 草稿态编辑格式
  }
  export interface DeleteTemplateMetadataReq {
    $id: string // 模板文件元数据 ID
    $projectCode?: string // 项目编码
  }
  export interface GetTemplateVersionReq {
    $id: string // 模板文件版本 ID
    $projectCode?: string // 项目编码
  }
  export interface GetTemplateContentReq {
    $projectCode?: string // 项目编码
    templateSpace: string // 模板文件文件夹名称
    templateName: string // 模板文件元数据名称
    version: string // 模板文件版本
  }
  export interface ListTemplateVersionReq {
    $projectCode?: string // 项目编码
    $templateID: string // 模板文件元数据ID
  }
  export interface CreateTemplateVersionReq {
    $projectCode?: string // 项目编码
    description: string // 模板文件版本描述
    version: string // 模板文件版本
    editFormat: string // 编辑模式
    content: string // 模板文件版本Content
    $templateID: string // 模板文件元数据ID
    force: boolean // 能否被覆盖
  }
  export interface DeleteTemplateVersionReq {
    $id: string // 模板文件版本 ID
    $projectCode?: string // 项目编码
  }
  export interface CreateTemplateSetReq {
    name: string // 模板集名称
    description: string // 模板集描述
    $projectCode?: string // 项目编码
    version: string // 模板集版本
    category: string // 模板集分类
    keywords: string[] // 模板集关键字
    readme: string // 模板集 README
    templates: TemplateID[] // 模板集包含的模板
    values: string // 模板集默认值
    force: boolean // 是否覆盖
  }
  export interface TemplateID {
    templateSpace: string // 模板文件文件夹名称
    templateName: string // 模板文件元数据名称
    version: string // 模板文件版本
  }
  export interface ListTemplateFileVariablesReq {
    $projectCode?: string // 项目编码
    templateVersions: string[] // 模板文件列表
    clusterID: string // 集群 ID
    namespace: string // 命名空间
  }
  export interface DeployTemplateFileReq {
    $projectCode?: string // 项目编码
    templateVersions: string[] // 模板文件列表
    variables: Record<string, string> // 模板文件变量值
    clusterID: string // 集群 ID
    namespace: string // 命名空间
  }
  export interface GetEnvManageReq {
    $id: string // 环境管理 ID
    $projectCode?: string // 项目编码
  }
  export interface ListEnvManagesReq {
    $projectCode?: string // 项目编码
  }
  export interface CreateEnvManageReq {
    $projectCode?: string // 项目编码
    env: string // 环境名称
    clusterNamespaces: ClusterNamespaces[] // 关联命名空间
  }
  export interface UpdateEnvManageReq {
    $id: string // 环境管理 ID
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 关联命名空间
    env: string // 环境名称
  }
  export interface RenameEnvManageReq {
    $id: string // 环境管理 ID
    $projectCode?: string // 项目编码
    env: string // 环境名称
  }
  export interface DeleteEnvManageReq {
    $id: string // 环境管理 ID
    $projectCode?: string // 项目编码
  }
  export interface FetchMultiClusterResourceReq {
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    $kind: string // 资源类型
    viewID: string // viewID
    creator: string[] // creator
    labelSelector: LabelSelector[] // 标签选择器
    name: string // name
    ip: string // ip
    status: string[] // status
    sortBy: string // sortBy
    order: string // desc
    limit: number // limit
    offset: number // offset
  }
  export interface FetchMultiClusterCustomResourceReq {
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    $crd: string // crd
    viewID: string // viewID
    creator: string[] // creator
    labelSelector: LabelSelector[] // 标签选择器
    name: string // name
    ip: string // ip
    status: string[] // status
    sortBy: string // sortBy
    order: string // desc
    limit: number // limit
    offset: number // offset
  }
  export interface MultiClusterResourceCountReq {
    $projectCode?: string // 项目编码
    clusterNamespaces: ClusterNamespaces[] // 集群和命名空间
    viewID: string // viewID
    creator: string[] // creator
    labelSelector: LabelSelector[] // 标签选择器
    name: string // name
  }
  export interface ExportTemplateReq {
    templateSpaceNames: string[] // 模板空间名称
  }
}
