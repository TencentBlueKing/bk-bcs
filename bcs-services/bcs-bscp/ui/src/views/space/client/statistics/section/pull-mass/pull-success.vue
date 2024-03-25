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
    { type: '分类一', value: 27 },
    { type: '分类二', value: 25 },
    { type: '分类三', value: 18 },
    { type: '分类四', value: 15 },
    { type: '分类五', value: 10 },
    { type: '其他', value: 5 },
  ];

  onMounted(() => {
    const piePlot = new Pie(canvasRef.value!, {
      appendPadding: 10,
      data,
      angleField: 'value',
      colorField: 'type',
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
    });

    piePlot.render();
  });
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
