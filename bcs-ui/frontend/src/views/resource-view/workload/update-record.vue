<template>
  <div>
    <Header :title="$t('deploy.helm.history')" :desc="name" :cluster-id="clusterId" :namespace="namespace" />
    <div class="px-[24px] pt-[8px]">
      <Row>
        <template #right>
          <bcs-input
            class="w-[480px]"
            right-icon="bk-icon icon-search"
            :placeholder="$t('updateRecord.placeholder.search')"
            clearable
            v-model="searchValue">
          </bcs-input>
        </template>
      </Row>
      <bcs-table
        class="mt-[16px]"
        :data="curPageData"
        :pagination="pagination"
        v-bkloading="{ isLoading }"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column :label="$t('updateRecord.label.revision')" width="120">
          <template #default="{ row }">
            <bcs-button text @click="showDetail(row)">{{ row.revision }}</bcs-button>
            <bcs-tag
              theme="success"
              v-if="onlineVersion === row.revision">
              {{ $t('updateRecord.label.onlineVersion') }}
            </bcs-tag>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('updateRecord.label.images')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.images.join(',') }}</span>
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('updateRecord.label.updater')"
          width="180"
          prop="updater"
          show-overflow-tooltip>
        </bcs-table-column>
        <bcs-table-column :label="$t('updateRecord.label.createTime')" prop="createTime" width="180"></bcs-table-column>
        <bcs-table-column :label="$t('updateRecord.label.operator')" width="120">
          <template #default="{ row }">
            <bcs-button text @click="showDiff(row)">{{ $t('updateRecord.button.diff') }}</bcs-button>
            <bcs-button
              text
              class="ml-[10px]"
              v-if="onlineVersion !== row.revision"
              @click="handleRollback(row)"
            >{{ $t('updateRecord.button.rollout') }}</bcs-button>
          </template>
        </bcs-table-column>
      </bcs-table>
    </div>
    <!-- 详情 -->
    <bcs-sideslider :is-show.sync="showDetailSideslider" :width="1000" quick-close>
      <template #header>
        <div class="flex items-center">
          <span>{{ $t('updateRecord.label.versionDetail') }}</span>
          <bcs-divider direction="vertical"></bcs-divider>
          <span class="text-[12px]">{{ curRow.revision }}</span>
        </div>
      </template>
      <template #content>
        <div class="h-[calc(100vh-108px)] p-[24px]" v-bkloading="{ isLoading: detailLoading }">
          <CodeEditor
            readonly
            :value="detail"
            full-screen />
        </div>
      </template>
      <template #footer>
        <div class="h-[48px] bg-[#FAFBFD] flex items-center px-[24px] w-full footer">
          <bcs-button
            theme="primary"
            @click="showDiffSideslider = true">
            {{ $t('updateRecord.button.diff2') }}
          </bcs-button>
          <bcs-button
            v-if="onlineVersion !== curRow.revision"
            @click="handleRollback(curRow)">{{ $t('updateRecord.button.rollout') }}</bcs-button>
          <bcs-button @click="showDetailSideslider = false">{{ $t('updateRecord.button.close') }}</bcs-button>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 对比/回滚 -->
    <Rollback
      :name="name"
      :namespace="namespace"
      :category="category"
      :cluster-id="clusterId"
      :revision="curRow.revision"
      :value="showDiffSideslider"
      :rollback="rollback"
      :crd="crd"
      @hidden="handleRollbackSidesilderHide"
      @rollback-success="handleRollbackSuccess" />
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';

import Rollback from './rollback.vue';
import useRecords, { IRevisionData } from './use-records';

import Header from '@/components/layout/Header.vue';
import Row from '@/components/layout/Row.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import usePage from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';

const props = defineProps({
  namespace: {
    type: String,
    default: '',
    required: true,
  },
  crd: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
    required: true,
  },
  category: {
    type: String,
    default: '',
    required: true,
  },
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
});

const { workloadHistory, gameWorkloadHistory, revisionDetail, revisionGameDetail } = useRecords();

const tableData = ref<IRevisionData[]>([]);

const curRow = ref<Partial<IRevisionData>>({});

// 列表历史数据
const isLoading = ref(false);
const keys = ref(['images', 'updater']); // 模糊搜索字段
const { tableDataMatchSearch, searchValue } = useTableSearch(tableData, keys);
const { pageChange, pageSizeChange, curPageData, pagination } = usePage(tableDataMatchSearch);
const onlineVersion = ref('');
const handleGetHistory = async () => {
  isLoading.value = true;
  if (props.category === 'custom_objects') {
    tableData.value = await gameWorkloadHistory({
      $crd: props.crd,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      namespace: props.namespace,
    });
  } else {
    tableData.value = await workloadHistory({
      $namespaceId: props.namespace,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
    });
  }

  onlineVersion.value = tableData.value[0]?.revision;
  isLoading.value = false;
};

// 详情
const detailLoading = ref(false);
const detail = ref('');
const showDetailSideslider = ref(false);
const showDetail = async (row: IRevisionData) => {
  curRow.value = row;
  showDetailSideslider.value = true;
  detailLoading.value = true;
  let data = { rollout_revision: '' };
  if (props.category === 'custom_objects') {
    data = await revisionGameDetail({
      $crd: props.crd,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      $revision: row.revision,
      namespace: props.namespace,
    });
  } else {
    data = await revisionDetail({
      $namespaceId: props.namespace,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      $revision: row.revision,
    });
  }

  detail.value = data.rollout_revision;
  detailLoading.value = false;
};

// 对比
const showDiffSideslider = ref(false);
const showDiff = (row) => {
  curRow.value = row;
  showDiffSideslider.value = true;
};

// 隐藏回滚
const handleRollbackSidesilderHide = () => {
  showDiffSideslider.value = false;
  rollback.value = false;
};

// 回滚成功
const handleRollbackSuccess = () => {
  showDiffSideslider.value = false;
  showDetailSideslider.value = false;
  handleGetHistory();
};

// 回滚
const rollback = ref(false);
const handleRollback = (row = curRow.value) => {
  showDiff(row);
  rollback.value = true;
};

onBeforeMount(() => {
  handleGetHistory();
});
</script>
<style lang="postcss" scoped>
>>> .footer {
  border-top: 1px solid #DCDEE5;
}
</style>
