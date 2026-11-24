<template>
  <div class="shared-manage">
    <!-- 基本信息：共享状态 / 可见范围 / 使用项目 / 已分配给其他项目 / 未分配 -->
    <div
      class="min-w-[700px] flex py-[14px] px-[20px] bg-white border border-[#DCDEE5] rounded-[2px] mb-[16px] divide-x divide-[#f0f1f5] shadow-[0_2px_4px_rgba(25,25,41,.05)] overflow-hidden"
      v-bkloading="{ isLoading: loading, opacity: 1 }">
      <div class="flex-1 min-w-fit first:pl-0 px-[20px] flex flex-col gap-[4px] whitespace-nowrap">
        <span class="text-[12px] text-[#979BA5]">{{ $t('cluster.sharedManage.summary.sharedStatus') }}</span>
        <span class="flex items-center gap-[6px] text-[18px] leading-[26px] font-semibold text-[#313238]">
          <i :class="['w-[8px] h-[8px] rounded-full inline-block shrink-0', isShared ? 'bg-[#2DCB56]' : 'bg-[#C4C6CC]']"></i>
          {{ isShared ? $t('cluster.sharedManage.summary.sharedOn') : $t('cluster.sharedManage.summary.sharedOff') }}
        </span>
      </div>
      <div class="flex-1 min-w-fit first:pl-0 px-[20px] flex flex-col gap-[4px] whitespace-nowrap">
        <span class="text-[12px] text-[#979BA5]">{{ $t('cluster.sharedManage.summary.visibleRange') }}</span>
        <span class="flex items-center gap-[6px] text-[18px] leading-[26px] font-semibold text-[#313238]">
          {{ rangeText }}
        </span>
      </div>
      <div class="flex-1 min-w-fit first:pl-0 px-[20px] flex flex-col gap-[4px] whitespace-nowrap">
        <span class="text-[12px] text-[#979BA5]">{{ $t('cluster.sharedManage.summary.usingProjects') }}</span>
        <span class="flex items-baseline gap-[6px] text-[18px] leading-[26px] font-semibold text-[#313238]">
          {{ usingProjectCount }}
          <small class="text-[12px] font-normal text-[#979BA5]">{{ $t('cluster.sharedManage.summary.unit') }}</small>
        </span>
      </div>
      <div class="flex-1 min-w-fit first:pl-0 px-[20px] flex flex-col gap-[4px] whitespace-nowrap">
        <span class="text-[12px] text-[#979BA5]">{{ $t('cluster.sharedManage.summary.assignedToOthers') }}</span>
        <span class="flex items-baseline gap-[6px] text-[18px] leading-[26px] font-semibold text-[#313238]">
          {{ assignedToOthersCount }}
          <small class="text-[12px] font-normal text-[#979BA5]">
            {{ $t('cluster.sharedManage.summary.unitNamespace') }}
          </small>
        </span>
      </div>
      <div class="flex-1 min-w-fit first:pl-0 px-[20px] flex flex-col gap-[4px] whitespace-nowrap">
        <span class="text-[12px] text-[#979BA5]">{{ $t('cluster.sharedManage.summary.unassigned') }}</span>
        <span
          :class="[
            'flex items-baseline gap-[6px] text-[18px] leading-[26px] font-semibold',
            unassignedCount ? 'text-[#FF9C01]' : 'text-[#313238]',
          ]">
          {{ unassignedCount }}
          <small class="text-[12px] font-normal text-[#979BA5]">
            {{ $t('cluster.sharedManage.summary.unitNamespace') }}
          </small>
          <bcs-button
            v-if="unassignedCount"
            class="!text-[12px] ml-[6px] shrink-0"
            text
            @click="handleGoAssign">
            {{ $t('cluster.sharedManage.summary.goAssign') }}
          </bcs-button>
        </span>
      </div>
    </div>

    <!-- 子页签：共享设置 / 命名空间分配 / 使用概览 -->
    <bcs-tab class="shared-manage-tab bg-white" :active.sync="activeSub" type="border-card" :label-height="42">
      <bcs-tab-panel
        name="settings"
        :label="$t('cluster.sharedManage.tabs.settings')"
        render-directive="if">
        <SharedSettings
          :cluster-data="clusterData"
          :namespace-list="namespaceList"
          @refresh="handleRefresh"
          @go-recycle="handleGoAssign" />
      </bcs-tab-panel>
      <!-- todo: 使用概览 后续实现 -->
      <bcs-tab-panel name="assign" :label="$t('cluster.sharedManage.tabs.assign')" render-directive="if">
        <SharedAssign
          :cluster-data="clusterData"
          :namespace-list="namespaceList"
          :loading="loading"
          @refresh="handleRefresh"
          @go-settings="activeSub = 'settings'" />
      </bcs-tab-panel>
      <bcs-tab-panel name="overview" :label="$t('cluster.sharedManage.tabs.overview')" render-directive="if">
        <SharedOverview
          :cluster-data="clusterData"
          :namespace-list="namespaceList"
          :loading="loading" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>

<script lang="ts" setup>
import { computed, getCurrentInstance, onMounted, ref } from 'vue';

import SharedAssign from './components/assign.vue';
import SharedOverview from './components/overview.vue';
import SharedSettings from './components/settings.vue';

import { clusterDetail } from '@/api/modules/cluster-manager';
import { INativeNamespace, listNamespaces } from '@/api/modules/shared-manage';
import { ICluster, useProject } from '@/composables/use-app';
import $router from '@/router';

// 共享管理页面
// 子页签：共享设置 / 命名空间分配 / 使用概览（均已实现）
const props = defineProps({
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
});

const { curProject } = useProject();

// 子页签激活项（共享设置 / 命名空间分配 / 使用概览）
// 完全由路由 query.sub 驱动（单一数据源）：get 从响应式 $route 读取，set 写回路由
// 离开分配页签时顺带清除分类参数；query 已一致时跳过，避免重复导航
const instance = getCurrentInstance();
const currentQuery = () => instance?.proxy?.$route.query || {};
const VALID_SUB_TABS = ['settings', 'assign', 'overview'];
const activeSub = computed({
  get: () => {
    const sub = currentQuery().sub as string;
    return VALID_SUB_TABS.includes(sub) ? sub : 'settings';
  },
  set: (value: string) => {
    const query = currentQuery();
    const targetSub = value === 'settings' ? undefined : value;
    const targetCategory = value === 'assign' ? query.category : undefined;
    if (query.sub === targetSub && query.category === targetCategory) return;
    $router.replace({
      query: { ...query, sub: targetSub, category: targetCategory },
    }).catch(() => {});
  },
});

// 共享状态 / 可见范围：与基本信息页保持一致，使用 clusterDetail（GetCluster）接口
// （clusterList 中的数据不含 sharedRanges 或已过期）
const clusterData = ref<Partial<ICluster>>({});
const isShared = computed(() => !!clusterData.value?.is_shared);

// 可见范围文案：指定项目 N 个项目 / 按业务 N 个业务 / 未配置即全平台（存量旧规则）
const rangeText = computed(() => {
  if (!isShared.value) return '';
  const ranges = clusterData.value?.sharedRanges;
  if (ranges?.projectIdOrCodes?.length) {
    return `${ranges.projectIdOrCodes.length} ${window.i18n.t('cluster.sharedManage.summary.unitProject')}`;
  }
  if (ranges?.bizs?.length) {
    return `${ranges.bizs.length} ${window.i18n.t('cluster.sharedManage.summary.unitBiz')}`;
  }
  return window.i18n.t('cluster.sharedManage.summary.wholePlatform');
});

const fetchClusterDetail = async () => {
  if (!props.clusterId) return;
  const res = await clusterDetail({
    $clusterId: props.clusterId,
    projectId: curProject.value?.projectID,
  }, { needRes: true, cancelWhenRouteChange: false }).catch(() => ({}));
  clusterData.value = res.data || {};
};

const namespaceList = ref<INativeNamespace[]>([]);
const loading = ref(false);

// 集群全量命名空间（含已分配 / 未分配）
// projectIDOrCode 传 '-'（集群视角）：共享集群归属项目与当前项目不一致时，
// 避免被权限层按归属项目过滤导致数据为空
const fetchNativeNamespaces = async () => {
  if (!props.clusterId) return;
  loading.value = true;
  const { data } = await listNamespaces(
    { $projectCode: '-', $clusterId: props.clusterId },
    { needRes: true, cancelWhenRouteChange: false },
  ).catch(() => ({ data: [] as INativeNamespace[] }));
  namespaceList.value = data || [];
  loading.value = false;
};

// 已分配给其他项目（归属项目非集群所属项目）的命名空间
const assignedToOthers = computed(() => namespaceList.value
  .filter(ns => ns.projectCode && ns.projectCode !== curProject.value?.projectCode));
const assignedToOthersCount = computed(() => assignedToOthers.value.length);
// 使用项目数（已分配命名空间归属项目去重）
const usingProjectCount = computed(() => new Set(assignedToOthers.value.map(ns => ns.projectCode)).size);
// 未分配（无归属且可操作）的命名空间数
// 兼容后端扩展字段 allocatable 未上线的场景：字段缺省视为可分配
const unassignedCount = computed(() => namespaceList.value
  .filter(ns => !ns.projectCode && ns.allocatable !== false).length);

// 切换到「命名空间分配」子页签（基本信息卡片「去分配」/ 共享设置「去回收」入口）
// 同时定位到「未分配」分类；activeSub 由路由驱动，导航确认后页签自动切换
const handleGoAssign = () => {
  $router.replace({
    query: { ...currentQuery(), sub: 'assign', category: 'unassigned' },
  }).catch(() => {});
};

// 共享设置保存后刷新数据
const handleRefresh = () => {
  fetchClusterDetail();
  fetchNativeNamespaces();
};

onMounted(() => {
  fetchClusterDetail();
  fetchNativeNamespaces();
});
</script>

<style lang="postcss" scoped>
.shared-manage {
  min-height: 200px;
}

/* 本页嵌套在外层 .cluster-detail-tab 内，其深度作用域样式（bk-tab-header 的
   padding-left: 24px、bk-tab-section 的 padding: 0 / 固定高度）会级联命中本页的
   bcs-tab，这里还原组件库 border-card 默认样式（header 无内边距、section 20px） */
:deep(.shared-manage-tab) {
  .bk-tab-header {
    padding-left: 0 !important;
  }

  .bk-tab-label-wrapper .bk-tab-label-list .bk-tab-label-item {
    border-right: 1px solid #dcdee5 !important;

    &.active {
      border-radius: 0;
    }

    &::after,
    &::before {
      width: 0 !important;
    }
  }


  .bk-tab-section {
    height: auto;
    padding: 20px;
    overflow: visible;

    .bk-tab-content {
      height: auto;
    }
  }
}
</style>
