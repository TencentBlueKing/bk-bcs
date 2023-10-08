<template>
  <bk-table-column label="配置项名称">
    <template #default="{ row }">
      <bk-checkbox></bk-checkbox>
      aaa
    </template>
  </bk-table-column>
</template>
<script lang="ts" setup>
const props = defineProps<{
    config: { label: string; prop: string; };
  }>();
const renderTh = () => '<bk-checkbox />';
</script>
<style lang="scss" scoped>

</style>
