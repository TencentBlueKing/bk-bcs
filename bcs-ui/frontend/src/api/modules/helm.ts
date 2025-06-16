import { createRequest } from '../request';

// 基础性能数据，标准输出日志
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/helmmanager/v1/projects/$projectCode',
});

// helm
export const reposList = request('get', '/repos');
export const createRepo = request('post', '/repos');
export const repoCharts = request('get', '/repos/$repoName/charts');
export const deleteRepoChart = request('delete', '/repos/$repoName/charts/$chartName');
export const repoChartVersions = request('get', '/repos/$repoName/charts/$chartName/versions');
export const repoChartVersionDetail = request('get', '/repos/$repoName/charts/$chartName/versions/$version');
export const deleteRepoChartVersion = request('delete', '/repos/$repoName/charts/$chartName/versions/$version');
export const releaseDetail = request('get', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName');
export const deleteRelease = request('delete', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName');
export const releaseChart = request('post', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName');
export const updateRelease = request('put', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName');
export const releaseHistory = request('get', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName/history');
export const previewRelease = request('post', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName/preview');
export const rollbackRelease = request('put', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName/rollback');
export const releaseStatus = request('get', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName/status');
export const releasesList = request('get', '/clusters/$clusterId/releases');
export const chartDetail = request('get', '/repos/$repoName/charts/$chartName');
export const downloadChartUrl = `${window.BCS_API_HOST}/bcsapi/v4/helmmanager/v1/projects/$projectCode/repos/$repoName/charts/$chartName/versions/$version/download`;
export const chartReleases = request('post', '/repos/$repoName/charts/$chartName/releases');
export const releasesManifest = request('get', '/clusters/$clusterId/namespaces/$namespaceId/releases/$releaseName/revisions/$revision/manifest');
// 日志采集 & 组件库
export const addonsDetail = request('get', '/clusters/$clusterId/addons/$name'); // 获取组件详情
export const updateOns = request('put', '/clusters/$clusterId/addons/$name'); // 更新组件
export const addonsList = request('get', '/clusters/$clusterId/addons');
export const addonsInstall = request('post', '/clusters/$clusterId/addons');
export const addonsUninstall = request('delete', '/clusters/$clusterId/addons/$name');
export const addonsStop = request('put', '/clusters/$clusterId/addons/$name/stop');
export const addonsPreview = request('post', '/clusters/$clusterId/addons/$name/preview');

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
