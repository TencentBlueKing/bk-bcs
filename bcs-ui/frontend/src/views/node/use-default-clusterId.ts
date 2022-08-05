import { computed } from '@vue/composition-api';
import store from '@/store';
import { BCS_CLUSTER } from '@/common/constant';

export default function useDefaultClusterId() {
  const curClusterId = computed(() => store.state.curClusterId);
  const clusterList = computed(() => (store.state as any).cluster.clusterList || []);
  // 单集群ID > sessionStorage >列表第一个
  const defaultClusterId = computed<string|undefined>(() => curClusterId.value
        || sessionStorage.getItem(BCS_CLUSTER)
        || clusterList.value[0]?.clusterID);
  // 是否是单集群
  const isSingleCluster = computed(() => !!curClusterId.value);

  return {
    curClusterId,
    defaultClusterId,
    clusterList,
    isSingleCluster,
  };
}
