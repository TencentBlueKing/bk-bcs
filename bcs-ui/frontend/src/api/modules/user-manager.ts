import { createRequest } from '../request';

const request = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '',
});
// auth
export const userPerms = request('post', '/api/iam/user_perms/');
export const userPermsByAction = request('post', '/api/iam/user_perms/actions/$actionId/');

const request2 = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '/api/cluster_manager/proxy/bcsapi/v4',
});
// token
export const createToken = request2('post', '/usermanager/v1/tokens');
export const updateToken = request2('put', '/usermanager/v1/tokens/$token');
export const deleteToken = request2('delete', '/usermanager/v1/tokens/$token');
export const getTokens = request2('get', '/usermanager/v1/users/$username/tokens');

const userRequest = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/usermanager/v1',
});
// user
export const userInfo = userRequest('get', '/users/info');
