import { createRequest } from '../request';
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/clusterresources/v1/projects/$projectId/clusters/$clusterId',
});

export const namespaceList = request('get', '/namespaces');
export const dashbordList = request('get', '/namespaces/$namespaceId/$type/$category');// 注意：HPA类型没有子分类$category
export const formSchema = request('get', '/form_schema');
export const dashbordListWithoutNamespace = request('get', '/$type/$category'); // PersistentVolume, StorageClass资源暂不支持命名空间
export const retrieveDetail = request('get', '/namespaces/$namespaceId/$type/$category/$name');
export const retrieveCustomObjectDetail = request('get', '/crds/$crdName/custom_objects/$name?namespace=$namespaceId');
export const retrieveContainerDetail = request('get', '/namespaces/$namespaceId/workloads/$category/$name/containers/$containerName');
export const listWorkloadPods = request('get', '/namespaces/$namespaceId/workloads/pods');
export const listStoragePods = request('get', '/namespaces/$namespaceId/workloads/pods/$podId/$type');
export const listContainers = request('get', '/namespaces/$namespaceId/workloads/pods/$podId/containers');
export const fetchContainerEnvInfo = request('get', '/namespaces/$namespaceId/workloads/pods/$podId/containers/$containerName/env_info');
export const resourceCreate = request('post', '/$type/$category');
export const resourceUpdate = request('put', '/namespaces/$namespaceId/$type/$category/$name');
export const resourceDelete = request('delete', '/namespaces/$namespaceId/$type/$category/$name');
export const exampleManifests = request('get', '/examples/manifests');
export const crdList = request('get', '/crds');// 获取CRD列表
export const customResourceList = request('get', '/crds/$crd/$category'); // 自定义资源
export const retrieveCustomResourceDetail = request('get', '/crds/$crd/$category/$name'); // 自定义资源详情
export const customResourceCreate = request('post', '/crds/$crd/$category'); // 自定义资源创建
export const customResourceUpdate = request('put', '/crds/$crd/$category/$name'); // 自定义资源更新
export const customResourceDelete = request('delete', '/crds/$crd/$category/$name'); // 自定义资源删除
export const reschedulePod = request('put', '/namespaces/$namespaceId/workloads/pods/$podId/reschedule'); // pod重新调度
export const renderManifestPreview = request('post', '/render_manifest_preview');
export const fetchNodePodsData = request('get', '/nodes/$nodename/workloads/pods');
export const enlargeCapacityChange = request('put', '/namespaces/$namespace/workloads/$category/$name/scale'); // 扩缩容
export const batchReschedulePod = request('put', '/namespaces/$namespace/workloads/$category/$name/reschedule'); // 批量重新调度
export const pvcMountInfo = request('get', '/namespaces/$namespace/storages/persistent_volume_claims/$pvcID/mount_info');
export const getNetworksEndpointsFlag = request('get', '/namespaces/$namespaces/networks/endpoints/$name/status');
export const getReplicasets = request('get', '/namespaces/$namespaceId/workloads/replicasets');// 获取deployment下rs资源
