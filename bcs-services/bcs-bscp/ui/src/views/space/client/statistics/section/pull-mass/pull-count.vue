<template>
  <Card title="拉取数量趋势（近一周）" :height="360">
    <div ref="canvasRef" class="canvas-wrap"></div>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import Card from '../../components/card.vue';
  import { DualAxes } from '@antv/g2plot';

  const canvasRef = ref<HTMLElement>();

  const uvBillData = [
    { time: '2019-03', value: 350, type: 'uv' },
    { time: '2019-04', value: 900, type: 'uv' },
    { time: '2019-05', value: 300, type: 'uv' },
    { time: '2019-06', value: 450, type: 'uv' },
    { time: '2019-07', value: 470, type: 'uv' },
    { time: '2019-08', value: 350, type: 'uv' },
    { time: '2019-09', value: 900, type: 'uv' },
    { time: '2019-10', value: 300, type: 'uv' },
    { time: '2019-11', value: 450, type: 'uv' },
    { time: '2019-12', value: 470, type: 'uv' },
    { time: '2019-13', value: 350, type: 'uv' },
    { time: '2019-14', value: 900, type: 'uv' },
    { time: '2019-15', value: 300, type: 'uv' },
    { time: '2019-17', value: 450, type: 'uv' },
    { time: '2019-17', value: 470, type: 'uv' },
    { time: '2019-03', value: 220, type: 'bill' },
    { time: '2019-04', value: 300, type: 'bill' },
    { time: '2019-05', value: 250, type: 'bill' },
    { time: '2019-06', value: 220, type: 'bill' },
    { time: '2019-07', value: 362, type: 'bill' },
    { time: '2019-08', value: 220, type: 'bill' },
    { time: '2019-09', value: 300, type: 'bill' },
    { time: '2019-10', value: 250, type: 'bill' },
    { time: '2019-11', value: 220, type: 'bill' },
    { time: '2019-12', value: 362, type: 'bill' },
    { time: '2019-13', value: 220, type: 'bill' },
    { time: '2019-14', value: 300, type: 'bill' },
    { time: '2019-15', value: 250, type: 'bill' },
    { time: '2019-16', value: 220, type: 'bill' },
    { time: '2019-17', value: 362, type: 'bill' },
    { time: '2019-03', value: 220, type: 'bil' },
    { time: '2019-04', value: 300, type: 'bil' },
    { time: '2019-05', value: 250, type: 'bil' },
    { time: '2019-06', value: 220, type: 'bil' },
    { time: '2019-07', value: 362, type: 'bil' },
    { time: '2019-08', value: 220, type: 'bil' },
    { time: '2019-09', value: 300, type: 'bil' },
    { time: '2019-10', value: 250, type: 'bil' },
    { time: '2019-11', value: 220, type: 'bil' },
    { time: '2019-12', value: 362, type: 'bil' },
    { time: '2019-13', value: 220, type: 'bil' },
    { time: '2019-14', value: 300, type: 'bil' },
    { time: '2019-15', value: 250, type: 'bil' },
    { time: '2019-16', value: 220, type: 'bil' },
    { time: '2019-17', value: 362, type: 'bil' },
  ];

  const transformData = [
    { time: '2019-03', count: 800 },
    { time: '2019-04', count: 600 },
    { time: '2019-05', count: 400 },
    { time: '2019-06', count: 380 },
    { time: '2019-07', count: 220 },
  ];

  onMounted(() => {
    const dualAxes = new DualAxes(canvasRef.value!, {
      data: [uvBillData, transformData],
      xField: 'time',
      yField: ['value', 'count'],
      geometryOptions: [
        {
          geometry: 'column',
          isGroup: true,
          seriesField: 'type',
        },
        {
          geometry: 'line',
          lineStyle: {
            lineWidth: 2,
          },
        },
      ],
    });

    dualAxes.render();
  });
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
