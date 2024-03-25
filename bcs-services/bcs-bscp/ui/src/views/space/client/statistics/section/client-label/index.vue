<template>
  <div>
    <SectionTitle :title="'客户端标签/附加信息分布'" />
    <div class="content-wrap">
      <div class="left">
        <Area />
      </div>
      <div class="right">
        <Clustre />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import SectionTitle from '../../components/section-title.vue';
  import Area from './area.vue';
  import Clustre from './clustre.vue';
</script>

<style scoped lang="scss">
  .content-wrap {
    display: flex;
    .left,
    .right {
      flex-grow: 1;
      flex-basis: 0;
    }
    .left {
      margin-right: 16px;
    }
  }
</style>
