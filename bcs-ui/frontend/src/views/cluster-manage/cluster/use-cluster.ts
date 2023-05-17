/* eslint-disable camelcase */
import { computed, ref, watch, Ref, set, reactive, toRef } from 'vue';
import useInterval from '@/composables/use-interval';
import { clusterDetail, cloudNodes } from '@/api/modules/cluster-manager';
import $store from '@/store';
import $router from '@/router';
import { ICluster } from '@/composables/use-app';

/**
 * 获取集群列表
 * @param ctx
 * @returns
 */
export function useClusterList() {
  const clusterList = computed<ICluster[]>(() => $store.state.cluster.clusterList);
  const curProjectId = computed(() => $store.getters.curProjectId);
  const clusterExtraInfo = ref({});
  const webAnnotations = ref({ perms: {} });
  // 获取集群列表
  const getClusterList = async () => {
    const res = await $store.dispatch('cluster/getClusterList', curProjectId.value);
    clusterExtraInfo.value = res.clusterExtraInfo || {};
    webAnnotations.value = res.web_annotations || { perms: {} };
  };
  // 开启轮询
  const { start, stop } = useInterval(getClusterList, 5000);
  const runningClusterIds = computed(() => clusterList.value.filter(item => [
    'INITIALIZATION',
    'DELETING'].includes(item.status)).map(item => item.cluster_id));
  watch(runningClusterIds, (newValue) => {
    if (!newValue.length) {
      stop();
    } else {
      start();
    }
  });

  return {
    webAnnotations,
    clusterList,
    curProjectId,
    clusterExtraInfo,
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
  const retryClusterTask = async (cluster): Promise<boolean> => {
    const { latestTask } = await getTaskData(cluster.clusterID);
    const result = await $store.dispatch('clustermanager/taskRetry', {
      $taskId: latestTask.taskID,
      updater: user.value.username,
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
    clusterNode,
    getCloudNodes,
  };
}

/**
 * 任务操作
 * @param ctx
 */
export function useTask() {
  // 查询任务列表
  const taskList = async (cluster) => {
    const data = await $store.dispatch('clustermanager/taskList', {
      clusterID: cluster.clusterID,
      projectID: cluster.projectID,
    });
    return data;
  };
  return {
    taskList,
  };
}

export function useClusterInfo() {
  const isLoading = ref(false);
  const clusterData = ref<Record<string, any>>({});
  const clusterOS = computed(() => clusterData.value?.clusterBasicSettings?.OS);
  const clusterAdvanceSettings = computed(() => clusterData.value?.clusterAdvanceSettings || {});
  const extraInfo = computed(() => clusterData.value?.extraInfo || {});
  const getClusterDetail = async ($clusterId: string, cloudInfo = false) => {
    isLoading.value = true;
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
