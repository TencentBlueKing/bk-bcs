<template>
  <Card title="拉取成功率" :height="416" :width="318">
    <bk-loading class="loading-wrap" :loading="loading">
      <div v-if="data.length" ref="canvasRef" class="canvas-wrap">
        <Tooltip ref="tooltipRef" @jump="jumpToSearch" />
      </div>
      <bk-exception v-else class="exception-wrap-item exception-part" type="empty" scene="part" description="暂无数据">
        <template #type>
          <span class="bk-bscp-icon icon-pie-chart exception-icon" />
        </template>
      </bk-exception>
    </bk-loading>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import Tooltip from '../../components/tooltip.vue';
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
        release_change_status: item.release_change_status === 'Failed' ? '拉取失败' : '拉取成功',
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
      color: ['#F5876C', '#DAEFE4'],
      radius: 0.9,
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
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
          jumpStatus.value = originalItems[0].data.release_change_status === '拉取成功' ? 'Success' : 'Failed';
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
</style>
