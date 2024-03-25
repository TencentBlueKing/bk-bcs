<template>
  <div>
    <SectionTitle :title="'客户端配置拉取质量'" />
    <PullCount />
    <section class="pull-wrap">
      <PullSuccess />
      <div class="pull-error">
        <PullError />
      </div>
      <div class="time-list">
        <Card v-for="item in 3" :key="item" :title="'平均拉取耗时'" :width="207" :height="128" />
      </div>
    </section>
  </div>
</template>

<script lang="ts" setup>
  import SectionTitle from '../../components/section-title.vue';
  import Card from '../../components/card.vue';
  import PullCount from './pull-count.vue';
  import PullError from './pull-error.vue';
  import PullSuccess from './pull-success.vue';
</script>

<style scoped lang="scss">
  .pull-wrap {
    display: flex;
    justify-content: space-between;
    .pull-error {
      flex-grow: 1;
      margin: 0 16px;
    }
  }
</style>
