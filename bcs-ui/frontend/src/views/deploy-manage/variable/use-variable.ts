import { ref } from 'vue';
import {
  createVariable,
  variableDefinitions,
  deleteDefinitions,
  updateVariable,
  importVariable,
  clusterVariable,
  updateClusterVariable,
  namespaceVariable,
  updateNamespaceVariable,
} from '@/api/base';
import {
  getClusterVariables,
  updateSpecifyClusterVariables,
  getClusterNamespaceVariable,
  updateClusterNamespaceVariable,
} from '@/api/modules/project';
import { usePage } from './use-table';

export interface IParams {
  limit: number
  offset: number
  searchKey: string
  scope: 'global' | 'cluster' | 'namespace' | ''
  all: boolean
}
export type Pick<T, K extends keyof T> = T[K];

export default function useVariable() {
  const isLoading = ref(false);
  const variableList = ref([]);
  const {
    pagination,
    handlePageChange,
    handlePageLimitChange,
  } = usePage();

  async function handleCreateVariable(params) {
    const result = await createVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleImportVariable(params) {
    const result = await importVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleUpdateVariable(params) {
    const result = await updateVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function handleDeleteDefinitions(params) {
    const result = await deleteDefinitions(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function getVariableDefinitions(params: IParams) {
    isLoading.value = true;
    const { results, total } = await variableDefinitions(params).catch(() => ({ results: [], total: 0 }));
    variableList.value = results;
    pagination.value.count = total;
    isLoading.value = false;
    return { results, total };
  }

  async function getClusterVariable(params) {
    const data = await clusterVariable(params).catch(() => ({ total: 0, results: [] }));
    return data;
  }
  async function handleUpdateClusterVariable(params) {
    const result = await updateClusterVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }
  async function getNamespaceVariable(params) {
    const data = await namespaceVariable(params).catch(() => ({ total: 0, results: [] }));
    return data;
  }
  async function handleUpdateNamespaceVariable(params) {
    const result = await updateNamespaceVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }
  async function  handleGetClusterVariables(params) {
    const data = await getClusterVariables(params).catch(() => ({ results: [], total: 0 }));
    return data;
  }
  async function handleUpdateSpecifyClusterVariables(params) {
    const result = await updateSpecifyClusterVariables(params).then(() => true)
      .catch(() => false);
    return result;
  }
  async function handleGetClusterNamespaceVariable(params) {
    const data = await getClusterNamespaceVariable(params).catch(() => ({ results: [], total: 0 }));
    return data;
  }

  async function handleUpdateClusterNamespaceVariable(params) {
    const result = await updateClusterNamespaceVariable(params).then(() => true)
      .catch(() => false);
    return result;
  }

  return {
    isLoading,
    variableList,
    pagination,
    handlePageChange,
    handlePageLimitChange,
    handleCreateVariable,
    getVariableDefinitions,
    handleUpdateVariable,
    handleDeleteDefinitions,
    handleImportVariable,
    getClusterVariable,
    handleUpdateClusterVariable,
    getNamespaceVariable,
    handleUpdateNamespaceVariable,
    handleGetClusterVariables,
    handleUpdateSpecifyClusterVariables,
    handleGetClusterNamespaceVariable,
    handleUpdateClusterNamespaceVariable,
  };
}
