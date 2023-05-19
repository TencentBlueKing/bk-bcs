<script setup lang="ts">
  import { ref, onMounted, nextTick } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Plus, Search, DownShape } from 'bkui-vue/lib/icon'
  import { InfoBox } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../store/global'
  import { getSpaceGroupList, deleteGroup } from '../../../api/group'
  import { IGroupItem, IGroupCategory, IGroupCategoryItem } from '../../../../types/group'
  import CreateGroup from './create-group.vue'
  import EditGroup from './edit-group.vue'
  import RuleTag from './components/rule-tag.vue'
  import ServicesToPublished from './services-to-published.vue'
  
  const { spaceId } = storeToRefs(useGlobalStore())

  const listLoading = ref(false)
  const groupList = ref<IGroupItem[]>([])
  const categorizedGroupList = ref<IGroupCategory[]>([])
  const tableData = ref<IGroupItem[]|IGroupCategoryItem[]>([])
  const isCategorizedView = ref(false) // 按规则分类查看
  const changeViewPending = ref(false)
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })
  const isCreateGroupShow = ref(false)
  const isEditGroupShow = ref(false)
  const editingGroup = ref<IGroupItem>({
    id: 0,
    name: '',
    public: true,
    bind_apps: [],
    released_apps_num: 0,
    selector: {
      labels_and: []
    }
  })
  const isPublishedSliderShow = ref(false)

  onMounted(async() => {
    await loadGroupList()
    refreshTableData()
  })

  // 加载全量分组数据
  const loadGroupList = async() => {
    try {
      listLoading.value = true
      const res = await getSpaceGroupList(spaceId.value)
      groupList.value = res.details
      categorizedGroupList.value = categorizingData(res.details)
      pagination.value.count = res.details.length
      refreshTableData()
    } catch(e) {
      console.error(e)
    } finally {
      listLoading.value = false
    }
  }

  // 刷新表格数据
  const refreshTableData = () => {
    if (isCategorizedView.value) {
      categorizedGroupList.value.forEach(item => {
        item.fold = false
        item.show = true
      })
      tableData.value = getCategorizedTableData()
    } else {
      const start = pagination.value.limit * (pagination.value.current - 1)
      tableData.value = groupList.value.slice(start, start + pagination.value.limit)
    }
  }

  // 将全量分组数据按照分类分组
  const categorizingData = (data: IGroupItem[]) => {
    const categoryList: IGroupCategory[] = []
    data.forEach(group => {
      const selector = group.selector.labels_and || group.selector.labels_or
      selector?.forEach(rule => {
        const data = categoryList.find(item => item.name === rule.key)
        if (data) {
          data.children.push({...group, CATEGORY_NAME: rule.key})
        } else {
          categoryList.push({
            name: rule.key,
            show: true,
            fold: false,
            children: [{...group, CATEGORY_NAME: rule.key}]
          })
        }
      })
    })
    return categoryList
  }

  // 分类视图下的table数据
  const getCategorizedTableData = () => {
    const list: IGroupCategoryItem[] = []
    categorizedGroupList.value.forEach(category => {
      if (category.show) {
        list.push({ IS_CATEORY_ROW: true, CATEGORY_NAME: category.name, fold: category.fold, bind_apps: [] })
        if (!category.fold) {
          list.push(...category.children)
        }
      }
    })
    return list
  }

  // 获取服务可见范围值
  const getBindAppsName = (apps: { id: number; name: string }[] = []) => {
    return apps.map(item => item && item.name).join('; ')
  }

  // 创建分组
  const openCreateGroupDialog = () => {
    isCreateGroupShow.value = true
  }

  // 编辑分组
  const openEditGroupDialog = (group: IGroupItem) => {
    isEditGroupShow.value = true
    editingGroup.value = group
  }

  // 切换分类查看视图
  const handleChangeView = () => {
    changeViewPending.value = true
    pagination.value.current = 1
    refreshTableData()
    nextTick(() => {
      changeViewPending.value = false
    })
  }

  // 搜索
  // @todo 规则搜索交互确定
  const handleSearch = () => {}

  //关联服务
  const handleOpenPublishedSlider = (group: IGroupItem) => {
    isPublishedSliderShow.value = true
    editingGroup.value = group
  }

  // 删除分组
  const handleDeleteGroup = (group: IGroupItem) => { 
    InfoBox({
      title: `确认是否删除分组【${group.name}?】`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteGroup(spaceId.value, group.id)
        if (tableData.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        loadGroupList()
      },
    } as any)
  }

  // 分类展开/收起
  const handleToggleCategoryFold = (name: string) => {
    const category = categorizedGroupList.value.find(category => category.name === name)
    if (category) {
      category.fold = !category.fold
      tableData.value = getCategorizedTableData()
    }
  }

  const handlePageChange = (val: number) => {
    pagination.value.current = val
    refreshTableData()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.current = 1
    pagination.value.limit = val
    refreshTableData()
  }

</script>
<template>
  <section class="groups-management-page">
    <div class="operate-area">
      <bk-button theme="primary" @click="openCreateGroupDialog"><Plus class="button-icon" />新增分组</bk-button>
      <div class="filter-actions">
        <bk-checkbox
          v-model="isCategorizedView"
          class="rule-filter-checkbox"
          size="small"
          :true-label="true"
          :false-label="false"
          :disabled="changeViewPending"
          @change="handleChangeView">
          按规则分类查看
        </bk-checkbox>
        <bk-input class="search-group-input" placeholder="分组名称/分组规则" @enter="handleSearch">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
    </div>
    <div class="group-table-wrapper">
      <bk-loading style="min-height: 300px;" :loading="listLoading">
        <template v-if="groupList.length > 0">
          <bk-table class="group-table" :border="['outer']" :data="tableData">
            <bk-table-column label="分组名称" :width="210" show-overflow-tooltip>
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
            <bk-table-column label="分组规则" show-overflow-tooltip>
              <template #default="{ row }">
                <template v-if="!row.IS_CATEORY_ROW">
                  <template v-if="row.selector">
                    <span v-for="(rule, index) in (row.selector.labels_or || row.selector.labels_and)" :key="index">
                      <span v-if="index > 0 "> & </span>
                      <rule-tag class="tag-item" :rule="rule"/>
                    </span>
                  </template>
                  <span v-else>-</span>
                </template>
              </template>
            </bk-table-column>
            <bk-table-column label="服务可见范围" :width="240">
              <template #default="{ row }">
                <template v-if="!row.IS_CATEORY_ROW">
                  <span v-if="row.public">公开</span>
                  <span v-else>{{ getBindAppsName(row.bind_apps) }}</span>
                </template>
              </template>
            </bk-table-column>
            <bk-table-column label="分组状态" :width="100">
              <template #default="{ row }">
                <span class="group-status">
                  <div :class="['dot', { 'published': row.released_apps_num > 0 }]"></div>
                  {{ row.released_apps_num > 0 ? '已上线': '未上线' }}
                </span>
              </template>
            </bk-table-column>
            <bk-table-column label="上线服务数" :width="110">
              <template #default="{ row }">
                <template v-if="!row.IS_CATEORY_ROW">
                  <template v-if="row.released_apps_num === 0">0</template>
                  <bk-button v-else text theme="primary" @click="handleOpenPublishedSlider(row)">{{ row.released_apps_num }}</bk-button>
                </template>
              </template>
            </bk-table-column>
            <bk-table-column label="操作" :width="120">
              <template #default="{ row }">
                <div v-if="!row.IS_CATEORY_ROW" class="action-btns">
                  <bk-button text theme="primary" :disabled="row.released_apps_num > 0" @click="openEditGroupDialog(row)">编辑分组</bk-button>
                  <bk-button text theme="primary" :disabled="row.released_apps_num > 0" @click="handleDeleteGroup(row)">删除</bk-button>
                </div>
              </template>
            </bk-table-column>
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
            @limit-change="handlePageLimitChange"/>
        </template>
        <bk-exception v-if="groupList.length === 0 && !listLoading" class="group-data-empty" type="empty" scene="part">
          当前暂无数据
          <div class="create-group-text-btn">
            <bk-button text theme="primary" @click="openCreateGroupDialog">立即创建</bk-button>
          </div>
        </bk-exception>
      </bk-loading>
    </div>
    <create-group v-model:show="isCreateGroupShow" @reload="loadGroupList"></create-group>
    <edit-group v-model:show="isEditGroupShow" :group="editingGroup" @reload="loadGroupList"></edit-group>
    <services-to-published v-model:show="isPublishedSliderShow" :id="editingGroup.id" :name="editingGroup.name"></services-to-published>
  </section>
</template>
<style lang="scss" scoped>
  .groups-management-page {
    height: 100%;
    padding: 24px;
    background: #f5f7fa;
    overflow: hidden;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
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
      max-height: calc(100vh - 200px);
      overflow: auto;
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
    .bk-button:not(:last-of-type) {
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
