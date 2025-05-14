<template>
  <div>
    <div class="overflow-auto text-[14px] flex flex-col h-[calc(100vh-100px)]">
      <div class="p-[24px]">
        <div>
          <div class="mb-[12px]">{{ $t('cluster.nodeList.label.selectedNode') }}</div>
          <div class="flex flex-wrap">
            <span
              v-for="item, index in nodes"
              :key="index + item.nodeName"
              class="bg-[#F0F1F5] rounded-[2px] px-[10px] mb-[8px] mr-[8px]">
              {{ item.nodeName }}</span>
          </div>
        </div>
        <div class="mt-[16px]">
          <div class="mb-[12px]">{{ $t('cluster.nodeList.label.drainRange.text') }}</div>
          <bk-radio-group v-model="range" @change="handleRadioChange">
            <bk-radio class="block" :value="'all'">
              <span class="pl-[4px]">{{ $t('cluster.nodeList.label.drainRange.all') }}</span>
            </bk-radio>
            <bk-radio class="block !ml-0 mt-[8px]" :value="'labels'">
              <span class="pl-[4px]">{{ $t('cluster.nodeList.label.drainRange.labels') }}</span>
            </bk-radio>
          </bk-radio-group>
          <KeyValue
            class="ml-[22px] mt-[8px] w-[500px]"
            ref="keyValueRef"
            :show-header="false"
            :show-operate="false"
            :show-footer="false"
            :model-value="podLabels"
            :key-rules="[
              {
                message: $i18n.t('generic.validate.labelKey1'),
                validator: KEY_REGEXP
              }
            ]"
            :value-rules="[
              {
                message: $i18n.t('generic.validate.labelValue'),
                validator: VALUE_REGEXP
              }
            ]"
            @data-change="handleLabelChange" />
        </div>
        <div class="mt-[16px]">
          <div class="mb-[12px]">{{ $t('cluster.nodeList.label.terminationGracePeriod.text') }}</div>
          <bk-radio-group v-model="grace" @change="handleGraceChange">
            <bk-radio class="block" :value="'default'">
              <span class="pl-[4px]">{{ $t('cluster.nodeList.label.terminationGracePeriod.default') }}</span>
            </bk-radio>
            <bk-radio class="block !ml-0 mt-[8px]" :value="'seconds'">
              <i18n
                path="cluster.nodeList.label.terminationGracePeriod.seconds"
                class="pl-[4px] flex items-center">
                <template #seconds>
                  <bcs-input
                    v-model="seconds"
                    type="number"
                    :min="0"
                    :max="max"
                    class="w-[120px] mx-[10px]"
                    @change="handleSecondsChange">
                    <template slot="append">
                      <div class="group-text">{{ $t('units.suffix.seconds') }}</div>
                    </template>
                  </bcs-input>
                </template>
              </i18n>
            </bk-radio>
          </bk-radio-group>
        </div>
      </div>
      <div class="px-[24px] py-[16px] bg-[#F5F7FA] flex-1">
        <div class="mb-[14px]">
          <div>
            <span class="text-[16px] font-bold">{{ $t('cluster.nodeList.label.preview') }}</span>
            <span v-show="isShowFilterTip">
              <i class="bcs-icon bcs-icon-alarm-insufficient text-[#FF9C01]"></i>
              <span class="text-[12px] text-[#979BA5]">( {{ $t('cluster.nodeList.tips.filterTip') }} )</span>
            </span>
          </div>
          <div class="flex justify-between">
            <i18n class="leading-[24px]" path="cluster.nodeList.label.drainNum">
              <template #podNum>
                <span class="font-bold">{{ podList.length }}</span>
              </template>
              <template #drainNum>
                <span class="font-bold">{{ drainNum }}</span>
              </template>
              <template #notDrainNum>
                <span class="font-bold">{{ podList.length - drainNum }}</span>
              </template>
            </i18n>
            <bk-checkbox
              class="select-none flex-shrink-0 leading-[24px]"
              v-model="isShowDrainOnly">
              {{ $t('cluster.nodeList.label.isDrainOnly') }}
            </bk-checkbox>
          </div>
        </div>
        <bcs-table
          :data="curPageData"
          :pagination="pagination"
          v-bkloading="{ isLoading: podLoading }"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange"
          @filter-change="handleFilterChange">
          <bcs-table-column
            :label="$t('generic.label.name')"
            :filters="filtersDataSource.podNames"
            :filter-method="filterMethod"
            column-key="podName"
            show-overflow-tooltip
            prop="podName">
            <template #default="{ row }">
              <span>{{ row.podName || '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.nodeList.label.namespace')"
            :filters="filtersDataSource.namespaces"
            :filter-method="filterMethod"
            column-key="nameSpace"
            show-overflow-tooltip
            prop="nameSpace">
            <template #default="{ row }">
              <span>{{ row.nameSpace || '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.label.status')"
            :filters="filtersDataSource.status"
            :filter-method="filterMethod"
            column-key="podStatus"
            show-overflow-tooltip
            prop="podStatus">
            <template #default="{ row }">
              <span>{{ row.podStatus || '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.label.kind')"
            :filters="filtersDataSource.types"
            :filter-method="filterMethod"
            show-overflow-tooltip
            column-key="podServiceAccount"
            prop="podServiceAccount">
            <template #default="{ row }">
              <span>{{ row.podServiceAccount || '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.nodeList.label.node')"
            :filters="filtersDataSource.nodes"
            :filter-method="filterMethod"
            column-key="node"
            show-overflow-tooltip
            prop="node">
            <template #default="{ row }">
              <span>{{ row.node || '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.nodeList.label.termination')"
            prop="gracePeriodSeconds">
            <template #default="{ row }">
              <span>{{ typeof row.gracePeriodSeconds === 'number' ? row.gracePeriodSeconds : '--' }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.nodeList.label.drainRisk')"
            prop="evictionRisk">
            <template #default="{ row }">
              <div class="py-[15px]" v-if="row.evictionRisk?.length">
                <div
                  v-for="risk, index in row.evictionRisk || []" :key="index"
                  :class="[
                    'overflow-auto',
                    risk?.riskParameter ? 'underline decoration-dashed underline-offset-[3px]' : ''
                  ]"
                  v-bk-tooltips="{
                    content: risk?.riskParameter,
                    disabled: !risk?.riskParameter,
                  }"
                >
                  {{ risk?.riskDescription }}
                </div>
              </div>
              <span v-else>--</span>
            </template>
          </bcs-table-column>
        </bcs-table>
      </div>
    </div>
    <div class="px-[24px] py-[8px] bg-[#fff]">
      <bcs-button
        theme="primary"
        class="min-w-[88px]"
        :loading="isSubmitting"
        @click="handleSubmit">
        {{ $t('cluster.nodeList.button.drainConfirm', { num: drainNum }) }}
      </bcs-button>
      <bcs-button
        theme="default"
        class="min-w-[88px]"
        @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
    </div>
  </div>
</template>
<script lang="ts">
import { debounce } from 'lodash';
import { computed, defineComponent, onBeforeMount, PropType, ref, watch } from 'vue';

import { drainCheckList, schedulerNode } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { KEY_REGEXP, VALUE_REGEXP } from '@/common/constant';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import KeyValue from '@/components/key-value.vue';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

type Risk = {
  riskDescription: string
  riskParameter: string
};

interface IPod {
  evictionRisk: Risk[]
  gracePeriodSeconds: number
  nameSpace: string
  node: string
  podName: string
  podServiceAccount: string
  podStatus: string
  willBeEvicted: boolean
}

export default defineComponent({
  name: 'PodDrain',
  components: { KeyValue },
  props: {
    nodes: {
      type: Array as PropType<any[]>,
      default: () => [],
    },
    clusterId: {
      type: String,
    },
  },
  setup(props, ctx) {
    const nodeNames = computed(() => props.nodes.map(v => v.nodeName));
    const nodeKey = ref();


    // 驱逐范围
    const keyValueRef = ref<InstanceType<typeof KeyValue> | null>(null);
    const range = ref('all');
    const labelValueStr = ref('');
    const podLabels = ref([]);
    function handleRadioChange(curRange) {
      if (curRange === 'all') {
        paramsData.value.podSelector = '';
      } else if (keyValueRef.value?.validate?.()) {
        paramsData.value.podSelector = labelValueStr.value;
      }
    };
    function labelChange(result) {
      const resultList = result.filter(item => !!item.key && !!item.value);
      const labelValues = resultList.map(item => `${item.key}=${item.value}`);
      labelValueStr.value = labelValues.join(',');
      if (range.value !== 'all' && keyValueRef.value?.validate?.()) {
        paramsData.value.podSelector = labelValueStr.value;
      }
    };
    const handleLabelChange = debounce(labelChange, 300);


    // Pod 终止宽限时间
    const grace = ref('default');
    const seconds = ref(0);
    const max = ref(600);
    function handleGraceChange(curGrace) {
      if (curGrace === 'default') {
        paramsData.value.gracePeriodSeconds = 0;
      } else {
        paramsData.value.gracePeriodSeconds = seconds.value;
      }
    }
    function secondsChange() {
      if (grace.value === 'seconds') {
        if (seconds.value > max.value) {
          seconds.value = max.value;
        }
        paramsData.value.gracePeriodSeconds = seconds.value;
      }
    }
    const handleSecondsChange = debounce(secondsChange, 300);


    // Pod 列表
    const podList = ref<IPod[]>([]);
    const drainNum = computed(() => podList.value.filter(v => v.willBeEvicted).length);
    const isShowDrainOnly = ref(false);
    const podLoading = ref(false);
    async function getPodList(params = {}) {
      podLoading.value = true;
      podList.value = await drainCheckList({
        clusterID: props.clusterId,
        nodes: nodeNames.value,
        ignoreAllDaemonSets: false,
        ...params,
      })
        .catch(() => [])
        .finally(() => {
          podLoading.value = false;
          handleResetPage();
        });
    }

    // 表格表头搜索项配置
    const filtersDataSource = computed(() => ({
      podNames: podNames.value,
      namespaces: namespaces.value,
      status: status.value,
      types: types.value,
      nodes: nodes.value,
    }));
    const isShowFilterTip = computed(() => isShowDrainOnly.value || hasFilterData.value);
    function dataConstructor(paramName) {
      const seen = new Set<string>();
      return podList.value
        .filter(item => item[paramName]) // 过滤无效项
        .reduce<Array<{ id: string; name: string; text: string; value: string }>>((acc, item) => {
        const data = item[paramName]!; // 非空断言（已由filter确保）
        if (!seen.has(data)) {
          seen.add(data);
          acc.push({
            id: item.podName,
            name: data,
            text: data,
            value: data,
          });
        }
        return acc;
      }, []);
    }
    const podNames = computed(() => dataConstructor('podName'));
    const namespaces = computed(() => dataConstructor('nameSpace'));
    const status = computed(() => dataConstructor('podStatus'));
    const types = computed(() => dataConstructor('podServiceAccount'));
    const nodes = computed(() => dataConstructor('node'));
    function filterMethod(value, row, column) {
      const { property } = column;
      return row[property] === value;
    }
    /**
     * 处理表头筛选项变化
     */
    const filterTableData = computed(() => {
      const list = Object.keys(filterData.value).filter(key => filterData.value[key].length > 0);
      if (!list.length && !isShowDrainOnly.value) return podList.value;

      return podList.value.filter(v => !isShowDrainOnly.value || v.willBeEvicted)
        .filter(row => list.every((key) => {
          const value = row[key];
          return filterData.value[key].includes(value);
        }));
    });
    const filterData = ref({});
    // 是否存在筛选项
    const hasFilterData = ref(false);
    function handleFilterChange(data, filterValue) {
      const key = Object.keys(data)[0];
      filterData.value = {
        ...filterData.value,
        [key]: data[key],
      };
      // eslint-disable-next-line no-restricted-syntax
      for (const key in filterValue) {
        if (filterValue[key].length > 0) {
          hasFilterData.value = true;
          return;
        }
      }
      hasFilterData.value = false;
    }
    // 分页后的表格数据
    const {
      curPageData,
      pagination,
      pageChange,
      pageSizeChange,
      handleResetPage,
    } = usePage(filterTableData);


    // Pod 驱逐操作
    const user = computed(() => $store.state.user);
    const paramsData = ref({
      clusterID: props.clusterId,
      nodes: nodeNames.value,
      gracePeriodSeconds: 0,
      podSelector: '',
      operator: user.value.username,
      ignoreAllDaemonSets: false,
    });
    const isSubmitting = ref(false);
    async function handleSubmit() {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.button.drain.title'),
        defaultInfo: true,
        theme: 'danger',
        confirmFn: async () => {
          const result = await schedulerNode(paramsData.value)
            .then(() => true)
            .catch(() => false);
          if (result) {
            ctx.emit('success');
            $bkMessage({
              theme: 'success',
              message: window.i18n.t('generic.msg.success.ok'),
            });
          }
        },
      });
    }
    function handleCancel() {
      ctx.emit('cancel');
    }

    watch(paramsData, () => {
      getPodList({
        gracePeriodSeconds: Number(paramsData.value.gracePeriodSeconds),
        podSelector: paramsData.value.podSelector,
      });
    }, { deep: true });

    watch(isShowFilterTip, () => {
      handleResetPage();
    });

    onBeforeMount(() => {
      getPodList();
    });

    return {
      range,
      nodeKey,
      seconds,
      grace,
      podList,
      podLoading,
      isSubmitting,
      filtersDataSource,
      isShowFilterTip,
      curPageData,
      pagination,
      max,
      paramsData,
      podLabels,
      drainNum,
      isShowDrainOnly,
      KEY_REGEXP,
      VALUE_REGEXP,
      keyValueRef,
      pageChange,
      pageSizeChange,
      handleSubmit,
      handleCancel,
      handleLabelChange,
      handleRadioChange,
      handleGraceChange,
      handleSecondsChange,
      filterMethod,
      handleFilterChange,
      handleResetPage,
    };
  },
});

</script>
