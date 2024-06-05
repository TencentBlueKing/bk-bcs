<template>
  <bk-table :data="tableData" :max-height="'280'" :border="['outer']" :key="primaryKey">
    <bk-table-column :label="primaryKey" prop="primary_val"></bk-table-column>
    <bk-table-column v-for="foreignKey in foreignKeyList" :key="foreignKey" :label="foreignKey">
      <template #default="{ row }">{{ row.foreign_key === foreignKey ? row.foreign_val : '--' }}</template>
    </bk-table-column>
    <bk-table-column :label="$t('数量')" prop="count"></bk-table-column>
    <bk-table-column :label="$t('占比')" prop="percent">
      <template #default="{ row }">{{ `${(row.percent * 100).toFixed(1)}%` }}</template>
    </bk-table-column>
  </bk-table>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  const props = defineProps<{
    data: IClientLabelItem[];
  }>();

  const tableData = ref<any>(props.data);
  const primaryKey = ref('');
  const foreignKeyList = ref<string[]>([]);

  watch(
    () => props.data,
    () => {
      primaryKey.value = props.data[0].primary_key;
      foreignKeyList.value = props.data.map((item) => item.foreign_key);
      tableData.value = props.data;
    },
    { immediate: true },
  );
</script>

<style scoped lang="scss"></style>
