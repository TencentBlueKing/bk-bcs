<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="emits('jump', jumpQuery)" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { Pie } from '@antv/g2plot';
  import { useI18n } from 'vue-i18n';
  import Tooltip from '../../../components/tooltip.vue';
  const { t } = useI18n();

  const props = defineProps<{
    data: any;
  }>();

  const emits = defineEmits(['jump']);

  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  let piePlot: Pie | null;
  const jumpQuery = ref<{ [key: string]: string }>({});

  onMounted(() => {
    initPieChart();
  });

  const initPieChart = () => {
    piePlot = new Pie(canvasRef.value!, {
      data: props.data,
      angleField: 'value',
      colorField: 'name',
      radius: 1,
      padding: [40, 40, 40, 40],
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(1)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      tooltip: {
        fields: ['value', 'percent'],
        showTitle: true,
        title: 'name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          console.log(originalItems);
          jumpQuery.value = { client_type: originalItems[0].data.client_type };
          originalItems[0].marker = false;
          originalItems[0].name = t('客户端数量');
          originalItems[1].name = t('占比');
          originalItems[1].value = `${(originalItems[1].data.percent * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      interactions: [{ type: 'element-highlight' }],
      legend: {
        layout: 'horizontal',
        position: 'right',
        flipPage: false,
        maxWidth: 300,
        offsetX: -200,
        reversed: true,
      },
    });
    piePlot.render();
  };
</script>

<style scoped lang="scss">
  :deep(.g2-tooltip) {
    visibility: hidden;
    .g2-tooltip-title {
      padding-left: 16px;
      font-size: 14px;
    }
    .g2-tooltip-list-item:nth-child(2) {
      .g2-tooltip-marker {
        display: none !important;
      }
      .g2-tooltip-name {
        margin-left: 16px;
      }
    }
    .g2-tooltip-list-item:nth-child(1) {
      .g2-tooltip-marker {
        position: absolute;
        top: 15px;
      }
      .g2-tooltip-name {
        margin-left: 16px;
      }
    }
  }
</style>
