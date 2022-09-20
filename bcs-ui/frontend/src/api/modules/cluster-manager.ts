import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '/api/cluster_manager/proxy/bcsapi/v4',
});

// nodetemplate
export const nodeTemplateList = request('get', '/clustermanager/v1/projects/$projectId/nodetemplates');
export const createNodeTemplate = request('post', '/clustermanager/v1/projects/$projectId/nodetemplates');
export const deleteNodeTemplate = request('delete', '/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId');
export const updateNodeTemplate = request('put', '/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId');
export const nodeTemplateDetail = request('get', '/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId');
export const bkSopsList = request('get', '/clustermanager/v1/bksops/business/$businessID/templates');
export const bkSopsParamsList = request('get', '/clustermanager/v1/bksops/business/$businessID/templates/$templateID');
export const cloudModulesParamsList = request('get', '/clustermanager/v1/clouds/$cloudID/versions/$version/modules/$moduleID');
export const bkSopsDebug = request('post', '/clustermanager/v1/bksops/debug');
export const bkSopsTemplatevalues = request('get', '/clustermanager/v1/bksops/templatevalues');
export const getNodeTemplateInfo = request('get', '/clustermanager/v1/node/$innerIP/info');

// Cluster Manager
export const cloudList = request('get', '/clustermanager/v1/cloud');
export const createCluster = request('post', '/clustermanager/v1/cluster');
export const cloudVpc = request('get', '/clustermanager/v1/cloudvpc');
export const cloudRegion = request('get', '/clustermanager/v1/cloudregion/$cloudId');
export const vpccidrList = request('get', '/clustermanager/v1/vpccidr/$vpcID');
export const fetchClusterList = request('get', '/clustermanager/v1/cluster');
export const deleteCluster = request('delete', '/clustermanager/v1/cluster/$clusterId');
export const retryCluster = request('post', '/clustermanager/v1/cluster/$clusterId/retry');
export const taskList = request('get', '/clustermanager/v1/task');
export const taskDetail = request('get', '/clustermanager/v1/task/$taskId');
export const clusterNode = request('get', '/clustermanager/v1/cluster/$clusterId/node');
export const addClusterNode = request('post', '/clustermanager/v1/cluster/$clusterId/node');
export const deleteClusterNode = request('delete', '/clustermanager/v1/cluster/$clusterId/node');
export const clusterDetail = request('get', '/clustermanager/v1/cluster/$clusterId');
export const modifyCluster = request('put', '/clustermanager/v1/cluster/$clusterId');
export const importCluster = request('post', '/clustermanager/v1/cluster/import');
export const kubeConfig = request('put', '/clustermanager/v1/cloud/kubeConfig');
export const nodeAvailable = request('post', '/clustermanager/v1/node/available');
export const cloudAccounts = request('get', '/clustermanager/v1/clouds/$cloudId/accounts');
export const createCloudAccounts = request('post', '/clustermanager/v1/clouds/$cloudId/accounts');
export const deleteCloudAccounts = request('delete', '/clustermanager/v1/clouds/$cloudId/accounts/$accountID');
export const cloudRegionByAccount = request('get', '/clustermanager/v1/clouds/$cloudId/regions');
export const cloudClusterList = request('get', '/clustermanager/v1/clouds/$cloudId/clusters');
export const taskRetry = request('put', '/clustermanager/v1/task/$taskId/retry');

const request2 = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '',
});
// node 操作
export const getK8sNodes = request2('get', '/api/cluster_mgr/projects/$projectId/clusters/$clusterId/nodes/');
export const fetchK8sNodeLabels = request2('post', '/api/cluster_mgr/projects/$projectId/clusters/$clusterId/nodes/labels/');
export const setK8sNodeLabels = request2('put', '/api/cluster_mgr/projects/$projectId/clusters/$clusterId/nodes/labels/');
export const getNodeTaints = request2('post', '/api/cluster_mgr/projects/$projectId/clusters/$clusterId/nodes/taints/');
export const setNodeTaints = request2('put', '/api/cluster_mgr/projects/$projectId/clusters/$clusterId/nodes/taints/');
export const schedulerNode = request2('put', '/api/projects/$projectId/clusters/$clusterId/pods/reschedule/');
