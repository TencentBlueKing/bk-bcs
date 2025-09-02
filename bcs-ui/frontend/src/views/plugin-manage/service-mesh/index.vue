<template>
  <BcsContent hide-back :title="$t('serviceMesh.title')" :padding="24">
    <Row>
      <template #left>
        <bcs-button
          theme="primary"
          icon="plus"
          @click="handleCreate">{{ $t('serviceMesh.create') }}</bcs-button>
      </template>
      <template #right>
        <bcs-search-select
          clearable
          class="min-w-[360px] bg-[#fff]"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('serviceMesh.placeholder.search')"
          ref="searchSelect"
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClear">
        </bcs-search-select>
      </template>
    </Row>
    <div class="mt-[16px]" v-bkloading="{ isLoading: loading }">
      <bcs-table
        :data="meshData"
        :pagination="pagination"
        row-auto-height
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column :label="$t('serviceMesh.label.name')" prop="name" sortable show-overflow-tooltip>
          <template #default="{ row }">
            <span
              v-authority="{
                clickable: web_annotations.perms[row.meshID]?.['MeshManager.GetIstioDetail'],
                disablePerms: true,
                originClick: true,
              }"
              class="cursor-pointer text-[#3A84FF]"
              @click="handleDetail(row)">{{ row.name || '--' }}</span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.version')" prop="version">
          <template #default="{ row }">
            {{ row.version || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.status')" show-overflow-tooltip>
          <template #default="{ row }">
            <StatusIcon
              :status="row.status"
              :message="row.statusMessage"
              :status-color-map="statusColorMap"
              :status-text-map="statusTextMap"
              :pending="loadingStatusList.includes(row.status)" />
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.cluster')" show-overflow-tooltip>
          <template #default="{ row }">
            <div
              class="flex flex-col"
              v-bk-tooltips="{
                content: `${(row.primaryClusters || []).join()},
                  ${(row.remoteClusters || []).map(v => v.clusterID).join()}`,
                disabled: row.remoteClusters?.length < 3,
              }">
              <span
                v-for="(item, index) in (row.primaryClusters || [])"
                :key="`${index}-${item}`">
                {{ item }}
              </span>
              <span
                v-for="item in (row.remoteClusters || []).slice(0, 2)"
                :key="item.clusterID">
                {{ item.clusterID || '--' }}
                <span v-if="row.remoteClusters?.length > 2">...</span>
              </span>
            </div>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.createTime')" prop="createTime" sortable show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatDate(Number(row.createTime || 0)) || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column fixed="right" :label="$t('generic.label.action')" width="280">
          <template #default="{ row }">
            <bcs-button
              text
              @click="handleLinkToMonitor(row)"
              v-if="row.monitoringLink">{{ $t('serviceMesh.label.scene') }}</bcs-button>
            <bcs-button
              v-authority="{
                clickable: web_annotations.perms[row.meshID]?.['MeshManager.DeleteIstio'],
                disablePerms: true,
                originClick: true,
              }"
              :class="[{ 'ml-[10px]': row.monitoringLink }]"
              text
              @click="deleteMesh(row)">
              {{ $t('generic.label.delete') }}</bcs-button>
          </template>
        </bcs-table-column>
        <template #empty>
          <BcsEmptyTableStatus
            :type="searchSelectValue?.length ? 'search-empty' : 'empty'"
            @clear="handleClear" />
        </template>
      </bcs-table>
    </div>
    <!-- 新增网格 -->
    <bcs-sideslider
      :is-show.sync="isShowCreateMesh"
      :title="$t('serviceMesh.label.create')"
      :width="680"
      :before-close="handleCreateBeforeClose"
      :quick-close="false">
      <template #content>
        <CreateMesh @cancel="handleCancel" />
      </template>
    </bcs-sideslider>
    <!-- 详情/编辑 -->
    <bcs-sideslider
      :is-show.sync="editSettings.isShow"
      :title="editSettings.title"
      :width="680"
      :before-close="handleEditBeforeClose"
      :quick-close="isQuickClose">
      <template #content>
        <DetailEdit
          ref="detailEditRef"
          :mesh-id="editSettings.meshID"
          :project-code="editSettings.projectCode"
          @cancel="handleEditCancel"
          @success="handleUpdateSuccess" />
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';

import CreateMesh from './create-mesh.vue';
import DetailEdit from './detail-edit.vue';
import useMesh from './use-mesh';

import $bkMessage from '@/common/bkmagic';
import { formatDate } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import StatusIcon from '@/components/status-icon';
import useInterval from '@/composables/use-interval';
import usePageConf from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';

// 搜索
const searchSelectData = ref<any[]>([]);
const searchSelectValue = ref<any[]>([]);
const searchParams = ref({});
// searchSelect数据源配置
const searchSelectDataSource = [
  {
    name: $i18n.t('serviceMesh.label.name'),
    id: 'name',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
  {
    name: $i18n.t('serviceMesh.label.version'),
    id: 'version',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
  {
    name: $i18n.t('serviceMesh.label.status'),
    id: 'status',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
  {
    name: $i18n.t('serviceMesh.label.cluster'),
    id: 'clusterID',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
];
// 网格
const {
  meshData,
  web_annotations,
  fetchMeshData,
  handleDelete,
  handleGetMeshDetail,
} = useMesh();
// 分页
const {
  pagination,
  pageChange,
  pageSizeChange,
  handleResetPage,
} = usePageConf(meshData, {
  current: 1,
  limit: 10,
  onPageChange: handleGetData,
  onPageSizeChange: handleGetData,
});
// 搜索项改变
function searchSelectChange(inputList) {
  const params = {};
  inputList.forEach((item) => {
    params[item.id] = item.values && item.values.length > 0 ? item?.values[0]?.id ?? '' : '';
  });
  searchParams.value = params;
}
// 清空搜索
function handleClear() {
  searchSelectValue.value = [];
  searchParams.value = {};
}

const loadingStatusList = ref(['installing', 'updating', 'uninstalling']);

// 点击新增
const isShowCreateMesh = ref<boolean>(false);
function handleCreate() {
  isShowCreateMesh.value = true;
};

// 获取数据
const loading = ref(false);
async function handleGetData() {
  await fetchMeshData({
    page: pagination.value.current,
    pageSize: pagination.value.limit,
    ...searchParams.value,
  });
}
const { start, stop } = useInterval(() => handleGetData(), 5000);

const detailEditRef = ref<InstanceType<typeof DetailEdit> | null>(null);
const isQuickClose = computed(() => !detailEditRef.value?.isEdit);

// 状态信息
const statusTextMap = ref({
  running: $i18n.t('generic.status.running'),
  failed: $i18n.t('generic.status.installFailed'),
  installing: $i18n.t('generic.status.installing'),
  updating: $i18n.t('generic.status.updating'),
  'update-failed': $i18n.t('generic.status.updateFailed'),
  uninstalling: $i18n.t('generic.status.uninstalling'),
  'uninstall-failed': $i18n.t('generic.status.uninstallFailed'),
});
const statusColorMap = ref({
  running: 'green',
  failed: 'red',
  'uninstalling-failed': 'red',
  'update-failed': 'red',
});

// 网格详情
const editSettings = ref({
  isShow: false,
  title: '',
  meshID: '',
  projectCode: '',
});
async function handleDetail(row) {
  if (!web_annotations.value.perms[row.meshID]?.['MeshManager.GetIstioDetail']) {
    // 没有权限时接口会触发权限弹窗，类似用户点击申请权限
    handleGetMeshDetail({
      meshID: row.meshID,
      projectCode: row.projectCode,
    });
    return;
  };
  editSettings.value.isShow = true;
  editSettings.value.title = row.name;
  editSettings.value.meshID = row.meshID;
  editSettings.value.projectCode = row.projectCode;
}

// 点击删除
function deleteMesh(row) {
  if (!web_annotations.value.perms[row.meshID]?.['MeshManager.DeleteIstio']) {
    // 没有权限时接口会触发权限弹窗，类似用户点击申请权限
    handleDel(row);
    return;
  };
  $bkInfo({
    clsName: 'custom-info-confirm',
    theme: 'danger',
    title: $i18n.t('serviceMesh.tips.delete'),
    subTitle: $i18n.t('serviceMesh.tips.subTitle', { name: row.name }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      await handleDel(row);
      handleGetData();
    },
  });
}
// 跳转到监控
function handleLinkToMonitor(row) {
  window.open(row.monitoringLink, '_blank');
}
// 删除网格
async function handleDel(row) {
  const result = await handleDelete({
    meshID: row.meshID,
    projectCode: row.projectCode,
  });
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.ok'),
    });
    handleGetData();
  }
}
// 关闭新增网格抽屉
async function handleCancel() {
  isShowCreateMesh.value = false;
  await handleGetData();
}
function handleCreateBeforeClose() {
  $bkInfo({
    title: $i18n.t('generic.msg.info.exitTips.text'),
    subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
    defaultInfo: true,
    okText: $i18n.t('generic.button.exit'),
    confirmFn() {
      isShowCreateMesh.value = false;
    },
  });
}

// 编辑网格
function handleEditCancel() {
  editSettings.value.isShow = false;
  handleGetData();
}
function handleUpdateSuccess(data) {
  handleGetData();
  editSettings.value.title = data.name;
}
// 关闭前的钩子
function handleEditBeforeClose() {
  if (!detailEditRef.value?.isChange || !detailEditRef.value?.isEdit) return true;
  $bkInfo({
    title: $i18n.t('generic.msg.info.exitTips.text'),
    subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
    defaultInfo: true,
    okText: $i18n.t('generic.button.exit'),
    confirmFn() {
      editSettings.value.isShow = false;
    },
  });
}

// 搜索项有值后就不展示了
watch(searchSelectValue, () => {
  const ids = searchSelectValue.value.map(item => item.id);
  searchSelectData.value = searchSelectDataSource.filter(item => !ids.includes(item.id));
}, { immediate: true, deep: true });

watch(searchParams, () => {
  handleResetPage();
  handleGetData();
});

watch(meshData, () => {
  const result = meshData.value.some(item => loadingStatusList.value.includes(item?.status || ''));
  if (result) {
    start();
  } else {
    stop();
  }
}, { deep: true });

onMounted(async () => {
  loading.value = true;
  await handleGetData();
  loading.value = false;
  start();
});
onBeforeUnmount(() => {
  stop();
});
defineExpose({
  detailEditRef,
  isShowCreateMesh,
});

</script>
<style lang="postcss" scoped>
:deep(.bk-sideslider-wrapper) {
  overflow: hidden;
}
:deep(.bk-sideslider-content) {
  height: calc(100vh - 52px);
}
:deep(.bk-form .bk-label .bk-label-text) {
  font-size: 12px;
}
:deep(.bk-tab-header) {
  position: sticky;
  top: 0;
  z-index: 10;
}
:deep(.bk-tab-section) {
  height: calc(100vh - 102px);
  overflow: auto;
  padding-bottom: 70px;
}
:deep(.bk-tab-label-wrapper) {
  padding: 8px 16px 0;

  .bk-tab-label-list {
    height: 42px !important;
    line-height: 42px;

    .bk-tab-label-item {
      height: 42px !important;
      line-height: 42px !important;
    }
  }
}
</style>
