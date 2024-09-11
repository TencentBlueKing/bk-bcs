<template>
  <bk-table-column label="配置文件名">
    <bk-checkbox></bk-checkbox>
    aaa
  </bk-table-column>
</template>
<script lang="ts" setup>
  defineProps<{
    config: { label: string; prop: string };
  }>();
  // const renderTh = () => '<bk-checkbox />';
</script>
<style lang="scss" scoped></style>
