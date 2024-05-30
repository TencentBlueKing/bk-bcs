<template>
  <bk-table :data="data" :max-height="'280'" :border="['outer']">
    <bk-table-column :label="$t('组件类型')" prop="client_type"></bk-table-column>
    <bk-table-column :label="$t('组件版本')" prop="client_version"></bk-table-column>
    <bk-table-column :label="$t('数量')" prop="value"></bk-table-column>
    <bk-table-column :label="$t('占比')">
      <template #default="{ row }">{{ `${(parseFloat(row.percent) * 100).toFixed(1)}%` }}</template>
    </bk-table-column>
  </bk-table>
</template>

<script lang="ts" setup>
  import { IVersionDistributionItem } from '../../../../../../../types/client';
  defineProps<{
    data: IVersionDistributionItem[];
    label: string;
  }>();
</script>

<style scoped lang="scss"></style>
