<script lang="ts" setup>
  const props = defineProps<{
    pkgs: {template_set_id: number; template_set_name: string;}[];
  }>()

</script>
<template>
  <div class="packages-tag">
    <div v-for="pkg in props.pkgs" class="tag-item" :key="pkg.template_set_id">{{ pkg.template_set_name }}</div>
  </div>
</template>
<style lang="scss" scoped>
  .tag-item {
    display: inline-block;
    padding: 0 8px;
    height: 22px;
    line-height: 22px;
    font-size: 12px;
    color: #63656e;
    background: #f0f1f5;
    border-radius: 2px;
    &:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
