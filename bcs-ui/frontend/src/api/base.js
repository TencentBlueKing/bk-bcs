/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
import { request } from './request';
// todo 当前文件要废弃，请使用modules下面的API定义文件 !!!
// log
export const LOG_API_URL = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4/monitor/api/projects/$projectId/clusters/$clusterId`;
export const podContainersList = request('get', `${LOG_API_URL}/namespaces/$namespaceId/pods/$podId/containers`);
export const podLogs = request('get', `${LOG_API_URL}/namespaces/$namespaceId/pods/$podId/logs`);
export const podLogsStreamURL = `${LOG_API_URL}/namespaces/$namespaceId/pods/$podId/logs/stream?container_name=$containerName&started_at=$startedAt`;

// dashbord
export const podMetric = request('post', '/api/metrics/projects/$projectId/clusters/$clusterId/pods/$metric/');
export const containerMetric = request('post', '/api/metrics/projects/$projectId/clusters/$clusterId/pods/$podId/containers/$metric/');

// cluster resource
// todo
export const crPrefix = '/bcsapi/v4/clusterresources/v1';
export const CR_API_URL = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}${crPrefix}`;
export const namespaceList = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces`);
export const dashbordList = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/$type/$category`);// 注意：HPA类型没有子分类$category
export const formSchema = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/form_schema`);
export const dashbordListWithoutNamespace = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/$type/$category`); // PersistentVolume, StorageClass资源暂不支持命名空间
export const retrieveDetail = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/$type/$category/$name`);
export const retrieveCustomObjectDetail = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crdName/custom_objects/$name?namespace=$namespaceId`);
export const retrieveContainerDetail = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/$category/$name/containers/$containerName`);
export const listWorkloadPods = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/pods`);
export const listStoragePods = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/pods/$podId/$type`);
export const listContainers = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/pods/$podId/containers`);
export const fetchContainerEnvInfo = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/pods/$podId/containers/$containerName/env_info`);
export const resourceCreate = request('post', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/$type/$category`);
export const resourceUpdate = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/$type/$category/$name`);
export const resourceDelete = request('delete', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/$type/$category/$name`);
export const exampleManifests = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/examples/manifests`);
export const crdList = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds`);// 获取CRD列表
export const customResourceList = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crd/$category`); // 自定义资源
export const retrieveCustomResourceDetail = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crd/$category/$name`); // 自定义资源详情
export const customResourceCreate = request('post', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crd/$category`); // 自定义资源创建
export const customResourceUpdate = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crd/$category/$name`); // 自定义资源更新
export const customResourceDelete = request('delete', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crd/$category/$name`); // 自定义资源删除
export const reschedulePod = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespaceId/workloads/pods/$podId/reschedule`); // pod重新调度
export const renderManifestPreview = request('post', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/render_manifest_preview`);
export const fetchNodePodsData = request('get', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/nodes/$nodename/workloads/pods`);
export const enlargeCapacityChange = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/$category/$name/scale`); // 扩缩容
export const batchReschedulePod = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/namespaces/$namespace/workloads/$category/$name/reschedule`); // pod批量重新调度
export const crdEnlargeCapacityChange = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crdName/custom_objects/$cobjName/scale`); // crd扩缩容
export const batchRescheduleCrdPod = request('put', `${CR_API_URL}/projects/$projectId/clusters/$clusterId/crds/$crdName/custom_objects/$cobjName/reschedule`); // crd-pod批量重新调度

// node
export const fetchBizTopo = request('get', '/api/projects/$projectId/cc/topology/');
export const fetchBizHosts = request('post', '/api/projects/$projectId/cc/hosts/');

// project
export const createProject = request('post', '/api/nav/projects/');
export const editProject = request('put', '/api/nav/projects/$projectId/');
export const logLinks = request('post', '/api/datalog/projects/$projectId/log_links/');

// Cluster Manager
const prefix = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4`;
export const cloudList = request('get', `${prefix}/clustermanager/v1/cloud`);
export const createCluster = request('post', `${prefix}/clustermanager/v1/cluster`);
export const cloudVpc = request('get', `${prefix}/clustermanager/v1/cloudvpc`);
export const cloudRegion = request('get', `${prefix}/clustermanager/v1/cloudregion/$cloudId`);
export const vpccidrList = request('get', `${prefix}/clustermanager/v1/vpccidr/$vpcID`);
export const deleteCluster = request('delete', `${prefix}/clustermanager/v1/cluster/$clusterId`);
export const retryCluster = request('post', `${prefix}/clustermanager/v1/cluster/$clusterId/retry`);
export const taskList = request('get', `${prefix}/clustermanager/v1/task`);
export const taskDetail = request('get', `${prefix}/clustermanager/v1/task/$taskId`);
export const clusterNode = request('get', `${prefix}/clustermanager/v1/cluster/$clusterId/node`);
export const addClusterNode = request('post', `${prefix}/clustermanager/v1/cluster/$clusterId/node`);
export const deleteClusterNode = request('delete', `${prefix}/clustermanager/v1/cluster/$clusterId/node`);
export const clusterDetail = request('get', `${prefix}/clustermanager/v1/cluster/$clusterId`);
export const modifyCluster = request('put', `${prefix}/clustermanager/v1/cluster/$clusterId`);
export const importCluster = request('post', `${prefix}/clustermanager/v1/cluster/import`);
export const kubeConfig = request('put', `${prefix}/clustermanager/v1/cloud/kubeConfig`);
export const nodeAvailable = request('post', `${prefix}/clustermanager/v1/node/available`);
export const cloudAccounts = request('get', `${prefix}/clustermanager/v1/clouds/$cloudId/accounts`);
export const createCloudAccounts = request('post', `${prefix}/clustermanager/v1/clouds/$cloudId/accounts`);
export const deleteCloudAccounts = request('delete', `${prefix}/clustermanager/v1/clouds/$cloudId/accounts/$accountID`);
export const cloudResourceGroupByAccount = request('get', `${prefix}/clustermanager/v1/clouds/$cloudId/resourcegroups`);
export const cloudRegionByAccount = request('get', `${prefix}/clustermanager/v1/clouds/$cloudId/regions`);
export const cloudClusterList = request('get', `${prefix}/clustermanager/v1/clouds/$cloudId/clusters`);
export const taskRetry = request('put', `${prefix}/clustermanager/v1/task/$taskId/retry`);
export const nodemanCloudList = request('get', `${prefix}/clustermanager/v1/nodeman/cloud`);
export const ccTopology = request('get', `${prefix}/clustermanager/v1/cluster/$clusterId/cc/topology`);
// token
export const createToken = request('post', `${prefix}/usermanager/v1/tokens`);
export const updateToken = request('put', `${prefix}/usermanager/v1/tokens/$token`);
export const deleteToken = request('delete', `${prefix}/usermanager/v1/tokens/$token`);
export const getTokens = request('get', `${prefix}/usermanager/v1/users/$username/tokens`);

// cluster tools
export const clusterTools = request('get', '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools/');
export const clusterToolsInstall = request('post', '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools/$toolId/');
export const clusterToolsUpgrade = request('put', '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools/$toolId/');
export const clusterToolsUninstall = request('delete', '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools/$toolId/');
export const clusterToolsInstalledDetail = request('get', '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools/$toolId/installed_detail/');

// nodetemplate
export const nodeTemplateList = request('get', `${prefix}/clustermanager/v1/projects/$projectId/nodetemplates`);
export const createNodeTemplate = request('post', `${prefix}/clustermanager/v1/projects/$projectId/nodetemplates`);
export const deleteNodeTemplate = request('delete', `${prefix}/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId`);
export const updateNodeTemplate = request('put', `${prefix}/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId`);
export const nodeTemplateDetail = request('get', `${prefix}/clustermanager/v1/projects/$projectId/nodetemplates/$nodeTemplateId`);
export const bkSopsList = request('get', `${prefix}/clustermanager/v1/bksops/business/$businessID/templates`);
export const bkSopsParamsList = request('get', `${prefix}/clustermanager/v1/bksops/business/$businessID/templates/$templateID`);
export const cloudModulesParamsList = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/versions/$version/modules/$moduleID`);
export const bkSopsDebug = request('post', `${prefix}/clustermanager/v1/bksops/debug`);
export const bkSopsTemplatevalues = request('get', `${prefix}/clustermanager/v1/bksops/templatevalues`);
export const getNodeTemplateInfo = request('get', `${prefix}/clustermanager/v1/node/$innerIP/info`);

// helm
const helmPrefix = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4/helmmanager/v1/projects/$projectCode/clusters/$clusterId`;
export const helmReleaseHistory = request('get', `${helmPrefix}/namespaces/$namespaceId/releases/$name/history`);

// metric
const metricPrefix = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4/monitor/api/metrics/projects/$projectCode/clusters/$clusterId`;
export const clusterCpuUsage = request('get', `${metricPrefix}/cpu_usage`);
export const clusterDiskUsage = request('get', `${metricPrefix}/disk_usage`);
export const clusterMemoryUsage = request('get', `${metricPrefix}/memory_usage`);
export const clusterOverview = request('get', `${metricPrefix}/overview`);
export const clusterNodeCpuUsage = request('get', `${metricPrefix}/nodes/$nodeIP/cpu_usage`);
export const clusterNodeDiskIOUsage = request('get', `${metricPrefix}/nodes/$nodeIP/diskio_usage`);
export const clusterNodeInfo = request('get', `${metricPrefix}/nodes/$nodeIP/info`);
export const clusterNodeMemoryUsage = request('get', `${metricPrefix}/nodes/$nodeIP/memory_usage`);
export const clusterNodeNetworkReceive = request('get', `${metricPrefix}/nodes/$nodeIP/network_receive`);
export const clusterNodeNetworkTransmit = request('get', `${metricPrefix}/nodes/$nodeIP/network_transmit`);
export const clusterNodeOverview = request('get', `${metricPrefix}/nodes/$nodeIP/overview`);
export const clusterPodMetric = request('post', `${metricPrefix}/namespaces/$namespaceId/pods/$metric`);
export const clusterContainersMetric = request('get', `${metricPrefix}/namespaces/$namespaceId/pods/$podId/containers/$containerId/$metric`);

// variable
const variablePrefix = `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4/bcsproject/v1/projects/$projectCode`;
export const createVariable = request('post', `${variablePrefix}/variables`);
export const variableDefinitions = request('get', `${variablePrefix}/variables`);
export const deleteDefinitions = request('delete', `${variablePrefix}/variables`);
export const updateVariable = request('put', `${variablePrefix}/variables/$variableID`);
export const importVariable = request('post', `${variablePrefix}/variables/import`);
export const clusterVariable = request('get', `${variablePrefix}/variables/$variableID/cluster`);
export const updateClusterVariable = request('put', `${variablePrefix}/variables/$variableID/cluster`);
export const namespaceVariable = request('get', `${variablePrefix}/variables/$variableID/namespace`);
export const updateNamespaceVariable = request('put', `${variablePrefix}/variables/$variableID/namespace`);

// log
export const createLogCollect = request('post', '/api/log_collect/projects/$projectId/clusters/$clusterId/configs/');
export const logCollectList = request('get', '/api/log_collect/projects/$projectId/clusters/$clusterId/configs/');
export const updateLogCollect = request('put', '/api/log_collect/projects/$projectId/clusters/$clusterId/configs/$configId/');
export const deleteLogCollect = request('delete', '/api/log_collect/projects/$projectId/clusters/$clusterId/configs/$configId/');
export const retrieveLogCollect = request('get', '/api/log_collect/projects/$projectId/clusters/$clusterId/configs/$configId/');

// node group(pool)
export const nodeGroup = request('get', `${prefix}/clustermanager/v1/nodegroup`);
export const createNodeGroup = request('post', `${prefix}/clustermanager/v1/nodegroup`);
export const nodeGroupDetail = request('get', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID`);
export const updateNodeGroup = request('put', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID`);
export const deleteNodeGroup = request('delete', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID`);
export const disableNodeGroupAutoScale = request('post', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID/autoscale/disable`);
export const enableNodeGroupAutoScale = request('post', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID/autoscale/enable`);
export const nodeGroupNodeList = request('get', `${prefix}/clustermanager/v2/nodegroup/$nodeGroupID/node`);
export const deleteNodeGroupNode = request('delete', `${prefix}/clustermanager/v2/nodegroup/$nodeGroupID/groupnode`);
export const addNodeGroupNode = request('post', `${prefix}/clustermanager/v1/nodegroup/$nodeGroupID/node`);
export const resourceSchema = request('get', `${prefix}/clustermanager/v1/resourceschema/$cloudID/$name`);
export const cloudOsImage = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/osimage`);
export const cloudNoderoles = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/serviceroles`);
export const cloudInstanceTypes = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/instancetypes`);
export const cloudSecurityGroups = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/securitygroups`);
export const cloudSubnets = request('get', `${prefix}/clustermanager/v1/clouds/$cloudID/subnets`);
export const clusterAutoScaling = request('get', `${prefix}/clustermanager/v1/autoscalingoption/$clusterId`);
export const updateClusterAutoScaling = request('put', `${prefix}/clustermanager/v1/autoscalingoption/$clusterId`);
export const toggleClusterAutoScalingStatus = request('put', `${prefix}/clustermanager/v1/autoscalingoption/$clusterId/status `);
export const clusterAutoScalingLogs = request('get', `${prefix}/clustermanager/v1/operationlogs`);
export const clusterNodeDrain = request('post', `${prefix}/clustermanager/v1/node/drain`);
export const nodeCordon = request('put', `${prefix}/clustermanager/v1/node/cordon`);
export const nodeUnCordon = request('put', `${prefix}/clustermanager/v1/node/uncordon`);

export default {
  dashbordList,
  podMetric,
  containerMetric,
  retrieveDetail,
  listWorkloadPods,
  listStoragePods,
  listContainers,
  retrieveContainerDetail,
  fetchContainerEnvInfo,
  resourceCreate,
  resourceUpdate,
  resourceDelete,
  exampleManifests,
  namespaceList,
  createProject,
  crdList,
  customResourceList,
  retrieveCustomResourceDetail,
  customResourceCreate,
  customResourceUpdate,
  customResourceDelete,
  reschedulePod,
  logLinks,
  editProject,
  fetchBizTopo,
  fetchBizHosts,
};
