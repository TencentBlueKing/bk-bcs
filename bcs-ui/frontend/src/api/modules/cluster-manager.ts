import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/clustermanager/v1',
});

// nodetemplate
export const nodeTemplateList = request('get', '/projects/$projectId/nodetemplates');
export const createNodeTemplate = request('post', '/projects/$projectId/nodetemplates');
export const deleteNodeTemplate = request('delete', '/projects/$projectId/nodetemplates/$nodeTemplateId');
export const updateNodeTemplate = request('put', '/projects/$projectId/nodetemplates/$nodeTemplateId');
export const nodeTemplateDetail = request('get', '/projects/$projectId/nodetemplates/$nodeTemplateId');
export const bkSopsList = request('get', '/bksops/business/$businessID/templates');
export const bkSopsParamsList = request('get', '/bksops/business/$businessID/templates/$templateID');
export const cloudModulesParamsList = request('get', '/clouds/$cloudID/versions/$version/modules/$moduleID');
export const bkSopsDebug = request('post', '/bksops/debug');
export const bkSopsTemplatevalues = request('get', '/bksops/templatevalues');
export const getNodeTemplateInfo = request('get', '/node/$innerIP/info');

// Cluster Manager
export const cloudList = request('get', '/cloud');
export const createCluster = request('post', '/cluster');
export const cloudVpc = request('get', '/cloudvpc');
export const cloudRegion = request('get', '/cloudregion/$cloudId');
export const vpccidrList = request('get', '/vpccidr/$vpcID');
export const fetchClusterList = request('get', '/cluster');
export const deleteCluster = request('delete', '/cluster/$clusterId');
export const retryCluster = request('post', '/cluster/$clusterId/retry');
export const taskList = request('get', '/task');
export const taskDetail = request('get', '/task/$taskId');
export const clusterNode = request('get', '/cluster/$clusterId/node');
export const addClusterNode = request('post', '/cluster/$clusterId/node');
export const deleteClusterNode = request('delete', '/cluster/$clusterId/node');
export const clusterDetail = request('get', '/cluster/$clusterId');
export const modifyCluster = request('put', '/cluster/$clusterId');
export const importCluster = request('post', '/cluster/import');
export const kubeConfig = request('put', '/cloud/kubeConfig');
export const nodeAvailable = request('post', '/node/available');
export const cloudAccounts = request('get', '/clouds/$cloudId/accounts');
export const createCloudAccounts = request('post', '/clouds/$cloudId/accounts');
export const deleteCloudAccounts = request('delete', '/clouds/$cloudId/accounts/$accountID');
export const cloudRegionByAccount = request('get', '/clouds/$cloudId/regions');
export const cloudClusterList = request('get', '/clouds/$cloudId/clusters');
export const taskRetry = request('put', '/task/$taskId/retry');
export const cloudDetail = request('get', '/cloud/$cloudId');
export const cloudNodes = request('get', '/clouds/$cloudId/instances');

// node 操作
export const getK8sNodes = request('get', '/cluster/$clusterId/node');
export const uncordonNodes = request('put', '/node/uncordon');
export const cordonNodes = request('put', '/node/cordon');
export const schedulerNode = request('post', '/node/drain');
export const setNodeLabels = request('put', '/node/labels');
export const setNodeTaints = request('put', '/node/taints');

// 集群管理
export const masterList = request('get', '/cluster/$clusterId/master');

// auth
export const newUserPermsByAction = request('post', '/perms/actions/$actionId');

// CA
export const clusterAutoScalingLogsV2 = request('get', '/operationlogs');
export const cloudsZones = request('get', '/clouds/$cloudId/zones');
