import { ref } from 'vue';
import { ISubscribeData } from './use-subscribe';
import $store from '@/store';

/**
 * 加载表格数据
 * @param ctx
 * @returns
 */
export default function useTableData() {
  const isLoading = ref(false);
  const data = ref<ISubscribeData>({
    manifestExt: {},
    manifest: {},
  });
  const webAnnotations = ref<any>({});

  const fetchList = async (type: string, category: string, namespaceId: string, clusterId: string) => {
    const action = namespaceId ? 'dashboard/getTableData' : 'dashboard/getTableDataWithoutNamespace';
    const res = await $store.dispatch(action, {
      $type: type,
      $category: category,
      $namespaceId: namespaceId,
      $clusterId: clusterId,
    });
    return res;
  };
  const handleFetchList = async (
    type: string,
    category: string,
    namespaceId: string,
    clusterId: string,
  ): Promise<ISubscribeData|undefined> => {
    // persistent_volumes、storage_classes资源和命名空间无关，其余资源必须传命名空间
    if (!namespaceId && !['persistent_volumes', 'storage_classes'].includes(category)) return;
    isLoading.value = true;
    const res = await fetchList(type, category, namespaceId, clusterId);
    data.value = res.data;
    webAnnotations.value = res.webAnnotations || {};
    isLoading.value = false;
    return res.data;
  };

  const fetchCRDData = async (clusterId: string) => {
    const res = await $store.dispatch('dashboard/crdList', { $clusterId: clusterId });
    return res;
  };
  const handleFetchCustomResourceList = async (
    clusterId: string,
    crd?: string,
    category?: string,
    namespace?: string,
  ): Promise<ISubscribeData|undefined> => {
    // crd 和 category 必须同时存在（同时不存在：crd列表，同时存在：特定类型自定义资源列表）
    if ((crd && !category) || (!crd && category)) return;
    isLoading.value = true;
    const res = await $store.dispatch('dashboard/customResourceList', {
      $crd: crd,
      $category: category,
      $clusterId: clusterId,
      namespace,
    });
    data.value = res.data;
    webAnnotations.value = res.webAnnotations || {};
    isLoading.value = false;
    return res.data;
  };

  return {
    isLoading,
    data,
    webAnnotations,
    fetchList,
    handleFetchList,
    fetchCRDData,
    handleFetchCustomResourceList,
  };
}
