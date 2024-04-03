<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column } from '@antv/g2plot';
  import Tooltip from '../../../components/tooltip.vue';

  const props = defineProps<{
    data: any;
  }>();
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  let columnPlot: Column;
  const data = ref(props.data || []);

  watch(
    () => props.data,
    () => {
      data.value = props.data;
      columnPlot.changeData(data.value);
    },
  );

  onMounted(() => {
    createChart();
  });

  const createChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: data.value,
      xField: 'value',
      yField: 'count',
      padding: [20, 10, 50, 20],
      limitInPlot: false,
      color: '#3E96C2',
      seriesField: 'count',
      maxColumnWidth: 40,
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
        title: 'value',
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
