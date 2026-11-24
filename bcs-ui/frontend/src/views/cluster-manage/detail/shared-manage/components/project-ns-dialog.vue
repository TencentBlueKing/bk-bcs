<template>
  <!-- 使用概览「查看命名空间」弹窗：对齐原型 projectNsHtml（标题 + 汇总描述 + 明细表） -->
  <bcs-dialog
    class="shared-ns-dialog"
    :value="value"
    width="860"
    :close-icon="false"
    :draggable="false"
    header-position="left"
    render-directive="if"
    @input="handleInput">
    <template #header>
      <div class="!text-[18px] leading-[26px] text-[#313238]">
        {{ $t('cluster.sharedManage.overview.dialog.title', { name: project.name }) }}
      </div>
    </template>
    <p class="text-[12px] text-[#979ba5] leading-[20px] mb-[10px]">
      {{ summary }}
    </p>
    <bcs-table :data="items" :max-height="420">
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.nsName')" min-width="220">
        <template #default="{ row }">
          <div class="flex items-center gap-[6px] min-w-0">
            <span class="truncate text-[#313238]">{{ row.name }}</span>
            <bcs-tag v-if="!row.quota" class="shrink-0 !h-[20px] !leading-[20px]">
              {{ $t('cluster.sharedManage.overview.tag.noQuota') }}
            </bcs-tag>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.cpu')" width="170">
        <template #default="{ row }">
          <span v-if="row.quota" class="text-[#63656E]">
            {{ parseCpu(row.used?.cpuRequests).toFixed(1) }} / {{ parseCpu(row.quota?.cpuLimits) }}
            {{ $t('cluster.sharedManage.overview.unit.core') }}
          </span>
          <span v-else class="text-[#979BA5]">{{ $t('cluster.sharedManage.overview.unset') }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.mem')" width="180">
        <template #default="{ row }">
          <span v-if="row.quota" class="text-[#63656E]">
            {{ parseMemoryGiB(row.used?.memoryRequests).toFixed(1) }}
            / {{ parseMemoryGiB(row.quota?.memoryLimits) }} GiB
          </span>
          <span v-else class="text-[#979BA5]">{{ $t('cluster.sharedManage.overview.unset') }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.sharedManage.overview.table.cpuRate')" width="160">
        <template #default="{ row }">
          <div v-if="row.quota" class="flex items-center gap-[8px]">
            <div class="flex-1 h-[6px] rounded-[3px] bg-[#F0F1F5] overflow-hidden">
              <i
                class="block h-full rounded-[3px]"
                :class="nsRate(row) > 80 ? 'bg-[#FF9C01]' : 'bg-[#3A84FF]'"
                :style="{ width: `${barWidth(nsRate(row))}%` }"></i>
            </div>
            <span class="w-[38px] text-[12px] text-[#63656E]">{{ nsRate(row) }}%</span>
          </div>
          <span v-else class="text-[#979BA5]">--</span>
        </template>
      </bcs-table-column>
    </bcs-table>
    <template #footer>
      <bcs-button theme="primary" @click="handleInput(false)">
        {{ $t('generic.button.close') }}
      </bcs-button>
    </template>
  </bcs-dialog>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

import { INativeNamespace, parseCpu, parseMemoryGiB } from '@/api/modules/shared-manage';

// 按项目查看命名空间弹窗：展示该项目在此集群下的命名空间及配额 / 用量
const props = defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  // 集群名（描述行展示）
  clusterName: {
    type: String,
    default: '',
  },
  // 使用方项目（名称 + 编码）
  project: {
    type: Object as () => { name: string, code: string },
    default: () => ({ name: '', code: '' }),
  },
  // 该项目在此集群下的命名空间
  items: {
    type: Array as () => INativeNamespace[],
    default: () => [],
  },
});

const emits = defineEmits(['input']);

const handleInput = (value: boolean) => {
  emits('input', value);
};

// 数值工具（CPU 支持 "8"/"100m"，内存支持 "16Gi"/"128Mi"，统一换算）
const formatNum = (num: number) => (Number.isInteger(num) ? num : num.toFixed(1));
const pct = (a: number, b: number) => (b ? Math.round((a / b) * 100) : 0);
const barWidth = (rate: number) => Math.min(Math.max(rate, 2), 100);
const nsRate = (ns: INativeNamespace) => pct(parseCpu(ns.used?.cpuRequests), parseCpu(ns.quota?.cpuLimits));

// 汇总（描述行）：CPU / 内存统计仅取设置了配额的命名空间
const sums = computed(() => props.items.reduce((acc, ns) => ({
  cpu: acc.cpu + (ns.quota ? parseCpu(ns.quota.cpuLimits) : 0),
  mem: acc.mem + (ns.quota ? parseMemoryGiB(ns.quota.memoryLimits) : 0),
  ucpu: acc.ucpu + parseCpu(ns.used?.cpuRequests),
  umem: acc.umem + parseMemoryGiB(ns.used?.memoryRequests),
}), { cpu: 0, mem: 0, ucpu: 0, umem: 0 }));

const summary = computed(() => window.i18n.t('cluster.sharedManage.overview.dialog.summary', {
  cluster: props.clusterName,
  num: props.items.length,
  cpuText: sums.value.cpu
    ? `${sums.value.ucpu.toFixed(1)} / ${window.i18n.t('cluster.sharedManage.overview.dialog.quota')} ${formatNum(sums.value.cpu)}`
    : window.i18n.t('cluster.sharedManage.overview.dialog.unset'),
  memText: sums.value.mem
    ? `${sums.value.umem.toFixed(1)} / ${window.i18n.t('cluster.sharedManage.overview.dialog.quota')} ${formatNum(sums.value.mem)}`
    : window.i18n.t('cluster.sharedManage.overview.dialog.unset'),
}));
</script>

<style lang="postcss">
/* 弹窗垂直居中（同 confirm-dialog：transfer-dom 挂载到 body，需全局作用域） */
.shared-ns-dialog .bk-dialog-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.shared-ns-dialog .bk-dialog {
  position: relative !important;
  top: 0 !important;

  .bk-dialog-tool {
    display: none;
  }
}

.shared-ns-dialog .bk-dialog-header {
  padding: 20px 24px 12px;
}
</style>
