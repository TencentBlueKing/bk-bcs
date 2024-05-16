<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip
      :need-down-icon="!!drillDownDemension"
      :down="drillDownDemension"
      ref="tooltipRef"
      @jump="emits('jump', labelValue)" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column, Datum } from '@antv/g2plot';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import Tooltip from '../../../components/tooltip.vue';

  const props = defineProps<{
    data: IClientLabelItem[];
    bkBizId: string;
    appId: number;
    chartShowType: string;
    drillDownDemension: string;
  }>();
  const emits = defineEmits(['jump', 'drillDown']);

  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const labelValue = ref('');
  let columnPlot: Column;

  watch(
    () => props.data,
    () => {
      columnPlot.changeData(props.data);
    },
  );

  watch(
    () => props.chartShowType,
    (val) => {
      if (val === 'tile') {
        columnPlot.update({
          isGroup: true,
          isStack: false,
          label: {
            // 可手动配置 label 数据标签位置
            position: 'top', // 'top', 'bottom', 'middle',
            // 配置样式
            style: {
              fill: '#979BA5',
            },
          },
        });
      } else {
        columnPlot.update({
          isGroup: false,
          isStack: true,
          label: {
            // 可手动配置 label 数据标签位置
            position: 'middle', // 'top', 'bottom', 'middle',
            // 配置样式
            style: {
              fill: '#fff',
            },
          },
        });
      }
    },
  );

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: props.data,
      xField: 'primary_val',
      yField: 'count',
      padding: [30, 10, 50, 20],
      isGroup: true,
      limitInPlot: false,
      seriesField: 'foreign_val',
      maxColumnWidth: 40,
      legend: {
        position: 'bottom',
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
        fields: ['foreign_val', 'count'],
        formatter: (datum: Datum) => {
          return { name: datum.foreign_val, value: datum.count };
        },
        showTitle: true,
        title: 'primary_val',
        container: tooltipRef.value?.getDom(),
        enterable: true,
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
    columnPlot.on('plot:click', (e: any) => {
      if (!e.data) return;
      emits('drillDown', e.data.data as IClientLabelItem);
    });
    columnPlot.render();
  };
</script>

<style scoped lang="scss">
  :deep(.g2-tooltip) {
    visibility: hidden;
    .g2-tooltip-list-item {
      .g2-tooltip-marker {
        border-radius: initial !important;
      }
    }
  }
</style>
