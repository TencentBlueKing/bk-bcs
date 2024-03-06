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
export const fetchClusterList = request('get', '/projects/$projectId/clusters');
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
export const updateCloudAccounts = request('put', '/clouds/$cloudId/accounts/$accountID');
export const validateCloudAccounts = request('post', '/clouds/$cloudId/accounts/available');
export const clusterConnect = request('get', '/clouds/$cloudId/clusters/$clusterID/connect');
export const cloudRegionByAccount = request('get', '/clouds/$cloudId/regions');
export const cloudClusterList = request('get', '/clouds/$cloudId/clusters');
export const taskRetry = request('put', '/task/$taskId/retry');
export const taskSkip = request('put', '/task/$taskId/skip');
export const cloudDetail = request('get', '/cloud/$cloudId');
export const cloudNodes = request('get', '/clouds/$cloudId/instances');
export const cloudKeyPairs = request('get', '/clouds/$cloudId/keypairs');
export const cloudAccountType = request('get', '/clouds/$cloudId/accounttype');
export const cloudBwps = request('get', '/clouds/$cloudId/bwps');
export const cloudVersionModules = request('get', '/clouds/$cloudId/versions/$version/modules/$module');// 查询云组件参数
export const cloudConnect = request('get', '/clouds/$cloudId/clusters/$clusterID/connect');
export const cloudProjects = request('get', '/clouds/$cloudId/projects');
export const nodemanCloud = request('get', '/nodeman/cloud');
export const cloudOsImage = request('get', '/clouds/$cloudId/osimage');
export const cloudVPC = request('get', '/clouds/$cloudId/vpcs');
export const cloudSecurityGroups = request('get', '/clouds/$cloudId/securitygroups');
export const cloudSubnets = request('get', '/clouds/$cloudId/subnets');
export const cloudInstanceTypes = request('get', '/clouds/$cloudId/instancetypes');
export const cloudInstanceTypesByLevel = request('get', '/clouds/$cloudId/regions/$region/clusterlevels/$level/instancetypes');
export const cloudCidrconflict = request('get', '/clouds/$cloudId/vpcs/$vpc/cidrconflict');
export const addSubnets = request('post', '/clusters/$clusterId/subnets');

// node 操作
export const getK8sNodes = request('get', '/cluster/$clusterId/node');
export const uncordonNodes = request('put', '/node/uncordon');
export const cordonNodes = request('put', '/node/cordon');
export const schedulerNode = request('post', '/node/drain');
export const setNodeLabels = request('put', '/node/labels');
export const setNodeTaints = request('put', '/node/taints');

// 集群管理
export const masterList = request('get', '/cluster/$clusterId/master');

// CA
export const clusterAutoScalingLogsV2 = request('get', '/operationlogs');
export const cloudsZones = request('get', '/clouds/$cloudId/zones');
export const updateClusterAutoScalingProviders = request('put', '/autoscalingoption/$clusterId/providers/$provider');

// vCluster
export const sharedclusters = request('get', '/sharedclusters');
export const deleteVCluster = request('delete', '/vcluster/$clusterId');
export const createVCluster = request('post', '/vcluster');

// nodegroup
export const nodeGroups = request('get', '/clusters/$clusterId/nodegroups');
export const desirednode = request('post', '/nodegroup/$id/desirednode');

export const batchDeleteNodes = request('delete', '/clusters/$clusterId/nodes/-/batch');

// ip selector（ip选择器）
export const customSettings = request('post', '/web/customSettings/projects/$projectId/scope/$scope/$biz/batchGet');
export const topologyHostCount = request('post', '/web/projects/$projectId/scope/$scope/$biz/topology/hostCount');
export const hostCheck = request('post', '/web/projects/$projectId/scope/$scope/$biz/host/check');
export const topologyHostsNodes = request('post', '/web/projects/$projectId/scope/$scope/$biz/topology/hosts/nodes');
export const topologyHostIdList = request('post', '/web/projects/$projectId/scope/$scope/$biz/topology/hostids/nodes');
export const hostInfoByHostId = request('post', '/web/projects/$projectId/scope/$scope/$biz/hosts/details');

export const setClusterModule = request('put', '/clusters/$clusterId/module');
export const ccTopology = request('get', '/cluster/$clusterId/cc/topology');
