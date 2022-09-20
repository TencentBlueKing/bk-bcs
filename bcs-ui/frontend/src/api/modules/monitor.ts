import { createRequest } from '../request';

// 基础性能数据，标准输出日志
const request = createRequest({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/monitor/api/metrics/projects/$projectCode/clusters/$clusterId',
});

// metric 指标接口
export const clusterCpuUsage = request('get', '/cpu_usage');
export const clusterDiskUsage = request('get', '/disk_usage');
export const clusterMemoryUsage = request('get', '/memory_usage');
export const clusterOverview = request('get', '/overview');
export const clusterNodeCpuUsage = request('get', '/nodes/$nodeIP/cpu_usage');
export const clusterNodeDiskIOUsage = request('get', '/nodes/$nodeIP/diskio_usage');
export const clusterNodeInfo = request('get', '/nodes/$nodeIP/info');
export const clusterNodeMemoryUsage = request('get', '/nodes/$nodeIP/memory_usage');
export const clusterNodeNetworkReceive = request('get', '/nodes/$nodeIP/network_receive');
export const clusterNodeNetworkTransmit = request('get', '/nodes/$nodeIP/network_transmit');
export const clusterNodeOverview = request('get', '/nodes/$nodeIP/overview');
export const clusterPodMetric = request('post', '/namespaces/$namespaceId/pods/$metric');
export const clusterContainersMetric = request('get', '/namespaces/$namespaceId/pods/$podId/containers/$containerId/$metric');

const LOG_API_URL = '/bcsapi/v4/monitor/api/projects/$projectId/clusters/$clusterId';
const request2 = createRequest({
  domain: window.BCS_API_HOST,
  prefix: LOG_API_URL,
});
// 日志
export const podContainersList = request2('get', '/namespaces/$namespaceId/pods/$podId/containers');
export const podLogs = request2('get', '/namespaces/$namespaceId/pods/$podId/logs');
export const podLogsDownloadURL = `${LOG_API_URL}/namespaces/$namespaceId/pods/$podId/logs/download?container_name=$containerName`;
export const podLogsStreamURL = `${LOG_API_URL}/namespaces/$namespaceId/pods/$podId/logs/stream?container_name=$containerName&started_at=$startedAt`;
