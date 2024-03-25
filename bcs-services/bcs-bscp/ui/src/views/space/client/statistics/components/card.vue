<template>
  <div class="card" :style="{ height: height + 'px', width: width + 'px' }">
    <div class="head">
      <span>{{ title }}</span>
      <slot name="head-suffix" />
    </div>
    <div class="container">
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
  defineProps<{
    width?: number;
    height?: number;
    title: string;
  }>();
</script>

<style scoped lang="scss">
  .card {
    padding: 12px 16px;
    margin-bottom: 16px;
    background: #ffffff;
    box-shadow:
      0 2px 4px 0 #0000001a,
      0 2px 4px 0 #1919290d;
    border-radius: 2px;
    .head {
      display: flex;
      align-items: center;
      height: 24px;
      .title {
        color: #313238;
        margin-right: 16px;
      }
    }
    .container {
      height: calc(100% - 24px);
    }
  }
</style>
