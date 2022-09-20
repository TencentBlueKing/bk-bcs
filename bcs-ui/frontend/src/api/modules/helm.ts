import { createRequest } from '../request';

// 基础性能数据，标准输出日志
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/helmmanager/v1/projects/$projectCode/clusters/$clusterId',
});

// helm
export const helmReleaseHistory = request('get', '/namespaces/$namespaceId/releases/$name/history');

// cluster tools
const request2 = createRequest({
  domain: window.DEVOPS_BCS_API_URL,
  prefix: '/api/cluster_tools/projects/$projectId/clusters/$clusterId/tools',
});

// 组件库
export const clusterTools = request2('get', '/');
export const clusterToolsInstall = request2('post', '/$toolId/');
export const clusterToolsUpgrade = request2('put', '/$toolId/');
export const clusterToolsUninstall = request2('delete', '/$toolId/');
export const clusterToolsInstalledDetail = request2('get', '/$toolId/installed_detail/');
