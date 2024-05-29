<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip :need-down-icon="true" ref="tooltipRef" @jump="emits('jump', jumpQuery)" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column } from '@antv/g2plot';
  import { useI18n } from 'vue-i18n';
  import Tooltip from '../../../components/tooltip.vue';

  const { t } = useI18n();
  const props = defineProps<{
    data: any;
  }>();
  const emits = defineEmits(['drillDown', 'jump']);

  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  let columnPlot: Column | null;
  const data = ref(props.data.children || []);
  const jumpQuery = ref<{ [key: string]: string }>({});

  watch(
    () => props.data,
    () => {
      data.value = props.data.children;
      columnPlot!.changeData(data.value);
    },
    { deep: true },
  );

  onMounted(() => {
    initColumnChart();
  });

  const initColumnChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: data.value,
      padding: [30, 10, 50, 30],
      color: '#3E96C2',
      xField: 'name',
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
      },
      maxColumnWidth: 80,
      seriesField: 'name',
      label: {
        // 可手动配置 label 数据标签位置
        position: 'top', // 'top', 'bottom', 'middle'
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
        fields: ['value'],
        showTitle: false,
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          jumpQuery.value = { client_type: originalItems[0].data.client_type };
          originalItems.forEach((item) => {
            item.name = item.data.name;
          });
          return originalItems;
        },
      },
    });
    columnPlot.on('plot:click', (e: any) => {
      if (!e.data?.data.children) return;
      emits('drillDown', e.data?.data);
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
  .nav {
    color: #313238;
    .group-dimension {
      cursor: pointer;
    }
    .drill-down-data {
      color: #979ba5;
    }
  }
</style>
