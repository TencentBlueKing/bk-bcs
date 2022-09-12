import Vue from 'vue';
import store from '@/store';
import { computed } from '@vue/composition-api';

const { $bkMessage } = Vue.prototype;

export interface INodesParams {
  clusterId: string;
  nodeIps: string[];
  nodeTemplateID?: string;
}
export interface INodeParams {
  clusterId: string;
  nodeIP: string[];
  status: 'REMOVABLE' | 'RUNNING';
}
export interface INodeDispatchParams {
  clusterId: string;
  nodeName: string[];
  status: 'REMOVABLE' | 'RUNNING';
}
export interface IBatchDispatchParams {
  clusterId: string;
  nodeNameList: string[];
  status: 'REMOVABLE' | 'RUNNING';
}

export default function useNode() {
  const projectId = computed(() => store.state.curProjectId);
  const projectList = computed<any[]>(() => store.state.sideMenu.onlineProjectList);
  const curProject = computed(() => projectList.value.find(item => item.project_id === projectId.value));
  const user = computed(() => store.state.user);
  // 获取节点列表
  const getNodeList = async (clusterId) => {
    if (!clusterId) {
      console.warn('clusterId is empty');
      return [];
    }
    const data = await store.dispatch('cluster/getK8sNodes', {
      $clusterId: clusterId,
    }).catch(() => []);
    return data;
  };
  // 添加节点
  const addNode = async (params: INodesParams) => {
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
  const getTaskData = async (params: Omit<INodeParams, 'status'>) => {
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
  const retryTask = async (params: Omit<INodeParams, 'status'>) => {
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
  // 停止/允许 调度
  const toggleNodeDispatch = async (params: INodeDispatchParams) => {
    const { clusterId, nodeName, status } = params;
    if (!clusterId || !nodeName || !['REMOVABLE', 'RUNNING'].includes(status)) {
      console.warn('clusterId or nodeName or status is empty');
      return;
    }
    const result = await store.dispatch('cluster/updateNodeStatus', {
      projectId: projectId.value,
      clusterId,
      nodeName,
      status,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: status === 'REMOVABLE' ? window.i18n.t('停止调度成功') : window.i18n.t('允许调度成功'),
    });
    return result;
  };
  // 批量调度
  const batchToggleNodeDispatch = async (params: IBatchDispatchParams) => {
    const { clusterId, nodeNameList, status } = params;
    const result = await store.dispatch('cluster/batchUpdateNodeStatus', {
      projectId: projectId.value,
      clusterId,
      nodeNameList,
      status,
    });
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('操作成功'),
    });
    return result;
  };
  // Pod驱逐
  const schedulerNode = async (params: INodesParams) => {
    const { clusterId, nodeIps = [] } = params;
    if (!clusterId || !nodeIps.length) {
      console.warn('clusterId or nodeIps or status is empty');
      return;
    }
    const result = await store.dispatch('cluster/schedulerNode', {
      $clusterId: clusterId,
      host_ips: nodeIps,
    }).catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: window.i18n.t('Pod驱逐成功'),
    });
    return result;
  };
  // 删除节点
  const deleteNode = async (params: INodesParams) => {
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
  const getNodeOverview = async (params: Omit<INodeParams, 'status'>) => {
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
  return {
    getNodeList,
    getTaskData,
    toggleNodeDispatch,
    schedulerNode,
    deleteNode,
    addNode,
    getNodeOverview,
    batchToggleNodeDispatch,
    retryTask,
  };
}
