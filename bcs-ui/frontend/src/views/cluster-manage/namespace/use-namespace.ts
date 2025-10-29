import { ref, watch } from 'vue';

import {
  createdNamespace,
  deleteNamespace,
  fetchNamespaceInfo,
  getClusterNamespaceVariable,
  getNamespaceList,
  syncNamespaceList,
  updateClusterNamespaceVariable,
  updateNamespace,
  withdrawNamespace,
} from '@/api/modules/project';
import $store from '@/store';

// 是否为首次进入资源管理页面
export const isEntry = ref(false);

export function useNamespace() {
  const namespaceData = ref<any[]>([]);
  const variablesList = ref<any[]>([]);
  const variableLoading = ref(false);
  const namespaceLoading = ref(false);
  const webAnnotations = ref({ perms: {} });

  async function getNamespaceData(params, loading = true) {
    if (!params?.$clusterId) return;
    namespaceLoading.value = loading;
    const { data, web_annotations: _webAnnotations } = await getNamespaceList(
      params,
      { needRes: true, cancelWhenRouteChange: false },
    )
      .catch(() => ({ data: [], web_annotations: [] }));
    namespaceData.value = data;
    webAnnotations.value = _webAnnotations;
    namespaceLoading.value = false;
    return data;
  }

  async function handleGetVariablesList(params) {
    variableLoading.value = true;
    const { results, total } = await getClusterNamespaceVariable(params)
      .catch(() => ({ results: [], total: 0 }));
    variablesList.value = results;
    variableLoading.value = false;
    return { results, total };
  }

  async function handleUpdateNameSpace(params) {
    const result = await updateNamespace(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleDeleteNameSpace(params) {
    const result = await deleteNamespace(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleUpdateVariablesList(params) {
    const result = await updateClusterNamespaceVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleCreatedNamespace(params) {
    const result = await createdNamespace(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function getNamespaceInfo(params) {
    const result = await fetchNamespaceInfo(params).catch(() => {});
    return result;
  }

  async function handleSyncNamespaceList(params) {
    const result = await syncNamespaceList(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleWithdrawNamespace(params) {
    const result = await withdrawNamespace(params).then(() => true)
      .catch(() => false);
    return result;
  }

  return {
    namespaceLoading,
    namespaceData,
    webAnnotations,
    variablesList,
    variableLoading,
    getNamespaceData,
    handleGetVariablesList,
    handleUpdateNameSpace,
    handleDeleteNameSpace,
    handleUpdateVariablesList,
    handleCreatedNamespace,
    getNamespaceInfo,
    handleSyncNamespaceList,
    handleWithdrawNamespace,
  };
}

export interface INamespace {
  annotations: any[]
  cpuUseRate: number
  createTime: string
  labels: any[]
  name: string
  status: string
  isSystem: boolean
}
export function useSelectItemsNamespace() {
  const namespaceValue = ref('');
  const namespaceLoading = ref(false);
  const namespaceList = ref<INamespace[]>([]);

  const getNamespaceData = async ({ clusterId }, initNsValue = true) => {
    if (!clusterId) return;
    namespaceLoading.value = true;
    const data = await getNamespaceList({
      $clusterId: clusterId,
    }, {
      cancelWhenRouteChange: false,
    }).catch(() => []);
    // 过滤未创建成功的命名空间
    namespaceList.value = (data || []).filter(item => item.itsmTicketType !== 'CREATE');
    if (initNsValue) {
      // 初始化默认选中命名空间
      const defaultSelectNamespace = namespaceList.value
        .find(data => data.name === $store.state.curNamespace);
      namespaceValue.value = defaultSelectNamespace?.name || namespaceList.value[0]?.name;
    }

    namespaceLoading.value = false;
    return data;
  };

  watch(namespaceValue, () => {
    $store.commit('updateCurNamespace', namespaceValue.value);
  });

  return {
    namespaceLoading,
    namespaceValue,
    namespaceList,
    getNamespaceData,
  };
}
