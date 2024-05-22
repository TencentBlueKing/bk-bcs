<template>
  <div class="bcs-dashboard-view h-full flex overflow-hidden" :style="style">
    <div class="bcs-border-right cursor-default bg-[#fff] h-full">
      <!-- 视图切换 -->
      <ViewSelector
        class="bcs-border-bottom w-full"
        :is-view-config-show="showViewConfig"
        @toggle-view-config="showViewConfig = !showViewConfig"
        @create-new-view="handleCreateView"
        @edit-view="handleEditView" />
      <div class="flex items-start h-[calc(100%-66px)]">
        <!-- 视图配置 -->
        <ViewConfig
          v-if="showViewConfig"
          class="bcs-border-right h-full"
          ref="viewConfigRef"
          @close="closeViewConfig" />
        <!-- 侧边导航 -->
        <ResourceMenu class="w-[260px] h-full" />
      </div>
    </div>
    <!-- 界面 -->
    <div class="relative w-[0] h-full flex-1">
      <!-- 预览态样式 -->
      <div
        :class="[
          'absolute h-full z-[1000] right-0',
          'border-solid border-[6px] border-[#FFB848] pointer-events-none',
          openSideMenu ? 'left-[-260px]' : 'left-[-60px]'
        ]"
        v-if="isViewEditable">
        <div class="flex items-center h-[32px] absolute right-[-6px] top-[-6px]">
          <svg class="icon svg-icon" width="32px" height="32px">
            <use xlink:href="#bcs-icon-color-edit-tag"></use>
          </svg>
          <div
            :class="[
              'flex items-center pr-[24px] p-[4px] h-[32px] ml-[-4px]',
              'text-[16px] text-[#fff] bg-[#FF9C01]'
            ]">
            {{ $t('view.tips.editMode') }}
          </div>
        </div>
      </div>
      <RouterView :key="$route.fullPath" v-if="isRouterAlive" />
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed, onBeforeMount, onBeforeUnmount, provide, ref, watch } from 'vue';

import ResourceMenu from './view-manage/resource-menu.vue';
import useViewConfig from './view-manage/use-view-config';
import ViewConfig from './view-manage/view-config.vue';
import ViewSelector from './view-manage/view-selector.vue';

import { bus } from '@/common/bus';
import { useCluster } from '@/composables/use-app';
import useCalcHeight from '@/composables/use-calc-height';
import $router from '@/router';
import $store from '@/store';

const openSideMenu = computed(() => $store.state.openSideMenu);

// 计算content高度
const { style } = useCalcHeight({
  prop: 'height',
  offset: 52, // 导航
  calc: ['#bcs-notice-com'],
});

const showViewConfig = ref(false);

watch(showViewConfig, () => {
  $store.commit('updateViewConfigStatus', showViewConfig.value);
});

const isViewEditable = computed(() => $store.state.isViewEditable);

// 关闭视图配置
const closeViewConfig = () => {
  showViewConfig.value = false;
};

// 创建视图
const viewConfigRef = ref<InstanceType<typeof ViewConfig>>();
const handleCreateView = () => {
  showViewConfig.value = true;
  setTimeout(() => {
    viewConfigRef.value?.createNewView();
  });
};

// 编辑视图
const handleEditView = (id: string) => {
  showViewConfig.value = true;
  setTimeout(() => {
    viewConfigRef.value?.editView(id);
  });
};

// 重载当前界面
const isRouterAlive = ref(true);
const reload = () => {
  isRouterAlive.value = false;
  setTimeout(() => {
    isRouterAlive.value = true;
  });
};

const { dashboardViewID, updateViewIDStore } = useViewConfig();

watch(
  dashboardViewID,
  () => {
    if ($router.currentRoute?.name === 'dashboardCustomObjects') {
      // 切换视图时如果在自定义视图菜单上就跳到首页(deploy)
      $router.replace({
        name: 'dashboardWorkloadDeployments',
        params: $router.currentRoute.params,
      });
      return;
    }
    // 处理路径上的集群参数
    if (dashboardViewID.value && $router.currentRoute?.params?.clusterId) {
      // 自定义视图清空集群ID参数
      $router.replace({
        name: $router.currentRoute?.name,
        params: {},
      });
    } else if (!dashboardViewID.value && !$router.currentRoute?.params?.clusterId) {
      // 没有集群ID, 也没有视图ID, 就默认回显一个集群
      const cluster = clusterList.value.find(item => item.status === 'RUNNING');
      $router.replace({
        name: $router.currentRoute?.name,
        params: {
          clusterId: cluster?.clusterID,
        },
      });
    }
  },
);

const { clusterList, curClusterId } = useCluster();
const initRoutePath = () => {
  const pathClusterID = $router.currentRoute?.params?.clusterId;
  let cluster;
  if (pathClusterID) {
    // 路径上带有集群ID，则更新全局集群缓存
    cluster = clusterList.value.find(item => item.clusterID === pathClusterID);
    updateViewIDStore('');// 删除视图缓存
  } else if (!dashboardViewID.value) {
    // 没有集群ID, 也没有视图ID, 就默认回显全局缓存的集群
    cluster = clusterList.value.find(item => item.clusterID === curClusterId.value)
      || clusterList.value.find(item => item.status === 'RUNNING');
    $router.replace({
      name: $router.currentRoute?.name,
      params: {
        clusterId: cluster?.clusterID,
      },
    });
  }
  $store.commit('updateCurCluster', cluster ?? clusterList.value.find(item => item.status === 'RUNNING'));
};
onBeforeMount(() => {
  bus.$on('toggle-show-view-config', () => {
    if (isViewEditable.value) return;
    showViewConfig.value = !showViewConfig.value;
  });
  initRoutePath();
});

onBeforeUnmount(() => {
  bus.$off('toggle-show-view-config');
});

provide('dashboard-view', {
  reload,
});
</script>
<style lang="postcss">
.bcs-dashboard-view {
  .dashboard-content {
    &::-webkit-scrollbar {
      width: 8px;
    }
    &::-webkit-scrollbar-thumb {
      background-color: rgb(151 155 165 / 60%);
      border-radius: 10px;
      &:hover {
        background-color: rgb(151 155 165 / 90%)
      }
    }
    &::-webkit-scrollbar-track {
      background-color: #f5f7fa;
      border-radius: 10px;
    }
  }
}
</style>
