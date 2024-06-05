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
      <div
        v-if="chartSelectDimensions?.length"
        v-for="selectDimension in chartSelectDimensions"
        :key="selectDimension.primaryDimension"
        class="chart">
        <Chart
          :primary-dimension="selectDimension.primaryDimension"
          :all-label="addChartData!.labels"
          :select-dimension="selectDimension.minorDimension"
          :drill-dimension="selectDimension.drillDownDimension"
          :bk-biz-id="bkBizId"
          :app-id="appId"
          @select="handleSelectDimension" />
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
  import { getClientLabelsAndAnnotations } from '../../../../../../api/client';
  import { storeToRefs } from 'pinia';
  import useClientStore from '../../../../../../store/client';
  import SectionTitle from '../../components/section-title.vue';
  import Chart from './chart/index.vue';
  import Card from '../../components/card.vue';

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const selectedLabel = ref<string[]>([]);
  const chartSelectDimensions = ref<
    {
      primaryDimension: string;
      minorDimension: string[];
      drillDownDimension: string;
    }[]
  >([]);

  const addChartData = ref<{
    annotations: string[];
    labels: string[];
  }>({
    annotations: [],
    labels: [],
  });

  watch(
    () => props.appId,
    async () => {
      await getAddChartDate();
      handleChartDimension();
    },
  );

  watch(
    () => searchQuery.value,
    async () => {
      await getAddChartDate();
      handleChartDimension();
    },
    { deep: true },
  );

  watch(
    () => selectedLabel.value,
    () => {
      selectedLabel.value.forEach((label) => {
        if (chartSelectDimensions.value?.findIndex((item) => item.primaryDimension === label) === -1) {
          chartSelectDimensions.value?.push({
            primaryDimension: label,
            minorDimension: [label],
            drillDownDimension: '',
          });
        }
      });
      chartSelectDimensions.value?.forEach((item, index) => {
        const label = selectedLabel.value.find((label) => label === item.primaryDimension);
        if (!label) {
          chartSelectDimensions.value?.splice(index, 1);
        }
      });
    },
  );

  // 缓存图表维度数据
  watch(
    () => chartSelectDimensions.value,
    () => {
      const localStorageKey = 'clientLabelAndAnnotationsSelectDimension';
      const jsonString = localStorage.getItem(localStorageKey);
      const allService = jsonString ? JSON.parse(jsonString) : [];
      const serviceIndex = allService.findIndex((item: any) => item.id === props.appId);
      if (serviceIndex !== -1) {
        allService[serviceIndex].selectedDimension = chartSelectDimensions.value;
      } else {
        allService.push({ id: props.appId, selectedDimension: chartSelectDimensions.value });
      }
      localStorage.setItem(localStorageKey, JSON.stringify(allService));
    },
    { deep: true },
  );

  onMounted(async () => {
    await getAddChartDate();
    handleChartDimension();
  });

  const getAddChartDate = async () => {
    try {
      const res = await getClientLabelsAndAnnotations(props.bkBizId, props.appId, {
        last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      });
      res.data.annotations.sort(sortByLowerCase);
      res.data.labels.sort(sortByLowerCase);
      addChartData.value = res.data;
    } catch (e) {
      console.error(e);
    }
  };

  const sortByLowerCase = (a: string, b: string) => {
    const str1 = a.toLowerCase();
    const str2 = b.toLowerCase();
    return str1.localeCompare(str2);
  };

  const handleSelectDimension = ({ primaryDimension, minorDimension, drillDownDimension }: any) => {
    const selectedItem = chartSelectDimensions.value!.find((item) => item.primaryDimension === primaryDimension);
    if (selectedItem) {
      Object.assign(selectedItem, { primaryDimension, minorDimension, drillDownDimension });
    }
  };

  // 获取缓存的维度数据
  const handleChartDimension = () => {
    const jsonString = localStorage.getItem('clientLabelAndAnnotationsSelectDimension');
    let selectedDimensions = [];
    if (jsonString) {
      const allService = JSON.parse(jsonString);
      const service = allService.find((item: any) => item.id === props.appId);
      selectedDimensions = service
        ? service.selectedDimension.filter((item: any) => addChartData.value.labels.includes(item.primaryDimension))
        : [];
    }
    selectedLabel.value =
      selectedDimensions.length > 0
        ? selectedDimensions.map((item: any) => item.primaryDimension)
        : addChartData.value.labels.slice(0, Math.min(2, addChartData.value.labels.length));
    chartSelectDimensions.value = selectedDimensions.map((item: any) => {
      return {
        primaryDimension: item.primaryDimension,
        minorDimension: item.minorDimension.filter((dimension: string) =>
          addChartData.value.labels.includes(dimension),
        ),
        drillDownDimension: addChartData.value.labels.includes(item.drillDownDimension) ? item.drillDownDimension : '',
      };
    });
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
