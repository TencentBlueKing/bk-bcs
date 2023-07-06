<template>
  <RouterView v-if="clusterID" />
</template>
<script lang="ts">
import { defineComponent, toRef, reactive, computed, onBeforeMount, ref } from 'vue';
import $router from '@/router';
import $store from '@/store';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'DashboardIndex',
  setup() {
    const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);
    const { curClusterId, clusterList } = useCluster();
    const clusterID = ref<string>('');
    // 判断路由上的集群ID是否正确
    if (!clusterList.value.find(item => item.clusterID === currentRoute.value.params.clusterId)) {
      // 判断缓存上的集群ID是否正确
      if (clusterList.value.find(item => item.clusterID === curClusterId.value)) {
        clusterID.value = curClusterId.value;
      } else {
        // 取默认第一个作为集群ID
        clusterID.value = clusterList.value[0]?.clusterID;
      }
    } else {
      clusterID.value = currentRoute.value.params.clusterId;
    }

    // 需要提前更新当前缓存的集群
    $store.commit('updateCurCluster', clusterList.value.find(item => item.clusterID === clusterID.value));
    onBeforeMount(() => {
      if (
        !currentRoute.value.name
        || !clusterID.value
        || (clusterID.value === currentRoute.value.params.clusterId) // 路由上集群ID正确就不需要执行replace，防止router warn
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
