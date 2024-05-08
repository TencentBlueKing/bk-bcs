<template>
  <div>
    <SectionTitle :title="$t('客户端标签/附加信息分布')">
      <template #suffix>
        <div class="line"></div>
        <bk-select
          v-model="selectedLabel"
          :popover-options="{ theme: 'light bk-select-popover add-chart-popover', placement: 'bottom-start' }"
          :popover-min-width="240"
          :filterable="false"
          multiple>
          <template #trigger>
            <div class="add-chart-wrap">
              <Plus class="add-icon" />
              <span class="text">{{ $t('添加图表') }}</span>
            </div>
          </template>
          <bk-option-group :label="$t('标签')" collapsible>
            <bk-option
              v-if="addChartData?.labels.length"
              v-for="item in addChartData?.labels"
              :key="item"
              :id="item"
              :name="item">
            </bk-option>
            <div v-else class="bk-select-option no-data">{{ $t('暂无标签') }}</div>
          </bk-option-group>
          <bk-option-group :label="$t('附加信息')" collapsible>
            <bk-option
              v-if="addChartData?.annotations.length"
              v-for="item in addChartData?.annotations"
              :key="item"
              :id="item"
              :name="item" />
            <div v-else class="bk-select-option no-data">{{ $t('暂无附加信息') }}</div>
          </bk-option-group>
        </bk-select>
      </template>
    </SectionTitle>
    <div class="chart-list">
      <div v-if="selectedChart?.length" v-for="item in selectedChart" :key="item.label" class="chart">
        <Chart
          :data="item.data as IClientLabelItem[]"
          :label="item.label"
          :bk-biz-id="bkBizId"
          :app-id="appId"
          :loading="loading"
          @refresh="loadLabelsAndAnnotationsData" />
      </div>
      <Card v-else :height="368">
        <bk-exception
          class="exception-wrap-item exception-part"
          type="empty"
          scene="part"
          :description="$t('暂无数据')">
          <template #type>
            <span class="bk-bscp-icon icon-bar-chart exception-icon" />
          </template>
        </bk-exception>
      </Card>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { Plus } from 'bkui-vue/lib/icon';
  import { getClientLabelsAndAnnotations, getClientLabelData } from '../../../../../../api/client';
  import { storeToRefs } from 'pinia';
  import { IClientLabelItem, IClinetCommonQuery } from '../../../../../../../types/client';
  import useClientStore from '../../../../../../store/client';
  import SectionTitle from '../../components/section-title.vue';
  import Chart from './chart/index.vue';
  import Card from '../../components/card.vue';

  interface ISelectedChart {
    label: string;
    data: IClientLabelItem[];
  }

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const selectedLabel = ref<string[]>([]);

  const allLabelData = ref<{ [key: string]: IClientLabelItem[] }>(); // 所有标签图表数据

  const selectedChart = ref<ISelectedChart[]>([]); // 选择展示的图表

  const loading = ref(false);

  const addChartData = ref<{
    annotations: string[];
    labels: string[];
  }>();

  watch(
    () => props.appId,
    () => {
      getAddChartDate();
      loadLabelsAndAnnotationsData();
    },
  );

  watch(
    () => searchQuery.value,
    () => {
      getAddChartDate();
      loadLabelsAndAnnotationsData();
    },
    { deep: true },
  );

  watch(
    () => selectedLabel.value,
    () => {
      selectedChart.value = [];
      selectedLabel.value.forEach((item) => {
        const data = allLabelData.value?.[item];
        if (data) {
          selectedChart.value.push({
            label: item,
            data,
          });
        }
      });
    },
  );

  onMounted(() => {
    loadLabelsAndAnnotationsData();
    getAddChartDate();
  });

  const getAddChartDate = async () => {
    try {
      const res = await getClientLabelsAndAnnotations(props.bkBizId, props.appId, {
        last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      });
      addChartData.value = res.data;
      selectedLabel.value = addChartData.value?.labels.slice(0, 2) || addChartData.value?.labels.slice(0, 1) || [];
    } catch (e) {
      console.error(e);
    }
  };

  const loadLabelsAndAnnotationsData = async () => {
    const params: IClinetCommonQuery = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      loading.value = true;
      const res = await getClientLabelData(props.bkBizId, props.appId, params);
      allLabelData.value = res;
      selectedChart.value = [];
      if (Object.keys(res).length) {
        selectedLabel.value.forEach((item) => {
          const data = allLabelData.value?.[item];
          if (data) {
            selectedChart.value.push({
              label: item,
              data,
            });
          }
        });
      }
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped lang="scss">
  .line {
    width: 1px;
    height: 16px;
    background-color: #dcdee5;
    margin: 0 16px;
  }
  .add-chart-wrap {
    display: flex;
    align-items: center;
    height: 16px;
    cursor: pointer;
    .add-icon {
      border-radius: 50%;
      background-color: #3a84ff;
      color: #eaebf0;
      margin-right: 5px;
    }
    .text {
      color: #3a84ff;
      font-size: 12px;
    }
  }
  .chart-list {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 16px;
    .chart {
      width: calc(50% - 8px);
    }
  }
  .add-chart-wrap {
    .bk-select-option {
      padding-left: 24px !important;
    }
    .no-data {
      color: #c4c6cc !important;
    }
  }
  :deep(.bk-exception) {
    height: 100%;
    justify-content: center;
    transform: translateY(-20px);
  }
</style>
