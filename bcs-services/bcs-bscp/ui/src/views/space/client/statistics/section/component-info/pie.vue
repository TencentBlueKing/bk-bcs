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
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

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
  const jumpQuery = ref<{ [key: string]: string }>({});

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
      color: ['#2C2599', '#FFA66B', '#85CCA8', '#3E96C2'],
      interactions: [{ type: 'element-highlight' }],
      state: {
        active: {
          style: {
            stroke: '#ffffff',
          },
        },
      },
      label: {
        content: ({ data }) => `${(data.percent * 100).toFixed(1)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      legend: {
        position: 'right',
        layout: 'vertical',
      },
      tooltip: {
        fields: ['value', 'name'],
        showTitle: true,
        title: 'name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          if (originalItems[0].data.childNodeCount > 0) {
            jumpQuery.value = { client_type: originalItems[0].data.data.client_type };
          } else {
            jumpQuery.value = { client_version: originalItems[0].data.name };
          }
          originalItems[0].marker = false;
          originalItems[0].name = t('客户端数量');
          originalItems[1].name = t('占比');
          originalItems[1].value = `${(originalItems[1].data.data.percent * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      hierarchyConfig: {
        padding: 0.003,
      },
    });
    piePlot.render();
  };

  const jumpToSearch = () => {
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
