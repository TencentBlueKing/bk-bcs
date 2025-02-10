import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/clusterresources/api/v1',
});

export const storageEvents = request('get', '/projects/$projectCode/clusters/$clusterID/events');
