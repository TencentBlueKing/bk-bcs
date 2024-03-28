<template>
  <bk-table :data="data" :max-height="'280'" :border="['outer']">
    <bk-table-column label="配置版本" prop="current_release_name"></bk-table-column>
    <bk-table-column label="数量" prop="count"></bk-table-column>
    <bk-table-column label="占比" prop="percent">
      <template #default="{ row }">{{ `${(parseFloat(row.percent) * 100).toFixed(1)}%` }}</template>
    </bk-table-column>
  </bk-table>
</template>

<script lang="ts" setup>
  defineProps<{
    data: any;
  }>();
</script>

<style scoped lang="scss"></style>
