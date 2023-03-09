<template>
  <RouterView v-if="clusterID" />
</template>
<script lang="ts">
import { defineComponent, toRef, reactive, computed, onBeforeMount, ref } from '@vue/composition-api';
import $router from '@/router';
import $store from '@/store';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'DashboardIndex',
  setup() {
    const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);
    const { curClusterId, clusterList } = useCluster();
    const clusterID = ref<string>(currentRoute.value.params.clusterId
    || curClusterId.value
    || clusterList.value[0]?.clusterID);

    // 需要提前更新当前缓存的集群
    $store.commit('updateCurCluster', clusterList.value.find(item => item.clusterID === clusterID.value));
    onBeforeMount(() => {
      if (
        !currentRoute.value.name
        || !clusterID.value
        || (clusterID.value === currentRoute.value.params.clusterId)
      ) return;


      $router.replace({
        name: currentRoute.value.name,
        params: {
          clusterId: clusterID.value,
        },
      });
    });

    return {
      clusterID,
    };
  },
});
</script>
