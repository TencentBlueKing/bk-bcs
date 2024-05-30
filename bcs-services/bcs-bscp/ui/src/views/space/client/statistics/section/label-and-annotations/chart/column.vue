<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip
      :need-down-icon="!!drillDownDemension"
      :down="drillDownDemension"
      ref="tooltipRef"
      @jump="emits('jump', { label: jumpLabels, drillDownVal: drillDownVal })" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column, Datum } from '@antv/g2plot';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import Tooltip from '../../../components/tooltip.vue';
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();

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
  const drillDownVal = ref('');
  const jumpLabels = ref<{ [key: string]: string }>();
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
          isStack: false,
          xField: 'x_field',
          color: ['#3E96C2'],
          label: {
            // 可手动配置 label 数据标签位置
            position: 'top', // 'top', 'bottom', 'middle',
            // 配置样式
            style: {
              fill: '#979BA5',
            },
          },
          tooltip: {
            formatter: (datum: Datum) => {
              if (datum.foreign_val === datum.primary_key) {
                jumpLabels.value = { [datum.primary_key]: datum.primary_val };
              } else {
                jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
              }
              drillDownVal.value = datum.foreign_val;
              return { name: t('客户端数量'), value: datum.count };
            },
          },
        });
      } else {
        columnPlot.update({
          isStack: true,
          xField: 'primary_val',
          color: ['#3E96C2', '#61B2C2', '#85CCA8'],
          label: {
            // 可手动配置 label 数据标签位置
            position: 'middle', // 'top', 'bottom', 'middle',
            // 配置样式
            style: {
              fill: '#fff',
            },
          },
          legend: {
            custom: false,
            position: 'bottom',
          },
          tooltip: {
            formatter: (datum: Datum) => {
              if (datum.foreign_val === datum.primary_key) {
                jumpLabels.value = { [datum.primary_key]: datum.primary_val };
              } else {
                jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
              }
              drillDownVal.value = datum.foreign_val;
              return { name: datum.foreign_val, value: datum.count };
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
      xField: 'x_field',
      yField: 'count',
      seriesField: 'x_field',
      color: ['#3E96C2'],
      padding: [30, 10, 50, 30],
      limitInPlot: false,
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
      state: {
        active: {
          style: {
            lineWidth: 0, // 通过设置 lineWidth 为 0 来去掉黑边
            stroke: null, // 确保没有边框颜色
          },
        },
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
        fields: ['foreign_val', 'count', 'foreign_key', 'primary_key', 'primary_val'],
        formatter: (datum: Datum) => {
          if (datum.foreign_val === datum.primary_key) {
            jumpLabels.value = { [datum.primary_key]: datum.primary_val };
          } else {
            jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
          }
          drillDownVal.value = datum.foreign_val;
          return { name: t('客户端数量'), value: datum.count };
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
    .g2-tooltip-list-item {
      .g2-tooltip-marker {
        border-radius: initial !important;
      }
    }
  }
</style>
