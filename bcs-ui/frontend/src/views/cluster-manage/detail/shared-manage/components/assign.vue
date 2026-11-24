<template>
  <div class="ns-assign py-[16px] px-[20px] border-[#dcdee5] border border-t-0">
    <bcs-alert type="info" :title="$t('cluster.sharedManage.assign.alert')" class="!mb-[16px]">
    </bcs-alert>

    <!-- 工具栏：左侧操作 + 已选统计，右侧分类筛选 / 归属项目 / 搜索
         （按钮间距由全局 .bk-button+.bk-button 提供，不再叠加 flex gap） -->
    <div class="flex items-center justify-between flex-wrap mb-[16px] gap-[8px]">
      <div class="flex items-center">
        <bcs-button theme="primary" @click="createShow = true">
          <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.toolbar.create') }}</span>
        </bcs-button>
        <bcs-button
          :disabled="!selection.length"
          v-bk-tooltips="{
            content: $t('cluster.sharedManage.assign.msg.selectFirst'),
            disabled: !!selection.length,
          }"
          @click="openAssign(selection, 'assign')">
          <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.toolbar.assign') }}</span>
        </bcs-button>
        <bcs-button
          :disabled="!canRecycle"
          v-bk-tooltips="{ content: recycleTip, disabled: !recycleTip }"
          @click="openRecycle(selection)">
          <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.toolbar.recycle') }}</span>
        </bcs-button>
        <template v-if="selection.length">
          <span class="ml-[8px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.assign.toolbar.selected', { num: selection.length }) }}
          </span>
          <bcs-button text theme="primary" class="ml-[8px]" @click="handleClearSelection">
            <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.toolbar.clear') }}</span>
          </bcs-button>
        </template>
      </div>
      <div class="flex items-center flex-wrap gap-[8px]">
        <bk-radio-group v-model="category">
          <bk-radio-button value="all">
            {{ $t('cluster.sharedManage.assign.filter.all') }} {{ counts.all }}
          </bk-radio-button>
          <bk-radio-button value="unassigned">
            {{ $t('cluster.sharedManage.assign.filter.unassigned') }} {{ counts.unassigned }}
          </bk-radio-button>
          <bk-radio-button value="assigned">
            {{ $t('cluster.sharedManage.assign.filter.assigned') }} {{ counts.assigned }}
          </bk-radio-button>
          <bk-radio-button value="locked">
            {{ $t('cluster.sharedManage.assign.filter.locked') }} {{ counts.locked }}
          </bk-radio-button>
        </bk-radio-group>
        <bcs-select
          class="w-[150px]"
          :value="projectFilter"
          clearable
          searchable
          :placeholder="$t('cluster.sharedManage.assign.placeholder.project')"
          @change="handleProjectFilter"
          @clear="handleProjectFilter('')">
          <bcs-option
            v-for="(name, code) in ownerProjectMap"
            :key="code"
            :id="code"
            :name="name" />
        </bcs-select>
        <bcs-input
          class="!w-[200px]"
          :value="keyword"
          clearable
          :placeholder="$t('cluster.sharedManage.assign.placeholder.search')"
          @input="handleKeyword" />
      </div>
    </div>

    <!-- 命名空间表格（全量数据前端过滤 + 分页） -->
    <bcs-table
      ref="tableRef"
      v-bkloading="{ isLoading: loading || innerLoading }"
      :data="curPageData"
      :pagination="tablePagination"
      :row-class-name="rowClassName"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange"
      @selection-change="handleSelectionChange">
      <bcs-table-column type="selection" width="44" :selectable="isSelectable" />
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.name')" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <span :class="lockReason(row) ? 'text-[#C4C6CC]' : 'text-[#313238]'">{{ row.name }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.owner')" width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <div v-if="row.projectCode" class="flex items-center gap-[6px] min-w-0">
            <span class="truncate" :class="lockReason(row) ? 'text-[#C4C6CC]' : 'text-[#313238]'">
              {{ projectName(row.projectCode) }}
            </span>
            <span class="shrink-0 text-[12px] text-[#979BA5]">{{ row.projectCode }}</span>
            <bcs-tag
              v-if="row.projectID === clusterData.projectID"
              theme="info"
              class="shrink-0 !h-[20px] !leading-[20px]">
              {{ $t('cluster.sharedManage.assign.tag.ownerProject') }}
            </bcs-tag>
          </div>
          <bcs-tag v-else class="!h-[20px] !leading-[20px] !m-[0px]">
            {{ $t('cluster.sharedManage.assign.tag.unassigned') }}
          </bcs-tag>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.quota')" width="130" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.quota" class="text-[#63656E]">{{ quotaText(row) }}</span>
          <span v-else class="text-[#979BA5]">{{ $t('cluster.sharedManage.assign.quota.unset') }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.used')" width="130" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.quota && row.used" class="text-[#63656E]">{{ usedText(row) }}</span>
          <span v-else class="text-[#979BA5]">--</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.source')" width="110" show-overflow-tooltip>
        <template #default="{ row }">{{ sourceText(row) }}</template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.updateTime')" width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="text-[#63656E]">{{ timeText(row.updateTime || row.createTime) }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.assign.table.actions')" width="160">
        <template #default="{ row }">
          <div class="flex items-center gap-[12px]">
            <!-- 创建审批中：ITSM 审批单状态非空即审批中，「单据」跳转 ITSM 审批页 -->
            <template v-if="row.itsmTicketStatus">
              <bcs-tag
                v-bk-tooltips="row.disableReason || $t('cluster.sharedManage.assign.status.pendingReview')"
                theme="warning"
                class="!h-[22px] !leading-[22px] !m-[0px]">
                {{ $t('cluster.sharedManage.assign.tag.pendingReview') }}
              </bcs-tag>
              <bcs-button
                v-if="row.itsmTicketURL"
                text
                theme="primary"
                @click="handleToITSM(row.itsmTicketURL)">
                {{ $t('cluster.sharedManage.assign.op.ticket') }}
              </bcs-button>
            </template>
            <span
              v-else-if="lockReason(row)"
              v-bk-tooltips="lockReason(row)"
              class="text-[#979BA5]">
              {{ $t('cluster.sharedManage.assign.op.locked') }}
              <i class="bcs-icon bcs-icon-exclamation-circle ml-[2px] align-[1px]"></i>
            </span>
            <template v-else-if="!row.projectCode">
              <bcs-button text theme="primary" @click="openAssign([row], 'assign')">
                {{ $t('cluster.sharedManage.assign.op.assign') }}
              </bcs-button>
            </template>
            <template v-else>
              <bcs-button text theme="primary" @click="openAssign([row], 'transfer')">
                {{ $t('cluster.sharedManage.assign.op.transfer') }}
              </bcs-button>
              <bcs-button text theme="primary" @click="openRecycle([row])">
                {{ $t('cluster.sharedManage.assign.op.recycle') }}
              </bcs-button>
            </template>
          </div>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="hasFilter ? 'search-empty' : 'empty'" @clear="handleClearFilter" />
      </template>
    </bcs-table>

    <!-- 回收确认（复用通用二次确认弹窗） -->
    <SharedConfirmDialog
      :value="recycleShow"
      :title="$t('cluster.sharedManage.assign.recycle.title', { num: recycleItems.length })"
      :confirm-text="$t('cluster.sharedManage.assign.op.recycle')"
      :namespaces="recycleNsItems"
      @input="recycleShow = $event"
      @confirm="confirmRecycle">
      {{ recycleBody }}<b>{{ recycleBodyKeep }}</b>。
      <template #append>
        <bcs-checkbox class="mt-[10px]" :value="deleteQuota" @change="deleteQuota = $event">
          <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.recycle.deleteQuota') }}</span>
        </bcs-checkbox>
      </template>
    </SharedConfirmDialog>

    <!-- 分配 / 变更归属侧边栏 -->
    <AssignSlider
      :value="assignShow"
      :cluster-id="clusterData.clusterID"
      :items="assignItems"
      :projects="assignableProjects"
      :project-map="projectMap"
      :mode="assignMode"
      @input="assignShow = $event"
      @refresh="handleRefresh"
      @go-settings="$emit('go-settings')" />

    <!-- 创建命名空间侧边栏 -->
    <CreateNamespaceSlider
      :value="createShow"
      :cluster-id="clusterData.clusterID"
      :projects="assignableProjects"
      :project-map="projectMap"
      :existing-names="namespaceList.map(ns => ns.name)"
      @input="createShow = $event"
      @refresh="handleRefresh"
      @go-settings="$emit('go-settings')" />
  </div>
</template>

<script lang="ts" setup>
import { computed, getCurrentInstance, ref, watch } from 'vue';

import AssignSlider from './assign-slider.vue';
import SharedConfirmDialog from './confirm-dialog.vue';
import CreateNamespaceSlider from './create-namespace-slider.vue';

import { getSharedprojects } from '@/api/modules/cluster-manager';
import {
  INativeNamespace,
  parseCpu,
  parseMemoryGiB,
  recycleNamespace,
} from '@/api/modules/shared-manage';
import $bkMessage from '@/common/bkmagic';
import { formatTimeWithTimezone } from '@/common/util';
import { ICluster, IProject } from '@/composables/use-app';
import usePageConf from '@/composables/use-page';
import $router from '@/router';
import useProjects from '@/views/project-manage/project/use-project';

// 命名空间分配：提示条 + 工具栏 + 表格 + 分页，交互对齐原型 assign 页
const props = defineProps({
  clusterData: {
    type: Object as () => Partial<ICluster>,
    default: () => ({}),
  },
  namespaceList: {
    type: Array as () => INativeNamespace[],
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['refresh', 'go-settings']);

const { getProjectList } = useProjects();

// 项目名回显：全量项目 + 可见范围项目
const fullProjectList = ref<IProject[]>([]);
const sharedProjectList = ref<IProject[]>([]);
const projectMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {};
  [...fullProjectList.value, ...sharedProjectList.value].forEach((project) => {
    map[project.projectCode] = project.name;
  });
  return map;
});
const projectName = (code: string) => projectMap.value[code] || code;

// 分页拉取全量项目（项目名回显）
const initProjectList = async () => {
  const list: IProject[] = [];
  let offset = 0;
  let total = 0;
  do {
    const { data } = await getProjectList({ offset, limit: 100 })
      .catch(() => ({ data: { results: [] as IProject[], total: 0 } }));
    list.push(...(data?.results || []));
    total = data?.total || 0;
    offset += 100;
  } while (offset < total);
  fullProjectList.value = list;
};
// 可见范围项目（GetClusterSharedProject）
const initSharedProjectList = async () => {
  if (!props.clusterData?.clusterID) return;
  sharedProjectList.value = await getSharedprojects({
    $clusterId: props.clusterData.clusterID,
  }).catch(() => [] as IProject[]);
};

// 可分配的目标项目：可见范围内的项目（排除集群所属项目）
// 存量集群（未转为指定项目）接口返回为空时回退为全量项目
const assignableProjects = computed(() => {
  const list = sharedProjectList.value.length ? sharedProjectList.value : fullProjectList.value;
  return list.filter(project => project.projectID && project.projectID !== props.clusterData?.projectID);
});

// 锁定原因：优先展示后端 disableReason，缺省时按前端兜底文案
// 审批中（itsmTicketStatus 非空）后端会返回 allocatable = false
const lockReason = (ns: INativeNamespace) => {
  if (ns.allocatable === false) {
    return ns.disableReason || window.i18n.t('cluster.sharedManage.assign.status.unallocatable');
  }
  return '';
};
// 分类（设计稿 §5.3.2）：全部不过滤；未分配 = 未归属且可分配；已分配 = 有归属；不可操作 = 未归属且不可分配
// 兼容后端扩展字段未上线：allocatable 缺省视为可分配
type ICategory = 'all' | 'unassigned' | 'assigned' | 'locked';
const categoryOf = (ns: INativeNamespace): ICategory => {
  if (ns.projectCode) return 'assigned';
  if (ns.allocatable === false) return 'locked';
  return 'unassigned';
};

// 筛选状态（分类与 URL query.category 双向同步，刷新后保持）
// 本地变更写回路由、路由外部变更（如「去分配」）同步到本地，双向守卫防回环与重复导航
const VALID_CATEGORIES: ICategory[] = ['all', 'unassigned', 'assigned', 'locked'];
const category = ref<ICategory>('all');
watch(category, (value) => {
  const target = value === 'all' ? undefined : value;
  if ($router.currentRoute.query.category === target) return;
  $router.replace({
    query: {
      ...$router.currentRoute.query,
      category: target,
    },
  }).catch(() => {});
});
// 组件实例的 $route 保证响应式（$router.currentRoute 在部分 vue-router 3.x 版本为 getter，watch 不到）
const instance = getCurrentInstance();
watch(() => instance?.proxy?.$route.query.category, (value) => {
  const cat = VALID_CATEGORIES.includes(value as ICategory) ? (value as ICategory) : 'all';
  if (cat !== category.value) category.value = cat;
}, { immediate: true });
const projectFilter = ref('');
const keyword = ref('');

const counts = computed(() => {
  const result = { all: props.namespaceList.length, unassigned: 0, assigned: 0, locked: 0 };
  props.namespaceList.forEach((ns) => {
    result[categoryOf(ns)] += 1;
  });
  return result;
});
const filtered = computed(() => {
  const kw = keyword.value.trim();
  return props.namespaceList.filter(ns => (
    (category.value === 'all' || categoryOf(ns) === category.value)
    && (!projectFilter.value || ns.projectCode === projectFilter.value)
    && (!kw || ns.name.includes(kw))
  ));
});
const hasFilter = computed(() => category.value !== 'all' || !!projectFilter.value || !!keyword.value.trim());

// 前端分页：复用项目标准 usePageConf，分页器由表格内建渲染
const {
  curPageData,
  pagination,
  pageChange,
  pageSizeChange,
  handleResetPage,
} = usePageConf(filtered, { current: 1, limit: 10 });
const tablePagination = computed(() => ({ ...pagination.value, limitList: [10, 20, 50, 100] }));

// 筛选条件变化时回到第一页并清空勾选（勾选仅保留当前页）
watch([category, projectFilter, keyword], () => {
  handleResetPage();
  handleClearSelection();
});

const handleKeyword = (value: string) => {
  keyword.value = value;
};
const handleProjectFilter = (value: string) => {
  projectFilter.value = value || '';
};
const handleClearFilter = () => {
  category.value = 'all';
  projectFilter.value = '';
  keyword.value = '';
};

// 归属项目筛选项（列表内归属项目去重）
const ownerProjectMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {};
  props.namespaceList.forEach((ns) => {
    if (ns.projectCode) map[ns.projectCode] = projectName(ns.projectCode);
  });
  return map;
});

// 勾选
const tableRef = ref();
const selection = ref<INativeNamespace[]>([]);
const handleSelectionChange = (rows: INativeNamespace[]) => {
  selection.value = rows || [];
};
const handleClearSelection = () => {
  tableRef.value?.clearSelection?.();
  selection.value = [];
};
const isSelectable = (row: INativeNamespace) => !lockReason(row);
const rowClassName = ({ row }: { row: INativeNamespace }) => (lockReason(row) ? 'locked-row' : '');

// 回收：仅当所选全部有归属时可用
const canRecycle = computed(() => selection.value.length > 0
  && selection.value.every(ns => !!ns.projectCode));
const recycleTip = computed(() => {
  if (!selection.value.length) return window.i18n.t('cluster.sharedManage.assign.msg.selectFirst');
  if (!canRecycle.value) return window.i18n.t('cluster.sharedManage.assign.msg.recycleMixed');
  return '';
});

// 配额 / 使用量文案（CPU 支持 "8"/"100m"，内存支持 "16Gi"/"128Mi"，统一换算）
const formatNum = (num: number) => (Number.isInteger(num) ? num : num.toFixed(1));
const unitCpu = () => window.i18n.t('cluster.sharedManage.assign.unit.cpu');
const quotaText = (ns: INativeNamespace) => `${formatNum(parseCpu(ns.quota?.cpuLimits))} ${unitCpu()} / ${formatNum(parseMemoryGiB(ns.quota?.memoryLimits))} GiB`;
const usedText = (ns: INativeNamespace) => `${formatNum(parseCpu(ns.used?.cpuRequests))} ${unitCpu()} / ${formatNum(parseMemoryGiB(ns.used?.memoryRequests))} GiB`;

// 时间格式化（ISO 时间转浏览器时区，与项目其他表格一致）
const timeText = (time?: string) => (time ? formatTimeWithTimezone(time, '', 'YYYY-MM-DD HH:mm:ss') : '--');

// 跳转 ITSM 审批单（新开页，同 namespace.vue handleToITSM）
const handleToITSM = (link: string) => {
  window.open(link, '_blank');
};

// 回收弹窗描述（"不会被删除"加粗）
const recycleBody = window.i18n.t('cluster.sharedManage.assign.recycle.body');
const recycleBodyKeep = window.i18n.t('cluster.sharedManage.assign.recycle.bodyKeep');

// 来源文案（后端字段未上线前展示 --）
const SOURCE_KEYS: Record<string, string> = {
  create: 'cluster.sharedManage.assign.source.create',
  assign: 'cluster.sharedManage.assign.source.assign',
  apply: 'cluster.sharedManage.assign.source.apply',
  self: 'cluster.sharedManage.assign.source.self',
};
const sourceText = (ns: INativeNamespace) => {
  if (!ns.source) return '--';
  return SOURCE_KEYS[ns.source] ? window.i18n.t(SOURCE_KEYS[ns.source]) : ns.source;
};

// 分配 / 变更归属
const assignShow = ref(false);
const assignMode = ref('assign');
const assignItems = ref<INativeNamespace[]>([]);
const openAssign = (items: INativeNamespace[], mode: string) => {
  assignItems.value = items;
  assignMode.value = mode;
  assignShow.value = true;
};

// 创建命名空间
const createShow = ref(false);

// 回收确认
const recycleShow = ref(false);
const recycleItems = ref<INativeNamespace[]>([]);
const deleteQuota = ref(false);
const recycleNsItems = computed(() => recycleItems.value.map(ns => ({
  name: ns.name,
  project: projectName(ns.projectCode) || ns.projectCode,
  tag: ns.itsmTicketStatus
    ? window.i18n.t('cluster.sharedManage.assign.tag.pendingReview')
    : '',
})));
const openRecycle = (items: INativeNamespace[]) => {
  if (!items.length) return;
  recycleItems.value = items;
  deleteQuota.value = false;
  recycleShow.value = true;
};
// 回收 = 解除归属，逐个调接口（部分失败时提示，不回滚已成功的）
const confirming = ref(false);
const confirmRecycle = async () => {
  if (confirming.value) return;
  confirming.value = true;
  const results = await Promise.all(recycleItems.value.map(ns => recycleNamespace({
    $clusterId: props.clusterData?.clusterID,
    $namespace: ns.name,
    deleteQuota: deleteQuota.value,
  }).then(() => true)
    .catch(() => false)));
  confirming.value = false;
  const okCount = results.filter(Boolean).length;
  if (okCount) {
    $bkMessage({
      theme: 'success',
      message: window.i18n.t('cluster.sharedManage.assign.msg.recycleSuccess', { num: okCount }),
    });
    recycleShow.value = false;
    handleClearSelection();
    emits('refresh');
  } else {
    $bkMessage({
      theme: 'error',
      message: window.i18n.t('cluster.sharedManage.assign.msg.recycleFailed'),
    });
  }
};

// 操作成功后刷新列表
const handleRefresh = () => {
  emits('refresh');
};

// 组件内初始化请求（项目名回显 / 可见范围）聚合 loading（与列表 loading 叠加展示）
const innerLoading = ref(false);
initLoading();
async function initLoading() {
  innerLoading.value = true;
  await Promise.all([initProjectList(), initSharedProjectList()]);
  innerLoading.value = false;
}
watch(() => props.clusterData?.clusterID, () => initSharedProjectList(), { immediate: true });
</script>

<style lang="postcss" scoped>
/* 锁定行置灰：背景 + 全部文字 #c4c6cc（覆盖单元格内 span / tag 的字色） */
:deep(.bk-table .locked-row) {
  background-color: #fafbfd;

  .cell,
  .cell>span,
  .cell>div {
    color: #c4c6cc;
  }
}

.ns-assign {
  :deep(.bk-form-control) {
    width: auto;
    white-space: nowrap;

    .bk-radio-button-text {
      font-size: 12px;
    }
  }

  :deep(.bk-form-radio-button:nth-child(n+1)) {
    margin-left: -2px;
  }

  :deep(.bk-select .bk-select-name) {
    height: 28px;
  }
}
</style>
