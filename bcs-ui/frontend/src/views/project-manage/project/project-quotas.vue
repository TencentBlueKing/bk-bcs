<template>
  <BcsContent :title="$t('projects.project.quota')" hide-back>
    <LayoutGroup collapsible class="mb10">
      <template #title>
        <div class="flex-1 flex item-center justify-between">
          <span class="text-[14px] font-bold">{{ $t('projects.quota.projectStatistics') }}</span>
          <div class="flex items-center">
            <span class="ml-[20px] flex items-center">
              <span class="size-[14px] bg-[#3a84ff] inline-block mr-[5px]"></span>
              {{ $t('projects.quota.caUsed') }}</span>
            <span class="ml-[20px] flex items-center">
              <span class="size-[13px] border-[1px] border-[#3a84ff] bg-[#e1ecff] inline-block mr-[5px]"></span>
              {{ $t('projects.quota.caAvailable') }}</span>
            <span class="ml-[20px] flex items-center">
              <span class="size-[14px] bg-[#f59500] inline-block mr-[5px]"></span>
              {{ $t('projects.quota.federationUsed') }}</span>
            <span class="ml-[20px] flex items-center">
              <span class="size-[13px] border-[1px] border-[#f59500] bg-[#fdeed8] inline-block mr-[5px]"></span>
              {{ $t('projects.quota.federationAvailable') }}</span>
          </div>
        </div>
      </template>
      <div class="flex justify-between">
        <div
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm mr-[20px]">
          <div>
            <div>{{ $t('projects.quota.cpuResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ cpuSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.cpuMsg', {
                used: statisticsObj?.cpu?.usedNum || 0,
                available: statisticsObj?.cpu?.availableNum || 0 })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', cpuSum === 0 ? 'grayscale' : '']"
            :options="cpuOptions"
            ref="chartRef">
          </ECharts>
        </div>
        <div
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm mr-[20px]">
          <div>
            <div>{{ $t('projects.quota.memResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ memSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.memMsg', {
                used: statisticsObj?.mem?.usedNum || 0,
                available: statisticsObj?.mem?.availableNum || 0 })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', memSum === 0 ? 'grayscale' : '']"
            :options="memOptions"
            ref="chartRef">
          </ECharts>
        </div>
        <div
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm">
          <div>
            <div>{{ $t('projects.quota.gpuResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ gpuSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.gpuMsg', {
                used: statisticsObj?.gpu?.usedNum || 0,
                available: statisticsObj?.gpu?.availableNum || 0 })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', gpuSum === 0 ? 'grayscale' : '']"
            :options="gpuOptions"
            ref="chartRef">
          </ECharts>
        </div>
      </div>
    </LayoutGroup>
    <!-- 容量统计 -->
    <div class="flex items-center justify-between mb10">
      <div class="w-full max-w-[400px]">
        <bk-radio-group v-model="statisticsType">
          <bk-radio-button value="host">{{ $t('projects.quota.CA') }}</bk-radio-button>
          <bk-radio-button value="federation">{{ $t('projects.quota.federation') }}</bk-radio-button>
          <bk-radio-button value="self_host" class="!ml-[-2px]">{{ $t('tkeCa.label.provider.self') }}</bk-radio-button>
        </bk-radio-group>
        <bcs-search-select
          :key="searchSelectKey"
          clearable
          class="bg-[#fff] mt-[8px]"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('projects.quota.placeholder')"
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClearSearchSelect">
        </bcs-search-select>
      </div>
      <div class="flex text-[14px]">
        <div
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]">
          <div>{{ $t('projects.quota.cpuResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ statisticsObj?.cpu?.useRate || '0.00' }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.cpuMsgA', {
                sum: statisticsObj?.cpu?.totalNum || 0,
                used: statisticsObj?.cpu?.usedNum || 0 }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="(statisticsObj?.cpu?.useRate || 0) / 100"
            :stroke-width="6"
            :color="getColor(statisticsObj?.cpu?.useRate || 0)">
          </bcs-progress>
        </div>
        <div
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]">
          <div>{{ $t('projects.quota.memResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ statisticsObj?.mem?.useRate || '0.00' }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.memMsgA', {
                sum: statisticsObj?.mem?.totalNum || 0,
                used: statisticsObj?.mem?.usedNum || 0 }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="(statisticsObj?.mem?.useRate || 0) / 100"
            :stroke-width="6"
            :color="getColor(statisticsObj?.mem?.useRate || 0)">
          </bcs-progress>
        </div>
        <div
          v-if="statisticsType === 'federation'"
          v-bkloading="{ isLoading: statisticLoading }"
          class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]">
          <div>{{ $t('projects.quota.gpuResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ statisticsObj?.gpu?.useRate || '0.00' }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.gpuMsgA', {
                sum: statisticsObj?.gpu?.totalNum || 0,
                used: statisticsObj?.gpu?.usedNum || 0 }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="(statisticsObj?.gpu?.useRate || 0) / 100"
            :stroke-width="6"
            :color="getColor(statisticsObj?.gpu?.useRate || 0)">
          </bcs-progress>
        </div>
      </div>
    </div>
    <bk-table
      v-bkloading="{ isLoading }"
      :size="tableSetting.size"
      :data="curPageData"
      :key="tableKey"
      :pagination="pagination"
      class="network-table"
      @filter-change="handleFilter"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <template v-if="hostTypes.includes(statisticsType)">
        <bk-table-column
          :label="$t('projects.quota.label.instance')"
          :filters="filtersDataSource.instanceTypes"
          :filtered-value="filteredValue.instanceType"
          filter-searchable
          column-key="instanceType"
          fixed
          min-width="150px"
          prop="instanceType"
          v-if="isColumnRender('instanceType')" />
        <bk-table-column
          :label="$t('projects.quota.label.region')"
          :filters="filtersDataSource.regions"
          :filtered-value="filteredValue.region"
          column-key="region"
          fixed
          width="100px"
          prop="region"
          v-if="isColumnRender('region')" />
        <bk-table-column
          :label="$t('projects.quota.label.zone')"
          :filters="filtersDataSource.zones"
          :filtered-value="filteredValue.zoneName"
          column-key="zoneName"
          fixed
          width="120px"
          prop="zoneName"
          v-if="isColumnRender('zoneName')" />
        <bk-table-column
          :label="$t('generic.label.timeLeft')"
          column-key="purchaseDurationSettings"
          fixed
          min-width="200px"
          show-overflow-tooltip
          v-if="isColumnRender('purchaseDurationSettings') && statisticsType === 'self_host'">
          <template #default="{ row }">
            <span
              class="bcs-border-tips"
              v-bk-tooltips="row.current?.quotaAttr?.endTime">
              {{ formatHours(row.current?.quotaAttr?.purchaseDurationSettings) }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          sortable
          :label="$t('projects.quota.label.num')"
          prop="quotaNum"
          v-if="isColumnRender('quotaNum')" />
        <bk-table-column
          sortable
          :label="$t('projects.quota.label.used')"
          prop="quotaUsed"
          v-if="isColumnRender('quotaUsed')">
          <template #default="{ row }">
            <bk-button @click="handleUsed(row)" text>{{ row?.quotaUsed }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column
          sortable
          :label="$t('projects.quota.label.available')"
          prop="quotaAvailable"
          v-if="isColumnRender('quotaAvailable')" />
      </template>
      <template v-else>
        <bk-table-column
          :label="$t('projects.quota.label.resource')"
          filter-searchable
          column-key="quotaName"
          fixed
          min-width="200px"
          prop="quotaName"
          show-overflow-tooltip
          v-if="isColumnRender('quotaName')" />
        <bk-table-column
          :label="$t('projects.quota.label.clusterID')"
          column-key="clusterId"
          fixed
          width="150px"
          prop="clusterId"
          v-if="isColumnRender('clusterId')">
          <template #default="{ row }">
            <div class="overflow-hidden text-nowrap text-ellipsis">{{ clusterMap[row.clusterId] || '--' }}</div>
            <div class="overflow-hidden text-nowrap text-ellipsis">{{ row.clusterId }}</div>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('projects.quota.label.namespace')"
          column-key="nameSpace"
          fixed
          width="120px"
          prop="nameSpace"
          v-if="isColumnRender('nameSpace')" />
        <bk-table-column
          :label="$t('projects.quota.label.annotations')"
          :show-overflow-tooltip="false"
          column-key="annotations"
          width="120px"
          prop="annotations"
          v-if="isColumnRender('annotations')">
          <template #default="{ row }">
            <!-- 多标签展示 -->
            <!-- <div
              v-if="row?.annotations.length"
              class="py-[15px] overflow-auto flex flex-wrap">
              <bcs-tag
                class="bcs-ellipsis"
                v-bk-overflow-tips
                v-for="item in row?.annotations?.slice(0, 2)"
                :key="item.key">
                {{ `${item.key}=${item.value}` }}
              </bcs-tag>
              <bcs-tag
                v-if="row?.annotations.length > 2"
                v-bk-tooltips="{
                  allowHTML: true,
                  content: '#more-annotations',
                  duration: 300
                }">
                +{{ row?.annotations.length - 2 }}
              </bcs-tag>
              <div id="more-annotations">
                <div v-for="item in row?.annotations?.slice(2)" :key="item.key">
                  {{ `${item.key}=${item.value}` }}
                </div>
              </div>
            </div> -->
            <!-- 定制标签展示 -->
            <div class="flex items-center" v-if="row?.annotations?.[IS_EXCLUSIVE]">
              <bcs-tag
                class="bcs-ellipsis"
                v-bk-overflow-tips>
                {{ row?.annotations?.[IS_EXCLUSIVE] === 'true' ?
                  $t('projects.quota.tags.reserve')
                  : $t('projects.quota.tags.demande')
                }}
              </bcs-tag>
            </div>
            <div v-else>--</div>
          </template>
        </bk-table-column>
        <bk-table-column
          sortable
          width="130px"
          :label="$t('projects.quota.label.gpuNum')"
          prop="gpuNum"
          v-if="isColumnRender('gpuNum')" />
        <bk-table-column
          sortable
          width="130px"
          :label="$t('projects.quota.label.gpuUsed')"
          prop="gpuUsed"
          v-if="isColumnRender('gpuUsed')">
          <template #default="{ row }">
            <span>{{ row?.gpuUsed }}</span>
          </template>
        </bk-table-column>
        <bk-table-column
          sortable
          width="130px"
          :label="$t('projects.quota.label.gpuAvailable')"
          prop="gpuAvailable"
          v-if="isColumnRender('gpuAvailable')">
          <template #default="{ row }">
            {{ smartFormat(row.gpuAvailable) }}
          </template>
        </bk-table-column>
      </template>
      <bk-table-column
        width="130px"
        :label="$t('generic.label.status')"
        prop="status"
        v-if="isColumnRender('status')">
        <template #default="{ row }">
          <StatusIcon
            :status-color-map="statusColorMap"
            :status-text-map="statusTextMap"
            :status="row.current?.status"
            :pending="loadingStatusList.includes(row.current?.status)"
          />
        </template>
      </bk-table-column>
      <bk-table-column
        sortable
        width="130px"
        :label="$t('projects.quota.label.cpuNum')"
        prop="cpuNum"
        v-if="isColumnRender('cpuNum')" />
      <bk-table-column
        sortable
        width="130px"
        :label="$t('projects.quota.label.cpuUsed')"
        prop="cpuUsed"
        v-if="isColumnRender('cpuUsed')" />
      <bk-table-column
        sortable
        width="130px"
        :label="$t('projects.quota.label.cpuAvailable')"
        prop="cpuAvailable"
        v-if="isColumnRender('cpuAvailable')">
        <template #default="{ row }">
          {{ smartFormat(row.cpuAvailable) }}
        </template>
      </bk-table-column>
      <bk-table-column
        sortable
        width="140px"
        :label="$t('projects.quota.label.memNum')"
        prop="memNum"
        v-if="isColumnRender('memNum')" />
      <bk-table-column
        sortable
        width="150px"
        :label="$t('projects.quota.label.memUsed')"
        prop="memUsed"
        v-if="isColumnRender('memUsed')" />
      <bk-table-column
        sortable
        width="150px"
        :label="$t('projects.quota.label.memAvailable')"
        prop="memAvailable"
        v-if="isColumnRender('memAvailable')">
        <template #default="{ row }">
          {{ smartFormat(row.memAvailable) }}
        </template>
      </bk-table-column>
      <bk-table-column
        sortable
        width="100px"
        :label="$t('projects.quota.label.usage')"
        prop="usageRate"
        v-if="isColumnRender('usageRate')">
        <template #default="{ row }">
          <span>{{ row?.usageRate || '--' }}%</span>
        </template>
      </bk-table-column>
      <bk-table-column type="setting" :tippy-options="{ zIndex: 3000 }">
        <bcs-table-setting-content
          :fields="tableSetting.fields"
          :selected="tableSetting.selectedFields"
          :size="tableSetting.size"
          @setting-change="handleSettingChange">
        </bcs-table-setting-content>
      </bk-table-column>
    </bk-table>
    <!-- 配额使用情况 -->
    <bcs-sideslider
      :is-show.sync="showQuotaUsage"
      quick-close
      :title="`${curCA?.quota?.zoneResources?.instanceType}
        ${curCA?.quota?.zoneResources?.region}
        ${curCA?.quota?.zoneResources?.zoneName}
        ${$t('projects.quota.usage')}`"
      :width="800">
      <template #content>
        <QuotaUsage :data="curCA" />
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script lang="ts">
import dayjs from 'dayjs';
import { cloneDeep } from 'lodash';
import { computed, defineComponent, nextTick, onBeforeMount, ref, watch } from 'vue';

import QuotaUsage from './quota-usage.vue';

import { fetchProjectQuotas, fetchProjectQuotasStatistics, fetchProjectQuotasV2 } from '@/api/modules/project';
import ECharts from '@/components/echarts.vue';
import BcsContent from '@/components/layout/Content.vue';
import StatusIcon from '@/components/status-icon';
import { useProject } from '@/composables/use-app';
import usePage from '@/composables/use-page';
import useTableSearchSelect, { ISearchSelectData }  from '@/composables/use-table-search-select';
import useTableSetting from '@/composables/use-table-setting';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';

interface Iquota {
  current?: any;
  quotaName?: string;
  clusterId?: string;
  clusterName?: string;
  nameSpace?: string;
  annotations?: Record<string, string>;
  instanceType?: string;
  region?: string;
  zoneName?: string;
  quotaNum?: number;
  quotaUsed?: number;
  quotaAvailable?: number;
  gpuNum?: number;
  gpuUsed?: number;
  gpuAvailable?: number;
  cpuNum?: number;
  cpuUsed?: number;
  cpuAvailable?: number;
  memNum?: number;
  memUsed?: number;
  memAvailable?: number;
  usageRate?: string;
};

export default defineComponent({
  name: 'ProjectQuotas',
  components: { BcsContent, LayoutGroup, ECharts, QuotaUsage, StatusIcon },
  setup() {
    const { curProject } = useProject();

    const IS_EXCLUSIVE = 'bkbcs.tencent.com/is-exclusive';

    const sourceData = ref<any[]>([]);
    const tableData = computed(() => (statisticsType.value === 'self_host'
      ? selfQuotaData.value
      : sourceData.value.filter(item => item.current.quotaType === statisticsType.value)
    ));
    const statisticsObj = ref<Partial<{
      cpu: Record<string, number>;
      gpu: Record<string, number>;
      mem: Record<string, number>;
    }>>({});
    const isLoading = ref(false);
    // 状态文案
    const statusTextMap = {
      'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
      'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
      CREATING: $i18n.t('generic.status.creating'),
      RUNNING: $i18n.t('generic.status.ready'),
      DELETING: $i18n.t('generic.status.deleting'),
      DELETED: $i18n.t('generic.status.deleted'),
    };
    // 状态icon颜色
    const statusColorMap = {
      'CREATE-FAILURE': 'red',
      'DELETE-FAILURE': 'red',
      RUNNING: 'green',
      DELETED: 'gray',
    };
    const loadingStatusList = ['CREATING', 'DELETING'];

    // 从路由查询参数初始化 statisticsType
    const validTypes = ['host', 'federation', 'self_host'] as const;
    const initType = validTypes.includes($router.currentRoute.query.type as any)
      ? $router.currentRoute.query.type as 'host' | 'federation' | 'self_host'
      : 'host';
    const statisticsType = ref<'host' | 'federation' | 'self_host'>(initType);
    const hostTypes = ['host', 'self_host'];
    const fields = computed(() => [
      {
        id: 'quotaName',
        label: $i18n.t('projects.quota.label.resource'),
        disabled: hostTypes.includes(statisticsType.value),
      },
      {
        id: 'clusterId',
        label: $i18n.t('projects.quota.label.clusterID'),
        disabled: hostTypes.includes(statisticsType.value),
      },
      {
        id: 'nameSpace',
        label: $i18n.t('projects.quota.label.namespace'),
        disabled: hostTypes.includes(statisticsType.value),
      },
      {
        id: 'annotations',
        label: $i18n.t('projects.quota.label.annotations'),
        disabled: hostTypes.includes(statisticsType.value),
      },
      {
        id: 'instanceType',
        label: $i18n.t('projects.quota.label.instance'),
        disabled: true,
      },
      {
        id: 'region',
        label: $i18n.t('projects.quota.label.region'),
        disabled: !hostTypes.includes(statisticsType.value),
      },
      {
        id: 'zoneName',
        label: $i18n.t('projects.quota.label.zone'),
        disabled: !hostTypes.includes(statisticsType.value),
      },
      {
        id: 'purchaseDurationSettings',
        label: $i18n.t('generic.label.timeLeft'),
        disabled: statisticsType.value !== 'self_host',
      },
      {
        id: 'quotaNum',
        label: $i18n.t('projects.quota.label.num'),
        disabled: !hostTypes.includes(statisticsType.value),
      },
      {
        id: 'quotaUsed',
        label: $i18n.t('projects.quota.label.used'),
        disabled: !hostTypes.includes(statisticsType.value),
      },
      {
        id: 'quotaAvailable',
        label: $i18n.t('projects.quota.label.available'),
        disabled: !hostTypes.includes(statisticsType.value),
      },
      {
        id: 'gpuNum',
        label: $i18n.t('projects.quota.label.gpuNum'),
        disabled: statisticsType.value !== 'federation',
      },
      {
        id: 'gpuUsed',
        label: $i18n.t('projects.quota.label.gpuUsed'),
        disabled: statisticsType.value !== 'federation',
      },
      {
        id: 'gpuAvailable',
        label: $i18n.t('projects.quota.label.gpuAvailable'),
        disabled: statisticsType.value !== 'federation',
      },
      {
        id: 'status',
        label: $i18n.t('generic.label.status'),
      },
      {
        id: 'cpuNum',
        label: $i18n.t('projects.quota.label.cpuNum'),
      },
      {
        id: 'cpuUsed',
        label: $i18n.t('projects.quota.label.cpuUsed'),
      },
      {
        id: 'cpuAvailable',
        label: $i18n.t('projects.quota.label.cpuAvailable'),
      },
      {
        id: 'memNum',
        label: $i18n.t('projects.quota.label.memNum'),
      },
      {
        id: 'memUsed',
        label: $i18n.t('projects.quota.label.memUsed'),
      },
      {
        id: 'memAvailable',
        label: $i18n.t('projects.quota.label.memAvailable'),
      },
      {
        id: 'usageRate',
        label: $i18n.t('projects.quota.label.usage'),
      },
    ]);
    const {
      tableSetting,
      fieldsDataClone,
      handleSettingChange,
      isColumnRender,
    } = useTableSetting(fields.value);
    // 获取自建资源池 项目配额
    const selfQuotaData = ref<any[]>([]);
    async function handleGetSelfQuotas() {
      if (!curProject.value.projectID) return;
      // if (selfQuotaData.value.length === 0) {
      isLoading.value = true;
      const res = await fetchProjectQuotasV2({
        $projectId: curProject.value.projectID,
        quotaType: 'self_host',
        provider: 'internal',
      })
        .catch(() => ({ results: [] }))
        .finally(() => {
          isLoading.value = false;
        });
      selfQuotaData.value = handleProcessQuotas(res.results);
      // };
    };
    // 获取项目配额
    async function handleGetProjectQuotas() {
      if (!curProject.value.projectID) return;
      // if (sourceData.value.length === 0) {
      isLoading.value = true;
      const res = await fetchProjectQuotas({
        projectID: curProject.value.projectID,
        provider: 'selfProvisionCloud',
      }).catch(() => ({ results: [] }));
      isLoading.value = false;
      sourceData.value = handleProcessQuotas(res.results);
      // }
    }
    // 获取项目统计(ECharts) 部分数据
    const statisticLoading = ref(false);
    async function handelGetStatistics() {
      if (!curProject.value.projectID) return;
      statisticLoading.value = true;
      statisticsObj.value = await fetchProjectQuotasStatistics({
        $projectID: curProject.value.projectID,
        quotaType: statisticsType.value,
      })
        .catch(() => ({}))
        .finally(() => {
          statisticLoading.value = false;
        });

      // 整理饼图数据
      getChartData();
    }
    // 整理接口数据
    function handleProcessQuotas(data) {
      // const list = data?.filter(item => item?.status === 'RUNNING') || [];
      const result = data.reduce((acc, cur) => {
        const isHost = cur.quotaType === 'host';
        const zoneResources = cur.quota?.zoneResources || {};
        const obj: Iquota = {};
        obj.current = cur;
        obj.instanceType = zoneResources?.instanceType || '--';
        obj.region = zoneResources?.region || '--';
        obj.zoneName = zoneResources?.zoneName || '--';
        obj.quotaName = cur?.quotaName || '--';
        obj.clusterId = cur?.clusterId || '--';
        obj.clusterName = cur?.clusterName || '--';
        obj.nameSpace = cur?.nameSpace || '--';
        obj.annotations = cur?.annotations || {};
        // 多标签展示
        // const annotationsObj = cur?.annotations || {};
        // obj.annotations = Object.entries(annotationsObj).map(([key, value]) => ({ key, value }));


        obj.quotaNum = zoneResources.quotaNum || 0;
        obj.quotaUsed = zoneResources.quotaUsed || 0;
        obj.quotaAvailable = (zoneResources.quotaNum || 0) - (obj.quotaUsed || 0);
        obj.cpuNum = isHost ? (zoneResources.cpu || 0) * (obj.quotaNum || 0) // cpu 总量
          : (Number(cur.quota.cpu?.deviceQuota ?? 0) || 0);
        obj.cpuUsed = isHost ? (zoneResources.cpu || 0) * (obj.quotaUsed || 0) // cpu 已用
          : (Number(cur.quota.cpu?.deviceQuotaUsed ?? 0) || 0);
        obj.cpuAvailable = isHost ? (zoneResources.cpu || 0) * (obj.quotaAvailable || 0) // cpu 剩余
          : Number(isNaN(cur.quota.cpu?.deviceQuota - cur.quota.cpu?.deviceQuotaUsed) ? 0
            : smartFormat(cur.quota.cpu?.deviceQuota - cur.quota.cpu?.deviceQuotaUsed));

        obj.memNum = isHost ? (zoneResources.mem || 0) * (obj.quotaNum || 0) // mem 总量
          : extractNumber(cur.quota.mem?.deviceQuota);
        obj.memUsed = isHost ? (zoneResources.mem || 0) * (obj.quotaUsed || 0) // mem 已用
          : extractNumber(cur.quota.mem?.deviceQuotaUsed);
        obj.memAvailable = isHost ? (zoneResources.mem || 0) * (obj.quotaAvailable || 0) // mem 剩余
          : Math.max(extractNumber(cur.quota.mem?.deviceQuota) - extractNumber(cur.quota.mem?.deviceQuotaUsed), 0);

        obj.gpuNum = Number(cur.quota.gpu?.deviceQuota ?? 0) || 0; // gpu 总量
        obj.gpuUsed = Number(cur.quota.gpu?.deviceQuotaUsed ?? 0) || 0; // gpu 已用
        obj.gpuAvailable = Number(isNaN(cur.quota.gpu?.deviceQuota - cur.quota.gpu?.deviceQuotaUsed) ? 0 // gpu 剩余
          : smartFormat(cur.quota.gpu?.deviceQuota - cur.quota.gpu?.deviceQuotaUsed));

        // eslint-disable-next-line no-nested-ternary
        const rate = isHost ? ((obj.quotaUsed || 0) / (obj.quotaNum || 1))
          : (obj.gpuNum
            ? (obj.gpuUsed / (obj.gpuNum || 1))
            : (obj.cpuUsed / (obj.cpuNum || 1)));
        obj.usageRate = ((rate) * 10000 / 100)
          .toFixed(2); // 保留两位小数

        acc.push(obj);
        return acc;
      }, []);
      return result;
    }
    // 处理可能带单位的数据，取数字部分
    function extractNumber(str) { // str = '40GiB'
      const digits = str?.match(/\d+\.?\d*/)?.[0]; // 安全处理空值
      return digits ? parseFloat(digits) : 0;
    };
    // 浮点数保留两位小数，整数不保留小数
    function smartFormat(value: number) {
      if (isNaN(value)) return 0;
      // 检测是否为整数（兼容浮点精度问题）
      const isInteger = Number.isInteger(value)
        || Math.abs(value - Math.round(value)) < 1e-10;

      return isInteger ? value.toFixed(0) : value.toFixed(2);
    };
    const parseSearchSelectValue = computed(() => {
      const searchValues: { id: string; value: string[] }[] = [];
      searchSelectValue.value.forEach((item) => {
        const tmp: string[] = item.values.map(v => v.id);
        searchValues.push({
          id: item.id,
          value: tmp,
        });
      });
      return searchValues;
    });
    // 过滤后的表格数据(todo: 搜索性能优化)
    const filterTableData = computed(() => {
      if (!parseSearchSelectValue.value.length) return tableData.value;

      return tableData.value.filter(row => parseSearchSelectValue.value.every((item) => {
        const val = row || {};
        return item.value.some(v => val[item.id] === v);
      }));
    });
    // 分页后的表格数据
    const {
      curPageData,
      pagination,
      pageChange,
      pageSizeChange,
      handleResetPage,
      pageConf,
    } = usePage(filterTableData);


    // 节点池详情
    const handleGotoDetail = (nodePool) => {
      $router.push({
        name: 'nodePoolDetail',
        params: {
          clusterId: nodePool.clusterId,
          nodeGroupID: nodePool.nodeGroupId,
        },
      }).catch((err) => {
        console.warn(err);
      });
    };
    // 配额使用情况
    const showQuotaUsage = ref(false);
    const curCA = ref();
    function handleUsed(row) {
      curCA.value = row.current;
      showQuotaUsage.value = true;
    }

    // 饼图item颜色
    const itemColors = ['#3a84ff', '#f59500', '#e1ecff', '#fdeed8'];
    // 饼图基础配置
    const options = {
      tooltip: {
        trigger: 'item',
      },
      legend: {
        show: false,
      },
      series: [
        {
          name: '',
          type: 'pie',
          radius: ['70%', '90%'],
          avoidLabelOverlap: false,
          padAngle: 0,
          itemStyle: {
            borderRadius: 0,
          },
          color: itemColors,
          tooltip: {
            formatter: params => `${params.name}: ${params.value}`,
          },
          label: {
            show: true,
            position: 'center',
            formatter: '{d}%',
            fontSize: 14,
            fontWeight: 'bold',
            width: 60,
            height: 60,
            backgroundColor: '#f5f7fa',
            borderRadius: 60,
            silent: true, // 阻止label触发鼠标事件
          },
          emphasis: {
            disabled: false,
            scale: true,
            scaleSize: 5,
            itemStyle: {
              color: 'inherit',
            },
          },
          labelLine: {
            show: false,
          },
          data: [] as Array<{ value: number; name: string }>,
        },
      ],
    };

    // 只有一项数据时，饼图不需要缺口
    const cpuData = ref([
      { value: 0, name: $i18n.t('projects.quota.caUsed') },
      { value: 0, name: $i18n.t('projects.quota.federationUsed') },
      { value: 0, name: $i18n.t('projects.quota.caAvailable') },
      { value: 0, name: $i18n.t('projects.quota.federationAvailable') },
    ]);
    const padAngleValue1 = computed(() => (cpuData.value.filter(item => item.value > 0).length === 1 ? 0 : 5));

    function cpuLabelFormatter(params) {
      if (params.dataIndex < 2) {
        return `${params.name}<br />
        <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
        ${$i18n.t('projects.quota.cpuMsgB', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
      }
      return `${params.name}<br />
      <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
      ${$i18n.t('projects.quota.cpuMsgC', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
    };
    // CPU 配置
    const cpuOptions = computed(() => {
      const temp = cloneDeep(options);
      temp.series[0].padAngle = padAngleValue1.value;
      temp.series[0].tooltip.formatter = cpuLabelFormatter;
      temp.series[0].data = cpuData.value;
      return temp;
    });


    // 内存配置
    const memData = ref([
      { value: 0, name: $i18n.t('projects.quota.caUsed') },
      { value: 0, name: $i18n.t('projects.quota.federationUsed') },
      { value: 0, name: $i18n.t('projects.quota.caAvailable') },
      { value: 0, name: $i18n.t('projects.quota.federationAvailable') },
    ]);
    const padAngleValue2 = computed(() => (memData.value.filter(item => item.value > 0).length === 1 ? 0 : 5));
    function memLabelFormatter(params) {
      if (params.dataIndex < 2) {
        return `${params.name}<br />
        <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
        ${$i18n.t('projects.quota.memMsgB', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
      }
      return `${params.name}<br />
      <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
      ${$i18n.t('projects.quota.memMsgC', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
    };
    const memOptions = computed(() => {
      const temp = cloneDeep(options);
      temp.series[0].padAngle = padAngleValue2.value;
      temp.series[0].tooltip.formatter = memLabelFormatter;
      temp.series[0].data = memData.value;
      return temp;
    });


    // GPU 配置
    const gpuData = ref([
      { value: 0, name: $i18n.t('projects.quota.caUsed') },
      { value: 0, name: $i18n.t('projects.quota.federationUsed') },
      { value: 0, name: $i18n.t('projects.quota.caAvailable') },
      { value: 0, name: $i18n.t('projects.quota.federationAvailable') },
    ]);
    const padAngleValue3 = computed(() => (gpuData.value.filter(item => item.value > 0).length === 1 ? 0 : 5));
    function gpuLabelFormatter(params) {
      if (params.dataIndex < 2) {
        return `${params.name}<br />
        <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
        ${$i18n.t('projects.quota.gpuMsgB', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
      }
      return `${params.name}<br />
      <span style="display: inline-block; width: 12px; height: 12px;background-color: ${params.color}"></span>
      ${$i18n.t('projects.quota.gpuMsgC', { percent: (Number(params.percent) * 100 / 100).toFixed(2), counts: params.data.value })}`;
    };
    const gpuOptions = computed(() => {
      const temp = cloneDeep(options);
      temp.series[0].padAngle = padAngleValue3.value;
      temp.series[0].tooltip.formatter = gpuLabelFormatter;
      temp.series[0].data = gpuData.value;
      return temp;
    });

    const chartRef = ref(null);
    const cpuSum = computed(() => statisticsObj.value?.cpu?.totalNum || 0);
    const memSum = computed(() => statisticsObj.value?.mem?.totalNum || 0);
    const gpuSum = computed(() => statisticsObj.value?.gpu?.totalNum || 0);

    // 整理饼图数据
    function getChartData() {
      const usedIndex = hostTypes.includes(statisticsType.value) ? 0 : 1;
      const availableIndex = hostTypes.includes(statisticsType.value) ? 2 : 3;

      for (let i = 0; i < cpuOptions.value.series[0].data.length; i++) {
        cpuOptions.value.series[0].data[i].value = 0;
        memOptions.value.series[0].data[i].value = 0;
        gpuOptions.value.series[0].data[i].value = 0;
      }

      cpuOptions.value.series[0].data[usedIndex].value = statisticsObj.value?.cpu?.usedNum || 0;
      cpuOptions.value.series[0].data[availableIndex].value = statisticsObj.value?.cpu?.availableNum || 0;
      cpuOptions.value.series[0].label.formatter = `${statisticsObj.value?.cpu?.useRate || '0.00'}%`;

      memOptions.value.series[0].data[usedIndex].value = statisticsObj.value?.mem?.usedNum || 0;
      memOptions.value.series[0].data[availableIndex].value = statisticsObj.value?.mem?.availableNum || 0;
      memOptions.value.series[0].label.formatter = `${statisticsObj.value?.mem?.useRate || '0.00'}%`;

      gpuOptions.value.series[0].data[usedIndex].value = statisticsObj.value?.gpu?.usedNum || 0;
      gpuOptions.value.series[0].data[availableIndex].value = statisticsObj.value?.gpu?.availableNum || 0;
      gpuOptions.value.series[0].label.formatter = `${statisticsObj.value?.gpu?.useRate || '0.00'}%`;
    }

    // 统计类型
    function getColor(percent) {
      let color = '#fff';
      if (percent > 0 && percent < 70) {
        color = '#3a84ff';
      } else if (percent >= 70 && percent < 90) {
        color = '#f59500';
      } else if (percent >= 90) {
        color = '#e71818';
      }
      return color;
    }

    // 规格
    const instanceTypes = computed(() => tableData.value.reduce((pre, item) => {
      if (item?.instanceType && !pre.find(data => data.id === item?.instanceType)) {
        pre.push({
          id: item?.instanceType,
          name: item?.instanceType,
          text: item?.instanceType,
          value: item?.instanceType,
        });
      }
      return pre;
    }, []));
    // 城市
    const regions = computed(() => tableData.value.reduce((pre, item) => {
      if (item?.region && !pre.find(data => data.id === item?.region)) {
        pre.push({
          id: item?.region,
          name: item?.region,
          text: item?.region,
          value: item?.region,
        });
      }
      return pre;
    }, []));
    // 可用区
    const zones = computed(() => tableData.value.reduce((pre, item) => {
      if (item?.zoneName && !pre.find(data => data.id === item?.zoneName)) {
        pre.push({
          id: item?.zoneName,
          name: item?.zoneName,
          text: item?.zoneName,
          value: item?.zoneName,
        });
      }
      return pre;
    }, []));
    // 表格表头搜索项配置
    const filtersDataSource = computed(() => ({
      instanceTypes: instanceTypes.value,
      regions: regions.value,
      zones: zones.value,
    }));
    // 表格搜索项选中值
    const filteredValue = ref({
      instanceType: [],
      region: [],
      zoneName: [],
    });
    // searchSelect搜索框可选值
    const searchSelectDataSource = computed<ISearchSelectData[]>(() => [
      {
        name: $i18n.t('projects.quota.label.instance'),
        id: 'instanceType',
        multiable: true,
        children: instanceTypes.value,
      },
      {
        name: $i18n.t('projects.quota.label.region'),
        id: 'region',
        multiable: true,
        children: regions.value,
      },
      {
        name: $i18n.t('projects.quota.label.zone'),
        id: 'zoneName',
        multiable: true,
        children: zones.value,
      },
    ]);
    const {
      tableKey,
      searchSelectData,
      searchSelectValue,
      handleFilterChange,
      handleSearchSelectChange,
      handleClearSearchSelect,
    } = useTableSearchSelect({
      searchSelectDataSource,
      filteredValue,
    });
    const searchSelectChange = (list) => {
      // handleResetCheckStatus();
      handleSearchSelectChange(list);
    };
    const searchSelectKey = ref(0);
    function handleFilter(filters) {
      const filtersData = cloneDeep(filteredValue.value);
      Object.keys(filteredValue.value).forEach((key) => {
        if (Object.keys(filteredValue.value[key]).length === 0) {
          delete filtersData[key];
        }
        filters[key] && (filtersData[key] = filters[key]);
      });
      handleFilterChange(filtersData);
      searchSelectKey.value += 1;
    };

    // 保留两位小数
    function truncateToTwoDecimals(value) {
      const num = typeof value === 'string' ? parseFloat(value) : value;
      // 截断到两位小数（不四舍五入）
      return Math.trunc(num * 100) / 100;
    }

    // 获取对象前三个属性的数组
    function getProperties(row, num) {
      const obj = row?.annotations || {};
      if (Object.keys(obj).length === 0) return [];
      return Object.keys(obj)?.slice(0, num);
    }

    // 集群列表
    const {
      clusterList,
      getClusterList,
    } = useClusterList();
    const clusterMap = computed(() => clusterList.value.reduce((pre, cur) => {
      // eslint-disable-next-line no-param-reassign
      pre[cur.clusterID] = cur.clusterName;
      return pre;
    }, {}));

    /**
     * @description: 时间格式转换 传入小时，返回 余X小时/天/月/年
     * @param {string} hours - 小时数字符串
     * @return {string} 格式化后的时间字符串
     */
    function formatHours(hours) {
      // 参数校验并转换为数字
      const hoursNum = Number(hours);
      if (isNaN(hoursNum) || hoursNum <= 0) {
        return $i18n.t('units.time.leftHours', { num: 0 });
      }

      // 小于24小时，显示小时数
      if (hoursNum < 24) {
        return $i18n.t('units.time.leftHours', { num: Math.floor(hoursNum) });
      }

      const now = dayjs();
      const future = now.add(hoursNum, 'hour');

      // 计算年份差
      const years = future.diff(now, 'year');
      if (years > 0) {
        return $i18n.t('units.time.leftYears', { num: years });
      }

      // 计算月份差
      const months = future.diff(now, 'month');
      if (months > 0) {
        return $i18n.t('units.time.leftMonths', { num: months });
      }

      // 计算天数差
      const days = future.diff(now, 'day');
      // 至少显示0天，避免负数
      return $i18n.t('units.time.leftDays', { num: Math.max(0, days) });
    }

    watch(statisticsType, async () => {
      // 同步路由查询参数
      const currentType = String($router.currentRoute.query.type || '');
      if (currentType !== statisticsType.value) {
        $router.replace({
          query: {
            ...$router.currentRoute.query,
            type: statisticsType.value,
          },
        });
      }

      await nextTick(); // 放在下一个tick，避免 loading 状态不更新
      if (statisticsType.value === 'self_host') {
        handleGetSelfQuotas();
      } else {
        handleGetProjectQuotas();
      }
      handelGetStatistics();
      // 重置页码
      pageConf.current = 1;
      // 刷新table setting columns数据
      fieldsDataClone.value = fields.value;
      // 刷新表格
      tableKey.value = new Date().getTime();
    });

    onBeforeMount(async () => {
      statisticLoading.value = true;
      isLoading.value = true;
      // 获取集群列表
      await getClusterList();
      // 获取项目配额
      if (statisticsType.value === 'self_host') {
        handleGetSelfQuotas();
      } else {
        handleGetProjectQuotas();
      }
      handelGetStatistics();
    });

    return {
      isLoading,
      statisticLoading,
      statisticsObj,
      cpuSum,
      memSum,
      gpuSum,
      cpuOptions,
      memOptions,
      gpuOptions,
      chartRef,
      statisticsType,
      searchSelectValue,
      tableKey,
      searchSelectData,
      curPageData,
      pagination,
      handleResetPage,
      pageConf,
      filtersDataSource,
      filteredValue,
      showQuotaUsage,
      curCA,
      tableSetting,
      searchSelectKey,
      clusterMap,
      IS_EXCLUSIVE,
      hostTypes,
      statusColorMap,
      statusTextMap,
      loadingStatusList,
      handleGotoDetail,
      getColor,
      searchSelectChange,
      handleFilterChange,
      handleClearSearchSelect,
      pageChange,
      pageSizeChange,
      handleUsed,
      handleSettingChange,
      isColumnRender,
      handleFilter,
      truncateToTwoDecimals,
      getProperties,
      smartFormat,
      formatHours,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .layout-group {
  .content {
    padding: 10px 0;
  }
  .title .name {
    flex: 1;
  }
}
</style>
