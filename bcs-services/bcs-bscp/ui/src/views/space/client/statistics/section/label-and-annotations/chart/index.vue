<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div
      ref="containerRef"
      :class="{ fullscreen: isOpenFullScreen }"
      @mouseenter="isMouseEnter = true"
      @mouseleave="isMouseEnter = false">
      <Card :title="$t(`按 {n} 统计`, { n: primaryDimension })" :height="368">
        <template #operation>
          <OperationBtn
            v-show="isShowOperationBtn"
            :need-down="true"
            :is-open-full-screen="isOpenFullScreen"
            :all-label="allLabel"
            :primary-dimension="primaryDimension"
            @refresh="handleRefresh"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen"
            @toggle-show-btn="isOpenPopover = $event"
            @select-dimension="selectedDimension = $event"
            @select-down-dimension="selectedDownDimension = $event"
            @toggle-chart-show-type="chartShowType = $event" />
        </template>
        <template #head-suffix>
          <div class="head-suffix">
            <bk-tag theme="info" type="stroke"> {{ $t('标签') }} </bk-tag>
            <div v-if="selectedDownDimension" class="icon-wrap">
              <span
                class="action-icon bk-bscp-icon icon-download"
                v-bk-tooltips="{
                  content: `${$t('支持点击数据下钻')}\n${$t('下钻维度')}\: ${selectedDownDimension}`,
                }" />
            </div>
            <TriggerBtn v-model:currentType="currentType" />
          </div>
        </template>
        <bk-loading class="loading-wrap" :loading="loading">
          <div v-if="isDrillDown" class="nav">
            <span class="group-dimension" @click="handleCancelDrillDown">{{ primaryDimension }}</span> /
            <span class="drill-down-data">{{ navDrillDownData }}</span>
          </div>
          <component
            :bk-biz-id="bkBizId"
            :app-id="appId"
            :is="currentComponent"
            :label="primaryDimension"
            :data="data"
            :chart-show-type="chartShowType"
            :is-show-sunburst="isShowSunburst"
            :drill-down-demension="selectedDownDimension"
            @jump="jumpToSearch($event as any)"
            @drill-down="handleDrillDown" />
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import { getClientLabelData } from '../../../../../../../api/client';
  import { storeToRefs } from 'pinia';
  import useClientStore from '../../../../../../../store/client';
  import Card from '../../../components/card.vue';
  import TriggerBtn from '../../../components/trigger-btn.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import OperationBtn from '../../../components/operation-btn.vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import { useRouter } from 'vue-router';

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    primaryDimension: string;
    allLabel: string[];
  }>();

  const currentType = ref('column');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const isOpenFullScreen = ref(false);
  const containerRef = ref();
  const initialWidth = ref(0);
  const isMouseEnter = ref(false);
  const isOpenPopover = ref(false);
  const loading = ref(false);
  const data = ref<IClientLabelItem[]>([]);
  const selectedDimension = ref<string[]>([props.primaryDimension]);
  const selectedDownDimension = ref('');
  const navDrillDownData = ref('');
  const isDrillDown = ref(false);
  const chartShowType = ref('tile');
  const drillDownItem = ref<IClientLabelItem>();
  const jumpLabels = ref<{ [key: string]: string }>();

  const isShowOperationBtn = computed(() => isMouseEnter.value || isOpenPopover.value);

  const isShowSunburst = computed(() => selectedDimension.value.length > 1 && !isDrillDown.value);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);

  onMounted(() => {
    initialWidth.value = containerRef.value.offsetWidth;
    loadChartData();
  });

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      containerRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  watch(
    () => selectedDimension.value,
    () => {
      selectedDownDimension.value = '';
      isDrillDown.value = false;
      loadChartData();
    },
  );

  const loadChartData = async (drillDownData?: any) => {
    const allDimension: { [key: string]: string } = {};
    selectedDimension.value.forEach((item: string) => {
      if (item === props.primaryDimension) return;
      allDimension[item] = '';
    });
    const params = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
      primary_key: props.primaryDimension,
      foreign_keys: allDimension,
    };
    if (drillDownData) {
      params.primary_key = selectedDownDimension.value;
      params.foreign_keys = drillDownData;
    }
    try {
      loading.value = true;
      const res = await getClientLabelData(props.bkBizId, props.appId, params);
      if (drillDownData) {
        // 下钻后的数据组合为普通柱状图
        const drillDownData: IClientLabelItem[] = [];
        res[selectedDownDimension.value].forEach((item: IClientLabelItem) => {
          const index = drillDownData.findIndex((drillDownItem) => drillDownItem.primary_val === item.primary_val);
          if (index > -1) {
            drillDownData[index].count += item.count;
            drillDownData[index].percent += item.percent;
          } else {
            drillDownData.push(item);
          }
        });
        data.value = drillDownData;
      } else {
        data.value = res[props.primaryDimension];
      }
      if (selectedDimension.value.length > 1 && !isDrillDown.value) {
        data.value.forEach((item: IClientLabelItem) => {
          item.x_field = `${item.primary_val}, ${item.foreign_val}`;
        });
      } else {
        data.value.forEach((item: IClientLabelItem) => {
          item.x_field = item.primary_val;
        });
      }
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const jumpToSearch = ({ drillDownVal, label }: { label: any; drillDownVal: string }) => {
    let labels: any;
    // 判断是否是下钻
    if (isDrillDown.value) {
      labels = {
        ...jumpLabels.value,
        [selectedDownDimension.value]: drillDownVal,
      };
    } else {
      labels = label;
    }
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
      query: {
        label: JSON.stringify(labels),
        heartTime: searchQuery.value.last_heartbeat_time,
      },
    });
    window.open(routeData.href, '_blank');
  };

  // 下钻
  const handleDrillDown = (data: any) => {
    if (!selectedDownDimension.value || isDrillDown.value) return;
    drillDownItem.value = data;
    jumpLabels.value = {
      [data.foreign_key]: data.foreign_val,
      [data.primary_key]: data.primary_val,
    };
    loadChartData(jumpLabels.value);
    if (data.foreign_val === data.primary_val) {
      navDrillDownData.value = `${data.primary_val}`;
    } else {
      navDrillDownData.value = `${data.primary_val}, ${data.foreign_val}`;
    }
    isDrillDown.value = true;
  };

  const handleCancelDrillDown = () => {
    loadChartData();
    navDrillDownData.value = '';
    isDrillDown.value = false;
  };

  const handleRefresh = () => {
    if (isDrillDown.value) {
      loadChartData({
        [drillDownItem.value!.foreign_key]: drillDownItem.value!.foreign_val,
        [drillDownItem.value!.primary_key]: drillDownItem.value!.primary_val,
      });
    } else {
      loadChartData();
    }
  };
</script>

<style scoped lang="scss">
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
  .loading-wrap {
    height: 100%;
    .nav {
      font-size: 12px;
      color: #313238;
      .group-dimension {
        margin-right: 8px;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
      .drill-down-data {
        color: #979ba5;
        margin-left: 8px;
      }
    }
  }
  .head-suffix {
    margin-left: 8px;
    display: flex;
    align-items: center;
    gap: 8px;
    .icon-wrap {
      font-size: 12px;
      width: 18px;
      height: 18px;
      background: #f0f3ff;
      border-radius: 2px;
      text-align: center;
      line-height: 18px;
      color: #7594ef;
    }
  }
</style>
