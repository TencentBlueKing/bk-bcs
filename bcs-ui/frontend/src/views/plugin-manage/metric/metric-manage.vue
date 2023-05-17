<template>
  <BcsContent :title="$t('Metric管理')" hide-back>
    <Row>
      <template #left>
        <bk-button
          icon="plus"
          theme="primary"
          @click="showSideslider = true">
          {{ $t('新建Metric') }}
        </bk-button>
        <bk-button :disabled="!selections.length" @click="handleBatchDelete">{{ $t('批量删除') }}</bk-button>
      </template>
      <template #right>
        <ClusterSelectComb
          :cluster-id.sync="clusterID"
          :placeholder="$t('输入名称搜索')"
          :search.sync="searchValue"
          cluster-type="all"
          @refresh="getTableData" />
      </template>
    </Row>
    <bk-table
      class="mt-[20px]"
      :pagination="pagination"
      :data="curPageData"
      v-bkloading="{ isLoading }"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange"
      @selection-change="handleSelectionChange">
      <bk-table-column type="selection" width="50"></bk-table-column>
      <bk-table-column prop="metadata.name" :label="$t('名称')"></bk-table-column>
      <bk-table-column prop="metadata.namespace" :label="$t('命名空间')"></bk-table-column>
      <bk-table-column label="Service">
        <template #default="{ row }">
          {{getServiceName(row) || '--'}}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('Metric路径')">
        <template #default="{ row }">
          <template v-if="row._endpoints.path.length">
            <div v-for="item, index in row._endpoints.path" :key="index" class="overflow-hidden">
              {{ item }}
            </div>
          </template>
          <span v-else>--</span>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('端口')">
        <template #default="{ row }">
          <template v-if="row._endpoints.port.length">
            <div v-for="item, index in row._endpoints.port" :key="index" class="overflow-hidden">
              {{ item }}
            </div>
          </template>
          <span v-else>--</span>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('周期(s)')" width="120">
        <template #default="{ row }">
          <template v-if="row._endpoints.interval.length">
            <div v-for="item, index in row._endpoints.interval" :key="index" class="overflow-hidden">
              {{ item }}
            </div>
          </template>
          <span v-else>--</span>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('操作')" width="120">
        <template #default="{ row }">
          <span
            v-bk-tooltips="{
              disabled: getServiceName(row),
              content: $t('Service对象不存在'),
              placement: 'left'
            }">
            <bk-button
              text
              :disabled="!getServiceName(row)"
              @click="handleUpdateMetric(row)">
              {{ $t('更新') }}
            </bk-button>
          </span>
          <bk-button text class="ml10" @click="handleDeleteMetric(row)">{{ $t('删除') }}</bk-button>
        </template>
      </bk-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bk-table>
    <!-- 更新和编辑 -->
    <bk-sideslider
      :is-show.sync="showSideslider"
      :title="curOperateRow.metadata
        ? $t('编辑 {name}', { name: curOperateRow.metadata && curOperateRow.metadata.name })
        : $t('新建Metric')"
      :width="800"
      :before-close="handleBeforeClose"
      quick-close
      @hidden="handleHidden">
      <template #content>
        <div class="px-[32px] pt-[40px] pb-[20px]">
          <EditMetric
            :data="curOperateRow"
            @init-data="reset"
            @change="setChanged(true)"
            @submit="handleSubmit"
            @cancel="showSideslider = false" />
        </div>
      </template>
    </bk-sideslider>
  </BcsContent>
</template>
<script lang="ts" setup>
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import ClusterSelectComb from '@/components/cluster-selector/cluster-select-comb.vue';
import EditMetric from './edit-metric.vue';
import useMetric from './use-metric';
import { ref, watch } from 'vue';
import useSideslider from '@/composables/use-sideslider';
import usePageConf from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $i18n from '@/i18n/i18n-setup';

const {
  handleGetServiceMonitor,
  handleDeleteServiceMonitor,
  handleBatchDeleteServiceMonitor,
} = useMetric();

const { handleBeforeClose, setChanged, reset } = useSideslider();

const isLoading = ref(false);
const metricList = ref([]);
const clusterID = ref('');
const showSideslider = ref(false);
const selections = ref<any[]>([]);

const keys = ref(['metadata.name']);
const { tableDataMatchSearch, searchValue } = useTableSearch(metricList, keys);
const { curPageData, pagination, pageConf, pageChange, pageSizeChange } = usePageConf(tableDataMatchSearch);

watch(searchValue, () => {
  pageConf.current = 1;
});

const getTableData = async () => {
  isLoading.value = true;
  const data = await handleGetServiceMonitor(clusterID.value);
  metricList.value = data.map((row) => {
    const _endpoints =  row?.spec?.endpoints?.reduce((pre, item) => {
      item.interval && pre.interval.push(item.interval);
      item.path && pre.path.push(item.path);
      item.port && pre.port.push(item.port);
      return pre;
    }, {
      interval: [],
      path: [],
      port: [],
    }) || {
      interval: [],
      path: [],
      port: [],
    };
    return {
      ...row,
      _endpoints,
    };
  });
  isLoading.value = false;
};

const getServiceName = row => row?.metadata?.labels?.['io.tencent.bcs.service_name'];

const curOperateRow = ref<Record<string, any>>({});
const handleHidden = () => {
  curOperateRow.value = {};
};

// 表格勾选
const handleSelectionChange = (data) => {
  selections.value = data;
};
// 编辑metric
const handleUpdateMetric = (row) => {
  curOperateRow.value = row;
  showSideslider.value = true;
};
// 批量删除
const handleBatchDelete = async () => {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t(
      '确定删除 {name} 等 {count} 个Metric',
      { name: selections.value[0]?.metadata?.name, count: selections.value.length },
    ),
    defaultInfo: true,
    confirmFn: async () => {
      const result = await handleBatchDeleteServiceMonitor({
        $clusterId: clusterID.value,
        service_monitors: selections.value.map(item => ({
          name: item.metadata?.name,
          namespace: item.metadata?.namespace,
        })),
      });
      if (result) {
        pageConf.current = 1;
        getTableData();
      }
    },
  });
};
// 删除metric
const handleDeleteMetric = (row) => {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('确定删除Metric {name}', { name: row.metadata?.name }),
    defaultInfo: true,
    confirmFn: async () => {
      const result = await handleDeleteServiceMonitor({
        $clusterId: clusterID.value,
        $namespaceId: row.metadata?.namespace,
        $name: row.metadata.name,
      });
      if (result) {
        pageConf.current = 1;
        getTableData();
      }
    },
  });
};
// 提交成功
const handleSubmit = () => {
  showSideslider.value = false;
  getTableData();
};

watch(clusterID, () => {
  getTableData();
});
</script>
