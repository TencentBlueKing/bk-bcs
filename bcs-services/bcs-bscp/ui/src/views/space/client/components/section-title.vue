<template>
  <div class="title">{{ title }}</div>
</template>

<script lang="ts" setup>
  defineProps<{
    title: string;
  }>();
</script>

<style scoped lang="scss">
  .title {
    height: 28px;
    background: #eaebf0;
    border-radius: 2px;
    color: #313238;
    padding-left: 16px;
    margin-bottom: 16px;
    line-height: 28px;
  }
</style>
