<template>
  <Card title="拉取成功率" :height="416" :width="318">
    <div ref="canvasRef" class="canvas-wrap"></div>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Card from '../../components/card.vue';

  const canvasRef = ref<HTMLElement>();
  const data = [
    // 拉取成功率
    {
      count: 2,
      percent: 0.2857142857142857,
      release_change_status: '拉取失败',
    },
    {
      count: 5,
      percent: 0.7142857142857143,
      release_change_status: '拉取成功',
    },
  ];

  onMounted(() => {
    const piePlot = new Pie(canvasRef.value!, {
      appendPadding: 10,
      data,
      angleField: 'count',
      colorField: 'release_change_status',
      color: ['#F5876C', '#DAEFE4'],
      radius: 0.9,
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
      },
      interactions: [{ type: 'element-active' }],
      legend: {
        position: 'bottom',
      },
    });

    piePlot.render();
  });
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
