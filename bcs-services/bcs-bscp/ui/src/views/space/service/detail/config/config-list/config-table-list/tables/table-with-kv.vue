<template>
  <bk-loading :loading="loading">
    <bk-table
      class="config-table"
      :border="['outer']"
      :data="configList"
      :remote-pagination="true"
      :pagination="pagination"
      :key="versionData.id"
      :checked="checkedConfigs"
      selection-key="id"
      row-key="id"
      :row-class="getRowCls"
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageLimitChange"
      @page-value-change="refresh"
      @column-sort="handleSort"
      @column-filter="handleFilter"
      @selection-change="handleSelectionChange"
      @select-all="handleSelectAll">
      <bk-table-column v-if="versionData.id === 0" type="selection" :width="40" :min-width="40"></bk-table-column>
      <bk-table-column :label="t('配置项名称')" prop="spec.key" :min-width="240">
        <template #default="{ row }">
          <bk-button
            v-if="row.spec"
            text
            theme="primary"
            :disabled="row.kv_state === 'DELETE'"
            @click="handleView(row)">
            {{ row.spec.key }}
          </bk-button>
        </template>
      </bk-table-column>
      <bk-table-column :label="t('配置项值预览')" prop="spec.value">
        <template #default="{ row }">
          <kvValuePreview v-if="row.spec" :key="row.id" :value="row.spec.value" @view-all="handleView(row)" />
        </template>
      </bk-table-column>
      <bk-table-column :label="t('配置项描述')">
        <template #default="{ row }">
          <span v-if="row.spec">{{ row.spec.memo || '--' }}</span>
        </template>
      </bk-table-column>
      <bk-table-column
        prop="spec.kv_type"
        :label="t('数据类型')"
        :filter="{ filterFn: () => true, list: typeFilterList, checked: typeFilterChecked }"
        :width="120">
      </bk-table-column>
      <bk-table-column :label="t('创建人')" prop="revision.creator" :width="150"></bk-table-column>
      <bk-table-column :label="t('修改人')" prop="revision.reviser" :width="150"></bk-table-column>
      <bk-table-column :label="t('修改时间')" :sort="true" :width="180">
        <template #default="{ row }">
          <span v-if="row.revision">{{ datetimeFormat(row.revision.update_at) }}</span>
        </template>
      </bk-table-column>
      <bk-table-column
        v-if="versionData.id === 0"
        :label="t('变更状态')"
        :filter="{ filterFn: () => true, list: statusFilterList, checked: statusFilterChecked }"
        :width="140">
        <template #default="{ row }">
          <StatusTag :status="row.kv_state" />
        </template>
      </bk-table-column>
      <bk-table-column :label="t('操作')" fixed="right" :width="220">
        <template #default="{ row }">
          <div class="operate-action-btns">
            <bk-button
              v-if="row.kv_state === 'DELETE'"
              v-cursor="{ active: !hasEditServicePerm }"
              :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
              :disabled="!hasEditServicePerm"
              text
              theme="primary"
              @click="handleUndelete(row)">
              {{ t('恢复') }}
            </bk-button>
            <template v-else>
              <bk-button
                v-cursor="{ active: !hasEditServicePerm }"
                :class="{ 'bk-text-with-no-perm': versionData.id === 0 && !hasEditServicePerm }"
                :disabled="versionData.id === 0 && !hasEditServicePerm"
                text
                theme="primary"
                @click="handleEditOrView(row)">
                {{ versionData.id === 0 ? t('编辑') : t('查看') }}
              </bk-button>
              <bk-button
                v-cursor="{ active: !hasEditServicePerm }"
                v-if="row.kv_state === 'REVISE'"
                :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                :disabled="!hasEditServicePerm"
                text
                theme="primary"
                @click="handleUnModify(row)">
                {{ t('撤销') }}
              </bk-button>
              <bk-button
                v-if="versionData.status.publish_status !== 'editing'"
                text
                theme="primary"
                @click="handleDiff(row)">
                {{ t('对比') }}
              </bk-button>
              <bk-button
                v-cursor="{ active: !hasEditServicePerm }"
                v-if="versionData.id === 0"
                :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                :disabled="!hasEditServicePerm"
                text
                theme="primary"
                @click="handleDel(row)">
                {{ t('删除') }}
              </bk-button>
            </template>
          </div>
        </template>
      </bk-table-column>
      <template #empty>
        <TableEmpty :is-search-empty="isSearchEmpty" @clear="emits('clearStr')" style="width: 100%" />
      </template>
    </bk-table>
  </bk-loading>
  <edit-config
    v-model:show="editPanelShow"
    :config="activeConfig.spec as IConfigKvItem"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :editable="true"
    @confirm="getListData" />
  <ViewConfigKv v-model:show="viewPanelShow" :config="activeConfig" />
  <VersionDiff v-model:show="isDiffPanelShow" :current-version="versionData" :selected-kv-config-id="diffConfig" />
  <DeleteConfirmDialog
    v-model:isShow="isDeleteConfigDialogShow"
    :title="t('确认删除该配置项？')"
    @confirm="handleDeleteConfigConfirm">
    <div style="margin-bottom: 8px">
      {{ t('配置项') }}：<span style="color: #313238">{{ deleteConfig?.spec.key }}</span>
    </div>
    <div>{{ deleteConfigTips }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref, watch, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useConfigStore from '../../../../../../../../store/config';
  import useServiceStore from '../../../../../../../../store/service';
  import { ICommonQuery } from '../../../../../../../../../types/index';
  import { IConfigKvItem, IConfigKvType } from '../../../../../../../../../types/config';
  import { getKvList, deleteKv, getReleaseKvList, undeleteKv, unModifyKv } from '../../../../../../../../api/config';
  import { datetimeFormat } from '../../../../../../../../utils/index';
  import { getDefaultKvItem } from '../../../../../../../../utils/config';
  import { CONFIG_KV_TYPE } from '../../../../../../../../constants/config';
  import StatusTag from './status-tag';
  import EditConfig from '../edit-config-kv.vue';
  import ViewConfigKv from '../view-config-kv.vue';
  import kvValuePreview from './kv-value-preview.vue';
  import VersionDiff from '../../../components/version-diff/index.vue';
  import TableEmpty from '../../../../../../../../components/table/table-empty.vue';
  import DeleteConfirmDialog from '../../../../../../../../components/delete-confirm-dialog.vue';

  const configStore = useConfigStore();
  const serviceStore = useServiceStore();
  const { versionData } = storeToRefs(configStore);
  const { checkPermBeforeOperate } = serviceStore;
  const { permCheckLoading, hasEditServicePerm } = storeToRefs(serviceStore);
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    searchStr: string;
  }>();

  const emits = defineEmits(['clearStr', 'updateSelectedIds']);

  const loading = ref(false);
  const configList = ref<IConfigKvType[]>([]);
  const configsCount = ref(0);
  const editPanelShow = ref(false);
  const viewPanelShow = ref(false);
  const activeConfig = ref<IConfigKvType>(getDefaultKvItem());
  const deleteConfig = ref<IConfigKvType>();
  const selectedConfigIds = ref<number[]>([]);
  const isDiffPanelShow = ref(false);
  const diffConfig = ref(0);
  const isSearchEmpty = ref(false);
  const isDeleteConfigDialogShow = ref(false);
  const typeFilterChecked = ref<string[]>([]);
  const statusFilterChecked = ref<string[]>([]);
  const updateSortType = ref('null');
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  });
  const typeFilterList = computed(() =>
    CONFIG_KV_TYPE.map((item) => ({
      value: item.id,
      text: item.name,
    })),
  );
  const statusFilterList = computed(() => {
    return [
      {
        value: 'ADD',
        text: t('新增'),
      },
      {
        value: 'REVISE',
        text: t('修改'),
      },
      {
        value: 'DELETE',
        text: t('删除'),
      },
      {
        value: 'UNCHANGE',
        text: t('无修改'),
      },
    ];
  });

  const deleteConfigTips = computed(() => {
    if (deleteConfig.value) {
      return deleteConfig.value.kv_state === 'ADD'
        ? t('一旦删除，该操作将无法撤销，请谨慎操作')
        : t('配置项删除后，可以通过恢复按钮撤销删除');
    }
    return '';
  });

  const checkedConfigs = computed(() => {
    return configList.value.filter((config) => selectedConfigIds.value.includes(config.id));
  });

  watch(
    () => versionData.value.id,
    () => {
      refresh();
      selectedConfigIds.value = [];
      emits('updateSelectedIds', []);
    },
  );

  watch(
    () => props.searchStr,
    () => {
      props.searchStr ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
      refresh();
    },
  );

  watch(
    () => configsCount.value,
    () => {
      configStore.$patch((state) => {
        state.allConfigCount = configsCount.value;
      });
    },
  );

  const isUnNamedVersion = computed(() => versionData.value.id === 0);

  onMounted(() => {
    getListData();
  });

  const getListData = async () => {
    loading.value = true;
    try {
      const params: ICommonQuery = {
        start: (pagination.value.current - 1) * pagination.value.limit,
        limit: pagination.value.limit,
        with_status: true,
      };
      if (props.searchStr) {
        params.search_fields = 'key,revister,creator';
        params.search_key = props.searchStr;
      }
      if (typeFilterChecked.value!.length > 0) {
        params.kv_type = typeFilterChecked.value;
      }

      if (updateSortType.value !== 'null') {
        params.sort = 'updated_at';
        params.order = updateSortType.value.toUpperCase();
      }
      let res;
      if (isUnNamedVersion.value) {
        if (statusFilterChecked.value!.length > 0) {
          params.status = statusFilterChecked.value;
        }
        res = await getKvList(props.bkBizId, props.appId, params);
      } else {
        res = await getReleaseKvList(props.bkBizId, props.appId, versionData.value.id, params);
      }
      configList.value = res.details;
      configsCount.value = res.count;
      pagination.value.count = res.count;
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  // 表格行是否可以选中
  const isRowSelectEnable = ({ row, isCheckAll }: { row: IConfigKvType; isCheckAll: boolean }) => {
    return isCheckAll || row.kv_state !== 'DELETE';
  };

  // 表格行选择事件
  const handleSelectionChange = ({ checked, row }: { checked: boolean; row: IConfigKvType }) => {
    const index = selectedConfigIds.value.findIndex((id) => id === row.id);
    if (checked) {
      if (index === -1) {
        selectedConfigIds.value.push(row.id);
      }
    } else {
      selectedConfigIds.value.splice(index, 1);
    }
    emits('updateSelectedIds', selectedConfigIds.value);
  };

  // 全选
  const handleSelectAll = ({ checked }: { checked: boolean }) => {
    if (checked) {
      selectedConfigIds.value = configList.value.filter((item) => item.kv_state !== 'DELETE').map((item) => item.id);
    } else {
      selectedConfigIds.value = [];
    }
    emits('updateSelectedIds', selectedConfigIds.value);
  };

  const handleEditOrView = (config: IConfigKvType) => {
    activeConfig.value = config;
    if (isUnNamedVersion.value) {
      if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
        return;
      }
      editPanelShow.value = true;
    } else {
      viewPanelShow.value = true;
    }
  };

  const handleView = (config: IConfigKvType) => {
    activeConfig.value = config;
    viewPanelShow.value = true;
  };

  const handleDiff = (config: IConfigKvType) => {
    diffConfig.value = config.id;
    isDiffPanelShow.value = true;
  };

  const handleDel = (config: IConfigKvType) => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    isDeleteConfigDialogShow.value = true;
    deleteConfig.value = config;
  };

  const handleUnModify = async (config: IConfigKvType) => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    await unModifyKv(props.bkBizId, props.appId, config.spec.key);
    Message({ theme: 'success', message: t('撤销修改配置项成功') });
    refresh();
  };

  // 删除单个配置项
  const handleDeleteConfigConfirm = async () => {
    if (!deleteConfig.value) {
      return;
    }
    await deleteKv(props.bkBizId, props.appId, deleteConfig.value.id);

    // 删除的配置项如果在多选列表里，需要去掉
    const index = selectedConfigIds.value.findIndex((id) => id === deleteConfig.value?.id);
    if (index > -1) {
      selectedConfigIds.value.splice(index, 1);
    }

    // 新增的配置项被删除后，检查是否需要往前翻一页
    if (deleteConfig.value.kv_state === 'ADD') {
      if (configList.value.length === 1 && pagination.value.current > 1) {
        pagination.value.current -= 1;
      }
    }

    Message({
      theme: 'success',
      message: t('删除配置项成功'),
    });
    refresh(pagination.value.current);
    isDeleteConfigDialogShow.value = false;
  };

  // 撤销删除单个配置项
  const handleUndelete = async (config: IConfigKvType) => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    await undeleteKv(props.bkBizId, props.appId, config.spec.key);
    Message({ theme: 'success', message: t('恢复配置项成功') });
    refresh();
  };

  // 批量删除配置项后刷新配置项列表
  const refreshAfterBatchDelete = () => {
    if (selectedConfigIds.value.length === configList.value.length && pagination.value.current > 1) {
      pagination.value.current -= 1;
    }

    selectedConfigIds.value = [];
    emits('updateSelectedIds', []);
    refresh(pagination.value.current);
  };

  // page-limit
  const handlePageLimitChange = (limit: number) => {
    pagination.value.limit = limit;
    refresh();
  };

  const refresh = (current = 1) => {
    pagination.value.current = current;
    getListData();
  };

  const handleFilter = ({ checked, index }: any) => {
    if (index === 2) {
      // 调整数据类型筛选条件
      typeFilterChecked.value = checked;
    } else {
      // 调整状态筛选条件
      statusFilterChecked.value = checked;
    }
    refresh();
  };

  const handleSort = ({ type }: any) => {
    updateSortType.value = type;
    refresh();
  };

  // 判断当前行是否是删除行
  const getRowCls = (config: IConfigKvType) => {
    if (config.kv_state === 'DELETE') return 'delete-row';
  };

  defineExpose({
    refresh,
    refreshAfterBatchDelete,
  });
</script>
<style lang="scss" scoped>
  .operate-action-btns {
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
  .config-table {
    :deep(.bk-table-body) {
      tr.delete-row td {
        background: #fafbfd !important;
        .cell {
          color: #c4c6cc !important;
        }
      }
    }
  }
</style>
