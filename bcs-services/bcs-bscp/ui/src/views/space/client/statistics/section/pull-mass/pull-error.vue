<template>
  <div class="wrap">
    <div class="pull-error-wrap">
      <Card title="拉取失败原因" :height="416">
        <bk-loading class="loading-wrap" :loading="loading">
          <div v-if="data.length" ref="canvasRef" class="canvas-wrap"></div>
          <bk-exception
            v-else
            class="exception-wrap-item exception-part"
            type="empty"
            scene="part"
            description="暂无数据" />
        </bk-loading>
      </Card>
    </div>
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
  import { IPullErrorReason, IInfoCard, IClinetCommonQuery } from '../../../../../../../types/client';
  import { getClientPullStatusData } from '../../../../../../api/client';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';

  const clientStore = useClientStore();

  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  let columnPlot: Column | null;
  const canvasRef = ref<HTMLElement>();
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

  onMounted(async () => {
    await loadChartData();
    if (data.value.length) {
      initChart();
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
      label: {
        // 可手动配置 label 数据标签位置
        position: 'middle', // 'top', 'bottom', 'middle',
        // 配置样式
        style: {
          fill: '#FFFFFF',
          opacity: 0.6,
        },
      },
      xAxis: {
        label: {
          autoHide: true,
          autoRotate: false,
        },
      },
      legend: {
        custom: true,
        position: 'bottom',
        items: [
          {
            id: '1',
            name: '客户端数量',
            value: 'count',
            marker: {
              symbol: 'square',
            },
          },
        ],
      },
    });
    columnPlot!.render();
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
</style>
