import { createRequest } from '../request';

import { BCS_UI_PREFIX } from '@/common/constant';

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
export const syncNamespaceList = request('post', '/clusters/$clusterId/namespaces/sync');
export const withdrawNamespace = request('post', '/clusters/$clusterId/namespaces/$namespace/withdraw');

const projectRequest = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/bcsproject/v1',
});
// project
export const createProject = projectRequest('post', '/projects');
export const getProject = projectRequest('get', '/projects/$projectId');
export const editProject = projectRequest('put', '/projects/$projectId');
export const fetchProjectList = projectRequest('get', '/authorized_projects');
export const fetchAllProjectList = projectRequest('get', '/projects');
export const businessList = projectRequest('get', '/business');
export const projectBusiness = projectRequest('get', '/projects/$projectCode/business');

const uiRequest = createRequest({
  domain: window.BCS_API_HOST,
  prefix: BCS_UI_PREFIX,
});
export const releaseNote = uiRequest('get', '/release_note');
export const featureFlags = uiRequest('get', '/feature_flags');

// AI小鲸
export const assistant = uiRequest('post', '/assistant');
