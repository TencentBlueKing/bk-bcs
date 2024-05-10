<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="emits('jump', labelValue)" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column } from '@antv/g2plot';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import Tooltip from '../../../components/tooltip.vue';
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();

  const props = defineProps<{
    data: IClientLabelItem[];
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['jump']);

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

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: props.data,
      xField: 'value0',
      yField: 'count',
      padding: [30, 10, 50, 20],
      limitInPlot: false,
      isStack: true,
      color: ['#3E96C2', '#61B2C2', '#85CCA8', '#B5E0AB'],
      seriesField: 'value1',
      maxColumnWidth: 40,
      legend: {
        custom: true,
        position: 'bottom',
        items: [
          {
            id: '1',
            name: t('客户端数量'),
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
        title: 'value0',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          labelValue.value = originalItems[0].title;
          originalItems[0].name = t('客户端数量');
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
