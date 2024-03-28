<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { Column } from '@antv/g2plot';
  import Tooltip from '../../components/tooltip.vue';

  const props = defineProps<{
    data: any;
  }>();
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();

  onMounted(() => {
    show();
  });
  const show = () => {
    const columnPlot = new Column(canvasRef.value!, {
      data: props.data,
      xField: 'current_release_name',
      yField: 'count',
      limitInPlot: false,
      color: '#3E96C2',
      seriesField: 'count',
      legend: {
        custom: true,
        position: 'bottom',
        items: [
          {
            id: '1',
            name: '客户端数量',
            value: 'count',
            marker: {
              symbol: 'square',
            },
          },
        ],
      },
      label: {
        // 可手动配置 label 数据标签位置
        position: 'top', // 'top', 'bottom', 'middle',
        // 配置样式
        style: {
          fill: '#979BA5',
        },
      },
      tooltip: {
        fields: ['count'],
        showTitle: true,
        title: 'current_release_name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          // process originalItems,
          originalItems[0].name = '客户端数量';
          return originalItems;
        },
      },
      yAxis: {
        grid: {
          line: {
            style: {
              stroke: '#979BA5',
              lineDash: [4, 5],
            },
          },
        },
      },
    });

    columnPlot.render();
  };
</script>

<style scoped lang="scss"></style>
