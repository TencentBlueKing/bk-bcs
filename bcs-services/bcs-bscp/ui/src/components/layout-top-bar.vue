<template>
  <div :class="['layout-top-bar', { 'show-notice': showNotice }]">
    <section class="layout-top-section">
      <slot name="head"></slot>
    </section>
    <section class="layout-content-section">
      <section class="layout-content-container">
        <slot></slot>
      </section>
    </section>
  </div>
</template>

<script setup lang="ts">

import { storeToRefs } from 'pinia';
import useGlobalStore from '../store/global';

const { showNotice } = storeToRefs(useGlobalStore());

</script>

<style lang="scss" scoped>
.layout-top-bar {
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
  height: calc(100vh - 52px);
  width: 100%;
  &.show-notice {
    height: calc(100vh - 92px);
  }

  .layout-top-section {
    height: 52px;
    background: #ffffff;
    border-bottom: 1px solid #dcdee5;
    box-shadow: 0 3px 4px 0 rgba(0, 0, 0, 0.04);
    z-index: 1;
  }

  .layout-content-section {
    height: calc(100vh - 104px);
    overflow: auto;
  }

  .layout-content-container {
    height: 100%;
    min-width: 1366px;
  }
}
</style>
