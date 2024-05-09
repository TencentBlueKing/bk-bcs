<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { Column } from '@antv/g2plot';
  import { useRouter, useRoute } from 'vue-router';
  import Tooltip from '../../components/tooltip.vue';

  const props = defineProps<{
    data: any;
  }>();

  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  let columnPlot: Column;
  const data = ref(props.data || []);
  const jumpQuery = ref<{ [key: string]: string }>({});
  const router = useRouter();
  const route = useRoute();

  const bizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));

  watch(
    () => props.data,
    () => {
      data.value = props.data;
      columnPlot.changeData(data.value);
    },
  );

  onMounted(() => {
    initChart();
  });

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: props.data,
      isStack: true,
      color: ['#3E96C2', '#61B2C2', '#85CCA8', '#B5E0AB'],
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
      seriesField: 'client_version',
      maxColumnWidth: 80,
      label: {
        // 可手动配置 label 数据标签位置
        position: 'middle', // 'top', 'bottom', 'middle'
      },
      legend: {
        position: 'bottom',
      },
      tooltip: {
        fields: ['value'],
        showTitle: true,
        title: 'name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          jumpQuery.value = { client_type: originalItems[0].data.client_type };
          originalItems.forEach((item) => {
            item.name = item.data.client_version;
          });
          return originalItems;
        },
      },
    });
    columnPlot.render();
  };

  const jumpToSearch = () => {
    console.log(jumpQuery.value);
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: appId.value, bizId: bizId.value },
      query: jumpQuery.value,
    });
    window.open(routeData.href, '_blank');
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
