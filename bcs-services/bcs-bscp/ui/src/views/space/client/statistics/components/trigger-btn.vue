<template>
  <div class="btn-wrap">
    <div
      v-for="btn in btnList"
      :key="btn.type"
      :class="['icon-wrap', { current: currentType === btn.type }]"
      @click="currentType = btn.type">
      <span :class="['bk-bscp-icon', btn.icon]"></span>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  const currentType = ref('pie');
  const btnList = ref([
    {
      type: 'pie',
      icon: 'icon-help-fill',
    },
    {
      type: 'column',
      icon: 'icon-app-store',
    },
    {
      type: 'table',
      icon: 'icon-reduce',
    },
  ]);
</script>

<style scoped lang="scss">
  .btn-wrap {
    display: flex;
    align-items: center;
    padding: 2px;
    width: 64px;
    height: 24px;
    background: #f0f1f5;
    border-radius: 2px;
    .icon-wrap {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      cursor: pointer;
      color: #979ba5;
      &.current {
        background-color: #fff;
        color: #3a84ff;
      }
    }
  }
</style>
