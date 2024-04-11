<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column } from '@antv/g2plot';
  import Tooltip from '../../components/tooltip.vue';

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
    initChart();
  });

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: props.data,
      isStack: true,
      xField: 'client_type',
      yField: 'value',
      yAxis: {
        grid: {
          line: {
            style: {
              stroke: '#979BA5',
              lineDash: [4, 5],
            },
          },
        },
        tickInterval: 1,
      },
      seriesField: 'client_version',
      maxColumnWidth: 80,
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
      legend: {
        position: 'bottom',
      },
    });
    columnPlot.render();
  };
</script>

<style scoped lang="scss"></style>
