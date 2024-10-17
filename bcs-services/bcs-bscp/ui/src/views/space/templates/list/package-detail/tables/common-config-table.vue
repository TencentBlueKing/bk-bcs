<template>
  <div class="package-config-table">
    <div class="operate-area">
      <div class="table-operate-btns">
        <slot name="tableOperations"> </slot>
      </div>
      <bk-input
        v-model="searchStr"
        class="search-script-input"
        :placeholder="t('配置文件名/描述/创建人/更新人')"
        :clearable="true"
        @clear="refreshList()"
        @input="handleSearchInputChange"
        v-bk-tooltips="{ content: t('配置文件名/描述/创建人/更新人'), disabled: locale === 'zh-cn' }">
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
        @page-limit-change="handlePageLimitChange"
        @page-value-change="handlePageChange($event)">
        <template #prepend>
          <render-table-tip />
        </template>
        <bk-table-column :min-width="73" :width="73" :label="renderSelection">
          <template #default="{ row }">
            <across-check-box :checked="isChecked(row)" :handle-change="() => handleSelectionChange(row)" />
          </template>
        </bk-table-column>
        <bk-table-column :label="t('配置文件名')" :show-overflow-tooltip="true">
          <template #default="{ row }">
            <span v-if="row.spec" class="config-name" @click="handleViewConfig(row)">
              {{ fileAP(row) }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('配置文件描述')" prop="spec.memo" :show-overflow-tooltip="true">
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
                  @click="handleOpenAppBoundByTemplateSlider(row)">
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
        <bk-table-column :label="t('操作')" :width="locale === 'zh-cn' ? '160' : '200'" fixed="right">
          <template #default="{ row, index }">
            <div class="actions-wrapper">
              <slot name="columnOperations" :config="row">
                <bk-button theme="primary" text @click="handleEditConfig(row)">{{ t('编辑') }}</bk-button>
                <bk-button theme="primary" text @click="goToVersionManage(row.id)">{{ t('版本管理') }}</bk-button>
                <bk-popover
                  theme="light template-config-actions-popover"
                  placement="bottom-end"
                  :popover-delay="[0, 100]"
                  :arrow="false">
                  <div class="more-actions">
                    <Ellipsis class="ellipsis-icon" />
                  </div>
                  <template #content>
                    <div class="config-actions">
                      <div class="action-item" @click="handleOpenAddToPkgsDialog(row, index)">
                        {{ t('添加至套餐') }}
                      </div>
                      <div
                        v-if="citeByPkgsList[index]?.length > 0"
                        class="action-item"
                        @click="handleOpenMoveOutFromPkgsDialog(row)">
                        {{ t('移出套餐') }}
                      </div>
                      <DownloadConfig
                        class="action-item"
                        theme=""
                        :text="$t('下载模板文件')"
                        :space-id="spaceId"
                        :template-space-id="currentTemplateSpace"
                        :template-id="row.id" />
                      <div v-if="props.showDeleteAction" class="action-item" @click="handleDeleteClick(row)">
                        {{ t('删除模板文件') }}
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
    <AddToDialog
      v-model:show="isAddToPkgsDialogShow"
      :value="crtConfig"
      :cite-by-pkg-ids="configCiteByPkgIds"
      :is-across-checked="props.isAcrossChecked"
      :data-count="props.dataCount"
      @added="handleAdded" />
    <MoveOutFromPkgsDialog
      v-model:show="isMoveOutFromPkgsDialogShow"
      :id="crtConfig.length > 0 ? crtConfig[0].id : 0"
      :name="crtConfig.length > 0 ? fileAP(crtConfig[0]) : ''"
      :current-pkg="props.currentPkg"
      @moved-out="handleMovedOut" />
    <AppsBoundByTemplate
      v-model:show="appBoundByTemplateSliderData.open"
      :space-id="spaceId"
      :current-template-space="currentTemplateSpace"
      :config="appBoundByTemplateSliderData.data" />
    <DeleteConfigDialog
      v-model:show="isDeleteConfigDialogShow"
      :configs="crtConfig"
      :is-across-checked="props.isAcrossChecked"
      :data-count="props.dataCount"
      @deleted="handleConfigsDeleted" />
    <ViewConfig
      v-model:show="isViewConfigShow"
      :space-id="spaceId"
      :id="viewConfig?.id as number"
      :memo="selectConfigMemo"
      @open-edit="handleEditConfig(viewConfig as ITemplateConfigItem)" />
    <EditConfig
      v-model:show="isEditConfigShow"
      :memo="selectConfigMemo"
      :space-id="spaceId"
      :id="editConfigId"
      @edited="refreshList" />
  </div>
</template>
<script lang="ts" setup>
  import { onMounted, ref, watch, computed, toRef } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Ellipsis, Search, Spinner } from 'bkui-vue/lib/icon';
  import { debounce } from 'lodash';
  import useGlobalStore from '../../../../../../store/global';
  import useTemplateStore from '../../../../../../store/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import CheckType from '../../../../../../../types/across-checked';
  import useTablePagination from '../../../../../../utils/hooks/use-table-pagination';
  import useTableAcrossCheck from '../../../../../../utils/hooks/use-table-acrosscheck';
  import {
    ITemplateCitedCountDetailItem,
    ITemplateCitedByPkgs,
    ITemplateConfigItem,
  } from '../../../../../../../types/template';
  import { getPackagesByTemplateIds, getCountsByTemplateIds } from '../../../../../../api/template';
  import { datetimeFormat } from '../../../../../../utils/index';
  import AddToDialog from '../operations/add-to-pkgs/add-to-dialog.vue';
  import MoveOutFromPkgsDialog from '../operations/move-out-from-pkg/move-out-from-pkgs-dialog.vue';
  import PkgsTag from '../../components/packages-tag.vue';
  import AppsBoundByTemplate from '../apps-bound-by-template.vue';
  import TableEmpty from '../../../../../../components/table/table-empty.vue';
  import DownloadConfig from '../operations/download-config/download-config.vue';
  import DeleteConfigDialog from '../operations/delete-configs/delete-config-dialog.vue';
  import acrossCheckBox from '../../../../../../components/across-checkbox.vue';
  import ViewConfig from '../operations/view-config/view-config.vue';
  import EditConfig from '../operations/edit-config/edit-config.vue';

  const router = useRouter();
  const { t, locale } = useI18n();
  const { spaceId } = storeToRefs(useGlobalStore());
  const templateStore = useTemplateStore();
  const { currentTemplateSpace, topIds } = storeToRefs(templateStore);
  const { pagination, updatePagination } = useTablePagination('commonConfigTable');

  const props = defineProps<{
    currentPkg: number | string;
    selectedConfigs: ITemplateConfigItem[];
    showCitedByPkgsCol?: boolean; // 是否显示模板被套餐引用列
    showBoundByAppsCol?: boolean; // 是否显示模板被服务引用列
    showDeleteAction?: boolean; // 是否显示删除操作
    getConfigList: Function;
    isAcrossChecked: boolean;
    dataCount: number;
  }>();

  const emits = defineEmits(['update:selectedConfigs', 'sendAcrossCheckedType']);

  const listLoading = ref(false);
  const list = ref<ITemplateConfigItem[]>([]);
  const citedByPkgsLoading = ref(false);
  const citeByPkgsList = ref<ITemplateCitedByPkgs[][]>([]);
  const boundByAppsCountLoading = ref(false);
  const boundByAppsCountList = ref<ITemplateCitedCountDetailItem[]>([]);
  const searchStr = ref('');
  const isAddToPkgsDialogShow = ref(false); // 显示添加至套餐弹窗
  const isMoveOutFromPkgsDialogShow = ref(false); // 显示从套餐移除弹窗
  const isDeleteConfigDialogShow = ref(false); // 显示删除配置弹窗
  const appBoundByTemplateSliderData = ref<{ open: boolean; data: { id: number; name: string } }>({
    open: false,
    data: {
      id: 0,
      name: '',
    },
  });
  const crtConfig = ref<ITemplateConfigItem[]>([]);
  const isSearchEmpty = ref(false);
  const isViewConfigShow = ref(false);
  const viewConfig = ref<ITemplateConfigItem>();
  const isEditConfigShow = ref(false);
  const editConfigId = ref(0);
  const selectConfigMemo = ref('');
  const configCiteByPkgIds = ref<number[]>([]);
  const isAcrossChecked = ref(false);

  const crossPageSelect = computed(() => pagination.value.limit < pagination.value.count);

  const { selectType, selections, renderSelection, renderTableTip, handleRowCheckChange, handleClearSelection } =
    useTableAcrossCheck({
      dataCount: toRef(pagination.value, 'count'),
      curPageData: list, // 当前页数据
      rowKey: ['id'],
      crossPageSelect,
    });

  watch(
    () => props.currentPkg,
    () => {
      searchStr.value = '';
      loadConfigList();
    },
  );
  watch(selections, () => {
    isAcrossChecked.value = [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value);
    emits('update:selectedConfigs', selections.value);
    emits('sendAcrossCheckedType', isAcrossChecked.value, pagination.value.count);
  });

  onMounted(() => {
    handleClearSelection();
    loadConfigList();
  });

  // 配置文件名
  const fileAP = computed(() => (config: ITemplateConfigItem) => {
    const { path, name } = config.spec;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  });

  const loadConfigList = async (createConfig = false) => {
    listLoading.value = true;
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    if (!createConfig) {
      templateStore.$patch((state) => {
        state.topIds = [];
      });
    }
    if (topIds.value.length > 0) {
      params.ids = topIds.value;
    }
    if (searchStr.value) {
      params.search_fields = 'name,path,memo,creator,reviser';
      params.search_value = searchStr.value;
    }
    const res = await props.getConfigList(params);
    list.value = res.details;
    pagination.value.count = res.count;
    templateStore.$patch((state) => {
      state.countOfTemplatesForCurrentPackage = res.count;
    });
    listLoading.value = false;
    const ids = list.value.map((item) => item.id);
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

  // 配置项被套餐引用数据
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

  // 翻页
  const handlePageChange = (event: number) => {
    refreshList(event, true, true);
  };

  const refreshList = (current = 1, createConfig = false, pageChange = false) => {
    isSearchEmpty.value = searchStr.value !== '';
    pagination.value.current = current;
    // 非跨页全选/半选 需要重置全选状态
    if (![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value) || !pageChange) {
      handleClearSelection();
    }
    loadConfigList(createConfig);
  };

  // 模板移出或删除后刷新列表
  const refreshListAfterDeleted = (num: number) => {
    if (num === list.value.length && pagination.value.current > 1) {
      pagination.value.current -= 1;
    }
    templateStore.$patch((state) => {
      state.topIds = [];
    });
    refreshList();
  };

  const handleSearchInputChange = debounce(() => {
    refreshList();
  }, 300);

  const handleSelectionChange = (row: ITemplateConfigItem) => {
    const isSelected = selections.value.some((item) => item.id === row.id);
    // const isAcrossChecked = [CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value);
    const shouldBeChecked = isAcrossChecked.value ? isSelected : !isSelected;
    // 根据选择类型决定传递的状态
    handleRowCheckChange(shouldBeChecked, row);
    emits('update:selectedConfigs', selections.value);
    emits('sendAcrossCheckedType', isAcrossChecked.value, pagination.value.count);
  };

  // 选中状态
  const isChecked = (row: ITemplateConfigItem) => {
    if (![CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value)) {
      // 当前页状态传递
      return selections.value.some((item) => item.id === row.id);
    }
    // 跨页状态传递
    return !selections.value.some((item) => item.id === row.id);
  };

  // 添加至套餐
  const handleOpenAddToPkgsDialog = (config: ITemplateConfigItem, index: number) => {
    isAddToPkgsDialogShow.value = true;
    crtConfig.value = [config];
    configCiteByPkgIds.value = citeByPkgsList.value[index].map((pkg) => pkg.template_set_id);
  };

  // 从套餐移除
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

  // 删除配置项
  const handleDeleteClick = async (config: ITemplateConfigItem) => {
    isDeleteConfigDialogShow.value = true;
    crtConfig.value = [config];
  };

  const handleConfigsDeleted = () => {
    refreshListAfterDeleted(1);
    crtConfig.value = [];
    updateRefreshFlag();
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

  const handleViewConfig = (config: ITemplateConfigItem) => {
    isViewConfigShow.value = true;
    viewConfig.value = config;
    selectConfigMemo.value = config.spec.memo;
  };

  const handleEditConfig = (config: ITemplateConfigItem) => {
    isViewConfigShow.value = false;
    isEditConfigShow.value = true;
    editConfigId.value = config.id;
    selectConfigMemo.value = config.spec.memo;
  };

  // 设置新增行的标记class
  const getRowCls = (data: ITemplateConfigItem) => {
    if (topIds.value.includes(data.id)) {
      return 'new-row-marked';
    }
    return '';
  };

  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
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
      max-height: calc(100vh - 320px);
      overflow: auto;
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
