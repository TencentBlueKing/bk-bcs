<template>
  <ECharts
    class="!w-[100%]"
    ref="chartRef"
    :options="options"
    :auto-resize="false"
    v-bkloading="{ isLoading: loading }">
  </ECharts>
</template>
<script lang="ts">
import { throttle } from 'lodash';
import moment from 'moment';
import { defineComponent, onBeforeUnmount, onMounted, PropType, ref, toRefs } from 'vue';
import ECharts from 'vue-echarts/components/ECharts.vue';

import 'echarts/lib/chart/line';
import 'echarts/lib/component/tooltip';
import { clusterMetric } from '@/api/modules/monitor';
import $i18n from '@/i18n/i18n-setup';
export default defineComponent({
  name: 'ClusterOverviewChart',
  components: { ECharts },
  props: {
    metrics: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    colors: {
      type: Array,
      default: () => ['#30d878', '#3a84ff', '#853cff'],
    },
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { metrics, colors, clusterId } = toRefs(props);
    const metricMap = {
      cpu_usage: $i18n.t('metrics.cpuUsage'),
      disk_usage: $i18n.t('metrics.diskUsage'),
      memory_usage: $i18n.t('metrics.memUsage'),
      cpu_request_usage: $i18n.t('metrics.cpuRequestUsage.text'),
      memory_request_usage: $i18n.t('metrics.memRequestUsage.text'),
      diskio_usage: $i18n.t('metrics.diskIOUsage'),
      pod_usage: $i18n.t('metrics.podUsage'),
    };
    const options = ref<any>({
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'line',
          animation: false,
          label: {
            backgroundColor: '#6a7985',
          },
        },
        formatter(params) {
          let date = params[0].value[0];
          if (String(parseInt(date, 10)).length === 10) {
            date = `${parseInt(date, 10)}000`;
          }
          return `
              <div>${parseInt(date, 10) ? moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss') : '--'}</div>
              <div>${params[0].seriesName}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
          `;
        },
      },
      grid: {
        show: false,
        top: '4%',
        left: '4%',
        right: '5%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: [
        {
          type: 'time',
          boundaryGap: false,
          axisLine: {
            show: true,
            lineStyle: {
              color: '#dde4eb',
            },
          },
          axisTick: {
            alignWithLabel: true,
            length: 5,
            lineStyle: {
              color: '#ebf0f5',
            },
          },
          axisLabel: {
            color: '#868b97',
            formatter(value) {
              if (String(parseInt(value, 10)).length === 10) {
                value = `${parseInt(value, 10)}000`;
              }
              return moment(parseInt(value, 10)).format('HH:mm');
            },
          },
          splitLine: {
            show: true,
            lineStyle: {
              color: ['#ebf0f5'],
              type: 'dashed',
            },
          },
        },
      ],
      yAxis: [
        {
          boundaryGap: [0, '2%'],
          type: 'value',
          axisLine: {
            show: true,
            lineStyle: {
              color: '#dde4eb',
            },
          },
          axisTick: {
            alignWithLabel: true,
            length: 0,
            lineStyle: {
              color: 'red',
            },
          },
          axisLabel: {
            color: '#868b97',
            formatter(value) {
              return `${value.toFixed(1)}%`;
            },
          },
          splitLine: {
            show: true,
            lineStyle: {
              color: ['#ebf0f5'],
              type: 'dashed',
            },
          },
        },
      ],
      series: [],
    });
    const loading = ref(false);
    const handleGetChartData = async () => {
      if (!metrics.value.length) return;
      loading.value = true;
      const timeRange = {
        start_at: moment().subtract(60 * 60 * 1000, 'ms')
          .utc()
          .format(),
        end_at: moment().utc()
          .format(),
      };
      const data = await Promise.all(metrics.value
        .map($metric => clusterMetric({
          $metric,
          $clusterId: clusterId.value,
          ...timeRange,
        }).catch(() => ({}))));
      options.value.series = data.map((item, index) => ({
        name: metricMap[metrics.value[index]],
        type: 'line',
        smooth: true,
        showSymbol: false,
        hoverAnimation: false,
        areaStyle: {
          normal: {
            opacity: 0.2,
          },
        },
        itemStyle: {
          normal: {
            color: colors.value[index % colors.value.length],
          },
        },
        data: item.result?.[0]?.values || [[new Date(), 0]],
      }));
      loading.value = false;
    };
    const chartRef = ref<any>(null);
    const resizeHandler = () => {
      chartRef.value?.resize();
    };
    const throttleResize = throttle(resizeHandler, 100);

    const resizeObserver = new ResizeObserver(() => {
      window.requestAnimationFrame(() => {
        throttleResize();
      });
    });
    onMounted(() => {
      handleGetChartData();
      window.addEventListener('resize', resizeHandler);
      resizeObserver.observe(chartRef.value.$el);
    });
    onBeforeUnmount(() => {
      window.removeEventListener('resize', resizeHandler);
      resizeObserver.disconnect();
    });
    return {
      chartRef,
      loading,
      options,
    };
  },
});
</script>
