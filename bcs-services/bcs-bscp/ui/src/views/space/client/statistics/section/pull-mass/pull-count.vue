<template>
  <Card title="拉取数量趋势" :height="360">
    <template #head-suffix>
      <bk-select v-model="selectTime" class="time-selector" :clearable="false">
        <bk-option v-for="item in selectorTimeList" :id="item.value" :key="item.value" :name="item.label" />
      </bk-select>
    </template>
    <bk-loading class="loading-wrap" :loading="loading">
      <div v-if="data.time.length" ref="canvasRef" class="canvas-wrap">
        <Tooltip ref="tooltipRef" />
      </div>
      <bk-exception
        v-else
        class="exception-wrap-item exception-part"
        type="empty"
        scene="part"
        description="暂无数据" />
    </bk-loading>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import Card from '../../components/card.vue';
  import { DualAxes } from '@antv/g2plot';
  import { getClientPullCountData } from '../../../../../../api/client';
  import { IPullCount, IClinetCommonQuery } from '../../../../../../../types/client';
  import Tooltip from '../../components/tooltip.vue';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';

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
      label: '近7天',
    },
    {
      value: 15,
      label: '近15天',
    },
    {
      value: 30,
      label: '近30天',
    },
  ];
  const data = ref<IPullCount>({
    time: [],
    time_and_type: [],
  });
  const loading = ref(false);

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
    () => {
      loadChartData();
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
      padding: [10, 10, 30, 20],
      geometryOptions: [
        {
          geometry: 'column',
          isGroup: true,
          seriesField: 'type',
          columnWidthRatio: 0.2,
          color: ['#3E96C2', '#61B2C2', '#61B2C2'],
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
                item.name = 'SideCar 客户端';
                break;
              case 'sdk':
                item.name = 'SDK 客户端';
                break;
              case 'agent':
                item.name = '主机插件客户端';
                break;
              case 'command':
                item.name = 'command';
                break;
              default:
                item.name = '总量';
            }
          });
          return originalItems;
        },
      },
    });
    dualAxes!.render();
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
</style>
