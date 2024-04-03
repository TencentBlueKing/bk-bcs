<template>
  <Card title="拉取成功率" :height="416" :width="318">
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
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { Pie } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import { IPullSuccessRate } from '../../../../../../../types/client';
  import { getClientPullStatusData } from '../../../../../../api/client';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  let piePlot: Pie | null;
  const canvasRef = ref<HTMLElement>();
  const data = ref<IPullSuccessRate[]>([]);
  const loading = ref(false);

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
      }
    },
  );

  onMounted(async () => {
    await loadChartData();
    if (data.value.length) {
      initChart();
    }
  });

  const loadChartData = async () => {
    try {
      loading.value = true;
      const res = await getClientPullStatusData(props.bkBizId, props.appId, {});
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
    });
    piePlot!.render();
  };
</script>

<style scoped lang="scss">
  .loading-wrap {
    height: 100%;
  }
</style>
