import {
  revisionDetail as revisionDetailAPI,
  rollbackWorkload as rollbackWorkloadAPI,
  workloadHistory as workloadHistoryAPI,
} from '@/api/modules/cluster-resource';

// type Category = 'deployments'|'statefulsets'|'daemonsets';

export default function useRecords() {
  const revisionDetail = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $revision: string
    $clusterId: string
  }) => {
    const data = await revisionDetailAPI(params);
    return data;
  };
  const rollbackWorkload = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $revision: string
    $clusterId: string
  }) => {
    const data = await rollbackWorkloadAPI(params);
    return data;
  };
  const workloadHistory = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $clusterId: string
  }) => {
    const data = await workloadHistoryAPI(params);
    return data;
  };

  return {
    revisionDetail,
    rollbackWorkload,
    workloadHistory,
  };
}
