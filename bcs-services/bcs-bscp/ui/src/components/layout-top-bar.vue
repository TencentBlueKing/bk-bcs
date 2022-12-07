<script setup lang="ts">
</script>

<template>
  <div class="layout-top-bar">
    <section class="layout-top-section">
      <slot name="head"></slot>
    </section>
    <section class="layout-content-section">
      <slot></slot>
    </section>
  </div>
</template>


<style lang="scss" scoped>
.layout-top-bar {
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
  height: calc(100vh - 52px);
  width: 100%;

  .layout-top-section {
    height: 52px;
    background: #ffffff;
    box-shadow: 0 3px 4px 0 rgba(0, 0, 0, 0.04);
  }

  .layout-content-section {
    height: calc(100vh - 104px);
    overflow: auto;
  }
}
</style>