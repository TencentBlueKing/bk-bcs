import { computed, onBeforeMount, ref } from 'vue';

import { ICluster, useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';

export type ClusterType = 'independent' | 'managed' |'virtual' | 'shared' | 'all';
export default function useClusterSelector(
  emits: any,
  defaultClusterID: string,
  clusterType: ClusterType | Omit<ClusterType, 'all'>[] = ['independent', 'managed'],
  updateStore = true,
) {
  const { clusterList } = useCluster();
  const keyword = ref('');

  const collapseList = ref<Array<ClusterType>>([]);
  const handleToggleCollapse = (type: ClusterType) => {
    const index = collapseList.value.findIndex(item => item === type);
    if (index > -1) {
      collapseList.value.splice(index, 1);
    } else {
      collapseList.value.push(type);
    }
  };
  // 集群分类数据
  const clusterListByType = computed(() => clusterList.value
    .filter((item) => {
      const clusterID = item?.clusterID?.toLocaleLowerCase();
      const clusterName = item?.clusterName?.toLocaleLowerCase();
      const searchKey = keyword.value?.toLocaleLowerCase();
      return (clusterID?.includes(searchKey) || clusterName?.includes(searchKey));
    })
    .reduce<Array<{
    type: ClusterType
    list: Array<Partial<ICluster>>
    title: string
  }>>((list, item) => {
    if (item.clusterType === 'virtual') {
      const data = list.find(item => item.type === 'virtual');
      // 虚拟集群属于共享集群中的一种
      data?.list.push(item);
    } else if (item.is_shared) {
      // 共享集群
      const data = list.find(item => item.type === 'shared');
      data?.list.push(item);
    } else if (item.manageType === 'MANAGED_CLUSTER') {
      // 托管集群
      const data = list.find(item => item.type === 'managed');
      data?.list.push(item);
    } else {
      // 独立集群
      const data = list.find(item => item.type === 'independent');
      data?.list.push(item);
    }
    return list;
  }, [
    {
      type: 'virtual',
      list: [],
      title: 'vCluster',
    },
    {
      type: 'managed',
      list: [],
      title: $i18n.t('bcs.cluster.managed'),
    },
    {
      type: 'independent',
      list: [],
      title: $i18n.t('bcs.cluster.selfDeployed'),
    },
    {
      type: 'shared',
      list: [],
      title: $i18n.t('bcs.cluster.share'),
    },
  ])
    .filter(item => !!item.list.length));
  // 当前集群分类数据
  const clusterData = computed(() => {
    if (Array.isArray(clusterType)) {
      return clusterListByType.value.filter(item => clusterType.includes(item.type));
    }
    return clusterType === 'all'
      ? clusterListByType.value
      : clusterListByType.value.filter(item => item.type === clusterType);
  });
  const isClusterDataEmpty = computed(() => clusterData.value.every(item => !item.list.length));

  const localValue = ref<string>(defaultClusterID || $store.getters.curClusterId);

  const handleClusterChange = (clusterId = '') => {
    localValue.value = clusterId;
    updateStore && $store.commit('updateCurCluster', clusterList.value.find(item => item.clusterID === clusterId));
    emits('change', clusterId);
  };

  const handleValidateClusterID = () => {
    if (!clusterList.value.length) return;// 资源视图的左侧菜单是单独routerview，如果clusterList为空就不重置当前集群ID
    // 判断当前集群ID在当前场景中是否能使用
    const data = clusterData.value.find(data => data.list.some(item => item.clusterID === localValue.value));
    if (!data) {
      handleClusterChange(clusterData.value[0]?.list?.[0]?.clusterID);
    } else if (localValue.value !== defaultClusterID) {
      handleClusterChange(localValue.value);
    }
  };

  onBeforeMount(() => {
    handleValidateClusterID();
  });

  return {
    keyword,
    localValue,
    collapseList,
    isClusterDataEmpty,
    clusterData,
    handleToggleCollapse,
    handleClusterChange,
    handleValidateClusterID,
  };
}
