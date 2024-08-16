// 自动生成的, 请勿手动编辑!!!
import { createRequest } from '../request';
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4',
});
export const HelmManagerService = {
// 查询helm-manager服务是否可用
  Available: (params?: AvailableReq): Promise<AvailableResp extends { data: any } ? AvailableResp['data'] : AvailableResp> => request('get', '/helmmanager/v1/available')(params),
  // 创建仓库
  CreateRepository: (params?: CreateRepositoryReq): Promise<CreateRepositoryResp extends { data: any } ? CreateRepositoryResp['data'] : CreateRepositoryResp> => request('post', '/helmmanager/v1/projects/$projectCode/repos')(params),
  // 创建个人仓库
  CreatePersonalRepo: (params?: CreatePersonalRepoReq): Promise<CreatePersonalRepoResp extends { data: any } ? CreatePersonalRepoResp['data'] : CreatePersonalRepoResp> => request('post', '/helmmanager/v1/projects/$projectCode/repos/personal')(params),
  // 更新仓库
  UpdateRepository: (params?: UpdateRepositoryReq): Promise<UpdateRepositoryResp extends { data: any } ? UpdateRepositoryResp['data'] : UpdateRepositoryResp> => request('put', '/helmmanager/v1/projects/$projectCode/repos/$name')(params),
  // 查询仓库
  GetRepository: (params?: GetRepositoryReq): Promise<GetRepositoryResp extends { data: any } ? GetRepositoryResp['data'] : GetRepositoryResp> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$name')(params),
  // 删除仓库
  DeleteRepository: (params?: DeleteRepositoryReq): Promise<DeleteRepositoryResp extends { data: any } ? DeleteRepositoryResp['data'] : DeleteRepositoryResp> => request('delete', '/helmmanager/v1/projects/$projectCode/repos/$name')(params),
  // 查询仓库列表
  ListRepository: (params?: ListRepositoryReq): Promise<ListRepositoryResp extends { data: any } ? ListRepositoryResp['data'] : ListRepositoryResp> => request('get', '/helmmanager/v1/projects/$projectCode/repos')(params),
  // 批量查询chart包
  ListChartV1: (params?: ListChartV1Req): Promise<ListChartV1Resp extends { data: any } ? ListChartV1Resp['data'] : ListChartV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts')(params),
  // 查询chart包详细信息
  GetChartDetailV1: (params?: GetChartDetailV1Req): Promise<GetChartDetailV1Resp extends { data: any } ? GetChartDetailV1Resp['data'] : GetChartDetailV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name')(params),
  // 查询chart包的版本列表
  ListChartVersionV1: (params?: ListChartVersionV1Req): Promise<ListChartVersionV1Resp extends { data: any } ? ListChartVersionV1Resp['data'] : ListChartVersionV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name/versions')(params),
  // 查询指定版本的chart包详细信息
  GetVersionDetailV1: (params?: GetVersionDetailV1Req): Promise<GetVersionDetailV1Resp extends { data: any } ? GetVersionDetailV1Resp['data'] : GetVersionDetailV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name/versions/$version')(params),
  // 删除指定 chart，将会从 chart 仓库中删除
  DeleteChart: (params?: DeleteChartReq): Promise<DeleteChartResp extends { data: any } ? DeleteChartResp['data'] : DeleteChartResp> => request('delete', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name')(params),
  // 删除 chart 指定版本，将会从 chart 仓库中删除
  DeleteChartVersion: (params?: DeleteChartVersionReq): Promise<DeleteChartVersionResp extends { data: any } ? DeleteChartVersionResp['data'] : DeleteChartVersionResp> => request('delete', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name/versions/$version')(params),
  // 下载指定版本的chart
  DownloadChart: (params?: DownloadChartReq): Promise<HttpBody extends { data: any } ? HttpBody['data'] : HttpBody> => request('get', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name/versions/$version/download')(params),
  // 获取 chart 关联的 releases
  GetChartRelease: (params?: GetChartReleaseReq): Promise<GetChartReleaseResp extends { data: any } ? GetChartReleaseResp['data'] : GetChartReleaseResp> => request('post', '/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$name/releases')(params),
  // 查询指定集群的chart release信息
  ListReleaseV1: (params?: ListReleaseV1Req): Promise<ListReleaseV1Resp extends { data: any } ? ListReleaseV1Resp['data'] : ListReleaseV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/releases')(params),
  // 查询指定release的详细信息
  GetReleaseDetailV1: (params?: GetReleaseDetailV1Req): Promise<GetReleaseDetailV1Resp extends { data: any } ? GetReleaseDetailV1Resp['data'] : GetReleaseDetailV1Resp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name')(params),
  // 执行指定集群的chart release install
  InstallReleaseV1: (params?: InstallReleaseV1Req): Promise<InstallReleaseV1Resp extends { data: any } ? InstallReleaseV1Resp['data'] : InstallReleaseV1Resp> => request('post', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name')(params),
  // 执行指定集群的chart release uninstall
  UninstallReleaseV1: (params?: UninstallReleaseV1Req): Promise<UninstallReleaseV1Resp extends { data: any } ? UninstallReleaseV1Resp['data'] : UninstallReleaseV1Resp> => request('delete', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name')(params),
  // 执行指定集群的chart release upgrade
  UpgradeReleaseV1: (params?: UpgradeReleaseV1Req): Promise<UpgradeReleaseV1Resp extends { data: any } ? UpgradeReleaseV1Resp['data'] : UpgradeReleaseV1Resp> => request('put', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name')(params),
  // 执行指定集群的chart release rollback
  RollbackReleaseV1: (params?: RollbackReleaseV1Req): Promise<RollbackReleaseV1Resp extends { data: any } ? RollbackReleaseV1Resp['data'] : RollbackReleaseV1Resp> => request('put', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/rollback')(params),
  // 预览 Release 资源，如果已经部署则同时展示 diff
  ReleasePreview: (params?: ReleasePreviewReq): Promise<ReleasePreviewResp extends { data: any } ? ReleasePreviewResp['data'] : ReleasePreviewResp> => request('post', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/preview')(params),
  // 查询指定集群的chart release历史信息
  GetReleaseHistory: (params?: GetReleaseHistoryReq): Promise<GetReleaseHistoryResp extends { data: any } ? GetReleaseHistoryResp['data'] : GetReleaseHistoryResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/history')(params),
  // 查询指定集群的chart release manifest
  GetReleaseManifest: (params?: GetReleaseManifestReq): Promise<GetReleaseManifestResp extends { data: any } ? GetReleaseManifestResp['data'] : GetReleaseManifestResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/revisions/$revision/manifest')(params),
  // 获取 release 下所有资源的部署状态，如果是 workload 则增加 pod 状态
  GetReleaseStatus: (params?: GetReleaseStatusReq): Promise<CommonListResp extends { data: any } ? CommonListResp['data'] : CommonListResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/status')(params),
  // Helm Release 详情扩展(ingress/service/secret)
  GetReleaseDetailExtend: (params?: GetReleaseDetailExtendReq): Promise<CommonResp extends { data: any } ? CommonResp['data'] : CommonResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/expend')(params),
  // 获取 release 下所有pod, 支持根据时间筛选
  GetReleasePods: (params?: GetReleasePodsReq): Promise<CommonListResp extends { data: any } ? CommonListResp['data'] : CommonListResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/pods')(params),
  // 导出 cluster releases
  ImportClusterRelease: (params?: ImportClusterReleaseReq): Promise<ImportClusterReleaseResp extends { data: any } ? ImportClusterReleaseResp['data'] : ImportClusterReleaseResp> => request('post', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/namespaces/$namespace/releases/$name/import')(params),
};
export const ClusterAddonsService = {
// 获取集群组件列表
  ListAddons: (params?: ListAddonsReq): Promise<ListAddonsResp extends { data: any } ? ListAddonsResp['data'] : ListAddonsResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons')(params),
  // 获取集群组件详情
  GetAddonsDetail: (params?: GetAddonsDetailReq): Promise<GetAddonsDetailResp extends { data: any } ? GetAddonsDetailResp['data'] : GetAddonsDetailResp> => request('get', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons/$name')(params),
  // 安装集群组件
  InstallAddons: (params?: InstallAddonsReq): Promise<InstallAddonsResp extends { data: any } ? InstallAddonsResp['data'] : InstallAddonsResp> => request('post', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons')(params),
  // 更新集群组件
  UpgradeAddons: (params?: UpgradeAddonsReq): Promise<UpgradeAddonsResp extends { data: any } ? UpgradeAddonsResp['data'] : UpgradeAddonsResp> => request('put', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons/$name')(params),
  // 停止组件实例运行，保留配置相关信息
  StopAddons: (params?: StopAddonsReq): Promise<StopAddonsResp extends { data: any } ? StopAddonsResp['data'] : StopAddonsResp> => request('put', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons/$name/stop')(params),
  // 删除组件所有内容
  UninstallAddons: (params?: UninstallAddonsReq): Promise<UninstallAddonsResp extends { data: any } ? UninstallAddonsResp['data'] : UninstallAddonsResp> => request('delete', '/helmmanager/v1/projects/$projectCode/clusters/$clusterId/addons/$name')(params),
};
