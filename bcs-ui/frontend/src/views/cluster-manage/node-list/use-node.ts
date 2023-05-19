import store from '@/store';
import Vue, { computed } from 'vue';
import {
  getK8sNodes,
  schedulerNode as handleSchedulerNode,
  cordonNodes,
  uncordonNodes,
  setNodeLabels as handleSetNodeLabels,
  setNodeTaints as handleSetNodeTaints,
} from '@/api/modules/cluster-manager';

const { $bkMessage, $bkInfo } = Vue.prototype;

export interface INodesParams {
  clusterId: string;
  nodeIps: string[];
  nodeIP: string[];
  nodes: string[];
  nodeTemplateID?: string;
}

export interface INodeCordonParams {
  clusterID: string;
  nodes: string[];
}

export interface ILabelsItem {
  nodeName: string
  labels: Record<string, string>
}

export interface ITaint {
  key: string
  value: string
  effect: string
}

export interface ITaintsItem {
  nodeName: string
  taints: Array<ITaint>
}

export interface ILabelsAndTaintsParams<T> {
  nodes: Array<T>
  clusterID: string
}

export default function useNode() {
  const projectId = computed(() => store.getters.curProjectId);
  const projectList = computed<any[]>(() => store.state.projectList);
  const curProject = computed(() => projectList.value.find(item => item.project_id === projectId.value));
  const user = computed(() => store.state.user);
  // 获取节点列表
  const getNodeList = async (clusterId) => {
    if (!clusterId) {
      console.warn('clusterId is empty');
      return [];
    }
    const data = await getK8sNodes({
      $clusterId: clusterId,
    }).catch(() => []);
    return data.map(item => ({
      ...item,
      // 兼容就接口数据
      inner_ip: item.innerIP,
      name: item.nodeName,
      cluster_id: item.clusterID,
      // todo 方便前端搜索逻辑(节点来源: 节点池、手动添加)
      nodeSource: item.nodeGroupID ? 'nodepool' : 'custom',
    }));
  };
  // 添加节点
  const addNode = async (params: Pick<INodesParams, 'clusterId' | 'nodeIps' | 'nodeTemplateID'>) => {
    const { clusterId, nodeIps = [], nodeTemplateID = '' } = params;
    if (!clusterId || !nodeIps.length) {
      console.warn('clusterId or is nodes is empty');
      return;
    }
    const result = await store.dispatch('clustermanager/addClusterNode', {
      $clusterId: clusterId,
      nodes: nodeIps,
      nodeTemplateID,
      operator: store.state.user?.username,
    });
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('任务下发成功'),
    });
    return result;
  };
  // 任务数据
  const getTaskData = async (params: Pick<INodesParams, 'clusterId' | 'nodeIP'>) => {
    const { clusterId, nodeIP } = params;
    if (!clusterId || !nodeIP) {
      console.warn('clusterId or nodeIP is empty');
      return { taskData: null, latestTask: null };
    }
    const res = await store.dispatch('clustermanager/taskList', {
      clusterID: clusterId,
      projectID: projectId.value,
      nodeIP,
    });
    const { latestTask } = res;
    const steps = latestTask?.stepSequence || [];
    const taskData = steps.map(step => latestTask?.steps[step]);
    return {
      taskData,
      latestTask,
    };
  };
  // 任务重试
  const retryTask = async (params: Pick<INodesParams, 'clusterId' | 'nodeIP'>) => {
    const { latestTask } = await getTaskData({
      clusterId: params.clusterId,
      nodeIP: params.nodeIP,
    });
    const result = await store.dispatch('clustermanager/taskRetry', {
      $taskId: latestTask.taskID,
      updater: user.value.username,
    });
    return result;
  };
  // 停止调度
  const handleCordonNodes = async (params: INodeCordonParams) => {
    const { clusterID, nodes } = params;
    if (!clusterID || !nodes?.length) {
      console.warn('clusterId or innerIPs is empty');
      return;
    }
    const result = await cordonNodes({
      clusterID,
      nodes,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('停止调度成功'),
    });
    return result;
  };
  // 允许调度
  const handleUncordonNodes = async (params: INodeCordonParams) => {
    const { clusterID, nodes } = params;
    if (!clusterID || !nodes?.length) {
      console.warn('clusterId or innerIPs is empty');
      return;
    }
    const result = await uncordonNodes({
      clusterID,
      nodes,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('允许调度成功'),
    });
    return result;
  };
  // Pod驱逐
  const schedulerNode = async (params: Pick<INodesParams, 'clusterId' | 'nodes'>) => {
    const { clusterId, nodes = [] } = params;
    if (!clusterId || !nodes.length) {
      console.warn('clusterId or nodeIps or status is empty');
      return;
    }
    const data = await handleSchedulerNode({
      clusterID: clusterId,
      nodes,
    }).catch(() => null);
    if (data?.fail?.length) {
      $bkInfo({
        type: 'error',
        title: window.i18n.t('以下调度节点失败'),
        defaultInfo: true,
        clsName: 'custom-info-confirm',
        subTitle: data.fail.map(item => `${item.nodeName}(${item.message})`).join(', '),
      });
    } else if (data && !data.fail?.length) {
      $bkMessage({
        theme: 'success',
        message: window.i18n.t('Pod驱逐成功'),
      });
    }
    return data && !data.fail?.length;
  };
  // 删除节点
  const deleteNode = async (params: Pick<INodesParams, 'clusterId'|'nodeIps'>) => {
    const { clusterId = '', nodeIps = [] } = params;
    if (!clusterId || !nodeIps.length) {
      console.warn('clusterId or is nodes is empty');
      return;
    }
    const result = await store.dispatch('clustermanager/deleteClusterNode', {
      $clusterId: clusterId,
      nodes: nodeIps.join(','),
    });
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('任务下发成功'),
    });
    return result;
  };
  // 节点指标信息
  const getNodeOverview = async (params: Pick<INodesParams, 'clusterId'|'nodeIP'>) => {
    const { clusterId = '', nodeIP = '' } = params;
    if (!clusterId || !nodeIP) {
      console.warn('clusterId or nodeIP or status is empty');
      return;
    }
    const data = await store.dispatch('metric/clusterNodeOverview', {
      $projectCode: curProject.value.project_code,
      $clusterId: clusterId,
      $nodeIP: nodeIP,
    }).catch(() => ({}));
    return data;
  };
  // 设置节点标签
  const setNodeLabels = async (params: ILabelsAndTaintsParams<ILabelsItem>) => {
    const result = await handleSetNodeLabels(params).then(() => true)
      .catch(() => false);
    return result;
  };
  // 设置节点污点
  const setNodeTaints = async (params: ILabelsAndTaintsParams<ITaintsItem>) => {
    const result = await handleSetNodeTaints(params).then(() => true)
      .catch(() => false);
    return result;
  };
  return {
    getNodeList,
    getTaskData,
    handleUncordonNodes,
    handleCordonNodes,
    schedulerNode,
    deleteNode,
    addNode,
    getNodeOverview,
    retryTask,
    setNodeLabels,
    setNodeTaints,
  };
}
