<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="{ fullscreen: isOpenFullScreen }">
      <Card title="拉取成功率" :height="416" :width="318">
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
              <span class="bk-bscp-icon icon-pie-chart exception-icon" />
            </template>
          </bk-exception>
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import Tooltip from '../../components/tooltip.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import { IPullSuccessRate, IClinetCommonQuery } from '../../../../../../../types/client';
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

  let piePlot: Pie | null;
  const canvasRef = ref<HTMLElement>();
  const data = ref<IPullSuccessRate[]>([]);
  const loading = ref(false);
  const tooltipRef = ref();
  const jumpStatus = ref('');
  const isOpenFullScreen = ref(false);

  watch(
    () => props.appId,
    async () => {
      await loadChartData();
      if (data.value.length) {
        if (piePlot) {
          piePlot!.changeData(data.value);
        } else {
          initChart();
        }
      }
    },
  );

  watch(
    () => data.value,
    (val) => {
      if (!val.length && piePlot) {
        piePlot!.destroy();
        piePlot = null;
      } else {
        piePlot?.changeData(data.value);
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
      data.value = res.change_status.map((item: any) => ({
        count: item.count,
        percent: item.percent,
        release_change_status: item.release_change_status === 'Success' ? '拉取成功' : '拉取失败',
      }));
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const initChart = () => {
    piePlot = new Pie(canvasRef.value!, {
      appendPadding: 10,
      data: data.value,
      angleField: 'count',
      colorField: 'release_change_status',
      color: ({ release_change_status }) => {
        return release_change_status === '拉取成功' ? '#85CCA8' : '#F5876C';
      },
      radius: 0.9,
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      interactions: [{ type: 'element-active' }],
      legend: {
        position: 'bottom',
      },
      tooltip: {
        fields: ['count', 'percent'],
        showTitle: true,
        title: 'release_change_status',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          jumpStatus.value = originalItems[0].data.release_change_status;
          originalItems[0].name = '客户端数量';
          originalItems[1].name = '占比';
          originalItems[1].value = `${(parseFloat(originalItems[1].value) * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
    });
    piePlot!.render();
  };

  const jumpToSearch = () => {
    router.push({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
      query: { release_change_status: jumpStatus.value },
    });
  };
</script>

<style scoped lang="scss">
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
    background-color: rgba(0, 0, 0, 0.6);
    .card {
      position: absolute;
      width: 100% !important;
      height: 80vh !important;
      top: 50%;
      transform: translateY(-50%);
    }
  }
</style>
