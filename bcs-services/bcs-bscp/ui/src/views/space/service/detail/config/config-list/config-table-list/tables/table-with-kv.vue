<template>
  <bk-loading :loading="loading">
    <bk-table
      class="config-table"
      :border="['outer']"
      :data="configList"
      :remote-pagination="true"
      :pagination="pagination"
      :key="versionData.id"
      :row-class="getRowCls"
      @page-limit-change="handlePageLimitChange"
      @page-value-change="refresh"
      @column-sort="handleSort"
      @column-filter="handleFilter"
    >
      <bk-table-column label="配置项名称" prop="spec.key" :min-width="240">
        <template #default="{ row }">
          <bk-button
            v-if="row.spec"
            text
            theme="primary"
            :disabled="row.kv_state === 'DELETE'"
            @click="handleView(row)"
          >
            {{ row.spec.key }}
          </bk-button>
        </template>
      </bk-table-column>
      <bk-table-column label="配置项值预览" prop="spec.value">
        <template #default="{ row }">
          <div v-if="row.spec" class="preview">
            <div class="preview-value">{{ row.spec.value }}</div>
            <text-file @click="handleView(row)"/>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column
        label="数据类型"
        prop="spec.kv_type"
        :filter="{ filterFn: () => true, list: filterList, checked: filterChecked }"
      ></bk-table-column>
      <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
      <bk-table-column label="修改人" prop="revision.reviser"></bk-table-column>
      <bk-table-column label="修改时间" :sort="true" :width="220">
        <template #default="{ row }">
          <span v-if="row.revision">{{ datetimeFormat(row.revision.update_at) }}</span>
        </template>
      </bk-table-column>
      <bk-table-column v-if="versionData.id === 0" label="变更状态">
        <template #default="{ row }">
          <StatusTag :status="row.kv_state" />
        </template>
      </bk-table-column>
      <bk-table-column label="操作" fixed="right">
        <template #default="{ row }">
          <div class="operate-action-btns">
            <bk-button v-if="row.kv_state === 'DELETE'" text theme="primary" @click="handleUndelete(row)"
              >恢复</bk-button
            >
            <template v-else>
              <bk-button :disabled="row.kv_state === 'DELETE'" text theme="primary" @click="handleEditOrView(row)">{{
                versionData.id === 0 ? '编辑' : '查看'
              }}</bk-button>
              <bk-button
                v-if="versionData.status.publish_status !== 'editing'"
                text
                theme="primary"
                @click="handleDiff(row)"
                >对比</bk-button
              >
              <bk-button v-if="versionData.id === 0" text theme="primary" @click="handleDel(row)">删除</bk-button>
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
    :config="(activeConfig as IConfigKvItem)"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :editable="true"
    @confirm="getListData"
  />
  <ViewConfigKv v-model:show="viewPanelShow" :config="(activeConfig as IConfigKvItem)" />
  <VersionDiff v-model:show="isDiffPanelShow" :current-version="versionData" :selected-config-kv="diffConfig" />
  <DeleteConfirmDialog
    v-model:isShow="isDeleteConfigDialogShow"
    title="确认删除该配置项？"
    @confirm="handleDeleteConfigConfirm"
  >
    <div style="margin-bottom: 8px">
      配置项：<span style="color: #313238">{{ deleteConfig?.spec.key }}</span>
    </div>
    <div>{{ deleteConfigTips }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
import { ref, watch, onMounted, computed } from 'vue';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../../../store/config';
import useServiceStore from '../../../../../../../../store/service';
import { ICommonQuery } from '../../../../../../../../../types/index';
import { IConfigKvItem, IConfigKvType } from '../../../../../../../../../types/config';
import { getKvList, deleteKv, getReleaseKvList, undeleteKv } from '../../../../../../../../api/config';
import { datetimeFormat } from '../../../../../../../../utils/index';
import { CONFIG_KV_TYPE } from '../../../../../../../../constants/config';
import { TextFile } from 'bkui-vue/lib/icon';
import StatusTag from './status-tag';
import EditConfig from '../edit-config-kv.vue';
import ViewConfigKv from '../view-config-kv.vue';
import VersionDiff from '../../../components/version-diff/index.vue';
import TableEmpty from '../../../../../../../../components/table/table-empty.vue';
import DeleteConfirmDialog from '../../../../../../../../components/delete-confirm-dialog.vue';
import Message from 'bkui-vue/lib/message';

const configStore = useConfigStore();
const serviceStore = useServiceStore();
const { versionData } = storeToRefs(configStore);
const { checkPermBeforeOperate } = serviceStore;
const { permCheckLoading } = storeToRefs(serviceStore);

const props = defineProps<{
  bkBizId: string;
  appId: number;
  searchStr: string;
}>();

const emits = defineEmits(['clearStr']);

const loading = ref(false);
const configList = ref<IConfigKvType[]>([]);
const configsCount = ref(0);
const editPanelShow = ref(false);
const viewPanelShow = ref(false);
const activeConfig = ref<IConfigKvItem>();
const deleteConfig = ref<IConfigKvType>();
const isDiffPanelShow = ref(false);
const diffConfig = ref(0);
const isSearchEmpty = ref(false);
const isDeleteConfigDialogShow = ref(false);
const filterChecked = ref<string[]>([]);
const updateSortType = ref('null');
const pagination = ref({
  current: 1,
  count: 0,
  limit: 10,
});
const filterList = computed(() => CONFIG_KV_TYPE.map(item => ({
  value: item.id,
  text: item.name,
})));

const deleteConfigTips = computed(() => {
  if (deleteConfig.value) {
    return deleteConfig.value.kv_state === 'ADD'
      ? '一旦删除，该操作将无法撤销，请谨慎操作'
      : '配置项删除后，可以通过恢复按钮撤销删除';
  }
  return '';
});

watch(
  () => versionData.value.id,
  () => {
    refresh();
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
    if (filterChecked.value!.length > 0) {
      params.kv_type = filterChecked.value;
    }
    if (updateSortType.value !== 'null') {
      params.sort = 'updated_at';
      params.order = updateSortType.value.toUpperCase();
    }
    let res;
    if (isUnNamedVersion.value) {
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

const handleEditOrView = (config: IConfigKvType) => {
  activeConfig.value = config.spec;
  if (isUnNamedVersion.value) {
    editPanelShow.value = true;
  } else {
    viewPanelShow.value = true;
  }
};

const handleView = (config: IConfigKvType) => {
  activeConfig.value = config.spec;
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

const handleDeleteConfigConfirm = async () => {
  if (!deleteConfig.value) {
    return;
  }
  await deleteKv(props.bkBizId, props.appId, deleteConfig.value.id);
  if (configList.value.length === 1 && pagination.value.current > 1) {
    pagination.value.current -= 1;
  }
  Message({
    theme: 'success',
    message: '删除配置项成功',
  });
  refresh();
  isDeleteConfigDialogShow.value = false;
};

// 撤销删除
const handleUndelete = async (config: IConfigKvType) => {
  await undeleteKv(props.bkBizId, props.appId, config.spec.key);
  Message({ theme: 'success', message: '恢复配置项成功' });
  refresh();
};

const handlePageLimitChange = (limit: number) => {
  pagination.value.limit = limit;
  refresh();
};

const refresh = (current = 1) => {
  pagination.value.current = current;
  getListData();
};

const handleFilter = ({ checked }: any) => {
  filterChecked.value = checked;
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
.preview {
  display: flex;
  justify-content: space-between;
  .preview-value {
    width: 80%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  span {
    font-size: 12px;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
    }
  }
}
</style>
