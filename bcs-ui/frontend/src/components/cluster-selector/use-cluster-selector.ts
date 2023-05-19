import { useCluster } from '@/composables/use-app';
import { ref, computed, onBeforeMount } from 'vue';
import $store from '@/store';

export type ClusterType = 'independent' | 'all';
export default function useClusterSelector(
  emits: any,
  defaultClusterID: string,
  clusterType: ClusterType = 'independent',
  updateStore = true,
) {
  const { clusterList } = useCluster();
  const clusterData = computed(() => (clusterType === 'independent' ? independentClusterList.value : clusterList.value));
  const keyword = ref('');

  const sharedCollapse = ref(false);
  const sharedClusterList = computed(() => clusterList.value.filter(item => item.is_shared
    && (item.clusterID.includes(keyword.value) || item.clusterName.includes(keyword.value))));
  const independentCollapse = ref(false);
  const independentClusterList = computed(() => clusterList.value.filter(item => !item.is_shared
  && (item.clusterID.includes(keyword.value) || item.clusterName.includes(keyword.value))));

  const localValue = ref<string>(defaultClusterID || $store.getters.curClusterId);

  const handleClusterChange = (clusterId: string) => {
    localValue.value = clusterId;
    updateStore && $store.commit('updateCurCluster', clusterList.value.find(item => item.clusterID === clusterId));
    emits('change', clusterId);
  };

  const handleValidateClusterID = () => {
    if (!clusterList.value.length) return;// 资源视图的左侧菜单是单独routerview，如果clusterList为空就不重置当前集群ID
    // 判断当前集群ID在当前场景中是否能使用
    const data = clusterData.value.find(item => item.clusterID === localValue.value);
    if (!data) {
      handleClusterChange(clusterData.value[0]?.clusterID);
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
    clusterData,
    sharedCollapse,
    sharedClusterList,
    independentCollapse,
    independentClusterList,
    handleClusterChange,
    handleValidateClusterID,
  };
}
