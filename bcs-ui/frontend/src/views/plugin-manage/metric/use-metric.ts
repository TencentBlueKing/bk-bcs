import {
  getServiceMonitor,
  getServiceMonitorDetail,
  createServiceMonitor,
  updateServiceMonitor,
  deleteServiceMonitor,
  batchDeleteServiceMonitor,
} from '@/api/modules/monitor';
import { dashbordList } from '@/api/modules/cluster-resource';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';
import { ref } from 'vue';

export interface IMetricData {
  'service_name': string
  'path': string // 路径
  'selector': Record<string, string> // 关联label
  'interval': string // 采集周期
  'port': string // portName
  'sample_limit': Number// 允许最大Sample数
  'name': string // 名称
  'params': Record<string, string> // 参数
}

export default function useMetric() {
  const handleGetServiceMonitor = async (clusterID: string) => {
    const data = await getServiceMonitor({
      $clusterId: clusterID,
    }).catch(() => []);
    return data;
  };
  const handleGetServiceMonitorDetail = async (params: {
    $clusterId: string
    $namespaceId: string
    $name: string
  }) => {
    const data = await getServiceMonitorDetail(params).catch(() => ({}));
    return data;
  };
  const handleCreateServiceMonitor = async (params: IMetricData & {
    $namespaceId: string
    $clusterId: string
  }) => {
    const result = await createServiceMonitor(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('创建成功'),
    });
    return result;
  };
  const handleUpdateServiceMonitor = async (params: IMetricData & {
    $namespaceId: string
    $name: string
    $clusterId: string
  }) => {
    const result = await updateServiceMonitor(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('修改成功'),
    });
    return result;
  };
  const handleDeleteServiceMonitor = async (params: {
    $namespaceId: string
    $name: string
    $clusterId: string
  }) => {
    const result = await deleteServiceMonitor(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('删除成功'),
    });
    return result;
  };
  const handleBatchDeleteServiceMonitor = async (params: {
    $clusterId: string
    service_monitors: Array<{
      namespace: string
      name: string
    }>
  }) => {
    const result = await batchDeleteServiceMonitor(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('删除成功'),
    });
    return result;
  };
  const serviceList = ref<any[]>([]);
  const serviceLoading = ref(false);
  const handleGetServiceList = async ($namespaceId: string, $clusterId: string) => {
    if (!$namespaceId || !$clusterId) return;
    serviceLoading.value = true;
    const data = await dashbordList({
      $namespaceId,
      $clusterId,
      $type: 'networks',
      $category: 'services',
    }).catch(() => ({ manifest: { items: [] } }));
    serviceList.value = data?.manifest?.items || [];
    serviceLoading.value = false;
    return data;
  };

  return {
    handleGetServiceMonitor,
    handleGetServiceMonitorDetail,
    handleCreateServiceMonitor,
    handleUpdateServiceMonitor,
    handleDeleteServiceMonitor,
    handleBatchDeleteServiceMonitor,
    serviceLoading,
    serviceList,
    handleGetServiceList,
  };
}
