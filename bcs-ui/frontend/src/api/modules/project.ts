import { createRequest } from '../request';

// 项目管理，变量变量，命名空间, Quota 管理，权限
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/bcsproject/v1/projects/$projectCode',
});

// variable 变量管理
export const createVariable = request('post', '/variable');
export const variableDefinitions = request('get', '/variable/definitions');
export const deleteDefinitions = request('delete', '/variable/definitions');
export const updateVariable = request('put', '/variable/$variableID');
export const importVariable = request('post', '/variables/import');
export const clusterVariable = request('get', '/variables/$variableID/cluster');
export const updateClusterVariable = request('put', '/variables/$variableID/cluster');
export const namespaceVariable = request('get', '/variables/$variableID/namespace');
export const updateNamespaceVariable = request('put', '/variables/$variableID/namespace');
export const getClusterVariables = request('get', '/clusters/$clusterId/variables');
export const updateSpecifyClusterVariables = request('put', '/clusters/$clusterId/variables');
export const getClusterNamespaceVariable = request('get', '/clusters/$clusterId/namespaces/$namespace/variables');
export const updateClusterNamespaceVariable = request('put', '/clusters/$clusterId/namespaces/$namespace/variables');

// 命名空间
export const getNamespaceList = request('get', '/clusters/$clusterId/namespaces');
export const deleteNamespace = request('delete', '/clusters/$clusterId/namespaces/$namespace');
export const updateNamespace = request('put', '/clusters/$clusterId/namespaces/$namespace');
export const createdNamespace = request('post', '/clusters/$clusterId/namespaces');
export const fetchNamespaceInfo = request('get', '/clusters/$clusterId/namespaces/$name');

const request2 = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '',
});
// project
export const createProject = request2('post', '/api/nav/projects/');
export const editProject = request2('put', '/api/nav/projects/$projectId/');
export const logLinks = request2('post', '/api/datalog/projects/$projectId/log_links/');
export const getProjectList = request2('get', '/api/nav/projects/');
export const fetchBizTopo = request2('get', '/api/projects/$projectId/cc/topology/');
export const fetchBizHosts = request2('post', '/api/projects/$projectId/cc/hosts/');
export const getBizMaintainers = request2('get', '/api/projects/$projectId/biz_maintainers/');
export const projectFeatureFlag = request2('get', '/api/projects/$projectId/clusters/$clusterId/feature_flags/');
