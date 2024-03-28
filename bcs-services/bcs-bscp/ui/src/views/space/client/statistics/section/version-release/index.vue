<template>
  <div>
    <SectionTitle :title="'配置版本发布'" />
    <Card title="客户端配置版本" :height="344">
      <template #head-suffix>
        <TriggerBtn v-model:currentType="currentType" style="margin-left: 16px" />
      </template>
      <component :is="currentComponent" :data="data" />
    </Card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';
  import SectionTitle from '../../components/section-title.vue';
  import Card from '../../components/card.vue';
  import TriggerBtn from '../../components/trigger-btn.vue';
  // @ts-ignore
  import Pie from './pie.vue';
  // @ts-ignore
  import Column from './column.vue';
  // @ts-ignore
  import Table from './table.vue';

  const currentType = ref('pie');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  // 准备数据
  const data = [
    // 客户端配置版本  第一个图表
    {
      count: 122,
      current_release_id: 4, // 版本ID
      current_release_name: 'v3', // 版本名称
      percent: 0.2857142857142857,
    },
    {
      count: 13,
      current_release_id: 2,
      current_release_name: 'v2',
      percent: 0.42857142857142855,
    },
    {
      count: 125,
      current_release_id: 1,
      current_release_name: 'v1',
      percent: 0.2857142857142857,
    },
    {
      count: 150,
      current_release_id: 4, // 版本ID
      current_release_name: 'v4', // 版本名称
      percent: 0.2857142857142857,
    },
    {
      count: 156,
      current_release_id: 2,
      current_release_name: 'v5',
      percent: 0.42857142857142855,
    },
    {
      count: 2,
      current_release_id: 1,
      current_release_name: 'v6',
      percent: 0.2857142857142857,
    },
  ];
  onMounted(() => {});
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
