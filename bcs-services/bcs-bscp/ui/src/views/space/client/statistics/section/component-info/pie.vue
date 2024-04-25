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

  watch(
    () => props.data.children,
    () => {
      piePlot.changeData(props.data);
    },
  );

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    piePlot = new Sunburst(canvasRef.value!, {
      data: props.data,
      colorField: 'name',
      label: {
        content: ({ data }) => `${(data.percent * 100).toFixed(0)}%`,
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
      // tooltip: {
      //   fields: ['count', 'percent'],
      //   showTitle: true,
      //   title: 'current_release_name',
      //   container: tooltipRef.value?.getDom(),
      //   enterable: true,
      //   customItems: (originalItems: any[]) => {
      //     originalItems[1].value = `${(parseFloat(originalItems[1].value) * 100).toFixed(1)}%`;
      //     return originalItems;
      //   },
      // },
    });
    piePlot.render();
  };

  const jumpToSearch = () => {
    router.push({ name: 'client-search', params: { appId: appId.value, bizId: bizId.value } });
  };
</script>

<style lang="scss"></style>
