<template>
  <PopoverSelector
    class="flex items-center"
    trigger="click"
    placement="bottom-end"
    offset="0, 0"
    :on-hide="onHide"
    :on-show="onShow"
    :disabled="isViewEditable"
    theme="popover-select"
    ref="popoverSelectRef">
    <div
      :class="[
        'flex items-center justify-center h-[66px] v-m-view-selector',
        showViewInfo ? 'w-[260px]' : 'w-[60px]'
      ]">
      <!-- 视图信息 -->
      <template v-if="showViewInfo">
        <span
          :class="[
            'flex items-center justify-between flex-1 h-[50px] mx-[12px] px-[12px]',
            'rounded-sm bg-[#E1ECFF]/60',
            isViewEditable ? 'cursor-not-allowed' : 'cursor-pointer'
          ]">
          <span class="flex flex-col">
            <span class="inline-flex items-center justify-between text-[14px] text-[#313238] leading-[22px]">
              <span class="bcs-ellipsis" v-bk-overflow-tips>{{ curViewName || '--' }}</span>
              <i
                class="bcs-icon bcs-icon-alarm-insufficient text-[14px] text-[#FFB848] ml-[5px]"
                v-if="unknownClusterID"
                v-bk-tooltips="$t('view.tips.invalidate')">
              </i>
            </span>
            <span class="text-[12px] text-[#979BA5] leading-[20px]">{{ curViewType }}</span>
          </span>
          <i
            :class="[
              'text-[12px] text-[#979BA5] ml-[6px]',
              'bcs-icon bcs-icon-qiehuan',
              !isHide ? '!text-[#3A84FF]' : ''
            ]">
          </i>
        </span>
      </template>
      <!-- 打开视图配置面板 -->
      <span
        :class="[
          'relative inline-flex items-center justify-center cursor-pointer rounded-sm',
          'border border-solid border-[#C4C6CC] w-[32px] h-[32px]',
          showViewInfo ? 'mr-[16px]' : '',
          isViewConfigShow ? '!border-[#3A84FF] !text-[#3A84FF]' : 'text-[#979BA5] hover:!border-[#979BA5]'
        ]"
        v-bk-trace.click="{
          module: 'view',
          operation: 'filter',
          desc: '视图筛选按钮',
          username: $store.state.user.username,
          projectCode: $store.getters.curProjectCode,
        }"
        @click.stop="toggleViewConfig">
        <i class="bk-icon icon-funnel text-[14px]"></i>
      </span>
    </div>
    <template #content>
      <div class="px-[12px] my-[8px]">
        <div class="bg-[#F0F1F5] rounded-sm h-[32px] p-[4px] flex items-center">
          <div
            v-for="item in modeList"
            :key="item.id"
            :class="[
              'flex items-center justify-center h-[24px] flex-1 leading-[20px] rounded-sm cursor-pointer',
              viewMode === item.id ? 'bg-[#FFFFFF] text-[#3A84FF]' : ''
            ]"
            @click="changeViewMode(item.id)">
            {{ item.name }}
          </div>
        </div>
      </div>
      <!-- 视图模式 -->
      <div class="w-[320px]">
        <!-- 集群模式(集群选择器会自动更新全局集群缓存) -->
        <ClusterSelectPopover
          cluster-type="all"
          :value="clusterID"
          :selectable="isDefaultView"
          :is-show="!isHide"
          v-show="viewMode === 'cluster'"
          @click="changeClusterView"
          @init="changeClusterView" />
        <!-- 自定义视图 -->
        <ViewList v-show="viewMode === 'custom'" @change="changeCustomView" @edit="editCustomView" />
      </div>
      <div
        :class="[
          'bcs-border-top',
          'flex items-center justify-center',
          'mb-[-4px] text-[12px] h-[40px] bg-[#FAFBFD] cursor-pointer'
        ]"
        @click="handleCreateView">
        <i class="bk-icon icon-plus-circle-shape text-[#979BA5] text-[16px] mr-[4px]"></i>
        {{ $t('view.button.addCustomView') }}
      </div>
    </template>
  </PopoverSelector>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';

import useViewConfig from './use-view-config';
import ViewList from './view-list.vue';

import ClusterSelectPopover from '@/components/cluster-selector/cluster-select-popover.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import { useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

const props = defineProps({
  isViewConfigShow: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['toggle-view-config', 'create-new-view', 'edit-view']);

// 视图下拉列表
const isHide = ref(true);
const onHide = () => {
  isHide.value = true;
};
const onShow = () => {
  initViewMode();
  isHide.value = false;
};

// 集群ID
const clusterID = computed(() => {
  let pathClusterID = $router.currentRoute?.params?.clusterId;
  if (pathClusterID === '-') {
    pathClusterID = '';
  }
  return pathClusterID;
});

// 菜单折叠状态
const openSideMenu = computed(() => $store.state.openSideMenu);
// 是否显示视图信息
const showViewInfo = computed(() => openSideMenu.value || props.isViewConfigShow);

// 视图管理
const popoverSelectRef = ref<any>(null);
const {
  isViewEditable,
  isDefaultView,
  curViewType,
  curViewName,
  updateViewIDStore,
  curViewData,
} = useViewConfig();

const { clusterNameMap } = useCluster();
// 校验集群ID正确性
const unknownClusterID = computed(() => curViewData.value?.clusterNamespaces?.some((item) => {
  const { clusterID } = item;
  return !clusterNameMap.value[clusterID];
}));

// 集群列表页
const isDashboardHome = computed(() => $router.currentRoute?.matched?.some(item => item.name === 'dashboardHome'));

// 切换自定义视图
const changeCustomView = async (id: string) => {
  if (!isDashboardHome.value) {
    // 非列表页切换时跳转到列表页
    await $router.push({
      name: 'dashboardWorkloadDeployments',
      params: $router.currentRoute.params,
      query: {
        viewID: id,
      },
    });
  }
  updateViewIDStore(id);
  popoverSelectRef.value?.hide();
};

// 编辑自定义视图
const editCustomView = (id: string) => {
  emits('edit-view', id);
  popoverSelectRef.value?.hide();
};

// 切换集群视图
const changeClusterView = async (clusterID: string) => {
  if (clusterID === $router.currentRoute?.params.clusterId && isDashboardHome.value) {
    popoverSelectRef.value?.hide();
    return;
  };
  // 集群视图设置集群ID参数
  let name;
  if ($router.currentRoute?.name === 'dashboardCustomObjects' || !isDashboardHome.value) {
    // 切换集群时如果在自定义视图菜单就跳到首页(deploy)
    name = 'dashboardWorkloadDeployments';
  } else if ($router.currentRoute?.name) {
    name = $router.currentRoute?.name;
  } else {
    name = '404';
  }

  await $router.replace({
    name,
    params: {
      clusterId: clusterID,
    },
  }).catch(() => {});
  // 清空自定义视图ID
  updateViewIDStore('');// 设置为集群视图
  $store.commit('updateCrdData', {});
  popoverSelectRef.value?.hide();
};

// 显示和隐藏视图
const toggleViewConfig = () => {
  emits('toggle-view-config');
  popoverSelectRef.value?.hide();
};

// 切换集群视图 & 自定视图
const modeList = ref([
  {
    id: 'cluster',
    name: $i18n.t('view.button.clusterView'),
  },
  {
    id: 'custom',
    name: $i18n.t('view.button.customView'),
  },
]);
const viewMode = ref<'custom'|'cluster'>();
const changeViewMode = (mode) => {
  viewMode.value = mode;
};

// 创建视图
const handleCreateView = () => {
  emits('create-new-view');
  popoverSelectRef.value?.hide();
};

// 初始化视图模式
const initViewMode = () => {
  viewMode.value = isDefaultView.value ? 'cluster' : 'custom';
};

watch(isDefaultView, () => {
  initViewMode();
});
onBeforeMount(() => {
  initViewMode();
});
</script>
