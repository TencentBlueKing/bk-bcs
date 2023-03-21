<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { ECategoryType, EGroupRuleType } from '../../../../../types/group'
  import { IGroupCategoriesQuery, IGroupEditing, IGroupItem, IGroupRuleItem, IAllCategoryGroupItem } from '../../../../../types/group'
  import { getAllGroupList } from '../../../../api/group'
  import GroupEditDialog from './group-edit-dialog.vue'
  import CategoryGroup from './category-group.vue'

  const props = defineProps<{
    appId: number
  }>()

  const getDefaultGroupConfig = () => {
    return {
      name: '',
      group_category_id: '',
      mode: ECategoryType.Custom,
      rule_logic: 'AND',
      rules: [{ key: '', op: <EGroupRuleType>'', value: '' }],
      uid: ''
    }
  }

  const categoryTypes = [
    { id: ECategoryType.Custom, name: '普通分组' },
    { id: ECategoryType.Debug, name: '调试用分组' }
  ]
  const categoryList= ref<IAllCategoryGroupItem[]>([])
  const categoryListLoading = ref(true)
  const currentTab = ref<ECategoryType>(categoryTypes[0].id)
  const count = ref({
    [ECategoryType.Custom]: 0,
    [ECategoryType.Debug]: 0
  })
  const isGroupDialogShow = ref(false)
  const groupData = ref<IGroupEditing>(getDefaultGroupConfig())

  onMounted(() => {
    getAllGroupData()
  })

  const getAllGroupData = async() => {
    categoryListLoading.value = true
    const params: IGroupCategoriesQuery = {
      mode: currentTab.value,
      start: 0,
      limit: 100 // @todo 确认分页方式
    }
    const res = await getAllGroupList(props.appId, params)
    categoryList.value = res.details
    count.value[currentTab.value] = res.details.reduce((acc: number, crt: IAllCategoryGroupItem) => {
      return acc + crt.groups.length
    }, 0)
    categoryListLoading.value = false
  }

  const handleCreateGroup = () => {
    isGroupDialogShow.value = true
    groupData.value = getDefaultGroupConfig()
  }

  const handleEditGroup = (group: IGroupItem) => {
    const { id, attachment, spec } = group
    const { name, mode, selector, uid } = spec
    const logic = 'labels_and' in selector ? 'AND' : 'OR'
    const rules = logic === 'AND' ? selector.labels_and : selector.labels_or
    groupData.value = {
      id,
      name,
      mode,
      uid,
      rules: <IGroupRuleItem[]>rules,
      group_category_id: attachment.group_category_id,
      rule_logic: logic
    }
    isGroupDialogShow.value = true
  }

  const refreshCategoryList = () => {
    getAllGroupData()
  }

  const handleTabChange = (id: ECategoryType) => {
    currentTab.value = id
    categoryList.value = []
    getAllGroupData()
  }

</script>
<template>
    <section class="app-group-page">
      <div class="operate-area">
        <div class="action-btns">
          <bk-button theme="primary" @click="handleCreateGroup"><Plus class="button-icon" />创建分组</bk-button>
          <div class="group-tabs">
            <div
              v-for="(item, index) in categoryTypes"
              :key="item.id"
              :class="['tab-item', { active: currentTab === item.id }]"
              @click="handleTabChange(item.id)">
              <div class="tab-item-name">
                {{ item.name }}
                <span class="count">{{ count[item.id] }}</span>
              </div>
              <div v-if="index !== categoryTypes.length - 1" class="split-line"></div>
            </div>
          </div>
        </div>
        <bk-input class="search-group-input" placeholder="分组名称">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
      <bk-loading :loading="categoryListLoading" style="height: calc(100% - 34px);">
        <div class="group-list-wrapper">
          <template v-if="categoryList.length > 0">
            <CategoryGroup
              v-for="category in categoryList"
              :key="category.group_category_id"
              :app-id="props.appId"
              :mode="currentTab"
              :category-group="category"
              @edit="handleEditGroup" />
          </template>
          <bk-exception v-else type="empty">此服务下暂无分组</bk-exception>
        </div>
        <GroupEditDialog
          v-model:show="isGroupDialogShow"
          :category-list="categoryList"
          :app-id="props.appId"
          :group="groupData"
          @refreshCategoryList="refreshCategoryList"/>
      </bk-loading>
    </section>
</template>
<style lang="scss" scoped>
  .app-group-page {
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .action-btns {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .group-tabs {
    display: flex;
    align-items: center;
    margin-left: 24px;
    padding: 3px 4px;
    background: #f0f1f5;
    border-radius: 4px;
    .tab-item {
      display: flex;
      align-items: center;
      &.active {
        .tab-item-name {
          color: #3a84ff;
          background: #ffffff;
        }
        .count {
          background: #a3c5fd;
          color: #ffffff;
          line-height: 16px;
        }
      }
      &-name {
        padding: 6px 14px;
        font-size: 12px;
        line-height: 14px;
        color: #63656e;
        border-radius: 4px;
        cursor: pointer;
      }
      .count {
        padding: 0 8px;
        border-radius: 2px;
        color: #979ba5;
      }
    }
    .split-line {
      margin: 0 4px;
      height: 14px;
      width: 1px;
      background: #dcdee5;
    }
  }
  .button-icon {
    font-size: 18px;
  }
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .group-list-wrapper {
    margin-top: 19px;
    height: calc(100% - 34px);
    overflow: auto;
    .category-group:not(:last-child) {
      margin-bottom: 16px;
    }
  }
</style>
