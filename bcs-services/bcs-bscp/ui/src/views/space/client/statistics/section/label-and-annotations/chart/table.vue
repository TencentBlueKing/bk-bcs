<template>
  <bk-table :data="data" :max-height="'280'" :border="['outer']" :key="label">
    <bk-table-column :label="label" prop="primary_val"></bk-table-column>
    <bk-table-column :label="$t('数量')" prop="count"></bk-table-column>
    <bk-table-column :label="$t('占比')" prop="percent">
      <template #default="{ row }">{{ `${(row.percent * 100).toFixed(1)}%` }}</template>
    </bk-table-column>
  </bk-table>
</template>

<script lang="ts" setup>
  import { onMounted } from 'vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  const props = defineProps<{
    data: IClientLabelItem[];
    label: string;
  }>();

  onMounted(() => {
    console.log(props.data);
    // const dimension = [];
    // props
  });
</script>

<style scoped lang="scss"></style>
