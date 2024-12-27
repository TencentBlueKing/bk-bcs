/* eslint-disable camelcase */
import { computed, reactive, Ref, ref, set, toRef, watch } from 'vue';

import {
  cloudAccountType,
  cloudBwps,
  cloudNodes,
  clusterDetail,
  createVCluster,
  deleteVCluster,
  getFederalTask,
  retryFederalTask,
  sharedclusters,
  taskSkip,
} from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { ICluster } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

/**
 * 获取集群列表
 * @param ctx
 * @returns
 */
export const isAdding = ref(false); // 异步添加轮询标识
export function useClusterList() {
  const clusterList = computed<ICluster[]>(() => $store.state.cluster.clusterList);
  const curProjectId = computed(() => $store.getters.curProjectId);
  const clusterExtraInfo = ref({});
  const webAnnotations = ref({ perms: {} });
  // const clusterCurrentTaskDataMap = ref({});

  // 联邦集群
  const federationClusters = computed(() => clusterList.value.filter(item => item.clusterType === 'federation'));

  // const { taskList } = useTask();
  // 获取集群列表
  const getClusterList = async () => {
    const res = await $store.dispatch('cluster/getClusterList', curProjectId.value); // 会自动更新clusterList缓存
    clusterExtraInfo.value = res.clusterExtraInfo || {};
    webAnnotations.value = res.web_annotations || { perms: {} };

    // 获取当前运行中集群的任务详情
    // clusterList.value
    //   .forEach((item) => {
    //     if (['INITIALIZATION', 'DELETING'].includes(item.status)) {
    //       taskList(item).then(({ latestTask }) => {
    //         const steps = latestTask?.stepSequence || [];
    //         const task = steps.map(step => latestTask?.steps[step]).find(step => step.status === 'RUNNING');
    //         clusterCurrentTaskDataMap.value[item.clusterID] = task;
    //       });
    //     } else {
    //       clusterCurrentTaskDataMap.value[item.clusterID] = null;
    //     }
    //   });
  };
  // 开启轮询
  const { start, stop } = useInterval(getClusterList, 5000);
  const runningClusterIds = computed(() => clusterList.value.filter(item => [
    'INITIALIZATION',
    'DELETING'].includes(item.status)).map(item => item.cluster_id));
  watch(runningClusterIds, (newValue) => {
    if (!newValue.length && !isAdding.value) {
      stop();
    } else {
      start();
      if (newValue.length && isAdding.value) {
        isAdding.value = false;
      }
    }
  });

  return {
    webAnnotations,
    clusterList,
    curProjectId,
    clusterExtraInfo,
    federationClusters,
    getClusterList,
  };
}
export interface IOverviewMap {
  [key: string]: {
    cpu_usage: Record<string, string>;
    disk_usage: Record<string, string>;
    diskio_usage: Record<string, string>;
    memory_usage: Record<string, string>;
  };
}
/**
 * 获取集群指标数据
 * @param ctx
 * @param clusterList
 * @returns
 */
export function useClusterOverview(clusterList: Ref<any[]>) {
  const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

  const clusterOverviewMap = ref<IOverviewMap>({});
  const projectCode = computed(() => $route.value.params.projectCode);
  // 获取当前集群的指标信息
  const getClusterOverview = (clusterId) => {
    if (!clusterOverviewMap.value[clusterId]) return null;

    return clusterOverviewMap.value[clusterId];
  };
  // 获取集群指标项百分比
  const getMetricPercent = (data, metric) => {
    if (!data) return 0;

    let used = 0;
    let total = 0;
    if (metric === 'cpu_usage') {
      used = data?.[metric]?.used;
      total = data?.[metric]?.total;
    } else {
      used = data?.[metric]?.used_bytes;
      total = data?.[metric]?.total_bytes;
    }

    if (!Number(total)) {
      return 0;
    }
    let ret = Number(used) / Number(total) * 100;
    if (ret !== 0 && ret !== 100) {
      ret = Number(ret.toFixed(2));
    }

    return ret;
  };
  // 集群指标信息
  const fetchClusterOverview = async (cluster) => {
    if (!cluster.clusterID) return;
    const data = await $store.dispatch('metric/clusterOverview', {
      $projectCode: projectCode.value,
      $clusterId: cluster.clusterID,
    }).catch(() => ({ data: {} }));
    // 计算百分比
    Object.keys(data).forEach((metric) => {
      data[metric] = {
        ...data[metric],
        percent: getMetricPercent(data, metric),
      };
    });
    set(clusterOverviewMap.value, cluster.clusterID, data);
    return data;
  };

  watch(clusterList, (newValue) => {
    const newClusterList = newValue.filter(item => item.status === 'RUNNING' && !clusterOverviewMap.value?.[item.clusterID]);
    newClusterList.forEach((item) => {
      fetchClusterOverview(item);
    });
  });

  return {
    clusterOverviewMap,
    getClusterOverview,
  };
}
/**
 * 集群操作
 * @param ctx
 * @returns
 */
export function useClusterOperate() {
  const user = computed(() => $store.state.user);
  const projectId = computed(() => $store.getters.curProjectId);
  // 集群删除
  const deleteCluster = async (cluster): Promise<boolean> => {
    const result = await $store.dispatch('clustermanager/deleteCluster', {
      $clusterId: cluster.cluster_id,
    }).catch(() => false);
    return result;
  };
  // 集群重试重试
  const getTaskData = async (clusterId) => {
    const res = await $store.dispatch('clustermanager/taskList', {
      clusterID: clusterId,
      projectID: projectId.value,
    });
    const { latestTask } = res;
    const steps = latestTask?.stepSequence || [];
    const taskData = steps.map(step => latestTask?.steps[step]);
    return {
      taskData,
      latestTask,
    };
  };
  const retryClusterTask = async (clusterID: string): Promise<boolean> => {
    const { latestTask } = await getTaskData(clusterID);
    const result = await $store.dispatch('clustermanager/taskRetry', {
      $taskId: latestTask.taskID,
      updater: user.value.username,
    });
    return result;
  };
  const retryFederal = async (cluster): Promise<boolean> => {
    if (!cluster?.labels?.['federation.bkbcs.tencent.com/taskid']) return false;
    const result = await retryFederalTask({
      $taskId: cluster.labels['federation.bkbcs.tencent.com/taskid'],
    });
    return result;
  };
  // 获取集群下面所有的节点
  const clusterNode = async (cluster) => {
    const data = await $store.dispatch('clustermanager/clusterNode', {
      $clusterId: cluster.clusterID,
    });
    return data;
  };

  // 获取云上节点信息
  const getCloudNodes = async (params: {
    $cloudId: string
    region: string
    ipList: string
    accountID?: string
  }) => {
    const data = await cloudNodes(params).catch(() => []);
    return data;
  };

  return {
    deleteCluster,
    retryClusterTask,
    retryFederal,
    clusterNode,
    getCloudNodes,
  };
}

/**
 * 任务操作
 * @param ctx
 */
export function useTask() {
  const user = computed(() => $store.state.user);
  // 查询任务列表
  const taskList = async (cluster) => {
    // 统一数据格式
    let result;
    if (cluster?.labels?.['federation.bkbcs.tencent.com/taskid']) {
      const taskID = cluster.labels['federation.bkbcs.tencent.com/taskid'];
      const res = await getFederalTask({
        $taskId: taskID,
      }).catch(() => {});
      result = {
        data: [],
        latestTask: res,
      };
    } else {
      result = await $store.dispatch('clustermanager/taskList', {
        clusterID: cluster.clusterID,
        projectID: cluster.projectID,
      });
    }
    return result;
  };
  // 跳过任务
  const skipTask = async (taskID: string) => {
    const result = await taskSkip({
      $taskId: taskID,
      updater: user.value.username,
    }).then(() => true)
      .catch(() => false);
    return result;
  };
  return {
    taskList,
    skipTask,
  };
}

export function useClusterInfo() {
  const isLoading = ref(false);
  const clusterData = ref<ICluster>({} as unknown as any);
  const clusterOS = computed(() => clusterData.value?.clusterBasicSettings?.OS);
  const clusterAdvanceSettings = computed(() => clusterData.value?.clusterAdvanceSettings || {});
  const extraInfo = computed(() => clusterData.value?.extraInfo || {});
  const getClusterDetail = async ($clusterId: string, cloudInfo = false, loading = true) => {
    isLoading.value = loading;
    const res = await clusterDetail({
      $clusterId,
      cloudInfo,
    }, { needRes: true }).catch(() => ({}));
    clusterData.value = {
      ...res.data,
      providerType: res.extra?.providerType,
    };
    isLoading.value = false;
  };
  return {
    isLoading,
    clusterOS,
    getClusterDetail,
    clusterAdvanceSettings,
    clusterData,
    extraInfo,
  };
}

export function useVCluster() {
  const loading = ref(false);
  const sharedClusterList = ref<ICluster[]>([]);
  async function getSharedclusters() {
    loading.value = true;
    sharedClusterList.value = await sharedclusters({
      showVCluster: true,
    }).catch(() => []);
    loading.value = false;
  }
  async function handleCreateVCluster(params) {
    const result = await createVCluster(params).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.deliveryTask'),
    });
    return result;
  }

  async function handleDeleteVCluster(params) {
    const result = await deleteVCluster(params).then(() => true)
      .catch(() => false);
    return result;
  }

  return {
    loading,
    sharedClusterList,
    getSharedclusters,
    handleCreateVCluster,
    handleDeleteVCluster,
  };
}

export function useCloud() {
  const accountType = ref<'STANDARD'|'LEGACY'>();
  const getCloudAccountType = async (params: {
    $cloudId: string
    accountID: string
  }) => {
    const data = await cloudAccountType(params).catch(() => ({}));
    accountType.value = data.type;
    return accountType.value;
  };

  const getCloudBwps = async (params: {
    $cloudId: string
    accountID: string
    region: string
  }) => {
    const data = await cloudBwps(params).catch(() => []);
    return data;
  };

  return {
    accountType,
    getCloudAccountType,
    getCloudBwps,
  };
}

export function getClusterTypeName(clusterData) {
  if (clusterData.is_shared) return $i18n.t('bcs.cluster.share'); // 共享集群
  if (clusterData.clusterType === 'virtual') return 'vCluster'; // 虚拟集群
  if (clusterData.clusterType === 'federation') return $i18n.t('bcs.cluster.federation'); // 联邦集群

  if (clusterData.clusterCategory === 'builder' || clusterData.clusterCategory === '' || clusterData.clusterType === 'single') {
    // 托管和独立集群
    return clusterData.manageType === 'INDEPENDENT_CLUSTER' ? $i18n.t('bcs.cluster.selfDeployed') : $i18n.t('bcs.cluster.managed');
  }

  return '--';
}

export function getClusterImportCategory(clusterData: ICluster) {
  // 导入方式
  // kubeconfig
  if (clusterData?.importCategory === 'kubeConfig') return $i18n.t('bcs.cluster.kubeConfig');

  // machine
  if (clusterData?.importCategory === 'machine') return $i18n.t('bcs.cluster.machine');

  // 云凭证
  if (clusterData?.importCategory === 'cloud') return $i18n.t('bcs.cluster.cloud');

  return $i18n.t('bcs.cluster.platform');
}
