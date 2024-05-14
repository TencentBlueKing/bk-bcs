<template>
  <div class="title-wrap">
    <div class="title">{{ title }}</div>
    <slot name="suffix" />
  </div>
</template>

<script lang="ts" setup>
  defineProps<{
    title: string;
  }>();
</script>

<style scoped lang="scss">
  .title-wrap {
    display: flex;
    align-items: center;
    height: 28px;
    background: #eaebf0;
    border-radius: 2px;
    color: #313238;
    padding-left: 16px;
    margin-bottom: 16px;
    line-height: 22px;
    font-size: 14px;
  }
</style>
