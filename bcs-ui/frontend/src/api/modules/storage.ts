import { createRequest } from '../request';

// 集群管理，节点管理
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/storage',
});

export const storageEvents = request('post', '/events');

// todo 临时方案
const request2 = createRequest({
  domain: window.BCS_DEBUG_API_HOST,
  prefix: '/bcsapi/v4/storage',
});

export const uatStorageEvents = request2('post', '/events');
