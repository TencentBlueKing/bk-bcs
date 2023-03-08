<template>
  <div class="cursor-default">
    <!-- 集群切换 -->
    <div class="cluster-view">
      <span :class="['cluster-view-type', { 'shared': isSharedCluster }]">
        {{ isSharedCluster ? $t('共享') : $t('专用') }}
      </span>
      <span class="flex flex-1 flex-col ml-[10px]">
        <span class="mb-[4px] bcs-ellipsis">{{ curCluster.clusterName }}</span>
        <span class="text-[#979ba5] bcs-ellipsis">{{ curCluster.clusterID }}</span>
      </span>
      <span class="ml-[5px] cursor-pointer">
        <i class="biz-conf-btn bcs-icon bcs-icon-qiehuan f12" @click="showClusterSelector = true"></i>
      </span>
      <ClusterSelector v-model="showClusterSelector" />
    </div>
    <!-- 资源视图菜单 -->
    <div class="side-menu-wrapper"><SideMenu /></div>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref } from '@vue/composition-api';
import SideMenu from '@/views/app/side-menu.vue';
import ClusterSelector from '@/components/cluster-selector/index.vue';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'DashboardSideBar',
  components: { SideMenu, ClusterSelector },
  setup() {
    const { curCluster, isSharedCluster } = useCluster();
    const showClusterSelector = ref(false);

    return {
      curCluster,
      isSharedCluster,
      showClusterSelector,
    };
  },
});
</script>
<style lang="postcss" scoped>
.side-menu-wrapper {
  max-height: calc(100vh - 164px);
  overflow-y: auto;
  &::-webkit-scrollbar {
    display: none;
  }
}
.cluster-view {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 52px;
  margin: -12px 0 0 0;
  padding: 0 20px 0 10px;
  font-size: 12px;
  background-color: #fafbfd;
  border-bottom: 1px solid #dde4eb;
  min-width: 260px;
  overflow: hidden;
  &-type {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #3a84ff;
    border-radius: 4px;
    color: #fff;
    font-weight: bold;
    width: 38px;
    height: 28px;
    &.shared {
      background-color: #14a568;
    }
  }
}
</style>
