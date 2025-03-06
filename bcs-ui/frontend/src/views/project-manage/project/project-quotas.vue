<template>
  <BcsContent :title="$t('projects.project.quota')" hide-back v-bkloading="{ isLoading: isLoading }">
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
        <div class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm mr-[20px]">
          <div>
            <div>{{ $t('projects.quota.cpuResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ cpuSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.cpuMsg', {
                used: cpu.hostUsed + cpu.federationUsed,
                available: (cpu.hostSum + cpu.federationSum) - (cpu.hostUsed + cpu.federationUsed) })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', cpuSum === 0 ? 'grayscale' : '']"
            :options="cpuOptions"
            v-bkloading="{ isLoading: echartsLoading }"
            ref="chartRef">
          </ECharts>
        </div>
        <div class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm mr-[20px]">
          <div>
            <div>{{ $t('projects.quota.memResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ memSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.memMsg', {
                used: mem.hostUsed + mem.federationUsed,
                available: (mem.hostSum + mem.federationSum) - (mem.hostUsed + mem.federationUsed) })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', memSum === 0 ? 'grayscale' : '']"
            :options="memOptions"
            v-bkloading="{ isLoading: echartsLoading }"
            ref="chartRef">
          </ECharts>
        </div>
        <div class="bg-[#fff] flex-1 flex p-[10px] justify-between items-center shadow-sm">
          <div>
            <div>{{ $t('projects.quota.gpuResource') }}</div>
            <div class="my-[10px]">
              <span class="text-[30px] font-bold">{{ gpuSum }}</span>
              <span>{{ $t('projects.quota.sum') }}</span>
            </div>
            <div class="text-[12px] text-[#b7bac0]">
              {{ $t('projects.quota.gpuMsg', {
                used: gpu.hostUsed + gpu.federationUsed,
                available: (gpu.hostSum + gpu.federationSum) - (gpu.hostUsed + gpu.federationUsed) })
              }}
            </div>
          </div>
          <ECharts
            :class="['!size-[100px]', gpuSum === 0 ? 'grayscale' : '']"
            :options="gpuOptions"
            v-bkloading="{ isLoading: echartsLoading }"
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
        </bk-radio-group>
        <bcs-search-select
          :key="searchSelectKey"
          clearable
          class="bg-[#fff] mt-[8px]"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('projects.quota.placeholder')"
          default-focus
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClearSearchSelect">
        </bcs-search-select>
      </div>
      <div class="flex text-[14px]">
        <div class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]">
          <div>{{ $t('projects.quota.cpuResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ (cpuRateData.cupRate * 10000 / 100).toFixed(2) }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.cpuMsgA', { sum: cpuRateData.cupSum, used: cpuRateData.cupUsed }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="cpuRateData.cupRate || 0"
            :stroke-width="6"
            :color="getColor(cpuRateData.cupRate * 100)">
          </bcs-progress>
        </div>
        <div class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]">
          <div>{{ $t('projects.quota.memResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ (memRateData.memRate * 10000 / 100).toFixed(2) }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.memMsgA', { sum: memRateData.memSum, used: memRateData.memUsed }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="memRateData.memRate || 0"
            :stroke-width="6"
            :color="getColor(memRateData.memRate * 100)">
          </bcs-progress>
        </div>
        <div
          class="bg-[#fff] ml-[10px] flex-1 p-[10px] shadow-sm min-w-[200px] max-w-[300px]"
          v-if="statisticsType === 'federation'">
          <div>{{ $t('projects.quota.gpuResource') }}</div>
          <div class="bcs-ellipsis" v-bk-overflow-tips>
            <span>{{ (gpuRateData.gpuRate * 10000 / 100).toFixed(2) }}%</span>
            <span class="font-[400]">
              {{ $t('projects.quota.gpuMsgA', { sum: gpuRateData.gpuSum, used: gpuRateData.gpuUsed }) }}
            </span>
          </div>
          <bcs-progress
            class="mt-[10px]"
            :show-text="false"
            :percent="gpuRateData.gpuRate || 0"
            :stroke-width="6"
            :color="getColor(gpuRateData.gpuRate * 100)">
          </bcs-progress>
        </div>
      </div>
    </div>
    <bk-table
      :size="tableSetting.size"
      :data="curPageData"
      :key="tableKey"
      :pagination="pagination"
      class="network-table"
      @filter-change="handleFilter"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <template v-if="statisticsType === 'host'">
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
          v-if="isColumnRender('gpuAvailable')" />
      </template>
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
        v-if="isColumnRender('cpuAvailable')" />
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
        v-if="isColumnRender('memAvailable')" />
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
      <!-- <bk-table-column label="">
        <template #default="{ row }">
          <div
            v-for="item in row.nodeGroups"
            :key="item.nodeGroupId"
            class="bk-primary bk-button-normal bk-button-text"
            @click="handleGotoDetail(item)">
            {{ item.nodeGroupId }}
          </div>
          <span v-if="!row.nodeGroups?.length">--</span>
        </template>
      </bk-table-column> -->
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
import { cloneDeep } from 'lodash';
import { computed, defineComponent, onBeforeMount, ref, watch } from 'vue';

import QuotaUsage from './quota-usage.vue';

import { fetchProjectQuotas } from '@/api/modules/project';
import ECharts from '@/components/echarts.vue';
import BcsContent from '@/components/layout/Content.vue';
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
  components: { BcsContent, LayoutGroup, ECharts, QuotaUsage },
  setup() {
    const { curProject } = useProject();

    const sourceData = ref<any[]>([]);
    const tableData = computed(() => sourceData.value.filter(item => item.current.quotaType === statisticsType.value));
    const isLoading = ref(false);
    const statisticsType = ref<'host' | 'federation'>('host');
    const fields = computed(() => [
      {
        id: 'quotaName',
        label: $i18n.t('projects.quota.label.resource'),
        disabled: statisticsType.value === 'host',
      },
      {
        id: 'clusterId',
        label: $i18n.t('projects.quota.label.clusterID'),
        disabled: statisticsType.value === 'host',
      },
      {
        id: 'nameSpace',
        label: $i18n.t('projects.quota.label.namespace'),
        disabled: statisticsType.value === 'host',
      },
      {
        id: 'instanceType',
        label: $i18n.t('projects.quota.label.instance'),
        disabled: true,
      },
      {
        id: 'region',
        label: $i18n.t('projects.quota.label.region'),
        disabled: statisticsType.value !== 'host',
      },
      {
        id: 'zoneName',
        label: $i18n.t('projects.quota.label.zone'),
        disabled: statisticsType.value !== 'host',
      },
      {
        id: 'quotaNum',
        label: $i18n.t('projects.quota.label.num'),
        disabled: statisticsType.value !== 'host',
      },
      {
        id: 'quotaUsed',
        label: $i18n.t('projects.quota.label.used'),
        disabled: statisticsType.value !== 'host',
      },
      {
        id: 'quotaAvailable',
        label: $i18n.t('projects.quota.label.available'),
        disabled: statisticsType.value !== 'host',
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
    // 获取项目配额
    async function handleGetProjectQuotas() {
      if (!curProject.value.projectID) return;

      const res = await fetchProjectQuotas({
        projectID: curProject.value.projectID,
        provider: 'selfProvisionCloud',
      }).catch(() => ({ results: [] }));
      const list = res?.results?.filter(item => item?.status === 'RUNNING') || [];
      sourceData.value = list.reduce((acc, cur) => {
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
        obj.quotaNum = zoneResources.quotaNum || 0;
        obj.quotaUsed = zoneResources.quotaUsed || 0;
        obj.quotaAvailable = (zoneResources.quotaNum || 0) - (obj.quotaUsed || 0);
        obj.cpuNum = isHost ? (zoneResources.cpu || 0) * (obj.quotaNum || 0) // cpu 总量
          : Number(cur.quota.cpu?.deviceQuota || 0);
        obj.cpuUsed = isHost ? (zoneResources.cpu || 0) * (obj.quotaUsed || 0) // cpu 已用
          : Number(cur.quota.cpu?.deviceQuotaUsed || 0);
        obj.cpuAvailable = isHost ? (zoneResources.cpu || 0) * (obj.quotaAvailable || 0) // cpu 剩余
          : Number(isNaN(cur.quota.cpu?.deviceQuota - cur.quota.cpu?.deviceQuotaUsed) ? 0
            : cur.quota.cpu?.deviceQuota - cur.quota.cpu?.deviceQuotaUsed);

        obj.memNum = isHost ? (zoneResources.mem || 0) * (obj.quotaNum || 0) // mem 总量
          : extractNumber(cur.quota.mem?.deviceQuota);
        obj.memUsed = isHost ? (zoneResources.mem || 0) * (obj.quotaUsed || 0) // mem 已用
          : extractNumber(cur.quota.mem?.deviceQuotaUsed);
        obj.memAvailable = isHost ? (zoneResources.mem || 0) * (obj.quotaAvailable || 0) // mem 剩余
          : Math.max(extractNumber(cur.quota.mem?.deviceQuota) - extractNumber(cur.quota.mem?.deviceQuotaUsed), 0);

        obj.gpuNum = Number(cur.quota.gpu?.deviceQuota || 0); // gpu 总量
        obj.gpuUsed = Number(cur.quota.gpu?.deviceQuotaUsed || 0); // gpu 已用
        obj.gpuAvailable = Number(isNaN(cur.quota.gpu?.deviceQuota - cur.quota.gpu?.deviceQuotaUsed) ? 0 // gpu 剩余
          : cur.quota.gpu?.deviceQuota - cur.quota.gpu?.deviceQuotaUsed);

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
      // 整理饼图数据
      getChartData();
    }
    const extractNumber = (str) => { // str = '40GiB'
      const digits = str?.match(/\d+/g)?.join('') || '0'; // 安全处理空值
      return parseInt(digits, 10) || 0; // 转换为整数，失败则返回0
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

    const echartsLoading = ref(false);
    const chartRef = ref(null);
    const cpuSum = ref(0);
    const memSum = ref(0);
    const gpuSum = ref(0);
    const cpu = ref({
      hostSum: 0,
      hostUsed: 0,
      federationSum: 0,
      federationUsed: 0,
    });
    const mem = ref({
      hostSum: 0,
      hostUsed: 0,
      federationSum: 0,
      federationUsed: 0,
    });
    const gpu = ref({
      hostSum: 0,
      hostUsed: 0,
      federationSum: 0,
      federationUsed: 0,
    });
    const cpuRateData = computed(() => (statisticsType.value === 'host' ? {
      cupSum: cpu.value.hostSum,
      cupUsed: cpu.value.hostUsed,
      cupRate: cpu.value.hostUsed / cpu.value.hostSum || 0, // 已用/总和
    }
      : {
        cupSum: cpu.value.federationSum,
        cupUsed: cpu.value.federationUsed,
        cupRate: cpu.value.federationUsed / cpu.value.federationSum || 0, // 已用/总和
      }));
    const memRateData = computed(() => (statisticsType.value === 'host' ? {
      memSum: mem.value.hostSum,
      memUsed: mem.value.hostUsed,
      memRate: mem.value.hostUsed / mem.value.hostSum || 0, // 已用/总和
    }
      : {
        memSum: mem.value.federationSum,
        memUsed: mem.value.federationUsed,
        memRate: mem.value.federationUsed / mem.value.federationSum || 0, // 已用/总和
      }));
    const gpuRateData = computed(() => ({
      gpuSum: gpu.value.federationSum,
      gpuUsed: gpu.value.federationUsed,
      gpuRate: gpu.value.federationUsed / gpu.value.federationSum || 0, // 已用/总和
    }));
    function getChartData() {
      sourceData.value.forEach((item) => {
        if (item.current.quotaType === 'host') { // CA主机
          cpu.value.hostSum += item.cpuNum; // 总和
          cpu.value.hostUsed += item.cpuUsed; // 已用

          mem.value.hostSum += item.memNum; // 总和
          mem.value.hostUsed += item.memUsed; // 已用

          gpu.value.hostSum += item.gpuNum; // 总和
          gpu.value.hostUsed += item.gpuUsed; // 已用
        } else if (item.current.quotaType === 'federation') { // 弹性算力
          cpu.value.federationSum += item.cpuNum; // 总和
          cpu.value.federationUsed += item.cpuUsed; // 已用

          mem.value.federationSum += item.memNum; // 总和
          mem.value.federationUsed += item.memUsed; // 已用

          gpu.value.federationSum += item.gpuNum; // 总和
          gpu.value.federationUsed += item.gpuUsed; // 已用
        }
      });
      cpuSum.value = cpu.value.hostSum + cpu.value.federationSum; // 总和
      memSum.value = mem.value.hostSum + mem.value.federationSum; // 总和
      gpuSum.value = gpu.value.hostSum + gpu.value.federationSum; // 总和

      cpuOptions.value.series[0].data[0].value = cpu.value.hostUsed;
      cpuOptions.value.series[0].data[1].value = cpu.value.federationUsed;
      cpuOptions.value.series[0].data[2].value = cpu.value.hostSum - cpu.value.hostUsed;
      cpuOptions.value.series[0].data[3].value = cpu.value.federationSum - cpu.value.federationUsed;
      const cupUsed = cpu.value.hostUsed + cpu.value.federationUsed;
      cpuOptions.value.series[0].label.formatter = calculatePercentage(cupUsed, cpuSum.value);

      memOptions.value.series[0].data[0].value = mem.value.hostUsed;
      memOptions.value.series[0].data[1].value = mem.value.federationUsed;
      memOptions.value.series[0].data[2].value = mem.value.hostSum - mem.value.hostUsed;
      memOptions.value.series[0].data[3].value = mem.value.federationSum - mem.value.federationUsed;
      const memUsed = mem.value.hostUsed + mem.value.federationUsed;
      memOptions.value.series[0].label.formatter = calculatePercentage(memUsed, memSum.value);

      gpuOptions.value.series[0].data[0].value = gpu.value.hostUsed;
      gpuOptions.value.series[0].data[1].value = gpu.value.federationUsed;
      gpuOptions.value.series[0].data[2].value = gpu.value.hostSum - gpu.value.hostUsed;
      gpuOptions.value.series[0].data[3].value = gpu.value.federationSum - gpu.value.federationUsed;
      const gupUsed = gpu.value.hostUsed + gpu.value.federationUsed;
      gpuOptions.value.series[0].label.formatter = calculatePercentage(gupUsed, gpuSum.value);
    }
    const calculatePercentage = (num, den) => {
      if (den === 0) return '0.00%';
      return `${((num / den) * 100).toFixed(2)}%`;
    };

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

    watch(statisticsType, () => {
      // 重置页码
      pageConf.current = 1;
      // 刷新table setting columns数据
      fieldsDataClone.value = fields.value;
      // 刷新表格
      tableKey.value = new Date().getTime();
    });

    onBeforeMount(async () => {
      isLoading.value = true;
      // 获取集群列表
      await getClusterList();
      // 获取项目配额
      await handleGetProjectQuotas();
      isLoading.value = false;
    });

    return {
      isLoading,
      cpuSum,
      memSum,
      gpuSum,
      cpuOptions,
      memOptions,
      gpuOptions,
      echartsLoading,
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
      cpuRateData,
      memRateData,
      gpuRateData,
      cpu,
      mem,
      gpu,
      clusterMap,
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
