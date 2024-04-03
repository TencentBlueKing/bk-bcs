<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Tooltip from '../../../components/tooltip.vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import { useRouter, useRoute } from 'vue-router';

  const router = useRouter();
  const route = useRoute();

  const bizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));

  const props = defineProps<{
    data: IClientLabelItem[];
  }>();

  let piePlot: Pie;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const data = ref(props.data || []);
  const jumpId = ref('');

  watch(
    () => props.data,
    () => {
      data.value = props.data;
      piePlot.changeData(data.value);
    },
  );

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    piePlot = new Pie(canvasRef.value!, {
      data: data.value,
      angleField: 'count',
      colorField: 'value',
      radius: 1,
      autoFit: false,
      height: 184,
      width: 800,
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
      },
      tooltip: {
        fields: ['count', 'percent'],
        showTitle: true,
        title: 'value',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          jumpId.value = originalItems[0].title;
          originalItems[0].name = '客户端数量';
          originalItems[1].name = '占比';
          originalItems[1].value = `${(parseFloat(originalItems[1].value) * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      interactions: [{ type: 'element-active' }],
      legend: {
        position: 'right',
        offsetX: -200,
      },
    });
    piePlot.render();
  };

  const jumpToSearch = () => {
    router.push({ name: 'client-search', params: { appId: appId.value, bizId: bizId.value } });
  };
</script>

<style lang="scss">
  .canvas-wrap {
    position: relative;
    display: flex;
    align-items: center;
    height: 100%;
  }
  .g2-tooltip {
    visibility: hidden;
  }
</style>
