<script setup lang="ts">

  const props = withDefaults(defineProps<{
    showFooter?: boolean
  }>(), {
    showFooter: true
  })

</script>
<template>
  <section :class="['version-layout-container', { 'without-footer': !showFooter }]">
    <section class="layout-header">
      <slot name="header"></slot>
    </section>
    <section class="layout-content">
      <slot></slot>
    </section>
    <section v-if="showFooter" class="layout-footer">
      <slot name="footer"></slot>
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .version-layout-container {
    position: fixed;
    top: 52px;
    left: 0;
    bottom: 0;
    right: 0;
    height: calc(100vh - 52px);
    background: #ffffff;
    z-index: 2000;
    &.without-footer {
      .layout-content {
        height: calc(100% - 48px);
      }
    }
    .layout-header {
      position: relative;
      height: 52px;
      box-shadow: 0 3px 4px 0 rgba(0, 0, 0, 0.04);
      z-index: 1;
    }
    .layout-content {
      height: calc(100% - 100px);
      background: #f5f7fa;
      overflow: auto;
    }
    .layout-footer {
      height: 48px;
      border-top: 1px solid #dcdee5;
    }
  }
</style>
