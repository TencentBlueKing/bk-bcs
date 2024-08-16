// 自动生成的, 请勿手动编辑!!!
import { createRequest } from '../request';
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4',
});
export const BasicService = {
// Echo 接口，用于开发测试
  Echo: (params?: ClusterResource.EchoReq, config?: IFetchConfig): Promise<ClusterResource.EchoResp extends { data: any } ? ClusterResource.EchoResp['data'] : ClusterResource.EchoResp> => request('post', '/clusterresources/v1/echo')(params, config),
  // Ping 接口，用于检查服务是否存活
  Ping: (params?: ClusterResource.PingReq, config?: IFetchConfig): Promise<ClusterResource.PingResp extends { data: any } ? ClusterResource.PingResp['data'] : ClusterResource.PingResp> => request('get', '/clusterresources/v1/ping')(params, config),
  // Healthz 接口，用于检查服务健康状态
  Healthz: (params?: ClusterResource.HealthzReq, config?: IFetchConfig): Promise<ClusterResource.HealthzResp extends { data: any } ? ClusterResource.HealthzResp['data'] : ClusterResource.HealthzResp> => request('get', '/clusterresources/v1/healthz')(params, config),
  // Version 接口，用于获取服务版本信息
  Version: (params?: ClusterResource.VersionReq, config?: IFetchConfig): Promise<ClusterResource.VersionResp extends { data: any } ? ClusterResource.VersionResp['data'] : ClusterResource.VersionResp> => request('get', '/clusterresources/v1/version')(params, config),
};
export const NodeService = {
// 获取节点列表
  ListNode: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/nodes')(params, config),
};
export const NamespaceService = {
// 获取命名空间列表
  ListNS: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces')(params, config),
};
export const WorkloadService = {
// 获取 Deployment 列表
  ListDeploy: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments')(params, config),
  // 获取 Deployment
  GetDeploy: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name')(params, config),
  // 创建 Deployment
  CreateDeploy: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/deployments')(params, config),
  // 更新 Deployment
  UpdateDeploy: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name')(params, config),
  // 重新调度 Deployment
  RestartDeploy: (params?: ClusterResource.ResRestartReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/restart')(params, config),
  // 暂停或恢复 Deployment
  PauseOrResumeDeploy: (params?: ClusterResource.ResPauseOrResumeReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/pause_resume/$value')(params, config),
  // Deployment 扩缩容
  ScaleDeploy: (params?: ClusterResource.ResScaleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/scale')(params, config),
  // 重新调度 Deployment 下属的 Pod
  RescheduleDeployPo: (params?: ClusterResource.ResBatchRescheduleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/reschedule')(params, config),
  // 删除 Deployment
  DeleteDeploy: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name')(params, config),
  // 获取 Deployment Revision
  GetDeployHistoryRevision: (params?: ClusterResource.GetResHistoryReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/history')(params, config),
  // 获取deployment revision差异信息
  GetDeployRevisionDiff: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/revisions/$revision')(params, config),
  // 回滚 Deployment Revision
  RolloutDeployRevision: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/deployments/$name/rollout/$revision')(params, config),
  // 获取 ReplicasSet 列表
  ListRS: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/replicasets')(params, config),
  // 获取 DaemonSet 列表
  ListDS: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets')(params, config),
  // 获取 DaemonSet
  GetDS: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name')(params, config),
  // 创建 DaemonSet
  CreateDS: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/daemonsets')(params, config),
  // 更新 DaemonSet
  UpdateDS: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name')(params, config),
  // 重新调度 DaemonSet
  RestartDS: (params?: ClusterResource.ResRestartReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name/restart')(params, config),
  // 获取 DaemonSet Revision
  GetDSHistoryRevision: (params?: ClusterResource.GetResHistoryReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name/history')(params, config),
  // 获取 DaemonSet revision差异信息
  GetDSRevisionDiff: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name/revisions/$revision')(params, config),
  // 回滚 DaemonSet Revision
  RolloutDSRevision: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name/rollout/$revision')(params, config),
  // 删除 DaemonSet
  DeleteDS: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/daemonsets/$name')(params, config),
  // 获取 StatefulSet 列表
  ListSTS: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets')(params, config),
  // 获取 StatefulSet
  GetSTS: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name')(params, config),
  // 创建 StatefulSet
  CreateSTS: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/statefulsets')(params, config),
  // 更新 StatefulSet
  UpdateSTS: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name')(params, config),
  // 重新调度 StatefulSet
  RestartSTS: (params?: ClusterResource.ResRestartReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/restart')(params, config),
  // 获取 StatefulSet Revision
  GetSTSHistoryRevision: (params?: ClusterResource.GetResHistoryReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/history')(params, config),
  // 获取 StatefulSet revision差异信息
  GetSTSRevisionDiff: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/revisions/$revision')(params, config),
  // 回滚 StatefulSet Revision
  RolloutSTSRevision: (params?: ClusterResource.RolloutRevisionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/rollout/$revision')(params, config),
  // StatefulSet 扩缩容
  ScaleSTS: (params?: ClusterResource.ResScaleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/scale')(params, config),
  // 重新调度 StatefulSets 下属的 Pod
  RescheduleSTSPo: (params?: ClusterResource.ResBatchRescheduleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name/reschedule')(params, config),
  // 删除 StatefulSet
  DeleteSTS: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/statefulsets/$name')(params, config),
  // 获取 CronJob 列表
  ListCJ: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/cronjobs')(params, config),
  // 获取 CronJob
  GetCJ: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/cronjobs/$name')(params, config),
  // 创建 CronJob
  CreateCJ: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/cronjobs')(params, config),
  // 更新 CronJob
  UpdateCJ: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/cronjobs/$name')(params, config),
  // 删除 CronJob
  DeleteCJ: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/cronjobs/$name')(params, config),
  // 获取 Job 列表
  ListJob: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/jobs')(params, config),
  // 获取 Job
  GetJob: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/jobs/$name')(params, config),
  // 创建 Job
  CreateJob: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/jobs')(params, config),
  // 更新 Job
  UpdateJob: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/jobs/$name')(params, config),
  // 删除 Job
  DeleteJob: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/jobs/$name')(params, config),
  // 获取 Pod 列表
  ListPo: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods')(params, config),
  // 通过节点名称获取 Pod 列表
  ListPoByNode: (params?: ClusterResource.ListPoByNodeReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/nodes/$nodeName/workloads/pods')(params, config),
  // 获取 Pod
  GetPo: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name')(params, config),
  // 创建 Pod
  CreatePo: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/workloads/pods')(params, config),
  // 更新 Pod
  UpdatePo: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name')(params, config),
  // 删除 Pod
  DeletePo: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name')(params, config),
  // 获取 Pod 关联的 PVC
  ListPoPVC: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name/pvcs')(params, config),
  // 获取 Pod 关联的 ConfigMap
  ListPoCM: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name/configmaps')(params, config),
  // 获取 Pod 关联的 Secret
  ListPoSecret: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name/secrets')(params, config),
  // 重新调度 Pod
  ReschedulePo: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$name/reschedule')(params, config),
  // 获取 Pod 包含的容器列表
  ListContainer: (params?: ClusterResource.ContainerListReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$podName/containers')(params, config),
  // 获取指定 Pod 下单个容器信息
  GetContainer: (params?: ClusterResource.ContainerGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$podName/containers/$containerName')(params, config),
  // 获取指定 Pod 下单个容器环境变量信息
  GetContainerEnvInfo: (params?: ClusterResource.ContainerGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/pods/$podName/containers/$containerName/env_info')(params, config),
};
export const NetworkService = {
// 获取 Ingress 列表
  ListIng: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/ingresses')(params, config),
  // 获取 Ingress
  GetIng: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/ingresses/$name')(params, config),
  // 创建 Ingress
  CreateIng: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/networks/ingresses')(params, config),
  // 更新 Ingress
  UpdateIng: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/ingresses/$name')(params, config),
  // 删除 Ingress
  DeleteIng: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/ingresses/$name')(params, config),
  // 获取 Service 列表
  ListSVC: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/services')(params, config),
  // 获取 Service
  GetSVC: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/services/$name')(params, config),
  // 创建 Service
  CreateSVC: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/networks/services')(params, config),
  // 更新 Service
  UpdateSVC: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/services/$name')(params, config),
  // 删除 Service
  DeleteSVC: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/services/$name')(params, config),
  // 获取 Endpoints 列表
  ListEP: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/endpoints')(params, config),
  // 获取 Endpoints
  GetEP: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/endpoints/$name')(params, config),
  // 获取 Endpoints 状态
  GetEPStatus: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/endpoints/$name/status')(params, config),
  // 创建 Endpoints
  CreateEP: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/networks/endpoints')(params, config),
  // 更新 Endpoints
  UpdateEP: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/endpoints/$name')(params, config),
  // 删除 Endpoints
  DeleteEP: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/networks/endpoints/$name')(params, config),
};
export const ConfigService = {
// 获取 ConfigMap 列表
  ListCM: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/configmaps')(params, config),
  // 获取 ConfigMap
  GetCM: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/configmaps/$name')(params, config),
  // 创建 ConfigMap
  CreateCM: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/configs/configmaps')(params, config),
  // 更新 ConfigMap
  UpdateCM: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/configmaps/$name')(params, config),
  // 删除 ConfigMap
  DeleteCM: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/configmaps/$name')(params, config),
  // 获取 Secret 列表
  ListSecret: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/secrets')(params, config),
  // 获取 Secret
  GetSecret: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/secrets/$name')(params, config),
  // 创建 Secret
  CreateSecret: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/configs/secrets')(params, config),
  // 更新 Secret
  UpdateSecret: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/secrets/$name')(params, config),
  // 删除 Secret
  DeleteSecret: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/configs/secrets/$name')(params, config),
};
export const StorageService = {
// 获取 PersistentVolume 列表
  ListPV: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volumes')(params, config),
  // 获取 PersistentVolume
  GetPV: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volumes/$name')(params, config),
  // 创建 PersistentVolume
  CreatePV: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volumes')(params, config),
  // 更新 PersistentVolume
  UpdatePV: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volumes/$name')(params, config),
  // 删除 PersistentVolume
  DeletePV: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volumes/$name')(params, config),
  // 获取 PersistentVolumeClaim 列表
  ListPVC: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/storages/persistent_volume_claims')(params, config),
  // 获取 PersistentVolumeClaim
  GetPVC: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/storages/persistent_volume_claims/$name')(params, config),
  // 获取 PersistentVolumeClaim 被 Pod 挂载的信息
  GetPVCMountInfo: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/storages/persistent_volume_claims/$name/mount_info')(params, config),
  // 创建 PersistentVolumeClaim
  CreatePVC: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/persistent_volume_claims')(params, config),
  // 更新 PersistentVolumeClaim
  UpdatePVC: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/storages/persistent_volume_claims/$name')(params, config),
  // 删除 PersistentVolumeClaim
  DeletePVC: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/storages/persistent_volume_claims/$name')(params, config),
  // 获取 StorageClass 列表
  ListSC: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/storage_classes')(params, config),
  // 获取 StorageClass
  GetSC: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/storage_classes/$name')(params, config),
  // 创建 StorageClass
  CreateSC: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/storage_classes')(params, config),
  // 更新 StorageClass
  UpdateSC: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/storage_classes/$name')(params, config),
  // 删除 StorageClass
  DeleteSC: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/storages/storage_classes/$name')(params, config),
};
export const RBACService = {
// 获取 ServiceAccount 列表
  ListSA: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/rbac/service_accounts')(params, config),
  // 获取 ServiceAccount
  GetSA: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/rbac/service_accounts/$name')(params, config),
  // 创建 ServiceAccount
  CreateSA: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/rbac/service_accounts')(params, config),
  // 更新 ServiceAccount
  UpdateSA: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/rbac/service_accounts/$name')(params, config),
  // 删除 ServiceAccount
  DeleteSA: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/rbac/service_accounts/$name')(params, config),
};
export const HPAService = {
// 获取 HPA 列表
  ListHPA: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/hpa')(params, config),
  // 获取 HPA
  GetHPA: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/hpa/$name')(params, config),
  // 创建 HPA
  CreateHPA: (params?: ClusterResource.ResCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/hpa')(params, config),
  // 更新 HPA
  UpdateHPA: (params?: ClusterResource.ResUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/hpa/$name')(params, config),
  // 删除 HPA
  DeleteHPA: (params?: ClusterResource.ResDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/namespaces/$namespace/hpa/$name')(params, config),
};
export const CustomResService = {
// 获取 CRD 列表
  ListCRD: (params?: ClusterResource.ResListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds')(params, config),
  // 获取 CRD
  GetCRD: (params?: ClusterResource.ResGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$name')(params, config),
  // 获取 自定义资源 列表
  ListCObj: (params?: ClusterResource.CObjListReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects')(params, config),
  // 获取 自定义资源
  GetCObj: (params?: ClusterResource.CObjGetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName')(params, config),
  // 获取 自定义资源 Revision
  GetCObjHistoryRevision: (params?: ClusterResource.CObjHistoryReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/history')(params, config),
  // 获取 自定义资源 Revision diff
  GetCObjRevisionDiff: (params?: ClusterResource.CObjRolloutReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/revisions/$revision')(params, config),
  // 重新调度 自定义资源
  RestartCObj: (params?: ClusterResource.CObjRestartReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/restart')(params, config),
  // 回滚 自定义资源 Revision
  RolloutCObj: (params?: ClusterResource.CObjRolloutReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/rollout/$revision')(params, config),
  // 创建 自定义资源
  CreateCObj: (params?: ClusterResource.CObjCreateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects')(params, config),
  // 更新 自定义资源
  UpdateCObj: (params?: ClusterResource.CObjUpdateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName')(params, config),
  // 自定义资源扩缩容（仅 GameDeployment, GameStatefulSet 可用）
  ScaleCObj: (params?: ClusterResource.CObjScaleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/scale')(params, config),
  // 删除 自定义资源
  DeleteCObj: (params?: ClusterResource.CObjDeleteReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName')(params, config),
  // 重新调度自定义资源下属的 Pod（仅 GameDeployment, GameStatefulSet 可用）
  RescheduleCObjPo: (params?: ClusterResource.CObjBatchRescheduleReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/crds/$CRDName/custom_objects/$cobjName/reschedule')(params, config),
};
export const ResourceService = {
// 获取 K8S 资源模版
  GetK8SResTemplate: (params?: ClusterResource.GetK8SResTemplateReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/examples/manifests')(params, config),
  // 资源订阅
  Subscribe: (params?: ClusterResource.SubscribeReq, config?: IFetchConfig): Promise<ClusterResource.SubscribeResp extends { data: any } ? ClusterResource.SubscribeResp['data'] : ClusterResource.SubscribeResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/subscribe')(params, config),
  // 主动使 Discovery 缓存失效
  InvalidateDiscoveryCache: (params?: ClusterResource.InvalidateDiscoveryCacheReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/invalidate_discovery_cache')(params, config),
  // 表单化数据渲染预览
  FormDataRenderPreview: (params?: ClusterResource.FormRenderPreviewReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/render_manifest_preview')(params, config),
  // 表单化数据转换为 YAML
  FormToYAML: (params?: ClusterResource.FormToYAMLReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/form_to_yaml')(params, config),
  // YAML 转换为表单化数据
  YAMLToForm: (params?: ClusterResource.YAMLToFormReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/yaml_to_form')(params, config),
  // 获取指定资源表单 Schema，不带集群信息
  GetMultiResFormSchema: (params?: ClusterResource.GetMultiResFormSchemaReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('post', '/clusterresources/v1/projects/$projectCode/form_schema')(params, config),
  // 获取指定资源表单 Schema
  GetResFormSchema: (params?: ClusterResource.GetResFormSchemaReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/form_schema')(params, config),
  // 获取指定资源可用于表单化的 APIVersion
  GetFormSupportedAPIVersions: (params?: ClusterResource.GetFormSupportedApiVersionsReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/form_supported_api_versions')(params, config),
  // 获取用于下拉框选项的资源数据（selectItems）
  GetResSelectItems: (params?: ClusterResource.GetResSelectItemsReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectId/clusters/$clusterId/res_select_items')(params, config),
};
export const ViewConfigService = {
// 获取视图配置列表
  ListViewConfigs: (params?: ClusterResource.ListViewConfigsReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectCode/view_configs')(params, config),
  // 获取视图配置详情
  GetViewConfig: (params?: ClusterResource.GetViewConfigReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectCode/view_configs/$id')(params, config),
  // 创建视图配置1
  CreateViewConfig: (params?: ClusterResource.CreateViewConfigReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/view_configs')(params, config),
  // 更新视图配置
  UpdateViewConfig: (params?: ClusterResource.UpdateViewConfigReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/view_configs/$id')(params, config),
  // 视图重命名
  RenameViewConfig: (params?: ClusterResource.RenameViewConfigReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/view_configs/$id/rename')(params, config),
  // 删除单个视图配置
  DeleteViewConfig: (params?: ClusterResource.DeleteViewConfigReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectCode/view_configs/$id')(params, config),
  // 资源名称联想
  ResourceNameSuggest: (params?: ClusterResource.ViewSuggestReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/resource_name_suggest')(params, config),
  // label 联想
  LabelSuggest: (params?: ClusterResource.ViewSuggestReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/label_suggest')(params, config),
  // values 联想
  ValuesSuggest: (params?: ClusterResource.ViewSuggestReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/values_suggest')(params, config),
};
export const TemplateSetService = {
// 获取模板文件文件夹详情
  GetTemplateSpace: (params?: ClusterResource.GetTemplateSpaceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/spaces/$id')(params, config),
  // 获取模板文件文件夹列表
  ListTemplateSpace: (params?: ClusterResource.ListTemplateSpaceReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/spaces')(params, config),
  // 创建模板文件文件夹
  CreateTemplateSpace: (params?: ClusterResource.CreateTemplateSpaceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/spaces')(params, config),
  // 更新模板文件文件夹
  UpdateTemplateSpace: (params?: ClusterResource.UpdateTemplateSpaceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/template/spaces/$id')(params, config),
  // 删除模板文件文件夹
  DeleteTemplateSpace: (params?: ClusterResource.DeleteTemplateSpaceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectCode/template/spaces/$id')(params, config),
  // 获取模板文件元数据详情
  GetTemplateMetadata: (params?: ClusterResource.GetTemplateMetadataReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/metadatas/$id')(params, config),
  // 获取模板文件元数据列表
  ListTemplateMetadata: (params?: ClusterResource.ListTemplateMetadataReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/$templateSpaceID/metadatas')(params, config),
  // 创建模板文件元数据
  CreateTemplateMetadata: (params?: ClusterResource.CreateTemplateMetadataReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/$templateSpaceID/metadatas')(params, config),
  // 更新模板文件元数据
  UpdateTemplateMetadata: (params?: ClusterResource.UpdateTemplateMetadataReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/template/metadatas/$id')(params, config),
  // 删除模板文件元数据
  DeleteTemplateMetadata: (params?: ClusterResource.DeleteTemplateMetadataReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectCode/template/metadatas/$id')(params, config),
  // 获取模板文件版本详情
  GetTemplateVersion: (params?: ClusterResource.GetTemplateVersionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/versions/$id')(params, config),
  // 获取模板文件详情
  GetTemplateContent: (params?: ClusterResource.GetTemplateContentReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/detail')(params, config),
  // 获取模板文件版本列表
  ListTemplateVersion: (params?: ClusterResource.ListTemplateVersionReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectCode/template/$templateID/versions')(params, config),
  // 创建模板文件版本
  CreateTemplateVersion: (params?: ClusterResource.CreateTemplateVersionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/$templateID/versions')(params, config),
  // 删除模板文件版本
  DeleteTemplateVersion: (params?: ClusterResource.DeleteTemplateVersionReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectCode/template/versions/$id')(params, config),
  // 创建模板集
  CreateTemplateSet: (params?: ClusterResource.CreateTemplateSetReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/templatesets')(params, config),
  // 获取模板文件变量列表
  ListTemplateFileVariables: (params?: ClusterResource.ListTemplateFileVariablesReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/variables')(params, config),
  // 部署模板文件
  PreviewTemplateFile: (params?: ClusterResource.DeployTemplateFileReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/preview')(params, config),
  // 部署模板文件
  DeployTemplateFile: (params?: ClusterResource.DeployTemplateFileReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/template/deploy')(params, config),
  // 获取环境管理详情
  GetEnvManage: (params?: ClusterResource.GetEnvManageReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('get', '/clusterresources/v1/projects/$projectCode/envs/$id')(params, config),
  // 获取环境管理列表
  ListEnvManages: (params?: ClusterResource.ListEnvManagesReq, config?: IFetchConfig): Promise<ClusterResource.CommonListResp extends { data: any } ? ClusterResource.CommonListResp['data'] : ClusterResource.CommonListResp> => request('get', '/clusterresources/v1/projects/$projectCode/envs')(params, config),
  // 创建环境管理
  CreateEnvManage: (params?: ClusterResource.CreateEnvManageReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/envs')(params, config),
  // 更新环境管理
  UpdateEnvManage: (params?: ClusterResource.UpdateEnvManageReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/envs/$id')(params, config),
  // 环境管理重命名
  RenameEnvManage: (params?: ClusterResource.RenameEnvManageReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('put', '/clusterresources/v1/projects/$projectCode/envs/$id/rename')(params, config),
  // 删除环境管理
  DeleteEnvManage: (params?: ClusterResource.DeleteEnvManageReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('delete', '/clusterresources/v1/projects/$projectCode/envs/$id')(params, config),
};
export const MultiClusterService = {
// 获取多集群原生资源列表
  FetchMultiClusterResource: (params?: ClusterResource.FetchMultiClusterResourceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/multi_cluster_resources/$kind')(params, config),
  // 获取多集群自定义资源列表
  FetchMultiClusterCustomResource: (params?: ClusterResource.FetchMultiClusterCustomResourceReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/multi_cluster_resources/$crd/custom_objects')(params, config),
  // 获取集群资源数量
  MultiClusterResourceCount: (params?: ClusterResource.MultiClusterResourceCountReq, config?: IFetchConfig): Promise<ClusterResource.CommonResp extends { data: any } ? ClusterResource.CommonResp['data'] : ClusterResource.CommonResp> => request('post', '/clusterresources/v1/projects/$projectCode/multi_cluster_resources_count')(params, config),
};
