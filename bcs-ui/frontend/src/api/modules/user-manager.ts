import { createRequest } from '../request';

const iamRequest = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/usermanager/v1/iam',
});
// auth
export const userPerms = iamRequest('post', '/user_perms');
export const userPermsByAction = iamRequest('post', '/user_perms/actions/$actionId');

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

// 操作审计
const v3Request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/usermanager/v3',
});
export const activityLogsResourceTypes = v3Request('get', '/activity_logs/resource_types');
export const activityLogs = v3Request('get', '/projects/$projectCode/activity_logs');
