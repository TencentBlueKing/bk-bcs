<template>
  <bk-table :data="data" :max-height="'280'" :border="['outer']" :key="label">
    <bk-table-column :label="label" prop="value"></bk-table-column>
    <bk-table-column :label="$t('数量')" prop="count"></bk-table-column>
    <bk-table-column :label="$t('占比')" prop="percent">
      <template #default="{ row }">{{ `${(row.percent * 100).toFixed(1)}%` }}</template>
    </bk-table-column>
  </bk-table>
</template>

<script lang="ts" setup>
  import { IClientLabelItem } from '../../../../../../../../types/client';
  defineProps<{
    data: IClientLabelItem[];
    label: string;
  }>();
</script>

<style scoped lang="scss"></style>
