import Vue, { computed } from 'vue';

import {
  batchDeleteNodes as batchDeleteNodesAPI,
  cordonNodes,
  getK8sNodes,
  schedulerNode as handleSchedulerNode,
  setNodeAnnotations as handleSetNodeAnnotations,
  setNodeLabels as handleSetNodeLabels,
  setNodeTaints as handleSetNodeTaints,
  taskDetail as taskDetailAPI,
  uncordonNodes } from '@/api/modules/cluster-manager';
import {
  clusterAllNodeOverview,
} from '@/api/modules/monitor';
import store from '@/store';

const { $bkMessage, $bkInfo } = Vue.prototype;

export interface INodesParams {
  clusterId: string;
  nodeIps: string[];
  nodeIP: string[];
  nodes: string[];
  nodeTemplateID?: string;
  login?: any;
  advance?: any;
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

export interface IAnnotationsItem {
  nodeName: string
  annotations: Record<string, string>
}

export interface ILabelsAndTaintsParams<T> {
  nodes: Array<T>
  clusterID: string
}

export default function useNode() {
  const projectId = computed(() => store.getters.curProjectId);
  const curProject = computed(() => store.state.curProject);
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
    })).sort((pre, current) => {
      const runnings = ['INITIALIZATION', 'DELETING', 'APPLYING'];
      if (!runnings.includes(pre.status) && runnings.includes(current.status)) {
        return 1;
      } if (runnings.includes(pre.status) && !runnings.includes(current.status)) {
        return -1;
      }
      return 0;
    });
  };
  // 添加节点
  const addNode = async (params: Pick<INodesParams, 'clusterId' | 'nodeIps' | 'nodeTemplateID' | 'login' | 'advance'>) => {
    const { clusterId, nodeIps = [], nodeTemplateID = '', login, advance } = params;
    if (!clusterId || !nodeIps.length) {
      console.warn('clusterId or is nodes is empty');
      return;
    }
    const result = await store.dispatch('clustermanager/addClusterNode', {
      $clusterId: clusterId,
      nodes: nodeIps,
      nodeTemplateID,
      login,
      advance,
      operator: store.state.user?.username,
    });
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('generic.msg.success.deliveryTask'),
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
      operator: user.value.username,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('cluster.nodeList.msg.unCordonOK'),
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
      operator: user.value.username,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('cluster.nodeList.msg.cordonOK'),
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
      operator: user.value.username,
    }).catch(() => null);
    if (data?.fail?.length) {
      $bkInfo({
        type: 'error',
        title: window.i18n.t('cluster.nodeList.title.failedCordonData'),
        defaultInfo: true,
        clsName: 'custom-info-confirm',
        subTitle: data.fail.map(item => `${item.nodeName}(${item.message})`).join(', '),
      });
    } else if (data && !data.fail?.length) {
      $bkMessage({
        theme: 'success',
        message: window.i18n.t('cluster.nodeList.msg.drainOK'),
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
      message: window.i18n.t('generic.msg.success.deliveryTask'),
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
      $projectCode: curProject.value.projectCode,
      $clusterId: clusterId,
      $nodeIP: nodeIP,
    }).catch(() => ({}));
    return data;
  };
  // 节点指标全量信息
  const getAllNodeOverview = async (params: Pick<INodesParams, 'clusterId'|'nodes'>) => {
    const { clusterId = '', nodes = [] } = params;
    if (!clusterId) {
      console.warn('clusterId or status is empty');
      return;
    }
    const data = await clusterAllNodeOverview({
      $projectCode: curProject.value.projectCode,
      $clusterId: clusterId,
      node: nodes,
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
  // 设置节点注解
  const setNodeAnnotations = async (params: ILabelsAndTaintsParams<IAnnotationsItem>) => {
    const result = await handleSetNodeAnnotations(params).then(() => true)
      .catch(() => false);
    return result;
  };
  // 批量设置节点标签，返回data,以处理个别节点标签设置失败的情况
  const batchSetNodeLabels = async (params: ILabelsAndTaintsParams<ILabelsItem>) => {
    const result = await handleSetNodeLabels(params).then(data => data)
      .catch(() => false);

    return result;
  };
  // 设置节点污点，返回data,以处理个别节点污点设置失败的情况
  const batchSetNodeTaints = async (params: ILabelsAndTaintsParams<ITaintsItem>) => {
    const result = await handleSetNodeTaints(params).then(data => data)
      .catch(() => false);
    return result;
  };
  // 批量设置节点注解，返回data,以处理个别节点注解设置失败的情况
  const batchSetNodeAnnotations = async (params: ILabelsAndTaintsParams<IAnnotationsItem>) => {
    const result = await handleSetNodeAnnotations(params).then(data => data)
      .catch(() => false);
    return result;
  };
  // 批量删除节点（节点组、空节点）
  const batchDeleteNodes = async (params: {
    $clusterId: string
    nodeIPs?: string
    virtualNodeIDs?: string
    operator: string
    deleteMode?: 'terminate' | 'retain'
  }) => {
    const result = await batchDeleteNodesAPI(params).catch(() => false);
    return result;
  };
  // 任务详情
  const taskDetail = async ($taskId: string) => {
    const data = await taskDetailAPI({ $taskId }).catch(() => []);
    return data;
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
    setNodeAnnotations,
    batchSetNodeLabels,
    batchSetNodeTaints,
    batchSetNodeAnnotations,
    batchDeleteNodes,
    taskDetail,
    getAllNodeOverview,
  };
}
