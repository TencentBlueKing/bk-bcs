import { computed, onBeforeMount, ref } from 'vue';

import { ICluster, useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import useViewConfig from '@/views/resource-view/view-manage/use-view-config';

export type ClusterType = 'independent' | 'managed' |'virtual' | 'shared' | 'all';
/**
 * 集群选择器
 * @param emits 事件
 * @param defaultClusterID 默认集群ID
 * @param clusterType 集群类型，默认 ['independent', 'managed']，可选值：'independent' | 'managed' |'virtual' | 'shared' | 'all'
 * @param updateStore 是否更新store中的当前集群，默认true
 * @param validateClusterID 是否校验当前集群ID是否在当前场景中可用，默认true
 * @param syncQueryValue 是否同步路由query参数，注意！！！路由变化默认会取消之前未完成的请求，具体请查看 cancelWhenRouteChange 配置项
 */
export default function useClusterSelector(
  emits: any,
  defaultClusterID: string,
  clusterType: ClusterType | Omit<ClusterType, 'all'>[] = ['independent', 'managed'],
  updateStore = true,
  validateClusterID = true,
  syncQueryValue = false,
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
  // 当前场景的集群分类数据
  const clusterData = computed(() => {
    if (Array.isArray(clusterType)) {
      return clusterListByType.value.filter(item => clusterType.includes(item.type));
    }
    return clusterType === 'all'
      ? clusterListByType.value
      : clusterListByType.value.filter(item => item.type === clusterType);
  });
  const isClusterDataEmpty = computed(() => clusterData.value.every(item => !item.list.length));

  // 当 syncQueryValue 为 true 时，优先从路由query参数中读取 clusterId
  const queryClusterID = syncQueryValue ? ($router.currentRoute?.query?.clusterId as string) : '';
  const localValue = ref<string>(queryClusterID || defaultClusterID || $store.getters.curClusterId);

  const handleClusterChange = (clusterId = '') => {
    localValue.value = clusterId;
    updateStore && $store.commit('updateCurCluster', clusterList.value.find(item => item.clusterID === clusterId));
    // 当 syncQueryValue 为 true 时，将集群变更同步写入路由 query 参数
    if (syncQueryValue && localValue.value !== ($router.currentRoute?.query?.clusterId as string)) {
      const currentQuery = { ...$router.currentRoute?.query };
      $router.replace({ query: { ...currentQuery, clusterId } });
    }
    emits('change', clusterId);
  };

  const { dashboardViewID } = useViewConfig();
  const viewID = $router.currentRoute?.query?.viewID || dashboardViewID.value || '';
  const handleValidateClusterID = () => {
    if (!clusterList.value.length) return;// 资源视图的左侧菜单是单独routerview，如果clusterList为空就不重置当前集群ID
    // 判断当前集群ID在当前场景中是否能使用（当前场景下，不能从全量数据中判断）
    const data = clusterData.value.find(data => data.list.some(item => item.clusterID === localValue.value));
    let clusterID;
    if (!data) {
      handleClusterChange(clusterData.value[0]?.list?.[0]?.clusterID);
      clusterID = clusterData.value[0]?.list?.[0]?.clusterID;
    } else if (localValue.value !== defaultClusterID) {
      handleClusterChange(localValue.value);
      clusterID = localValue.value;
    }
    // 首页初始化集群模式id，视图模式下在 dashboard-view.vue 中初始化
    if (!viewID && clusterID) {
      clusterID && emits('init', clusterID);
    }
  };

  onBeforeMount(() => {
    validateClusterID && handleValidateClusterID();
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
