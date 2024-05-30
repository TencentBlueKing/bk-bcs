<template>
  <div class="wrap">
    <Teleport :disabled="!isOpenFullScreen" to="body">
      <div
        class="pull-error-wrap"
        :class="{ fullscreen: isOpenFullScreen }"
        @mouseenter="isShowOperationBtn = true"
        @mouseleave="isShowOperationBtn = false">
        <Card :title="t('拉取失败原因')" :height="416">
          <template #operation>
            <OperationBtn
              v-show="isShowOperationBtn"
              :is-open-full-screen="isOpenFullScreen"
              @refresh="refresh"
              @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
          </template>
          <template #head-suffix>
            <div v-if="!isShowSpecificReason" class="icon-wrap">
              <span class="action-icon bk-bscp-icon icon-download" v-bk-tooltips="{ content: $t('可下钻图表') }" />
            </div>
          </template>
          <bk-loading class="loading-wrap" :loading="loading">
            <div v-if="isShowSpecificReason" class="nav">
              <span class="main-reason" @click="refresh">{{ t('主要失败原因') }}</span> /
              <span class="reason">{{ selectFailedReason.mapName }}</span>
            </div>
            <div v-if="data.length && !isShowSpecificReason" ref="canvasRef" class="canvas-wrap">
              <Tooltip :need-down-icon="true" ref="tooltipRef" @jump="jumpToSearch" />
            </div>
            <div v-else-if="specificReason.length && isShowSpecificReason" class="specific-reason">
              <div ref="specificReasonRef" class="canvas-wrap"></div>
            </div>
            <bk-exception
              v-else
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              :description="t('暂无数据')">
              <template #type>
                <span class="bk-bscp-icon icon-bar-chart exception-icon" />
              </template>
            </bk-exception>
          </bk-loading>
        </Card>
      </div>
    </Teleport>
    <div class="time-list">
      <Card v-for="item in pullTime" :key="item.key" :title="item.name" :width="207" :height="128">
        <div class="time-info">
          <span v-if="item.value">
            <span class="time">{{ Math.round(item.value as number) }}</span>
            <span class="unit">{{ item.unit }}</span>
          </span>
          <span v-else class="empty">{{ t('暂无数据') }}</span>
        </div>
      </Card>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { Column, Pie } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import Tooltip from '../../components/tooltip.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import { IPullErrorReason, IInfoCard, IClinetCommonQuery } from '../../../../../../../types/client';
  import { getClientPullStatusData, getClientPullFailedReason } from '../../../../../../api/client';
  import { CLIENT_ERROR_CATEGORY_MAP, CLIENT_ERROR_SUBCLASSES_MAP } from '../../../../../../constants/client';
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

  let columnPlot: Column | null;
  let piePlot: Pie | null;
  const canvasRef = ref<HTMLElement>();
  const specificReasonRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const pullTime = ref<IInfoCard[]>([
    {
      value: 0,
      name: t('平均拉取耗时'),
      key: 'avg',
    },
    {
      value: 0,
      name: t('最大拉取耗时'),
      key: 'max',
    },
    {
      value: 0,
      name: t('最小拉取耗时'),
      key: 'min',
    },
  ]);
  const data = ref<IPullErrorReason[]>([]);
  const specificReason = ref<IPullErrorReason[]>([]);
  const loading = ref(false);
  const isOpenFullScreen = ref(false);
  const initialWidth = ref(0);
  const isShowSpecificReason = ref(false);
  const selectFailedReason = ref({
    name: '',
    mapName: '',
  });
  const isShowOperationBtn = ref(false);

  watch(
    () => props.appId,
    async () => {
      isShowSpecificReason.value = false;
      specificReason.value = [];
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
    { deep: true },
  );

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      if (isShowSpecificReason.value) {
        specificReasonRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
      } else {
        canvasRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
      }
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
      data.value = res.failed_reason?.map((item: IPullErrorReason) => {
        const mapName = CLIENT_ERROR_CATEGORY_MAP.find(
          (errorItem) => errorItem.value === item.release_change_failed_reason,
        )?.name;
        return {
          ...item,
          mapName,
        };
      });
      if (Object.keys(res.time_consuming).length) {
        Object.entries(res.time_consuming).forEach(([key, value]) => {
          const item = pullTime.value.find((item) => item.key === key) as IInfoCard;
          item.value = value as number;
          if (item.value > 1) {
            item.unit = 's';
          } else {
            item.value = item.value * 1000;
            item.unit = 'ms';
          }
        });
      } else {
        pullTime.value.forEach((item) => (item.value = 0));
      }
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const loadPullFailedReason = async () => {
    const params = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: { failed_reason: selectFailedReason.value.name },
    };
    try {
      loading.value = true;
      const res = await getClientPullFailedReason(props.bkBizId, props.appId, params);
      specificReason.value = res.data.failed_reason?.map((item: IPullErrorReason) => {
        const mapName = CLIENT_ERROR_SUBCLASSES_MAP.find(
          (errorItem) => errorItem.value === item.release_change_failed_reason,
        )?.name;
        return {
          ...item,
          mapName,
        };
      });
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const initChart = () => {
    columnPlot = new Column(canvasRef.value!, {
      data: data.value,
      xField: 'mapName',
      yField: 'count',
      seriesField: 'count',
      color: '#FFA66B',
      maxColumnWidth: 60,
      padding: [30, 10, 50, 30],
      legend: {
        custom: true,
        position: 'bottom',
        items: [
          {
            id: '1',
            name: t('拉取失败数量'),
            value: 'release_change_failed_reason',
            marker: {
              symbol: 'square',
              style: {
                fill: '#FFA66B',
              },
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
        showTitle: false,
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          originalItems[0].name = originalItems[0].data.mapName;
          return originalItems;
        },
      },
    });
    columnPlot!.render();
    columnPlot.on('plot:click', async (event: any) => {
      selectFailedReason.value.name = event.data?.data.release_change_failed_reason;
      selectFailedReason.value.mapName = event.data?.data.mapName;
      if (!selectFailedReason.value.name) return;
      isShowSpecificReason.value = true;
      await loadPullFailedReason();
      if (specificReason.value.length > 0) {
        initSpecificReasonChart();
      }
    });
  };

  const initSpecificReasonChart = () => {
    piePlot = new Pie(specificReasonRef.value!, {
      data: specificReason.value,
      angleField: 'count',
      colorField: 'mapName',
      padding: [20, 400, 60, 10],
      interactions: [{ type: 'element-highlight' }],
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
      legend: {
        position: 'right',
        offsetX: -200,
      },
    });
    piePlot.render();
  };

  const jumpToSearch = () => {
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
    });
    window.open(routeData.href, '_blank');
  };

  const refresh = async () => {
    if (!isShowSpecificReason.value) {
      loadChartData();
    } else {
      isShowSpecificReason.value = false;
      await loadChartData();
      if (data.value.length) {
        initChart();
        initialWidth.value = canvasRef.value!.offsetWidth;
      }
    }
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
    .time-list {
      display: flex;
      flex-direction: column;
      justify-content: space-between;
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
  .loading-wrap {
    height: 100%;
  }
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    .card {
      width: 100%;
      height: 100vh !important;
      :deep(.operation-btn) {
        top: 0 !important;
      }
    }
  }
  .specific-reason {
    height: 100%;
    z-index: 9999;
  }
  .nav {
    font-size: 12px;
    color: #313238;
    position: relative;
    z-index: 999;
    .main-reason {
      margin-right: 8px;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
    .reason {
      color: #979ba5;
      margin-left: 8px;
    }
  }
  :deep(.bk-exception) {
    height: 100%;
    justify-content: center;
    transform: translateY(-20px);
  }

  .icon-wrap {
    margin-left: 8px;
    font-size: 12px;
    width: 18px;
    height: 18px;
    background: #f0f3ff;
    border-radius: 2px;
    text-align: center;
    line-height: 18px;
    color: #7594ef;
  }
</style>
