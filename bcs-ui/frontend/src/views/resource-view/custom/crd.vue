<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="flex flex-col h-full">
    <ContentHeader class="flex-[0_0_auto] !h-[66px] !border-b-0 !shadow-none !bg-inherit">
      <span class="text-[16px] text-[#313238] font-bold">CRD</span>
      <template #right>
        <bcs-search-select
          class="bg-[#fff] w-[460px]"
          clearable
          :show-condition="false"
          :show-popover-tag-change="false"
          :data="searchSelectData"
          :values="searchSelectValue"
          :placeholder="$t('view.placeholder.searchNameOrCreator')"
          :key="searchSelectKey"
          @change="searchSelectChange"
          @clear="searchSelectChange()"
          v-if="!isViewEditable" />
      </template>
    </ContentHeader>
    <div
      class="dashboard-content flex-1 px-[24px] py-[16px] pt-0 overflow-auto"
      v-bkloading="{ isLoading, opacity: 1, color: '#f5f7fa' }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange"
        @filter-change="handleFilterChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" sortable fixed="left">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('cluster.labels.nameAndId')" v-if="!isClusterMode">
          <template #default="{ row }">
            <div class="flex flex-col py-[6px] h-[50px]">
              <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
              <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column label="Scope" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'scope') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Kind" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'kind') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="ApiVersion" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'apiVersion') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.source')" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <div class="flex items-center">
              <bk-popover
                class="size-[16px] mr-[4px]"
                :content="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconText"
                :tippy-options="{ interactive: false }">
                <i
                  class="text-[14px] p-[1px]"
                  :class="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconClass"></i>
              </bk-popover>
              <span
                v-bk-overflow-tips="{ interactive: false }"
                class="bcs-ellipsis" v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
                {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
                  handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
              </span>
              <span
                v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
                v-else-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Helm'">
                {{ handleGetExtData(row.metadata.uid, 'chart')
                  ?`${handleGetExtData(row.metadata.uid, 'chart') || '--'}`
                  : 'Helm' }}
              </span>
              <span
                v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
                v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template #default="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus
            :button-text="$t('generic.button.resetSearch')"
            type="search-empty"
            @clear="handleShowViewConfig" />
        </template>
      </bk-table>
    </div>
    <bcs-sideslider
      quick-close
      :is-show.sync="showDetailPanel"
      :width="800"
      :title="detailTitle">
      <template #content>
        <div class="h-[calc(100vh-60px)] overflow-auto">
          <CodeEditor
            v-full-screen="{ tools: ['fullscreen', 'copy'], content: yaml }"
            :options="{
              roundedSelection: false,
              scrollBeyondLastLine: false,
              renderLineHighlight: 'none',
            }"
            width="100%"
            height="100%"
            lang="yaml"
            readonly
            :value="yaml" />
        </div>
      </template>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts" setup>
import yamljs from 'js-yaml';
import { isEqual } from 'lodash';
import { computed, onBeforeMount, onBeforeUnmount, onMounted, ref, watch } from 'vue';

import useSearch from '../common/use-search';
import { ISubscribeData } from '../common/use-subscribe';
import useTableData from '../common/use-table-data';
import useViewConfig from '../view-manage/use-view-config';

import { bus } from '@/common/bus';
import ContentHeader from '@/components/layout/Header.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import { useCluster } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import fullScreen from '@/directives/full-screen';
import $store from '@/store';

const vFullScreen = fullScreen;
const { clusterNameMap } = useCluster();
const { curViewData, isClusterMode, isViewEditable } = useViewConfig();

const pageConf = ref({
  current: 1,
  limit: $store.state.globalPageSize,
  showTotalCount: true,
  count: 0,
});
const handlePageChange = (page: number) => {
  pageConf.value.current = page;
  handleGetTableData();
};
const handlePageSizeChange = (size: number) => {
  pageConf.value.current = 1;
  pageConf.value.limit = size;
  $store.commit('updatePageSize', size);
  handleGetTableData();
};
// 排序
const sortData = ref<{
  sortBy: string
  order: 'desc' | 'asc' | ''
}>({
  sortBy: '',
  order: '',
});
const propMap = {
  'metadata.name': 'name',
  'metadata.namespace': 'namespace',
  createTime: 'age',
};
const handleSortChange = ({ prop, order }) => {
  sortData.value = {
    sortBy: propMap[prop] || prop,
    order: order === 'ascending' ? 'asc' : 'desc',
  };
  handleGetTableData();
};
// 表头过滤
const filters = ref<Record<string, string[]>>({});
const handleFilterChange = (data) => {
  filters.value = data;
  handleGetTableData();
};

// 表格数据
const data = ref<ISubscribeData>({
  manifestExt: {},
  manifest: {},
  total: 0,
});
const {
  isLoading,
  getMultiClusterCustomResourceDefinition,
} = useTableData();
const curPageData = computed(() => data.value.manifest?.items || []);
// 获取表格数据
const handleGetTableData = async (loading = true) => {
  if (!curViewData.value) return;

  isLoading.value = loading;
  data.value = await getMultiClusterCustomResourceDefinition({
    ...curViewData.value,
    ...sortData.value,
    $crd: 'CustomResourceDefinition',
    offset: (pageConf.value.current - 1) * pageConf.value.limit,
    limit: pageConf.value.limit,
  });
  pageConf.value.count = data.value.total;
  isLoading.value = false;
};
// 重新搜索
watch(curViewData, (newValue, oldValue) => {
  if (!curViewData.value || isEqual(newValue, oldValue)) return;
  pageConf.value.current = 1;
  handleGetTableData();
}, { deep: true });

// 获取额外字段方法
const handleGetExtData = (uid: string, ext?: string, defaultData?: any) => {
  const extData = data.value.manifestExt?.[uid] || {};
  return ext ? (extData[ext] || defaultData) : extData;
};

// 来源类型
const sourceTypeMap = ref({
  Template: {
    iconClass: 'bcs-icon bcs-icon-templete',
    iconText: 'Template',
  },
  Helm: {
    iconClass: 'bcs-icon bcs-icon-helm',
    iconText: 'Helm',
  },
  Client: {
    iconClass: 'bcs-icon bcs-icon-client',
    iconText: 'Client',
  },
  Web: {
    iconClass: 'bcs-icon bcs-icon-web',
    iconText: 'Web',
  },
});

// 详情侧栏
const showDetailPanel = ref(false);
// 当前详情行数据
const curDetailRow = ref<{
  data: any
  extData: any
}>({
  data: {},
  extData: {},
});
const detailTitle = computed(() =>
  // const clusterID = curDetailRow.value.extData?.clusterID;
  curDetailRow.value.data?.metadata?.name);
// 显示侧栏详情
const handleShowDetail = (row) => {
  curDetailRow.value.data = row;
  curDetailRow.value.extData = handleGetExtData(row.metadata.uid);
  showDetailPanel.value = true;
};

// yaml内容
const yaml = computed(() => {
  // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
  const newDetailRow = {
    apiVersion: curDetailRow.value.data.apiVersion,
    kind: curDetailRow.value.data.kind,
    metadata: curDetailRow.value.data.metadata,
    ...curDetailRow.value.data,
  };
  return yamljs.dump(newDetailRow || {});
});

const handleShowViewConfig = () => {
  bus.$emit('toggle-show-view-config');
};

// 搜索
const {
  searchSelectData,
  searchSelectChange,
  searchSelectValue,
  searchSelectKey,
} = useSearch();

const { start, stop } = useInterval(() => handleGetTableData(false), 5000);
onBeforeMount(() => {
  // 轮询资源
  start();
});

onMounted(async () => {
  isLoading.value = true;
  await handleGetTableData();
  isLoading.value = false;
});

onBeforeUnmount(() => {
  stop();
});
</script>
