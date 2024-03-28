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
        <Card v-for="item in time" :key="item.key" :title="item.name" :width="207" :height="128">
          <span class="time">{{ Math.round(item.value) }}</span>
          <span class="unit">s</span>
        </Card>
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

  const time = [
    {
      value: 55.11492513033334,
      name: '平均拉取耗时',
      key: 'avg',
    },
    {
      value: 65.057842945,
      name: '最大拉取耗时',
      key: 'max',
    },
    {
      value: 39.322299293,
      name: '最小拉取耗时',
      key: 'min',
    },
  ];
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
  .time {
    font-size: 32px;
    color: #63656e;
    font-weight: 700;
  }
  .unit {
    font-size: 16px;
    color: #63656e;
  }
</style>
