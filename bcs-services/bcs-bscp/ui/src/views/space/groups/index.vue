<template>
  <bk-alert theme="info">
    {{ headInfo }}
    <span @click="goGroupDoc" class="hyperlink">{{ t('分组管理') }}</span>
  </bk-alert>
  <section class="groups-management-page">
    <div class="operate-area">
      <div class="btns">
        <bk-button theme="primary" @click="openCreateGroupDialog">
          <Plus class="button-icon" />
          {{ t('新增分组') }}
        </bk-button>
        <BatchDeleteBtn :bk-biz-id="spaceId" :selected-ids="selectedIds" @deleted="refreshAfterBatchDelete" />
      </div>
      <div class="filter-actions">
        <bk-checkbox
          v-model="isCategorizedView"
          class="rule-filter-checkbox"
          size="small"
          :true-label="true"
          :false-label="false"
          :disabled="changeViewPending"
          @change="handleChangeView">
          {{ t('按标签分类查看') }}
        </bk-checkbox>
        <bk-input
          class="search-group-input"
          :placeholder="t('分组名称/标签选择器')"
          @input="handleSearch"
          v-model.trim="searchInfo"
          @clear="handleSearch"
          :clearable="true">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
        </bk-input>
      </div>
    </div>
    <div class="group-table-wrapper">
      <bk-loading style="min-height: 300px" :loading="listLoading">
        <bk-table
          class="group-table"
          :row-class="getRowCls"
          show-overflow-tooltip
          :border="['outer']"
          :data="tableData">
          <template #prepend>
            <render-table-tip />
          </template>
          <bk-table-column :width="100" :label="renderSelection">
            <template #default="{ row }">
              <across-check-box
                :checked="selections.some((item) => item.name === row.name && item.id === row.id)"
                :disabled="row.released_apps_num > 0 || row.IS_CATEORY_ROW !== undefined"
                :handle-change="
                  () =>
                    handleRowCheckChange(!selections.some((item) => item.name === row.name && item.id === row.id), row)
                " />
            </template>
          </bk-table-column>
          <bk-table-column :label="t('分组名称')" :width="210" show-overflow-tooltip>
            <template #default="{ row }">
              <div v-if="isCategorizedView" class="categorized-view-name">
                <DownShape
                  v-if="row.IS_CATEORY_ROW"
                  :class="['fold-icon', { fold: row.fold }]"
                  @click="handleToggleCategoryFold(row.CATEGORY_NAME)" />
                {{ row.IS_CATEORY_ROW ? row.CATEGORY_NAME : row.name }}
              </div>
              <template v-else>{{ row.name }}</template>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('标签选择器')" show-overflow-tooltip>
            <template #default="{ row }">
              <template v-if="!row.IS_CATEORY_ROW">
                <template v-if="row.selector">
                  <span v-for="(rule, index) in row.selector.labels_or || row.selector.labels_and" :key="index">
                    <span v-if="index > 0"> & </span>
                    <rule-tag class="tag-item" :rule="rule" />
                  </span>
                </template>
                <span v-else>-</span>
              </template>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('服务可见范围')" :align="'center'" :width="240">
            <template #default="{ row }">
              <template v-if="!row.IS_CATEORY_ROW">
                <span v-if="row.public">{{ t('公开') }}</span>
                <span v-else>{{ getBindAppsName(row.bind_apps) }}</span>
              </template>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('分组状态')" :width="locale === 'zh-cn' ? '100' : '140'">
            <template #default="{ row }">
              <template v-if="!row.IS_CATEORY_ROW">
                <span class="group-status">
                  <div :class="['dot', { published: row.released_apps_num > 0 }]"></div>
                  {{ row.released_apps_num > 0 ? t('已上线') : t('未上线') }}
                </span>
              </template>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('上线服务数')" :align="'center'" :width="locale === 'zh-cn' ? '110' : '130'">
            <template #default="{ row }">
              <template v-if="!row.IS_CATEORY_ROW">
                <template v-if="row.released_apps_num === 0">0</template>
                <bk-button v-else text theme="primary" @click="handleOpenPublishedSlider(row)">{{
                  row.released_apps_num
                }}</bk-button>
              </template>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作')" :width="locale === 'zh-cn' ? '120' : '150'">
            <template #default="{ row }">
              <div v-if="!row.IS_CATEORY_ROW" class="action-btns">
                <div v-bk-tooltips="handleTooltip(row.released_apps_num, t('编辑'))" class="btn-item">
                  <bk-button
                    text
                    theme="primary"
                    :disabled="row.released_apps_num > 0"
                    @click="openEditGroupDialog(row)">
                    {{ t('编辑分组') }}
                  </bk-button>
                </div>
                <div v-bk-tooltips="handleTooltip(row.released_apps_num, t('删除'))" class="btn-item">
                  <bk-button text theme="primary" :disabled="row.released_apps_num > 0" @click="handleDeleteGroup(row)">
                    {{ t('删除') }}
                  </bk-button>
                </div>
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <tableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchInfo" />
          </template>
        </bk-table>
        <bk-pagination
          v-if="!isCategorizedView"
          v-model="pagination.current"
          class="table-list-pagination"
          location="left"
          :limit="pagination.limit"
          :layout="['total', 'limit', 'list']"
          :count="pagination.count"
          @change="handlePageChange"
          @limit-change="handlePageLimitChange" />
      </bk-loading>
    </div>
    <create-group v-model:show="isCreateGroupShow" @reload="loadGroupList($event)"></create-group>
    <edit-group v-model:show="isEditGroupShow" :group="editingGroup" @reload="loadGroupList"></edit-group>
    <services-to-published
      v-model:show="isPublishedSliderShow"
      :id="editingGroup.id"
      :name="editingGroup.name"></services-to-published>
  </section>
  <DeleteConfirmDialog
    v-model:is-show="isDeleteGroupDialogShow"
    :title="t('确认删除该分组？')"
    @confirm="handleDeleteGroupConfirm">
    <div style="margin-bottom: 8px">
      {{ t('分组名称') }}: <span style="color: #313238; font-weight: 600">{{ deleteGroupItem?.name }}</span>
    </div>
    <div>{{ t('一旦删除，该操作将无法撤销，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, nextTick, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { Plus, Search, DownShape } from 'bkui-vue/lib/icon';
  import Message from 'bkui-vue/lib/message';
  import { debounce } from 'lodash';
  import useGlobalStore from '../../../store/global';
  import CheckType from '../../../../types/across-checked';
  import { getSpaceGroupList, deleteGroup } from '../../../api/group';
  import useTablePagination from '../../../utils/hooks/use-table-pagination';
  import useTableAcrossCheck from '../../../utils/hooks/use-table-acrosscheck-fulldata';
  import { IGroupItem, IGroupCategory, IGroupCategoryItem } from '../../../../types/group';
  import CreateGroup from './create-group.vue';
  import EditGroup from './edit-group.vue';
  import BatchDeleteBtn from './batch-delete-btn.vue';
  import RuleTag from './components/rule-tag.vue';
  import ServicesToPublished from './services-to-published.vue';
  import tableEmpty from '../../../components/table/table-empty.vue';
  import DeleteConfirmDialog from '../../../components/delete-confirm-dialog.vue';
  import acrossCheckBox from '../../../components/across-checkbox.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t, locale } = useI18n();
  const { pagination, updatePagination } = useTablePagination('groupList');

  const listLoading = ref(false);
  const groupList = ref<IGroupItem[]>([]);
  const searchGroupList = ref<IGroupItem[]>([]);
  const categorizedGroupList = ref<IGroupCategory[]>([]);
  const tableData = ref<IGroupItem[] | IGroupCategoryItem[]>([]);
  const isCategorizedView = ref(false); // 按规则分类查看
  const searchInfo = ref('');
  const changeViewPending = ref(false);
  const isDeleteGroupDialogShow = ref(false);
  const deleteGroupItem = ref<IGroupItem>();
  const isCreateGroupShow = ref(false);
  const isEditGroupShow = ref(false);
  const editingGroup = ref<IGroupItem>({
    id: 0,
    name: '',
    public: true,
    bind_apps: [],
    released_apps_num: 0,
    selector: {
      labels_and: [],
    },
  });
  const isPublishedSliderShow = ref(false);
  const isSearchEmpty = ref(false);
  const topId = ref<number | undefined>(0);

  const headInfo = computed(() =>
    t(
      '分组由 1 个或多个标签选择器组成，服务配置版本选择分组上线结合客户端配置的标签用于灰度发布、A/B Test等运营场景，详情参考文档：',
    ),
  );
  // 跨页全选
  const selecTableData = computed(() => groupList.value.filter((item) => item.released_apps_num < 1));
  const pageSelectableData = computed(() => tableData.value.filter((item) => item.released_apps_num! < 1));
  const selectedIds = computed(() => {
    return selections.value.filter((item) => item.released_apps_num === 0).map((item) => item.id);
  });
  // 是否提供跨页全选功能
  const crossPageSelect = computed(() => {
    return (
      !isCategorizedView.value && pagination.value.limit < groupList.value.length && selecTableData.value.length !== 0
    );
  });
  const { selectType, selections, renderSelection, renderTableTip, handleRowCheckChange, handleClearSelection } =
    useTableAcrossCheck({
      tableData: selecTableData, // 全量数据，排除禁用
      curPageData: pageSelectableData, // 当前页数据，排除禁用
      crossPageSelect, // 展示跨页下拉框；按标签分类查看无分页，默认为当前页
    });

  watch(
    () => spaceId.value,
    async () => {
      pagination.value.current = 1;
      await loadGroupList();
      refreshTableData();
    },
  );

  onMounted(async () => {
    await loadGroupList();
    refreshTableData();
  });

  // 加载全量分组数据
  const loadGroupList = async (id?: number) => {
    try {
      listLoading.value = true;
      topId.value = id;
      const res = await getSpaceGroupList(spaceId.value, id);
      groupList.value = res.details;
      searchGroupList.value = res.details;
      categorizedGroupList.value = categorizingData(res.details);
      pagination.value.count = res.details.length;
      refreshTableData();
    } catch (e) {
      console.error(e);
    } finally {
      listLoading.value = false;
    }
  };

  // 刷新表格数据
  const refreshTableData = (pageChange = false) => {
    if (isCategorizedView.value) {
      categorizedGroupList.value = categorizingData(searchGroupList.value);
      categorizedGroupList.value.forEach((item) => {
        item.fold = false;
        item.show = true;
      });
      tableData.value = getCategorizedTableData();
    } else {
      const start = pagination.value.limit * (pagination.value.current - 1);
      tableData.value = searchGroupList.value.slice(start, start + pagination.value.limit);
      pagination.value.count = searchGroupList.value.length;
    }
    // 非跨页全选/半选 需要重置全选状态
    if (![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value) || !pageChange) {
      handleClearSelection();
    }
  };

  // 将全量分组数据按照分类分组
  const categorizingData = (data: IGroupItem[]) => {
    const categoryList: IGroupCategory[] = [];
    data.forEach((group) => {
      const selector = group.selector.labels_and || group.selector.labels_or;
      selector?.forEach((rule) => {
        const data = categoryList.find((item) => item.name === rule.key);
        if (data) {
          data.children.push({ ...group, CATEGORY_NAME: rule.key });
        } else {
          categoryList.push({
            name: rule.key,
            show: true,
            fold: false,
            children: [{ ...group, CATEGORY_NAME: rule.key }],
          });
        }
      });
    });
    return categoryList;
  };

  // 分类视图下的table数据
  const getCategorizedTableData = () => {
    const list: IGroupCategoryItem[] = [];
    categorizedGroupList.value.forEach((category) => {
      if (category.show) {
        list.push({ IS_CATEORY_ROW: true, CATEGORY_NAME: category.name, fold: category.fold, bind_apps: [] });
        if (!category.fold) {
          list.push(...category.children);
        }
      }
    });
    return list;
  };

  // 获取服务可见范围值
  const getBindAppsName = (apps: { id: number; name: string }[] = []) => apps.map((item) => item?.name).join('; ');

  // 创建分组
  const openCreateGroupDialog = () => {
    isCreateGroupShow.value = true;
  };

  // 编辑分组
  const openEditGroupDialog = (group: IGroupItem) => {
    isEditGroupShow.value = true;
    editingGroup.value = group;
  };

  // 切换分类查看视图
  const handleChangeView = () => {
    changeViewPending.value = true;
    pagination.value.current = 1;
    refreshTableData();
    nextTick(() => {
      changeViewPending.value = false;
    });
  };

  // 搜索
  const handleSearch = debounce(() => {
    if (!searchInfo.value) {
      searchGroupList.value = groupList.value;
      isSearchEmpty.value = false;
    } else {
      const lowercaseSearchStr = searchInfo.value.toLowerCase().replace(/\s/g, '');
      // 分组名称过滤出来的数据
      const groupNameList = groupList.value.filter((item) => item.name.toLowerCase().includes(lowercaseSearchStr));
      // 分组规则过滤出来的数据
      const groupRuleList = groupList.value.filter((item) => {
        let groupRuleMatch = false;
        item.selector?.labels_and?.forEach((labels) => {
          if (typeof labels.value === 'string') {
            const valueStr = labels.value as string;
            if (labels.key.includes(lowercaseSearchStr) || valueStr.includes(lowercaseSearchStr)) {
              groupRuleMatch = true;
            }
          }
        });
        return groupRuleMatch;
      });
      const searchArr = [...groupNameList, ...groupRuleList];
      const uniqueIds = new Set(); // 用于记录已经出现过的id
      searchGroupList.value = searchArr.filter((obj) => {
        if (uniqueIds.has(obj.id)) {
          return false;
        }
        uniqueIds.add(obj.id);
        return true;
      });
      isSearchEmpty.value = true;
    }
    refreshTableData();
  }, 300);

  // 关联服务
  const handleOpenPublishedSlider = (group: IGroupItem) => {
    isPublishedSliderShow.value = true;
    editingGroup.value = group;
  };

  // 删除分组
  const handleDeleteGroup = (group: IGroupItem) => {
    isDeleteGroupDialogShow.value = true;
    deleteGroupItem.value = group;
  };

  const handleDeleteGroupConfirm = async () => {
    await deleteGroup(spaceId.value, deleteGroupItem.value!.id);
    Message({
      theme: 'success',
      message: t('删除分组成功'),
    });
    if (tableData.value.length === 1 && pagination.value.current > 1) {
      pagination.value.current = pagination.value.current - 1;
    }
    loadGroupList();
    isDeleteGroupDialogShow.value = false;
  };

  // 分类展开/收起
  const handleToggleCategoryFold = (name: string) => {
    const category = categorizedGroupList.value.find((category) => category.name === name);
    if (category) {
      category.fold = !category.fold;
      tableData.value = getCategorizedTableData();
    }
  };

  const handlePageChange = (val: number) => {
    pagination.value.current = val;
    refreshTableData(true);
  };

  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
    refreshTableData();
  };

  // hover提示文字
  const handleTooltip = (flag: boolean, info: string) => {
    if (flag) {
      return {
        content: `${t('分组已上线，不能')}${info}`,
        placement: 'top',
      };
    }
    return { disabled: true };
  };

  // 清空搜索框
  const clearSearchInfo = () => {
    searchInfo.value = '';
    handleSearch();
  };

  // 批量删除后刷新表格数据
  const refreshAfterBatchDelete = () => {
    if (
      !isCategorizedView.value &&
      selections.value.length === tableData.value.length &&
      pagination.value.current > 1
    ) {
      pagination.value.current = pagination.value.current - 1;
    }
    loadGroupList();
  };

  // @ts-ignore
  // eslint-disable-next-line
  const goGroupDoc = () => window.open(BSCP_CONFIG.group_doc);

  const getRowCls = (group: IGroupItem) => {
    if (topId.value === group.id) {
      return 'new-row-marked';
    }
  };
</script>
<style lang="scss" scoped>
  .hyperlink {
    color: #3a84ff;
    cursor: pointer;
  }
  .groups-management-page {
    height: calc(100% - 33px);
    padding: 24px;
    background: #f5f7fa;
    overflow: hidden;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .btns {
      display: flex;
      align-items: center;
      :deep(.bk-button:not(:first-child)) {
        margin-left: 8px;
      }
    }
    .button-icon {
      font-size: 18px;
    }
  }
  .filter-actions {
    display: flex;
    align-items: center;
  }
  .rule-filter-checkbox {
    margin-right: 16px;
    font-size: 12px;
  }
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    font-size: 14px;
    background: #ffffff;
  }
  .group-table-wrapper {
    :deep(.bk-table-body) {
      max-height: calc(100vh - 280px);
      overflow: auto;
      tr.new-row-marked td {
        background: #f2fff4 !important;
      }
    }
  }
  .categorized-view-name {
    position: relative;
    padding-left: 20px;
    .fold-icon {
      position: absolute;
      top: 14px;
      left: 0;
      font-size: 14px;
      color: #3a84ff;
      // transition: transform 0.2s ease-in-out;
      cursor: pointer;
      &.fold {
        color: #c4c6cc;
        transform: rotate(-90deg);
      }
    }
  }
  .tag-item {
    padding: 0 10px;
    background: #f0f1f5;
    border-radius: 2px;
  }
  .group-status {
    display: flex;
    align-items: center;
    .dot {
      margin-right: 10px;
      width: 8px;
      height: 8px;
      background: #f0f1f5;
      border: 1px solid #c4c6cc;
      border-radius: 50%;
      &.published {
        background: #e5f6ea;
        border: 1px solid #3fc06d;
      }
    }
  }
  .group-data-empty {
    margin-top: 90px;
    color: #63656e;
    .create-group-text-btn {
      margin-top: 8px;
    }
  }
  .action-btns {
    .btn-item {
      display: inline-block;
    }
    .btn-item:not(:last-of-type) {
      margin-right: 8px;
    }
  }
  .table-list-pagination {
    padding: 12px;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    background: #ffffff;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>
