<template>
  <div class="package-config-table">
    <div class="operate-area">
      <div class="table-operate-btns">
        <slot name="tableOperations"> </slot>
      </div>
      <bk-input
        v-model="searchStr"
        class="search-script-input"
        :placeholder="t('配置文件名称/路径/描述/创建人/更新人')"
        :clearable="true"
        @clear="refreshList()"
        @input="handleSearchInputChange"
      >
        <template #suffix>
          <Search class="search-input-icon" />
        </template>
      </bk-input>
    </div>
    <bk-loading style="min-height: 200px" :loading="listLoading">
      <bk-table
        class="config-table"
        :border="['outer']"
        :data="list"
        :row-class="getRowCls"
        :remote-pagination="true"
        :pagination="pagination"
        :is-selected-fn="isSelectedFn"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="refreshList($event,true)"
        @selection-change="handleSelectionChange"
        @select-all="handleSelectAll"
      >
        <bk-table-column type="selection" :min-width="40" :width="40" class="aaaa"></bk-table-column>
        <bk-table-column :label="t('配置文件名称')">
          <template #default="{ row }">
            <div v-if="row.spec" v-overflow-title class="config-name" @click="goToViewVersionManage(row.id)">
              {{ row.spec.name }}
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('配置文件路径')" prop="spec.path"></bk-table-column>
        <bk-table-column :label="t('配置文件描述')" prop="spec.memo">
          <template #default="{ row }">
            <span v-if="row.spec">{{ row.spec.memo || '--' }}</span>
          </template>
        </bk-table-column>
        <template v-if="showCitedByPkgsCol">
          <bk-table-column :label="t('所在套餐')" :width="200">
            <template #default="{ index }">
              <template v-if="citedByPkgsLoading"><Spinner /></template>
              <template v-else-if="citeByPkgsList[index]">
                <PkgsTag v-if="citeByPkgsList[index].length > 0" :pkgs="citeByPkgsList[index]" />
                <span v-else>--</span>
              </template>
            </template>
          </bk-table-column>
        </template>
        <template v-if="showBoundByAppsCol">
          <bk-table-column :label="t('被引用')">
            <template #default="{ row, index }">
              <template v-if="boundByAppsCountLoading"><Spinner /></template>
              <template v-else-if="boundByAppsCountList[index]">
                <bk-button
                  v-if="boundByAppsCountList[index].bound_unnamed_app_count > 0"
                  text
                  theme="primary"
                  @click="handleOpenAppBoundByTemplateSlider(row)"
                >
                  {{ boundByAppsCountList[index].bound_unnamed_app_count }}
                </bk-button>
                <span v-else>0</span>
              </template>
            </template>
          </bk-table-column>
        </template>
        <bk-table-column :label="t('创建人')" prop="revision.creator" :width="100"></bk-table-column>
        <bk-table-column :label="t('更新人')" prop="revision.reviser" :width="100"></bk-table-column>
        <bk-table-column :label="t('更新时间')" prop="" :width="180">
          <template #default="{ row }">
            <template v-if="row.revision">
              {{ datetimeFormat(row.revision.update_at) }}
            </template>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('操作')" :width="locale === 'zh-CN' ? '140' : '200'" fixed="right">
          <template #default="{ row, index }">
            <div class="actions-wrapper">
              <slot name="columnOperations" :config="row">
                <bk-button theme="primary" text @click="goToCreateVersionManage(row.id)">{{ t('编辑') }}</bk-button>
                <bk-button theme="primary" text @click="goToVersionManage(row.id)">{{ t('版本管理') }}</bk-button>
                <bk-popover
                  theme="light template-config-actions-popover"
                  placement="bottom-end"
                  :popover-delay="[0, 100]"
                  :arrow="false"
                >
                  <div class="more-actions">
                    <Ellipsis class="ellipsis-icon" />
                  </div>
                  <template #content>
                    <div class="config-actions">
                      <div class="action-item" @click="handleOpenAddToPkgsDialog(row)">{{ t('添加至套餐') }}</div>
                      <div
                        v-if="citeByPkgsList[index].length > 0"
                        class="action-item"
                        @click="handleOpenMoveOutFromPkgsDialog(row)"
                      >
                        {{ t('移出套餐') }}
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </slot>
            </div>
          </template>
        </bk-table-column>
        <template #empty>
          <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchStr"></TableEmpty>
        </template>
      </bk-table>
    </bk-loading>
    <AddToDialog v-model:show="isAddToPkgsDialogShow" :value="crtConfig" @added="handleAdded" />
    <MoveOutFromPkgsDialog
      v-model:show="isMoveOutFromPkgsDialogShow"
      :id="crtConfig.length > 0 ? crtConfig[0].id : 0"
      :name="crtConfig.length > 0 ? crtConfig[0].spec.name : ''"
      :current-pkg="props.currentPkg"
      @moved-out="handleMovedOut"
    />
    <AppsBoundByTemplate
      v-model:show="appBoundByTemplateSliderData.open"
      :space-id="spaceId"
      :current-template-space="currentTemplateSpace"
      :config="appBoundByTemplateSliderData.data"
    />
  </div>
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { Ellipsis, Search, Spinner } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../../../store/global';
import useTemplateStore from '../../../../../../store/template';
import { ICommonQuery } from '../../../../../../../types/index';
import {
  ITemplateConfigItem,
  ITemplateCitedCountDetailItem,
  ITemplateCitedByPkgs,
} from '../../../../../../../types/template';
import { getPackagesByTemplateIds, getCountsByTemplateIds } from '../../../../../../api/template';
import { datetimeFormat } from '../../../../../../utils/index';
import AddToDialog from '../operations/add-to-pkgs/add-to-dialog.vue';
import MoveOutFromPkgsDialog from '../operations/move-out-from-pkg/move-out-from-pkgs-dialog.vue';
import PkgsTag from '../../components/packages-tag.vue';
import AppsBoundByTemplate from '../apps-bound-by-template.vue';
import TableEmpty from '../../../../../../components/table/table-empty.vue';
import { debounce } from 'lodash';

const router = useRouter();
const { t, locale } = useI18n();
const { spaceId } = storeToRefs(useGlobalStore());
const templateStore = useTemplateStore();
const { currentTemplateSpace, versionListPageShouldOpenEdit, versionListPageShouldOpenView, batchUploadIds } =
  storeToRefs(templateStore);

const props = defineProps<{
  currentPkg: number | string;
  selectedConfigs: ITemplateConfigItem[];
  showCitedByPkgsCol?: boolean; // 是否显示模板被套餐引用列
  showBoundByAppsCol?: boolean; // 是否显示模板被服务引用列
  getConfigList: Function;
}>();

const emits = defineEmits(['update:selectedConfigs']);

const listLoading = ref(false);
const list = ref<ITemplateConfigItem[]>([]);
const citedByPkgsLoading = ref(false);
const citeByPkgsList = ref<ITemplateCitedByPkgs[][]>([]);
const boundByAppsCountLoading = ref(false);
const boundByAppsCountList = ref<ITemplateCitedCountDetailItem[]>([]);
const searchStr = ref('');
const pagination = ref({
  current: 1,
  count: 0,
  limit: 10,
});
const isAddToPkgsDialogShow = ref(false);
const isMoveOutFromPkgsDialogShow = ref(false);
const appBoundByTemplateSliderData = ref<{ open: boolean; data: { id: number; name: string } }>({
  open: false,
  data: {
    id: 0,
    name: '',
  },
});
const crtConfig = ref<ITemplateConfigItem[]>([]);
const isSearchEmpty = ref(false);

watch(
  () => props.currentPkg,
  () => {
    searchStr.value = '';
    loadConfigList();
  },
);

onMounted(() => {
  loadConfigList();
});

const loadConfigList = async (isBatchUpload = false) => {
  listLoading.value = true;
  const params: ICommonQuery = {
    start: (pagination.value.current - 1) * pagination.value.limit,
    limit: pagination.value.limit,
  };
  if (!isBatchUpload) batchUploadIds.value = [];
  if (batchUploadIds.value.length > 0) {
    params.ids = batchUploadIds.value.join(',');
  }
  if (searchStr.value) {
    params.search_fields = 'name,path,memo,creator,reviser';
    params.search_value = searchStr.value;
  }
  const res = await props.getConfigList(params);
  list.value = res.details;
  pagination.value.count = res.count;
  listLoading.value = false;
  const ids = list.value.map(item => item.id);
  citeByPkgsList.value = [];
  boundByAppsCountList.value = [];
  if (ids.length > 0) {
    if (props.showCitedByPkgsCol) {
      loadCiteByPkgsCountList(ids);
    }

    if (props.showBoundByAppsCol) {
      loadBoundByAppsList(ids);
    }
  }
};

const loadCiteByPkgsCountList = async (ids: number[]) => {
  citedByPkgsLoading.value = true;
  const res = await getPackagesByTemplateIds(spaceId.value, currentTemplateSpace.value, ids);
  citeByPkgsList.value = res.details;
  citedByPkgsLoading.value = false;
};

const loadBoundByAppsList = async (ids: number[]) => {
  boundByAppsCountLoading.value = true;
  const res = await getCountsByTemplateIds(spaceId.value, currentTemplateSpace.value, ids);
  boundByAppsCountList.value = res.details;
  boundByAppsCountLoading.value = false;
};

const refreshList = (current = 1, isBatchUpload = false) => {
  isSearchEmpty.value = searchStr.value !== '';
  pagination.value.current = current;
  loadConfigList(isBatchUpload);
};

// 模板移出或删除后刷新列表
const refreshListAfterDeleted = (num: number) => {
  if (num === list.value.length && pagination.value.current > 1) {
    pagination.value.current -= 1;
  }
  batchUploadIds.value = [];
  refreshList();
};

const handleSearchInputChange = debounce(() => refreshList(), 300);
const handleSelectionChange = ({
  checked,
  row,
}: {
  checked: boolean;
  row: ITemplateConfigItem;
}) => {
  const configs = props.selectedConfigs.slice();
  if (checked) {
    if (!configs.find(item => item.id === row.id)) {
      configs.push(row);
    }
  } else {
    const index = configs.findIndex(item => item.id === row.id);
    if (index > -1) {
      configs.splice(index, 1);
    }
  }
  emits('update:selectedConfigs', configs);
};

const handleSelectAll = ({ checked }: {checked: boolean;}) => {
  if (checked)  {
    emits('update:selectedConfigs', list.value);
  } else {
    emits('update:selectedConfigs', []);
  }
};

const isSelectedFn = ({
  row,
}: {
  row: ITemplateConfigItem;
}) => props.selectedConfigs.findIndex(item => item.id === row.id) > -1;


const handleOpenAddToPkgsDialog = (config: ITemplateConfigItem) => {
  isAddToPkgsDialogShow.value = true;
  crtConfig.value = [config];
};

const handleOpenMoveOutFromPkgsDialog = (config: ITemplateConfigItem) => {
  isMoveOutFromPkgsDialogShow.value = true;
  crtConfig.value = [config];
};

const handleAdded = () => {
  refreshList();
  updateRefreshFlag();
};

const handleMovedOut = () => {
  refreshListAfterDeleted(1);
  crtConfig.value = [];
  updateRefreshFlag();
};

const handleOpenAppBoundByTemplateSlider = (config: ITemplateConfigItem) => {
  appBoundByTemplateSliderData.value = {
    open: true,
    data: {
      id: config.id,
      name: config.spec.name,
    },
  };
};

const updateRefreshFlag = () => {
  templateStore.$patch((state) => {
    state.needRefreshMenuFlag = true;
  });
};

const goToVersionManage = (id: number) => {
  router.push({
    name: 'template-version-manage',
    params: {
      templateSpaceId: currentTemplateSpace.value,
      packageId: props.currentPkg,
      templateId: id,
    },
  });
};

const goToViewVersionManage = (id: number) => {
  versionListPageShouldOpenView.value = true;
  goToVersionManage(id);
};

const goToCreateVersionManage = (id: number) => {
  versionListPageShouldOpenEdit.value = true;
  goToVersionManage(id);
};

// 设置新增行的标记class
const getRowCls = (data: ITemplateConfigItem) => {
  if (batchUploadIds.value.includes(data.id)) {
    return 'new-row-marked';
  }
  return '';
};

const handlePageLimitChange = (val: number) => {
  pagination.value.limit = val;
  refreshList(1, true);
};

const clearSearchStr = () => {
  searchStr.value = '';
  refreshList();
};

defineExpose({
  refreshList,
  refreshListAfterDeleted,
});
</script>
<style lang="scss" scoped>
.operate-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  .table-operate-btns {
    display: flex;
    align-items: center;
    :deep(.bk-button) {
      margin-right: 8px;
    }
  }
}
.search-script-input {
  width: 320px;
}
.search-input-icon {
  padding-right: 10px;
  color: #979ba5;
  background: #ffffff;
}
.config-name {
  color: #3a84ff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}
.actions-wrapper {
  display: flex;
  align-items: center;
  height: 100%;
  .more-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 8px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    cursor: pointer;
    &:hover {
      background: #dcdee5;
      color: #3a84ff;
    }
  }
  .ellipsis-icon {
    transform: rotate(90deg);
  }
  .bk-button {
    margin-right: 8px;
  }
}
.config-table {
  :deep(.bk-table-body) {
    tr.new-row-marked td {
      background: #f2fff4 !important;
    }
  }
}
</style>
<style lang="scss">
.template-config-actions-popover.bk-popover.bk-pop2-content {
  padding: 4px 0;
  border: 1px solid #dcdee5;
  box-shadow: 0 2px 6px 0 #0000001a;
  .config-actions {
    .action-item {
      padding: 0 12px;
      min-width: 58px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
}
.bk-table .bk-table-body table td .cell.selection {
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
