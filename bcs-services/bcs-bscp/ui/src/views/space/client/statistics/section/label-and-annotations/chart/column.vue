<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip
      :need-down-icon="!!drillDownDemension && !isDrillDown"
      :down="drillDownDemension"
      ref="tooltipRef"
      @jump="emits('jump', { label: jumpLabels, drillDownVal: drillDownVal })" />
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
    chartShowType: string;
    drillDownDemension: string;
    isDrillDown: boolean;
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
      updateChart(val);
    },
  );

  onMounted(() => {
    initChart();
    updateChart(props.chartShowType);
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
        fields: ['count', 'percent', 'primary_key', 'primary_val', 'foreign_key', 'foreign_val'],
        customItems: (originalItems: any[]) => {
          const datum = originalItems[0].data as IClientLabelItem;
          if (datum.foreign_val === datum.primary_key) {
            jumpLabels.value = { [datum.primary_key]: datum.primary_val };
          } else {
            jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
          }
          drillDownVal.value = originalItems[0].title;
          originalItems[0].name = t('客户端数量');
          return originalItems.slice(0, 1);
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

  const updateChart = (val: string) => {
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
        tooltip: {
          title: 'x_field',
          customItems: (originalItems: any[]) => {
            const datum = originalItems[0].data as IClientLabelItem;
            if (datum.foreign_val === datum.primary_key) {
              jumpLabels.value = { [datum.primary_key]: datum.primary_val };
            } else {
              jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
            }
            drillDownVal.value = originalItems[0].title;
            originalItems[0].name = t('客户端数量');
            originalItems[1].name = t('占比');
            originalItems[1].value = `${(originalItems[1].value * 100).toFixed(1)}%`;
            return originalItems.slice(0, 2);
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
          items: undefined,
        },
        tooltip: {
          title: 'primary_val',
          customItems: (originalItems: any[]) => {
            console.log(originalItems);
            const datum = originalItems[0].data as IClientLabelItem;
            if (datum.foreign_val === datum.primary_key) {
              jumpLabels.value = { [datum.primary_key]: datum.primary_val };
            } else {
              jumpLabels.value = { [datum.primary_key]: datum.primary_val, [datum.foreign_key]: datum.foreign_val };
            }
            drillDownVal.value = originalItems[0].title;
            let total = 0;
            const showItem = originalItems.filter((item) => item.name === 'foreign_val');
            showItem.forEach((item) => {
              item.name = item.value;
              item.value = item.data.count;
              total += item.data.count;
            });
            showItem.push({
              name: t('总和'),
              value: `${total}`,
              marker: true,
              color: '#C4C6CC',
            });
            return showItem;
          },
        },
      });
    }
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
