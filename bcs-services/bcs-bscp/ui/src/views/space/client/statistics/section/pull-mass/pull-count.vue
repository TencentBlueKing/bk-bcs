<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="{ fullscreen: isOpenFullScreen }">
      <Card :title="$t('拉取数量趋势')" :height="360">
        <template #operation>
          <OperationBtn
            :is-open-full-screen="isOpenFullScreen"
            @refresh="loadChartData"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
        </template>
        <template #head-suffix>
          <bk-select v-model="selectTime" class="time-selector" :filterable="false" :clearable="false">
            <bk-option v-for="item in selectorTimeList" :id="item.value" :key="item.value" :name="item.label" />
          </bk-select>
        </template>
        <bk-loading class="loading-wrap" :loading="loading">
          <div v-if="data.time.length" ref="canvasRef" class="canvas-wrap">
            <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
          </div>
          <bk-exception
            v-else
            class="exception-wrap-item exception-part"
            type="empty"
            scene="part"
            :description="$t('暂无数据')">
            <template #type>
              <span class="bk-bscp-icon icon-bar-chart exception-icon" />
            </template>
          </bk-exception>
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import Card from '../../components/card.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import { DualAxes } from '@antv/g2plot';
  import { getClientPullCountData } from '../../../../../../api/client';
  import { IPullCount, IClinetCommonQuery } from '../../../../../../../types/client';
  import Tooltip from '../../components/tooltip.vue';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';
  import { useRouter } from 'vue-router';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const router = useRouter();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  let dualAxes: DualAxes | null;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const selectTime = ref(7);
  const selectorTimeList = [
    {
      value: 7,
      label: t('近 {n} 天', { n: 7 }),
    },
    {
      value: 15,
      label: t('近 {n} 天', { n: 15 }),
    },
    {
      value: 30,
      label: t('近 {n} 天', { n: 30 }),
    },
  ];
  const data = ref<IPullCount>({
    time: [],
    time_and_type: [],
  });
  const loading = ref(false);
  const isOpenFullScreen = ref(false);

  watch([() => selectTime.value, () => props.appId], async () => {
    await loadChartData();
    if (data.value.time.length) {
      if (dualAxes) {
        dualAxes.changeData([data.value.time_and_type, data.value.time]);
      } else {
        initChart();
      }
    }
  });

  watch(
    () => searchQuery.value,
    async () => {
      await loadChartData();
      dualAxes!.changeData([data.value.time_and_type, data.value.time]);
    },
    { deep: true },
  );

  watch(
    () => data.value.time,
    (val) => {
      if (!val.length && dualAxes) {
        dualAxes!.destroy();
        dualAxes = null;
      }
    },
  );

  onMounted(async () => {
    await loadChartData();
    if (data.value.time.length) {
      initChart();
    }
  });

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      search: searchQuery.value.search,
      pull_time: selectTime.value,
    };
    try {
      loading.value = true;
      const res = await getClientPullCountData(props.bkBizId, props.appId, params);
      data.value.time = res.time || [];
      data.value.time_and_type = res.time_and_type || [];
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const initChart = () => {
    dualAxes = new DualAxes(canvasRef.value!, {
      data: [data.value.time_and_type, data.value.time],
      xField: 'time',
      yField: ['value', 'count'],
      yAxis: [
        {
          grid: {
            line: {
              style: {
                stroke: '#979BA5',
                lineDash: [4, 5],
              },
            },
          },
        },
        {
          min: 0,
        },
      ],
      padding: [10, 20, 30, 20],
      geometryOptions: [
        {
          geometry: 'column',
          isGroup: true,
          seriesField: 'type',
          columnWidthRatio: 0.2,
          color: ['#3E96C2', '#61B2C2', '#61B2C2'],
          // @ts-ignore
          maxColumnWidth: 80,
        },
        {
          geometry: 'line',
          lineStyle: {
            lineWidth: 2,
          },
          color: '#2C2599',
          label: {
            position: 'top',
          },
        },
      ],
      legend: {
        position: 'bottom',
      },
      tooltip: {
        fields: ['value', 'count'],
        showTitle: true,
        title: 'time',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          originalItems.forEach((item) => {
            switch (item.data.type) {
              case 'sidecar':
                item.name = `SideCar ${t('客户端')}`;
                break;
              case 'sdk':
                item.name = `SDK ${t('客户端')}`;
                break;
              case 'agent':
                item.name = t('主机插件客户端');
                break;
              case 'command':
                item.name = 'CLI';
                break;
              default:
                item.name = t('总量');
            }
          });
          return originalItems;
        },
      },
    });
    dualAxes!.render();
  };

  const jumpToSearch = () => {
    router.push({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
    });
  };
</script>

<style scoped lang="scss">
  .loading-wrap {
    height: 100%;
  }
  .time-selector {
    margin-left: 16px;
    width: 88px;
    :deep(.bk-input) {
      height: 26px;
    }
  }
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    .card {
      height: 100vh !important;
      :deep(.operation-btn) {
        top: 0 !important;
      }
    }
  }
</style>
