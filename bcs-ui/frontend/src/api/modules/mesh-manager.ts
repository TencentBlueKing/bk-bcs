import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/meshmanager/v1/mesh',
});

export const meshList = request('get', '/istio/list');
export const meshConfig = request('get', '/istio/config');
export const meshCreate = request('post', '/istio/install');
export const meshDelete = request('delete', '/istio/$meshID');
export const meshUpdate = request('put', '/istio/$meshID');
export const meshDetail = request('get', '/istio/detail/$meshID');
export const meshClusters = request('get', '/clusters');
