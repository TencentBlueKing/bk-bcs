<script setup lang="ts">
  import Group from './group.vue'
  import Preview from './preview.vue'

</script>
<template>
  <div class="select-group-wrapper">
    <div class="group-tree-area">
      <Group />
    </div>
    <div class="preview-area">
      <Preview />
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .select-group-wrapper {
    display: flex;
    align-items: center;
    height: 100%;
    background: #ffffff;
  }
  .group-tree-area {
    padding: 24px;
    width: 566px;
    height: 100%;
    border-right: 1px solid #dcdee5;
  }
  .preview-area {
    flex: 1;
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
</style>