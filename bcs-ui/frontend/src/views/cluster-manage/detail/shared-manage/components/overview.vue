<template>
  <div v-bkloading="{ isLoading: loading || innerLoading, opacity: innerLoading ? 1 : 0.6 }" class="py-[16px] px-[20px] border-[#dcdee5] border border-t-0">
    <!-- 无配额警告条 -->
    <bcs-alert
      v-if="noQuotaCount"
      type="warning"
      class="!mb-[16px]"
      :title="$t('cluster.sharedManage.overview.noQuotaAlert', { num: noQuotaCount })">
    </bcs-alert>

    <!-- 统计卡片：使用项目 / 已分配命名空间 / CPU 配额 / 内存配额 -->
    <div class="grid grid-cols-4 gap-[16px] mb-[16px]">
      <div class="bg-[#FAFBFD] border border-[#F0F1F5] rounded-[2px] px-[16px] py-[14px]">
        <p class="text-[12px] text-[#63656E]">{{ $t('cluster.sharedManage.overview.stat.projects') }}</p>
        <p class="mt-[4px] text-[#313238]">
          <span class="text-[24px] leading-[32px] font-semibold">{{ list.length }}</span>
          <span class="ml-[6px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.stat.ofRange', { num: rangeCount }) }}
          </span>
        </p>
      </div>
      <div class="bg-[#FAFBFD] border border-[#F0F1F5] rounded-[2px] px-[16px] py-[14px]">
        <p class="text-[12px] text-[#63656E]">{{ $t('cluster.sharedManage.overview.stat.namespaces') }}</p>
        <p class="mt-[4px] text-[#313238]">
          <span class="text-[24px] leading-[32px] font-semibold">{{ total.ns }}</span>
          <span class="ml-[6px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.stat.clusterTotal', { num: namespaceList.length }) }}
          </span>
        </p>
      </div>
      <div class="bg-[#FAFBFD] border border-[#F0F1F5] rounded-[2px] px-[16px] py-[14px]">
        <p class="text-[12px] text-[#63656E]">{{ $t('cluster.sharedManage.overview.stat.cpuQuota') }}</p>
        <p class="mt-[4px] text-[#313238]">
          <span class="text-[24px] leading-[32px] font-semibold">{{ formatNum(total.cpu) }}</span>
          <span class="ml-[6px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.stat.cpuOf', {
              num: formatNum(clusterCpuTotal), pct: pct(total.cpu, clusterCpuTotal) }) }}
          </span>
        </p>
        <div class="mt-[8px] h-[6px] rounded-[3px] bg-[#F0F1F5] overflow-hidden">
          <i
            class="block h-full rounded-[3px] bg-[#3A84FF]"
            :style="{ width: `${barWidth(pct(total.cpu, clusterCpuTotal))}%` }"></i>
        </div>
      </div>
      <div class="bg-[#FAFBFD] border border-[#F0F1F5] rounded-[2px] px-[16px] py-[14px]">
        <p class="text-[12px] text-[#63656E]">{{ $t('cluster.sharedManage.overview.stat.memQuota') }}</p>
        <p class="mt-[4px] text-[#313238]">
          <span class="text-[24px] leading-[32px] font-semibold">{{ formatNum(total.mem) }}</span>
          <span class="ml-[6px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.stat.memOf', {
              num: formatNum(clusterMemTotal), pct: pct(total.mem, clusterMemTotal) }) }}
          </span>
        </p>
        <div class="mt-[8px] h-[6px] rounded-[3px] bg-[#F0F1F5] overflow-hidden">
          <i
            class="block h-full rounded-[3px] bg-[#3A84FF]"
            :style="{ width: `${barWidth(pct(total.mem, clusterMemTotal))}%` }"></i>
        </div>
      </div>
    </div>

    <!-- 使用项目表格 -->
    <bcs-table :data="list">
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.project')" min-width="220">
        <template #default="{ row }">
          <div class="flex items-center gap-[6px] min-w-0">
            <span class="truncate text-[#313238]">{{ projectName(row.projectCode) }}</span>
            <span class="shrink-0 text-[12px] text-[#979BA5]">{{ row.projectCode }}</span>
            <bcs-tag
              v-if="!inRange(row.projectCode)"
              theme="danger"
              class="shrink-0 !h-[20px] !leading-[20px]">
              {{ $t('cluster.sharedManage.overview.tag.outOfRange') }}
            </bcs-tag>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.namespaces')" width="120">
        <template #default="{ row }">
          <span class="text-[#63656E]">
            {{ row.items.length }} {{ $t('cluster.sharedManage.overview.unit.nsCount') }}
          </span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.cpu')" width="160">
        <template #default="{ row }">
          <span v-if="row.cpu" class="text-[#63656E]">
            {{ row.ucpu.toFixed(1) }} / {{ formatNum(row.cpu) }} {{ $t('cluster.sharedManage.overview.unit.core') }}
          </span>
          <span v-else class="text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.unset') }}
          </span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.mem')" width="180">
        <template #default="{ row }">
          <span v-if="row.mem" class="text-[#63656E]">
            {{ row.umem.toFixed(1) }} / {{ formatNum(row.mem) }} GiB
          </span>
          <span v-else class="text-[#979BA5]">
            {{ $t('cluster.sharedManage.overview.unset') }}
          </span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.cpuRate')" width="200">
        <template #default="{ row }">
          <div v-if="row.cpu" class="flex items-center gap-[8px]">
            <div class="flex-1 h-[6px] rounded-[3px] bg-[#F0F1F5] overflow-hidden">
              <i
                class="block h-full rounded-[3px]"
                :class="pct(row.ucpu, row.cpu) > 80 ? 'bg-[#FF9C01]' : 'bg-[#3A84FF]'"
                :style="{ width: `${barWidth(pct(row.ucpu, row.cpu))}%` }"></i>
            </div>
            <span class="w-[38px] text-[12px] text-[#63656E]">{{ pct(row.ucpu, row.cpu) }}%</span>
          </div>
          <span v-else class="text-[#979BA5]">--</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.actions')" width="120">
        <template #default="{ row }">
          <bcs-button text theme="primary" @click="openProjectNs(row)">
            {{ $t('cluster.sharedManage.overview.table.viewNs') }}
          </bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus type="empty" :title="$t('cluster.sharedManage.overview.empty.title')" />
      </template>
    </bcs-table>

    <!-- 查看命名空间弹窗（封装组件） -->
    <ProjectNsDialog
      :value="nsShow"
      :cluster-name="clusterData.clusterName || ''"
      :project="{
        name: projectName(currentGroup?.projectCode || ''),
        code: currentGroup?.projectCode || '',
      }"
      :items="currentGroup?.items || []"
      @input="nsShow = $event" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import ProjectNsDialog from './project-ns-dialog.vue';

import { getSharedprojects } from '@/api/modules/cluster-manager';
import { INativeNamespace, parseCpu, parseMemoryGiB } from '@/api/modules/shared-manage';
import { ICluster, IProject, useProject } from '@/composables/use-app';
import useProjects from '@/views/project-manage/project/use-project';

// 使用概览：统计共享集群被哪些使用方项目占用、配额与用量（不统计集群所属项目自建命名空间）
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

const { curProject } = useProject();
const { getProjectList } = useProjects();

// 项目名回显（全量项目 + 可见范围项目）
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

// 可见范围项目数（存量集群接口返回为空时回退到共享设置中的项目清单）
const rangeCount = computed(() => sharedProjectList.value.length
  || props.clusterData?.sharedRanges?.projectIdOrCodes?.length || 0);
const inRange = (code: string) => {
  if (!sharedProjectList.value.length) return true;
  return sharedProjectList.value.some(project => project.projectCode === code);
};

// 组件内初始化请求（项目名回显 / 可见范围 / 集群总量）聚合 loading
const innerLoading = ref(false);

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
const initSharedProjectList = async () => {
  if (!props.clusterData?.clusterID) return;
  sharedProjectList.value = await getSharedprojects({
    $clusterId: props.clusterData.clusterID,
  }).catch(() => [] as IProject[]);
};

// 集群资源总量：不请求接口，按命名空间列表统计（所有已设置配额命名空间的 limit 之和）
const clusterCpuTotal = computed(() => props.namespaceList.reduce(
  (acc, ns) => acc + parseCpu(ns.quota?.cpuLimits), 0,
));
const clusterMemTotal = computed(() => props.namespaceList.reduce(
  (acc, ns) => acc + parseMemoryGiB(ns.quota?.memoryLimits), 0,
));

// 首次进入聚合加载（任一请求进行中即展示 loading，全部结束后关闭）
const initAll = async () => {
  innerLoading.value = true;
  await Promise.all([initProjectList(), initSharedProjectList()]);
  innerLoading.value = false;
};

// 数值工具（CPU 支持 "8"/"100m"，内存支持 "16Gi"/"128Mi"，统一换算）
const formatNum = (num: number) => (Number.isInteger(num) ? num : num.toFixed(1));
const pct = (a: number, b: number) => (b ? Math.round((a / b) * 100) : 0);
const barWidth = (rate: number) => Math.min(Math.max(rate, 2), 100);

// 锁定行不参与统计（与分配页口径一致，审批中后端返回 allocatable = false）
const isLocked = (ns: INativeNamespace) => ns.allocatable === false;

interface IUsageGroup {
  projectCode: string
  items: INativeNamespace[]
  cpu: number
  mem: number
  ucpu: number
  umem: number
}

// 按使用方项目聚合：排除未分配、集群所属项目自建、锁定命名空间
const list = computed<IUsageGroup[]>(() => {
  const groups: Record<string, IUsageGroup> = {};
  props.namespaceList.forEach((ns) => {
    if (!ns.projectCode
      || ns.projectCode === curProject.value?.projectCode
      || isLocked(ns)) return;
    const group = groups[ns.projectCode] || {
      projectCode: ns.projectCode,
      items: [],
      cpu: 0,
      mem: 0,
      ucpu: 0,
      umem: 0,
    };
    groups[ns.projectCode] = group;
    group.items.push(ns);
    if (ns.quota) {
      group.cpu += parseCpu(ns.quota.cpuLimits);
      group.mem += parseMemoryGiB(ns.quota.memoryLimits);
    }
    if (ns.used) {
      group.ucpu += parseCpu(ns.used.cpuRequests);
      group.umem += parseMemoryGiB(ns.used.memoryRequests);
    }
  });
  return Object.values(groups);
});

// 汇总（卡片）
const total = computed(() => list.value.reduce((acc, group) => ({
  ns: acc.ns + group.items.length,
  cpu: acc.cpu + group.cpu,
  mem: acc.mem + group.mem,
}), { ns: 0, cpu: 0, mem: 0 }));
// 已分配但未设置配额的命名空间数（警告条）
const noQuotaCount = computed(() => list.value.reduce(
  (acc, group) => acc + group.items.filter(ns => !ns.quota).length,
  0,
));

// 查看命名空间弹窗
const nsShow = ref(false);
const currentGroup = ref<IUsageGroup>();
const openProjectNs = (group: IUsageGroup) => {
  currentGroup.value = group;
  nsShow.value = true;
};

initAll();
watch(() => props.clusterData?.clusterID, () => initSharedProjectList(), { immediate: true });
</script>
