<template>
  <Card title="按 Cluster_id 统计" :height="368">
    <div ref="canvasRef" class="canvas-wrap"></div>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { Column } from '@antv/g2plot';
  import Card from '../../components/card.vue';

  const canvasRef = ref<HTMLElement>();

  onMounted(() => {
    const stackedColumnPlot = new Column(canvasRef.value!, {
      data,
      isStack: true,
      xField: '月份',
      yField: '月均降雨量',
      seriesField: 'name',
      label: {
        // 可手动配置 label 数据标签位置
        position: 'middle', // 'top', 'bottom', 'middle'
        // 可配置附加的布局方法
        layout: [
          // 柱形图数据标签位置自动调整
          { type: 'interval-adjust-position' },
          // 数据标签防遮挡
          { type: 'interval-hide-overlap' },
          // 数据标签文颜色自动调整
          { type: 'adjust-color' },
        ],
      },
    });

    stackedColumnPlot.render();
  });

  const data = [
    { name: 'London', 月份: 'Jan.', 月均降雨量: 18.9 },
    { name: 'London', 月份: 'Feb.', 月均降雨量: 28.8 },
    { name: 'London', 月份: 'Mar.', 月均降雨量: 39.3 },
    { name: 'London', 月份: 'Apr.', 月均降雨量: 81.4 },
    { name: 'London', 月份: 'May', 月均降雨量: 47 },
    { name: 'London', 月份: 'Jun.', 月均降雨量: 20.3 },
    { name: 'London', 月份: 'Jul.', 月均降雨量: 24 },
    { name: 'London', 月份: 'Aug.', 月均降雨量: 35.6 },
    { name: 'Berlin', 月份: 'Jan.', 月均降雨量: 12.4 },
    { name: 'Berlin', 月份: 'Feb.', 月均降雨量: 23.2 },
    { name: 'Berlin', 月份: 'Mar.', 月均降雨量: 34.5 },
    { name: 'Berlin', 月份: 'Apr.', 月均降雨量: 99.7 },
    { name: 'Berlin', 月份: 'May', 月均降雨量: 52.6 },
    { name: 'Berlin', 月份: 'Jun.', 月均降雨量: 35.5 },
    { name: 'Berlin', 月份: 'Jul.', 月均降雨量: 37.4 },
    { name: 'Berlin', 月份: 'Aug.', 月均降雨量: 42.4 },
    { name: 'Berl', 月份: 'Jan.', 月均降雨量: 12.4 },
    { name: 'Berl', 月份: 'Feb.', 月均降雨量: 23.2 },
    { name: 'Berl', 月份: 'Mar.', 月均降雨量: 34.5 },
    { name: 'Berl', 月份: 'Apr.', 月均降雨量: 99.7 },
    { name: 'Berl', 月份: 'May', 月均降雨量: 52.6 },
    { name: 'Berl', 月份: 'Jun.', 月均降雨量: 35.5 },
    { name: 'Berl', 月份: 'Jul.', 月均降雨量: 37.4 },
    { name: 'Berl', 月份: 'Aug.', 月均降雨量: 42.4 },
  ];
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
