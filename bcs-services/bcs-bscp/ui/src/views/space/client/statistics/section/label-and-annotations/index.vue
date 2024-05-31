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
      <div v-if="selectedLabel?.length" v-for="primaryDimension in selectedLabel" :key="primaryDimension" class="chart">
        <Chart
          :primary-dimension="primaryDimension"
          :all-label="addChartData!.labels"
          :bk-biz-id="bkBizId"
          :app-id="appId" />
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

  const addChartData = ref<{
    annotations: string[];
    labels: string[];
  }>();

  watch(
    () => props.appId,
    () => {
      getAddChartDate();
    },
  );

  watch(
    () => searchQuery.value,
    () => {
      getAddChartDate();
    },
    { deep: true },
  );

  watch(
    () => selectedLabel.value,
    () => {},
  );

  onMounted(() => {
    getAddChartDate();
  });

  const getAddChartDate = async () => {
    try {
      const res = await getClientLabelsAndAnnotations(props.bkBizId, props.appId, {
        last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      });
      res.data.annotations.sort(sortByLowerCase);
      res.data.labels.sort(sortByLowerCase);
      addChartData.value = res.data;
      selectedLabel.value = addChartData.value?.labels.slice(0, 2) || addChartData.value?.labels.slice(0, 1) || [];
    } catch (e) {
      console.error(e);
    }
  };

  const sortByLowerCase = (a: string, b: string) => {
    const str1 = a.toLowerCase();
    const str2 = b.toLowerCase();
    return str1.localeCompare(str2);
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
