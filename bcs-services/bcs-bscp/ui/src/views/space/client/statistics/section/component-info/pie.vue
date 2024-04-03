<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { Sunburst } from '@antv/g2plot';

  import Tooltip from '../../components/tooltip.vue';
  import { IVersionDistributionPie } from '../../../../../../../types/client';
  import { useRouter, useRoute } from 'vue-router';

  const router = useRouter();
  const route = useRoute();

  const bizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));

  const props = defineProps<{
    data: IVersionDistributionPie;
  }>();

  let piePlot: Sunburst;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const data = ref(props.data || []);

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
    piePlot = new Sunburst(canvasRef.value!, {
      data: data.value,
      height: 300,
      width: 800,
      colorField: 'name',
      label: {
        content: ({ data }) => `${data.percent.toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      legend: {
        position: 'right',
        offsetX: -200,
        marker: {
          symbol: 'circle',
        },
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
