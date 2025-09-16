
<template>
  <div class="metric-item" v-bkloading="{ isLoading, zIndex: 10 }">
    <div class="metric-item-title">
      <span class="title">
        {{ title }}
        <span class="icon ml5" v-if="desc" v-bk-tooltips="desc">
          <i class="bcs-icon bcs-icon-info-circle-shape"></i>
        </span>
      </span>
      <bk-dropdown-menu trigger="click" @show="isDropdownShow = true" @hide="isDropdownShow = false">
        <div class="dropdown-trigger-text" slot="dropdown-trigger">
          <span class="name">{{ activeTime.name }}</span>
          <i :class="['bk-icon icon-angle-down',{ 'icon-flip': isDropdownShow }]"></i>
        </div>
        <ul class="bk-dropdown-list" slot="dropdown-content">
          <li v-for="(item, index) in timeRange" :key="index" @click="handleTimeRangeChange(item)">
            {{ item.name }}
          </li>
        </ul>
      </bk-dropdown-menu>
    </div>
    <ECharts ref="eChartsRef" class="vue-echarts" :options="echartsOptions" auto-resize v-if="!isNoData"></ECharts>
    <bk-exception class="echarts-empty" type="empty" scene="part" v-else> </bk-exception>
  </div>
</template>
<script lang="ts">
import { throttle  } from 'lodash';
import moment from 'moment';
import { computed, defineComponent, getCurrentInstance, onMounted, onUnmounted, PropType, reactive, ref, toRef, toRefs, watch } from 'vue';

import defaultChartOption from '../views/resource-view/common/default-echarts-option';

import 'echarts/lib/chart/line';
import 'echarts/lib/component/tooltip';
import 'echarts/lib/component/legend';
import ECharts from '@/components/echarts.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

interface ITimeRange {
  name: string
  range: number
}

export default defineComponent({
  name: 'ResourceMetric',
  components: {
    ECharts,
  },
  props: {
    title: {
      type: String,
      default: '',
    },
    timeRange: {
      type: Array as PropType<ITimeRange[]>,
      default: () => [
        {
          name: window.i18n.t('units.time.1h'),
          range: 60 * 60 * 1000,
        },
        {
          name: window.i18n.t('units.time.24h'),
          range: 60 * 24 * 60 * 1000,
        },
        {
          name: window.i18n.t('units.time.lastDays'),
          range: 7 * 24 * 60 * 60 * 1000,
        },
      ],
    },
    options: {
      type: Object,
      default: () => ({}),
    },
    category: {
      type: String,
      default: '',
      required: true,
    },
    metric: {
      type: [String, Array],
      default: '',
      required: true,
    },
    params: {
      type: Object as PropType<Record<string, any>|null>,
      default: () => ({}),
    },
    colors: {
      type: [String, Array],
      default: '#3a84ff',
    },
    unit: {
      type: String,
      default: 'percent',
    },
    // series后缀（数组时要和metric一一对应）
    suffix: {
      type: [String, Array],
      default: '',
    },
    desc: {
      type: String,
      default: '',
    },
    series: {
      type: Array,
      default: () => [],
    },
  },
  setup(props) {
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

    const metricMap = {
      cpu_usage: $i18n.t('metrics.cpuUsage'),
      disk_usage: $i18n.t('metrics.diskUsage'),
      memory_usage: $i18n.t('metrics.memUsage'),
      cpu_request_usage: $i18n.t('metrics.cpuRequestUsage.text'),
      memory_request_usage: $i18n.t('metrics.memRequestUsage.text'),
      diskio_usage: $i18n.t('metrics.diskIOUsage'),
      network_receive: $i18n.t('metrics.network.receive'),
      network_transmit: $i18n.t('metrics.network.transmit'),
    };
    const state = reactive({
      isDropdownShow: false,
      activeTime: {
        name: $i18n.t('units.time.1h'),
        range: 60 * 60 * 1000,
      },
      isLoading: false,
    });
    const echartsOptions = ref<any>({});
    const isNoData = computed(() => !echartsOptions.value?.series?.some(series => !!series.data?.length));
    const metricNameProp = computed(() => {
      let prop = '';
      switch (props.category) {
        case 'pods':
          prop = 'pod_name';
          break;
        case 'containers':
          prop = 'container_name';
      }
      return prop;
    });
    const metricSuffix = computed(() => {
      if (!props.suffix) return [];

      return Array.isArray(props.suffix) ? props.suffix : [props.suffix];
    });

    const handleTimeRangeChange = (item) => {
      if (state.activeTime.range === item.range) return;

      state.activeTime = item;
      handleGetMetricData();
    };
    // 设置图表options
    const handleSetChartOptions = (data) => {
      if (!data) return;

      const metrics: any[] = Array.isArray(props.metric) ? props.metric : [props.metric];
      const series: any[] = [];
      data.forEach((item, index) => {
        const suffix = metricSuffix.value[index];
        const list = item?.result?.map((result) => {
          // series 配置
          const name = result.metric?.[metricNameProp.value] || metricMap[metrics[index]];
          const defaultSeries = props.series[index] || {};
          return Object.assign({
            name: suffix ? `${name} ${suffix}` : name,
            type: 'line',
            showSymbol: false,
            smooth: true,
            hoverAnimation: false,
            areaStyle: {
              opacity: 0.2,
            },
            itemStyle: {
              color: Array.isArray(props.colors)
                ? props.colors[index % props.colors.length]
                : props.colors,
            },
            data: result?.values || [],
          }, defaultSeries);
        }) || [];
        series.push(...list);
      });
      echartsOptions.value = Object.assign(defaultChartOption(props.unit), props.options, { series });
    };
    // 获取图表数据
    const projectCode = computed(() => $route.value.params.projectCode);
    const handleGetMetricData = async () => {
      const timeRange = {
        start_at: moment().subtract(state.activeTime.range, 'ms')
          .utc()
          .format(),
        end_at: moment().utc()
          .format(),
      };

      let action = '';
      switch (props.category) {
        case 'pods':
          if (!props.params) break;
          action = 'metric/clusterPodMetric';
          break;
        case 'containers':
          if (!props.params) break;
          action = 'metric/clusterContainersMetric';
        case 'nodes':
          if (!props.params?.$nodeIP) break;
          action = 'metric/clusterNodeMetric';
      }
      if (!action) return [];

      const metrics = Array.isArray(props.metric) ? props.metric : [props.metric];
      const promises: Promise<any>[] = [];
      metrics.forEach((metric) => {
        const params = {
          $metric: metric,
          $projectCode: projectCode.value,
          ...timeRange,
          ...props.params,
        };
        promises.push($store.dispatch(action, params));
      });

      state.isLoading = true;
      const metricData = await Promise.all(promises);
      state.isLoading = false;

      handleSetChartOptions(metricData);

      return metricData;
    };


    /**
     * 由于ECharts容器已被撑开，页面缩小无法触发自身resize事件
     * 通过监听父盒子尺寸变化，隐藏/显示图表达到重新渲染的效果
     */
    const resizeObserver = ref<ResizeObserver | null>(null);
    const containerRef = ref<HTMLElement>();
    const eChartsRef = ref();
    // 初始化
    function init() {
      containerRef.value = getCurrentInstance()?.proxy?.$el?.parentElement || undefined;
      if (containerRef.value) {
        resizeObserver.value = new ResizeObserver(() => {
          throttleFn();
        });
        resizeObserver.value.observe(containerRef.value);
      }
    };
    const throttleFn = throttle(() => {
      eChartsRef.value?.resize();
    }, 300);

    const { params } = toRefs(props);
    watch(params, (newValue, oldValue) => {
      if ((newValue && !oldValue)
        || (newValue && oldValue && JSON.stringify(newValue) !== JSON.stringify(oldValue))) {
        handleGetMetricData();
      }
    });

    onMounted(() => {
      handleGetMetricData();
      init();
    });
    onUnmounted(() => {
      resizeObserver.value?.disconnect();
      containerRef.value && resizeObserver.value?.unobserve?.(containerRef.value);
      resizeObserver.value = null;
    });

    return {
      ...toRefs(state),
      isNoData,
      eChartsRef,
      metricNameProp,
      echartsOptions,
      handleTimeRangeChange,
      handleGetMetricData,
      handleSetChartOptions,
    };
  },
});
</script>
<style lang="postcss" scoped>
.metric-item {
    width: 100%;
    padding: 20px 18px;
    &-title {
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-size: 14px;
        /deep/ .dropdown-trigger-text {
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            height: 32px;
            .icon-angle-down {
                font-size: 20px;
            }
        }
        /deep/ .bk-dropdown-list {
            li {
                height: 32px;
                line-height: 32px;
                padding: 0 16px;
                color: #63656e;
                font-size: 12px;
                white-space: nowrap;
                cursor: pointer;
                &:hover {
                    background-color: #eaf3ff;
                    color: #3a84ff;
                }
            }
        }
        .title {
            display: flex;
            align-items: center;
        }
        .icon {
            display: inline-block;
            font-size: 14px;
            color: #C4C6CC;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }
    .vue-echarts {
        padding-top: 12px;
        width: 100% !important;
        height: 180px;
    }
    /deep/ .echarts-empty {
        margin: 0;
        height: 180px;
        justify-content: center;
        width: 100%;
    }
}
</style>
