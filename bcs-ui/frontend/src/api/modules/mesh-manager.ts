import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/meshmanager/v1/mesh/istio',
});

export const meshList = request('get', '/list');
export const meshConfig = request('get', '/config');
export const meshCreate = request('post', '/install');
export const meshDelete = request('delete', '/$meshID');
export const meshUpdate = request('put', '/$meshID');
export const meshDetail = request('get', '/detail/$meshID');
