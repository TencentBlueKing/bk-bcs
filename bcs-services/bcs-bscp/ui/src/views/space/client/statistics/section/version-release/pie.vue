<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="emits('jump')" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Tooltip from '../../components/tooltip.vue';
  import { IClientConfigVersionItem } from '../../../../../../../types/client';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const props = defineProps<{
    data: IClientConfigVersionItem[];
    bkBizId: string;
    appId: number;
  }>();

  const emits = defineEmits(['update', 'jump']);

  let piePlot: Pie;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();

  watch(
    () => props.data,
    () => {
      piePlot.changeData(props.data);
    },
  );

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    piePlot = new Pie(canvasRef.value!, {
      data: props.data,
      angleField: 'count',
      colorField: 'current_release_name',
      radius: 1,
      padding: [20, 800, 20, 50],
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
        fields: ['count', 'percent'],
        showTitle: true,
        title: 'current_release_name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          emits('update', originalItems[0].title);
          originalItems[0].name = t('客户端数量');
          originalItems[1].name = t('占比');
          originalItems[1].value = `${(parseFloat(originalItems[1].value) * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      interactions: [{ type: 'element-highlight' }],
      legend: {
        layout: 'horizontal',
        position: 'right',
        flipPage: false,
        offsetX: -400,
      },
    });
    piePlot.render();
  };
</script>

<style lang="scss">
  .canvas-wrap {
    position: relative;
    height: 100%;
  }
  .g2-tooltip {
    visibility: hidden;
  }
</style>
