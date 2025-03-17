<template>
  <div class="bcs-dashboard-view h-full flex">
    <div class="bcs-border-right cursor-default bg-[#fff] h-full">
      <!-- 视图切换 -->
      <ViewSelector
        class="bcs-border-bottom w-full"
        :is-view-config-show="showViewConfig"
        @toggle-view-config="showViewConfig = !showViewConfig"
        @create-new-view="handleCreateView"
        @edit-view="handleEditView" />
      <div class="flex items-start v-m-menu-box overflow-auto">
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
import { cloneDeep } from 'lodash';
import { computed, onBeforeMount, onBeforeUnmount, provide, reactive, ref, toRef, watch } from 'vue';

import ResourceMenu from './view-manage/resource-menu.vue';
import useViewConfig from './view-manage/use-view-config';
import ViewConfig from './view-manage/view-config.vue';
import ViewSelector from './view-manage/view-selector.vue';

import { bus } from '@/common/bus';
import { useCluster } from '@/composables/use-app';
import $router from '@/router';
import $store from '@/store';

const openSideMenu = computed(() => $store.state.openSideMenu);

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

const { dashboardViewID } = useViewConfig();

// 集群列表页
const isDashboardHome = computed(() => $router.currentRoute?.matched?.some(item => item.name === 'dashboardHome'));

// 切换集群视图和自定义视图时更改路径参数
watch(
  dashboardViewID,
  (newVal, oldVal) => {
    if (!newVal && !oldVal) return;

    if ($router.currentRoute?.name === 'dashboardCustomObjects'
      || (!isDashboardHome.value && newVal && oldVal && newVal !== oldVal)) {
      // 切换视图时如果在customObject菜单或在二级详情界面切换视图时就跳到首页(deploy)，
      $router.replace({
        name: 'dashboardWorkloadDeployments',
        params: $router.currentRoute.params,
        query: {
          viewID: dashboardViewID.value,
        },
      });
      return;
    }

    if (dashboardViewID.value && $router.currentRoute?.params?.clusterId) {
      // 自定义视图时清空集群ID参数
      $router.replace({
        name: $router.currentRoute?.name,
        params: {
          clusterId: '-',
        },
        query: {
          viewID: dashboardViewID.value,
        },
      });
    } else if (!dashboardViewID.value && !$router.currentRoute?.params?.clusterId) {
      // 没有集群ID, 也没有视图ID, 就默认回显一个集群
      const cluster = clusterList.value.find(item => item.status === 'RUNNING');
      $router.replace({
        name: $router.currentRoute?.name,
        params: {
          clusterId: cluster?.clusterID,
        },
        query: {
          viewID: dashboardViewID.value,
        },
      });
    } else if (dashboardViewID.value && dashboardViewID.value !== $router.currentRoute?.query?.viewID) {
      // 替换query上的试图ID
      $router.replace({
        name: $router.currentRoute?.name,
        params: $router.currentRoute.params,
        query: {
          viewID: dashboardViewID.value,
        },
      });
    }
  },
);

const { clusterList, curClusterId } = useCluster();
// 初始化当前集群和试图ID信息
const initClusterAndViewID = () => {
  let cluster;
  let viewID = $router.currentRoute?.query?.viewID || dashboardViewID.value || '';
  let pathClusterID = $router.currentRoute?.params?.clusterId;
  if (pathClusterID === '-') {
    pathClusterID = '';
  }
  if (pathClusterID) {
    // 路径上带有集群ID, 以路径集群ID为主
    cluster = clusterList.value.find(item => item.clusterID === pathClusterID);
    if (isDashboardHome.value) {
      // 在列表页时如果有集群ID，则要清空视图ID（集群模式和视图模式只能二选一）
      viewID = '';
    } else {
      // 详情页视图ID根据query参数确定
      viewID = $router.currentRoute?.query?.viewID || '';
    }
  } else if (!viewID) {
    // 路径上没有集群ID, 也没有视图ID, 就默认跳转到一个集群中
    cluster = clusterList.value.find(item => item.clusterID === curClusterId.value)
      || clusterList.value.find(item => item.status === 'RUNNING');
    $router.replace({
      name: $router.currentRoute?.name,
      params: {
        clusterId: cluster?.clusterID,
      },
    });
  }
  $store.commit('updateDashboardViewID', viewID);// 更新当前视图ID
  $store.commit('updateCurCluster', cluster ?? clusterList.value.find(item => item.status === 'RUNNING'));
};

const curNsList = computed(() => $store.state.viewNsList);
// 解析并获取url参数
const propertis = ['name', 'creator', 'source', 'templateName', 'templateVersion', 'chartName', 'labelSelector'];
function handleGetQuery(query) {
  // 从query中获取命名空间
  query?.namespace && $store.commit('updateViewNsList', query.namespace.split(','));
  // query中不存在propertis中的参数，不显示showViewConfig
  if (!propertis.some(key => key in query)) return;

  $store.commit('updateTmpViewData', {
    filter: {
      name: query?.name,
      creator: query?.creator ? query?.creator.split(',') : [],
      createSource: {
        source: query?.source,
        template: {
          templateName: query?.templateName,
          templateVersion: query?.templateVersion,
        },
        chart: {
          chartName: query?.chartName,
        },
      },
      labelSelector: query?.labelSelector ? JSON.parse(decodeURIComponent(query?.labelSelector)) : [],
    },
  });
  // 自定义资源
  $store.commit('updateCrdData', {
    crd: query?.crd,
    kind: query?.kind,
    scope: query?.scope,
  });
  showViewConfig.value = true;
}
const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);

// 同步命名空间到url
watch(curNsList, () => {
  const queryData = cloneDeep(currentRoute.value.query);
  if (!curNsList.value?.length) {
    delete queryData.namespace;
  } else {
    queryData.namespace = curNsList.value.join(',');
  }
  $router.replace({
    query: queryData,
  }).catch(() => {});
});

onBeforeMount(() => {
  bus.$on('toggle-show-view-config', () => {
    if (isViewEditable.value) return;
    showViewConfig.value = !showViewConfig.value;
  });
  initClusterAndViewID();
});

onBeforeUnmount(() => {
  bus.$off('toggle-show-view-config');
});

provide('dashboard-view', {
  reload,
});

defineExpose({
  handleGetQuery,
});
</script>
<script lang="ts">
export default {
  beforeRouteEnter(to, from, next) {
    next((vm) => {
      (vm as any).handleGetQuery(to.query);
    });
  },
};
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
