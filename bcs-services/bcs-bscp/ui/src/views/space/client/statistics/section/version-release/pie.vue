<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Tooltip from '../../components/tooltip.vue';

  const props = defineProps<{
    data: any;
  }>();
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  let piePlot: Pie;
  onMounted(() => {
    show();
  });

  const show = () => {
    piePlot = new Pie(canvasRef.value!, {
      data: props.data,
      angleField: 'count',
      colorField: 'current_release_name',
      radius: 1,
      autoFit: false,
      height: 184,
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
      },
      tooltip: {
        fields: ['count', 'percent'],
        showTitle: true,
        title: 'current_release_name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          // process originalItems,
          originalItems[0].name = '客户端数量';
          originalItems[1].name = '占比';
          originalItems[1].value = `${(parseFloat(originalItems[1].value) * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      interactions: [{ type: 'element-active' }],
      legend: {
        offsetX: -500,
      },
    });
    piePlot.render();
  };
</script>

<style lang="scss">
  .canvas-wrap {
    position: relative;
    height: 280px;
  }
  .g2-tooltip {
    visibility: hidden;
  }
</style>
