/* eslint-disable @typescript-eslint/no-empty-interface */
// 自动生成的, 请勿手动编辑!!!
// interface CommonResp {
//   code: number | undefined // 返回错误码
//   message: string | undefined // 返回错误信息
//   result: boolean | undefined // 返回结果
//   data: any | undefined // 返回数据
//   requestID: string | undefined // requestID
//   webAnnotations: WebAnnotations | undefined // 权限信息
// }
// interface CommonListResp {
//   code: number | undefined // 返回错误码
//   message: string | undefined // 返回错误信息
//   result: boolean | undefined // 返回结果
//   data: any[] | undefined // 返回数据
//   requestID: string | undefined // requestID
//   webAnnotations: WebAnnotations | undefined // 权限信息
// }
interface WebAnnotations {
  perms: any | undefined // 权限信息
}
interface AvailableReq {
}
interface AvailableResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
}
interface CreateRepositoryReq {
  $projectCode?: string | undefined // 项目代码
  name: string | undefined // 仓库名称
  type: string | undefined // 仓库类型，HELM(Helm仓库), GENERIC(通用二进制文件仓库)
  takeover: boolean | undefined // 是否为接管已存在的仓库
  repoURL: string | undefined // 接管仓库的仓库地址
  username: string | undefined // 接管仓库的用户名
  password: string | undefined // 接管仓库的密码
  remote: boolean | undefined // 是否为远程仓库
  remoteURL: string | undefined // 远程仓库地址
  remoteUsername: string | undefined // 远程仓库用户名
  remotePassword: string | undefined // 远程仓库密码
  displayName: string | undefined // 展示名称
}
interface CreateRepositoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Repository | undefined // 创建的仓库信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface CreatePersonalRepoReq {
  $projectCode?: string | undefined // 项目代码
}
interface CreatePersonalRepoResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Repository | undefined // 创建的仓库信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface UpdateRepositoryReq {
  $projectCode?: string | undefined // 项目代码
  $name: string | undefined // 仓库名称
  type: string | undefined // 仓库类型
  remote: boolean | undefined // 是否为远程仓库
  remoteURL: string | undefined // 远程仓库地址
  username: string | undefined // 用户名
  password: string | undefined // 密码
}
interface UpdateRepositoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Repository | undefined // 更新的仓库信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface GetRepositoryReq {
  $projectCode?: string | undefined // 项目代码
  $name: string | undefined // 仓库名称
}
interface GetRepositoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Repository | undefined // 查询的仓库信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ListRepositoryReq {
  $projectCode?: string | undefined // 项目代码
}
interface ListRepositoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Repository[] // 仓库列表
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface DeleteRepositoryReq {
  $projectCode?: string | undefined // 项目代码
  $name: string | undefined // 仓库名称
}
interface DeleteRepositoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface Repository {
  $projectCode?: string | undefined // 项目代码
  name: string | undefined // 仓库名称
  type: string | undefined // 仓库类型
  repoURL: string | undefined // 当前仓库的url
  username: string | undefined // 用户名
  password: string | undefined // 密码
  remote: boolean | undefined // 是否为远程仓库
  remoteURL: string | undefined // 远程仓库地址
  remoteUsername: string | undefined // 远程仓库username
  remotePassword: string | undefined // 远程仓库password
  createBy: string | undefined // 创建者
  updateBy: string | undefined // 更新者
  createTime: string | undefined // 创建时间
  updateTime: string | undefined // 更新时间
  displayName: string | undefined // 展示名称
  public: boolean | undefined // 是否是公共仓库
}
interface ChartListData {
  page: number | undefined // 页数
  size: number | undefined // 每页数量
  total: number | undefined // 总数
  data: Chart[] // 查询的仓库信息
}
interface Chart {
  $projectId?: string | undefined // 项目id
  repository: string | undefined // 仓库名称
  type: string | undefined // 仓库类型
  key: string | undefined // chart key
  name: string | undefined // chart名称
  latestVersion: string | undefined // 最新的chart version
  latestAppVersion: string | undefined // 最新的app version
  latestDescription: string | undefined // 最新的description
  createBy: string | undefined // 创建者
  updateBy: string | undefined // 更新者
  createTime: string | undefined // 创建时间
  updateTime: string | undefined // 更新时间
  $projectCode?: string | undefined // 项目代号
  icon: string | undefined // chart图标
}
interface ChartVersionListData {
  page: number | undefined // 页数
  size: number | undefined // 每页数量
  total: number | undefined // 总数
  data: ChartVersion[] // 查询的chart版本列表
}
interface ChartVersion {
  name: string | undefined // chart name
  version: string | undefined // chart version
  appVersion: string | undefined // chart app version
  description: string | undefined // chart description
  createBy: string | undefined // 创建者
  updateBy: string | undefined // 更新者
  createTime: string | undefined // 创建时间
  updateTime: string | undefined // 更新时间
  url: string | undefined // chart url
}
interface ChartDetail {
  name: string | undefined // chart名称
  version: string | undefined // chart版本
  readme: string | undefined // chart自述
  valuesFile: string[] // values文件列表
  contents: Record<string, FileContent | undefined> // chart包所含文件
  url: string | undefined // chart url
}
interface FileContent {
  name: string | undefined // 文件名
  path: string | undefined // 文件相对chart包入口的路径
  content: string | undefined // 文件内容
}
interface ListChartV1Req {
  page: number | undefined // 页数
  size: number | undefined // 每页数量
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  name: string | undefined // chart 名称，支持模糊搜索
}
interface ListChartV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ChartListData | undefined // 查询的chart的信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface GetChartDetailV1Req {
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
}
interface GetChartDetailV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Chart | undefined // chart包详细信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ListChartVersionV1Req {
  page: number | undefined // 页数
  size: number | undefined // 每页数量
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
}
interface ListChartVersionV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ChartVersionListData | undefined // 查询的chart的版本信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface GetVersionDetailV1Req {
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
  $version: string // chart版本
}
interface GetVersionDetailV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ChartDetail | undefined // 查询指定版本的chart包详细信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface DeleteChartReq {
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
}
interface DeleteChartResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface DeleteChartVersionReq {
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
  $version: string // chart版本
}
interface DeleteChartVersionResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface DownloadChartReq {
  $projectCode?: string // 项目代码
  $repoName: string // 仓库名称
  $name: string // chart名称
  $version: string // chart版本
}
interface UploadChartReq {
  $projectCode?: string | undefined // 项目代码
  repoName: string | undefined // 仓库名称
  file: number | undefined // file 上传文件
  version: string | undefined // chart版本
  force: boolean | undefined // 是否强制上传
}
interface UploadChartResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface GetChartReleaseReq {
  $projectCode?: string | undefined // 项目代码
  $repoName: string | undefined // 仓库名称
  $name: string | undefined // chart名称
  versions: string[] // chart版本, 不传版本则获取所有版本的 release
}
interface GetChartReleaseResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Release[] // release信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ImportClusterReleaseReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 集群ID
  $namespace: string | undefined // 查询目标namespace
  $name: string | undefined // 查询目标name
  repoName: string | undefined // 仓库名称
  chartName: string | undefined // chart名称
}
interface ImportClusterReleaseResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ReleaseListData {
  page: number | undefined // 页数
  size: number | undefined // 每页数量
  total: number | undefined // 总数
  data: Release[] // 查询的chart release列表
}
interface Release {
  name: string | undefined // chart release名称
  namespace: string | undefined // 所在的namespace
  revision: number | undefined // 所处于的版本
  status: string | undefined // 当前的状态
  chart: string | undefined // chart名称
  appVersion: string | undefined // 所处于的app version
  updateTime: string | undefined // 更新时间
  chartVersion: string | undefined // chart的版本
  createBy: string | undefined // 创建者
  updateBy: string | undefined // 更新者
  message: string | undefined // 报错消息
  repo: string | undefined // chart 仓库
  iamNamespaceID: string | undefined // iamNamespaceID
  $projectCode?: string | undefined // 项目编码
  $clusterId: string | undefined // 集群 ID
  env: string | undefined // 环境
}
interface ReleaseDetail {
  name: string | undefined // chart release名称
  namespace: string | undefined // 所在的namespace
  revision: number | undefined // 所处于的版本
  status: string | undefined // 当前的状态
  chart: string | undefined // chart名称
  appVersion: string | undefined // 所处于的app version
  updateTime: string | undefined // 更新时间
  chartVersion: string | undefined // chart的版本
  values: string[] // 当前revision部署时的values文件
  description: string | undefined // release 描述
  notes: string | undefined // release notes
  args: string[] // 当前revision部署时的 helm 参数
  createBy: string | undefined // 创建者
  updateBy: string | undefined // 更新者
  message: string | undefined // 报错消息
  repo: string | undefined // chart 仓库
  valueFile: string | undefined // value 文件，上一次更新使用的文件
}
interface ListReleaseV1Req {
  $projectCode?: string // 项目代码
  $clusterId: string // 集群ID
  namespace: string | undefined // 指定命名空间查询，默认全部命名空间
  name: string | undefined // 指定 release 名称查询，支持正则表达式，如 'ara[a-z]+' 可以搜索到 maudlin-arachnid
  page: number | undefined // 页数
  size: number | undefined // 每页数量
}
interface ListReleaseV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ReleaseListData | undefined // 指定集群的chart release信息
  webAnnotations: WebAnnotations | undefined // 返回数据
  requestID: string | undefined // requestID
}
interface GetReleaseDetailV1Req {
  $projectCode?: string // 项目代码
  $clusterId: string // 集群ID
  $namespace: string // 查询目标namespace
  $name: string // 查询目标name
}
interface GetReleaseDetailV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ReleaseDetail | undefined // 指定release信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface InstallReleaseV1Req {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $namespace: string | undefined // 所在的namespace
  $name: string | undefined // chart release名称
  repository: string | undefined // chart所属的仓库
  chart: string | undefined // chart名称
  version: string | undefined // chart版本
  values: string[] // values文件内容, 越靠后优先级越高
  args: string[] // 额外的参数
  valueFile: string | undefined // value 文件，上一次更新使用的文件
  operator: string | undefined // 操作者
  env: string | undefined // 环境
}
interface InstallReleaseV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface UninstallReleaseV1Req {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $namespace: string | undefined // 所在的namespace
  $name: string | undefined // chart release名称
}
interface UninstallReleaseV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface UpgradeReleaseV1Req {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $namespace: string | undefined // 所在的namespace
  $name: string | undefined // chart release名称
  repository: string | undefined // chart所属的仓库
  chart: string | undefined // chart名称
  version: string | undefined // chart版本
  values: string[] // values文件, 越靠后优先级越高
  args: string[] // 额外的参数
  valueFile: string | undefined // value 文件，上一次更新使用的文件
  operator: string | undefined // 操作者
  env: string | undefined // 环境
}
interface UpgradeReleaseV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface RollbackReleaseV1Req {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $namespace: string | undefined // 所在的namespace
  $name: string | undefined // chart release名称
  revision: number | undefined // 要回滚到的revision
}
interface RollbackReleaseV1Resp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ReleasePreviewReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $namespace: string | undefined // 所在的namespace
  $name: string | undefined // chart release名称
  repository: string | undefined // chart所属的仓库
  chart: string | undefined // chart名称
  version: string | undefined // chart版本
  values: string[] // values文件, 越靠后优先级越高
  args: string[] // 额外的参数
  revision: number | undefined // release revision 版本, 如果revision为0, 则对比当前渲染结果和已部署的最新版本, 如果不为0, 则对比最新版本和指定版本
}
interface ReleasePreviewResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ReleasePreview | undefined // 返回数据
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ReleasePreview {
  newContents: Record<string, FileContent | undefined> // 新版本 release manifest 内容，根据资源类型分类，键格式：{Kind}/{name}
  oldContents: Record<string, FileContent | undefined> // 旧版本 release manifest 内容，根据资源类型分类，键格式：{Kind}/{name}
  newContent: string | undefined // 新版本 release manifest 内容
  oldContent: string | undefined // 旧版本 release manifest 内容
}
interface GetReleaseHistoryReq {
  $name: string // chart release名称
  $namespace: string // 所在的namespace
  $clusterId: string // 所在的集群ID
  $projectCode?: string // 项目代码
  filter: string | undefined // 状态过滤
}
interface GetReleaseHistoryResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: ReleaseHistory[] // release history
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface ReleaseHistory {
  revision: number | undefined // release revision
  name: string | undefined // release name
  namespace: string | undefined // release namespace
  updateTime: string | undefined // 更新时间
  description: string | undefined // release 描述
  status: string | undefined // release 状态
  chart: string | undefined // release chart
  chartVersion: string | undefined // chart 版本
  appVersion: string | undefined // release appVersion
  values: string | undefined // release values
}
interface GetReleaseManifestReq {
  $name: string // chart release名称
  $namespace: string // 所在的namespace
  $clusterId: string // 所在的集群ID
  $projectCode?: string // 项目代码
  $revision: number // release 版本
}
interface GetReleaseManifestResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Record<string, FileContent | undefined> // release manifest
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface GetReleaseStatusReq {
  $name: string // chart release名称
  $namespace: string // 所在的namespace
  $clusterId: string // 所在的集群ID
  $projectCode?: string // 项目代码
}
interface GetReleaseDetailExtendReq {
  $name: string // chart release名称
  $namespace: string // 所在的namespace
  $clusterId: string // 所在的集群ID
  $projectCode?: string // 项目代码
}
interface GetReleasePodsReq {
  $name: string // chart release名称
  $namespace: string // 所在的namespace
  $clusterId: string // 所在的集群ID
  $projectCode?: string // 项目代码
  after: number | undefined // 查询指定时间之后的 pods
}
interface ListAddonsReq {
  $projectCode?: string // 项目代码
  $clusterId: string // 所在的集群ID
}
interface ListAddonsResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Addons[] // 组件列表
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface Addons {
  name: string // 组件名称
  chartName: string // chart name
  description: string | undefined // 组件描述
  logo: string | undefined // logo
  docsLink: string | undefined // 文档链接
  version: string // 组件最新版本
  currentVersion: string | undefined // 组件当前安装版本，空代表没安装
  namespace: string // 部署的命名空间
  defaultValues: string | undefined // 默认配置，空代表可以直接安装，不需要填写自定义配置
  currentValues: string | undefined // 当前部署配置
  status: string | undefined // 部署状态，同 Helm release 状态，空代表没安装
  message: string | undefined // 部署信息，部署异常则显示报错信息
  supportedActions: string[] // 组件支持的操作，目前有 install, upgrade, stop, uninstall
  releaseName: string | undefined // 组件在集群中的 release name
}
interface GetAddonsDetailReq {
  $projectCode?: string // 项目代码
  $clusterId: string // 所在的集群ID
  $name: string // 组件名称
}
interface GetAddonsDetailResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  data: Addons | undefined // 组件信息
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface InstallAddonsReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  name: string // 组件名称
  version: string // 组件版本
  values: string | undefined // values
}
interface InstallAddonsResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface UpgradeAddonsReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $name: string | undefined // 组件名称
  version: string | undefined // 组件版本
  values: string | undefined // values
}
interface UpgradeAddonsResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface StopAddonsReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $name: string | undefined // 组件名称
}
interface StopAddonsResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
interface UninstallAddonsReq {
  $projectCode?: string | undefined // 项目代码
  $clusterId: string | undefined // 所在的集群ID
  $name: string | undefined // 组件名称
}
interface UninstallAddonsResp {
  code: number | undefined // 返回错误码
  message: string | undefined // 返回错误信息
  result: boolean | undefined // 返回结果
  requestID: string | undefined // requestID
  webAnnotations: WebAnnotations | undefined // 权限信息
}
