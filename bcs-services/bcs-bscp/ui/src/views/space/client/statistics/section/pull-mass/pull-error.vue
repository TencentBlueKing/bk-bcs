<template>
  <div class="wrap">
    <Teleport :disabled="!isOpenFullScreen" to="body">
      <div class="pull-error-wrap" :class="{ fullscreen: isOpenFullScreen }">
        <Card title="拉取失败原因" :height="416">
          <template #operation>
            <OperationBtn
              :is-open-full-screen="isOpenFullScreen"
              @refresh="loadChartData"
              @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
          </template>
          <bk-loading class="loading-wrap" :loading="loading">
            <div v-if="data.length" ref="canvasRef" class="canvas-wrap">
              <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
            </div>
            <bk-exception
              v-else
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              description="暂无数据">
              <template #type>
                <span class="bk-bscp-icon icon-bar-chart exception-icon" />
              </template>
            </bk-exception>
          </bk-loading>
        </Card>
      </div>
    </Teleport>
    <div>
      <Card v-for="item in pullTime" :key="item.key" :title="item.name" :width="207" :height="128">
        <div class="time-info">
          <span v-if="item.value">
            <span class="time">{{ Math.round(item.value) }}</span>
            <span class="unit">s</span>
          </span>
          <span v-else class="empty">暂无数据</span>
        </div>
      </Card>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { Column } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import Tooltip from '../../components/tooltip.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import { IPullErrorReason, IInfoCard, IClinetCommonQuery } from '../../../../../../../types/client';
  import { getClientPullStatusData } from '../../../../../../api/client';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';
  import { useRouter } from 'vue-router';

  const router = useRouter();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  let columnPlot: Column | null;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const pullTime = ref<IInfoCard[]>([
    {
      value: 0,
      name: '平均拉取耗时',
      key: 'avg',
    },
    {
      value: 0,
      name: '最大拉取耗时',
      key: 'max',
    },
    {
      value: 0,
      name: '最小拉取耗时',
      key: 'min',
    },
  ]);
  const data = ref<IPullErrorReason[]>([]);
  const loading = ref(false);
  const isOpenFullScreen = ref(false);
  const initialWidth = ref(0);

  watch(
    () => props.appId,
    async () => {
      await loadChartData();
      if (data.value.length) {
        if (columnPlot) {
          columnPlot!.changeData(data.value);
        } else {
          initChart();
        }
      }
    },
  );

  watch(
    () => data.value,
    (val) => {
      if (!val.length && columnPlot) {
        columnPlot!.destroy();
        columnPlot = null;
      } else {
        columnPlot?.changeData(data.value);
      }
    },
  );

  watch(
    () => searchQuery.value,
    () => {
      loadChartData();
    },
    { deep: true },
  );

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      canvasRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  onMounted(async () => {
    await loadChartData();
    if (data.value.length) {
      initChart();
      initialWidth.value = canvasRef.value!.offsetWidth;
    }
  });

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      loading.value = true;
      const res = await getClientPullStatusData(props.bkBizId, props.appId, params);
      data.value = res.failed_reason;
      Object.entries(res.time_consuming).map(
        ([key, value]) => (pullTime.value.find((item) => item.key === key)!.value = value as number),
      );
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: data.value,
      xField: 'release_change_failed_reason',
      yField: 'count',
      color: '#FFA66B',
      maxColumnWidth: 60,
      padding: [10, 10, 40, 20],
      legend: {
        layout: 'horizontal',
        custom: true,
        position: 'bottom',
        items: [
          {
            id: '1',
            name: '拉取失败数量',
            value: 'count',
            marker: {
              symbol: 'square',
            },
          },
        ],
      },
      yAxis: {
        grid: {
          line: {
            style: {
              stroke: '#979BA5',
              lineDash: [4, 5],
            },
          },
        },
        tickInterval: 1,
      },
      label: {
        // 可手动配置 label 数据标签位置
        position: 'top', // 'top', 'bottom', 'middle',
        // 配置样式
        style: {
          fill: '#979BA5',
        },
      },
      tooltip: {
        fields: ['value', 'count'],
        showTitle: true,
        title: 'release_change_failed_reason',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          originalItems[0].name = '客户端数量';
          return originalItems;
        },
      },
    });
    columnPlot!.render();
  };

  const jumpToSearch = () => {
    router.push({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
    });
  };
</script>

<style scoped lang="scss">
  .wrap {
    display: flex;
    .pull-error-wrap {
      height: 100%;
      flex: 1;
      margin: 0 16px;
      .loading-wrap {
        height: 100%;
      }
    }
    .time-info {
      margin-left: 8px;
      color: #63656e;
      .time {
        font-size: 32px;
        color: #63656e;
        font-weight: 700;
      }
      .unit {
        font-size: 16px;
        margin-left: 2px;
      }
      .empty {
        font-size: 12px;
        color: #979ba5;
      }
    }
  }
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    background-color: rgba(0, 0, 0, 0.6);
    .card {
      position: absolute;
      width: 100%;
      height: 80vh !important;
      top: 50%;
      transform: translateY(-50%);
      .loading-wrap {
        height: 100%;
      }
    }
  }
</style>
