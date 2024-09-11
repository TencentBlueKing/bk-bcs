<template>
  <div class="max-w-[600px]">
    <bcs-table :data="tableData">
      <bcs-table-column :label="$t('cluster.create.label.quotaAllocation')">
        <template #default="{ row }">
          <bcs-select :clearable="false" v-model="row.proportion" class="bg-[#fff]">
            <bcs-option
              v-for="item in proportionList"
              :id="item"
              :name="item"
              :key="item"
              v-show="curProportionList.includes(item) || item === row.proportion" />
          </bcs-select>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.create.label.coefficient')" width="120">
        <template #default="{ row }">
          <bcs-input type="number" :min="1" :max="1000" :precision="0" v-model="row.coefficient"></bcs-input>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.create.label.cpuQuota')" width="130" align="right">
        <template #default="{ row }">
          {{ getCpuQuota(row) }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.create.label.memQuota')" width="130" align="right">
        <template #default="{ row }">
          {{ getMemQuota(row) }}
        </template>
      </bcs-table-column>
      <bcs-table-column width="80">
        <template #default="{ $index }">
          <span
            :class="[
              'text-[14px] text-[#979BA5] cursor-pointer',
              {
                '!cursor-not-allowed text-[#dcdee5]': !curProportionList.length
              }
            ]"
            @click="handleAddQuota">
            <i class="bk-icon icon-plus-circle"></i>
          </span>
          <span
            :class="[
              'text-[14px] text-[#979BA5] cursor-pointer ml-[10px]',
              {
                '!cursor-not-allowed text-[#dcdee5]': tableData.length <= 1
              }
            ]"
            @click="handleDeleteQuota($index)">
            <i class="bk-icon icon-minus-circle"></i>
          </span>
        </template>
      </bcs-table-column>
    </bcs-table>
    <div
      class="bcs-border mt-[-1px]
      flex items-center text-[#313238] text-[12px] font-bold h-[42px] bg-[#F0F1F5] px-[16px]">
      <div class="flex-1">{{ $t('generic.label.totalCount') }}</div>
      <div :class="['px-[15px]', { 'text-[#ea3636]': totalCpuAndMem.cpu < 40 }]">
        {{ $t('units.cores', [totalCpuAndMem.cpu]) }}
      </div>
      <div
        :class="[
          'mr-[48px] w-[130px] text-right px-[15px]',
          { 'text-[#ea3636]': totalCpuAndMem.mem < 40 }
        ]">
        {{ `${totalCpuAndMem.mem} GiB` }}
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, PropType, ref, watch } from 'vue';

defineProps({
  value: {
    type: Object as PropType<{ cpu: number, mem: number }>,
    default: () => ({ cpu: 0, mem: 0 }),
  },
});
const emits = defineEmits(['input']);
// 配额比例
const tableData = ref([{
  proportion: '1:1',
  coefficient: 40,
}]);
const getCpuQuota = row => (row.proportion?.split(':')?.[0] || 0) * (row.coefficient || 0);
const getMemQuota = row => (row.proportion?.split(':')?.[1] || 0) * (row.coefficient || 0);
const totalCpuAndMem = computed(() => tableData.value.reduce<{
  cpu: number
  mem: number
}>((data, row) => {
  data.cpu += getCpuQuota(row);
  data.mem += getMemQuota(row);
  return data;
}, { cpu: 0, mem: 0 }));
watch(totalCpuAndMem, () => {
  emits('input', totalCpuAndMem.value);
}, { deep: true });
const proportionList = ref(['1:1', '1:2', '1:3', '1:4']);// API配置
const curProportionList = computed(() => proportionList.value
  .filter(proportion => !tableData.value.some(item => item.proportion === proportion)));

const handleAddQuota = () => {
  if (!curProportionList.value.length) return;

  tableData.value.push({
    proportion: curProportionList.value[0],
    coefficient: 40,
  });
};
const handleDeleteQuota = (index: number) => {
  if (tableData.value.length <= 1) return;

  tableData.value.splice(index, 1);
};

onBeforeMount(() => {
  emits('input', totalCpuAndMem.value);
});
</script>
