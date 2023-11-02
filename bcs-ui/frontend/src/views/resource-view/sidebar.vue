<template>
  <div class="cursor-default">
    <div class="cluster-view">
      <span :class="['cluster-view-type', { 'shared': isSharedCluster }]">
        {{ isSharedCluster ? $t('bcs.cluster.publicIcon') : $t('bcs.cluster.privateIcon') }}
      </span>
      <span class="flex flex-1 flex-col ml-[10px]">
        <span
          class="mb-[4px] bcs-ellipsis-word"
          :title="curCluster.clusterName">
          {{ curCluster.clusterName }}
        </span>
        <span class="text-[#979ba5] bcs-ellipsis">{{ curCluster.clusterID }}</span>
      </span>
      <!-- 集群切换 -->
      <PopoverSelector trigger="click" placement="right-start" offset="0, 12" ref="popoverSelectRef">
        <span
          class="ml-[5px] cursor-pointer w-[24px] h-[24px] flex items-center justify-center">
          <i class="biz-conf-btn bcs-icon bcs-icon-qiehuan f12"></i>
        </span>
        <template #content>
          <!-- 监听change事件会多次触发，这里监听click事件 -->
          <ClusterSelectPopover cluster-type="all" :key="curCluster.clusterID" @click="handleClusterClick" />
        </template>
      </PopoverSelector>
    </div>
    <!-- 资源视图菜单 -->
    <div class="side-menu-wrapper"><SideMenu /></div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, reactive, ref, toRef } from 'vue';

import useMenu from '../app/use-menu';

import ClusterSelectPopover from '@/components/cluster-selector/cluster-select-popover.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import { useCluster } from '@/composables/use-app';
import $router from '@/router';
import $store from '@/store';
import SideMenu from '@/views/app/side-menu.vue';

export default defineComponent({
  name: 'DashboardSideBar',
  components: { SideMenu, PopoverSelector, ClusterSelectPopover },
  setup() {
    const popoverSelectRef = ref<any>(null);
    const { curCluster, isSharedCluster } = useCluster();
    const { disabledMenuIDs } = useMenu();
    const curSideMenu = computed(() => $store.state.curSideMenu);
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

    const handleClusterClick = (clusterID: string) => {
      if (!clusterID) return;

      let routeName = '';
      if (disabledMenuIDs.value.includes(curSideMenu.value?.id)) {
        routeName = 'dashboardNamespace';
      } else {
        routeName = curSideMenu.value?.route || $route.value.name;
      }
      $router.replace({
        name: routeName,
        params: {
          clusterId: clusterID,
        },
      });
      popoverSelectRef.value?.hide();
    };

    return {
      popoverSelectRef,
      curCluster,
      isSharedCluster,
      handleClusterClick,
    };
  },
});
</script>
<style lang="postcss" scoped>
.side-menu-wrapper {
  max-height: calc(100vh - 170px);
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
  margin: -6px 0 6px 0;
  padding: 0 20px 0 12px;
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
