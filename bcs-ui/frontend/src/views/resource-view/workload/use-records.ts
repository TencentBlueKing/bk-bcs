import {
  gameWorkloadHistory as gameWorkloadHistoryAPI,
  revisionDetail as revisionDetailAPI,
  revisionGameDetail as revisionGameDetailAPI,
  rollbackGameWorkload as rollbackGameWorkloadAPI,
  rollbackWorkload as rollbackWorkloadAPI,
  workloadHistory as workloadHistoryAPI } from '@/api/modules/cluster-resource';

// type Category = 'deployments'|'statefulsets'|'daemonsets';

export interface IRevisionData {
  age: string
  createTime: string
  editMode: 'yaml'|'form'
  images: string[]
  revision: string
  updater: string
  resources: any
}

export default function useRecords() {
  const revisionDetail = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $revision: string|number
    $clusterId: string
  }) => {
    const data: {
      current_revision: string
      rollout_revision: string
    } = await revisionDetailAPI(params).catch(() => ({
      current_revision: '',
      rollout_revision: '',
    }));
    return data;
  };
  const revisionGameDetail = async (params: {
    $crd: string
    $category: string
    $name: string
    $revision: string|number
    $clusterId: string
    namespace: string
  }) => {
    const data: {
      current_revision: string
      rollout_revision: string
    } = await revisionGameDetailAPI(params).catch(() => ({
      current_revision: '',
      rollout_revision: '',
    }));
    return data;
  };
  const rollbackWorkload = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $revision: string|number
    $clusterId: string
  }) => {
    const data = await rollbackWorkloadAPI(params).then(() => true)
      .catch(() => false);
    return data;
  };
  const rollbackGameWorkload = async (params: {
    $crd: string
    $category: string
    $name: string
    $revision: string|number
    $clusterId: string
    namespace: string
  }) => {
    const data = await rollbackGameWorkloadAPI(params).then(() => true)
      .catch(() => false);
    return data;
  };
  const workloadHistory = async (params: {
    $category: string
    $namespaceId: string
    $name: string
    $clusterId: string
  }) => {
    const data = await workloadHistoryAPI(params).catch(() => []);
    return data as IRevisionData[];
  };
  const gameWorkloadHistory = async (params: {
    $crd: string
    $category: string
    $name: string
    $clusterId: string
    namespace: string
  }) => {
    const data = await gameWorkloadHistoryAPI(params).catch(() => []);
    return data as IRevisionData[];
  };

  return {
    revisionDetail,
    revisionGameDetail,
    rollbackWorkload,
    rollbackGameWorkload,
    workloadHistory,
    gameWorkloadHistory,
  };
}
