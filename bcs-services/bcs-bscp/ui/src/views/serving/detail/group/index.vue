<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { ECategoryType } from '../../../../../types/group'
  import { ICategoryItem, IGroupCategoriesQuery, IGroupItem, ICategoryGroup } from '../../../../../types/group'
  import { getGroupCategories } from '../../../../api/group'
  import CreateGroup from './create-group.vue'
  import CategoryGroup from './category-group.vue'

  const props = defineProps<{
    appId: number
  }>()

  const categoryTypes = [
    { id: ECategoryType.Custom, name: '普通分组' },
    { id: ECategoryType.Debug, name: '调试用分组' }
  ]
  const categoryList= ref<Array<ICategoryGroup>>([])
  const categoryListLoading = ref(true)
  const currentTab = ref<ECategoryType>(categoryTypes[0].id)
  const isCreateDialogShow = ref(false)

  onMounted(() => {
    getCategoryList()
  })

  const getCategoryList = async() => {
    categoryListLoading.value = true
    const params: IGroupCategoriesQuery = {
      mode: currentTab.value,
      start: 0,
      limit: 100 // @todo 确认分页方式
    }
    const res = await getGroupCategories(props.appId, params)
    categoryList.value = res.details.map((item: ICategoryItem) => {
      return { config: item, groups: { count: 0, data: [] } }
    })
    categoryListLoading.value = false
  }

  const handleTabChange = (id: ECategoryType) => {
    currentTab.value = id
    getCategoryList()
  }

</script>
<template>
    <section class="app-group-page">
      <div class="operate-area">
        <div class="action-btns">
          <bk-button theme="primary" @click="isCreateDialogShow = true"><Plus class="button-icon" />创建分组</bk-button>
          <div class="group-tabs">
            <div
              v-for="item in categoryTypes"
              :key="item.id"
              :class="['tab-item', { active: currentTab === item.id }]"
              @click="handleTabChange(item.id)">
              {{ item.name }}
              <span class="count">0</span>
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
            <CategoryGroup v-for="category in categoryList" :key="category.config.id" :category-group="category" />
          </template>
          <bk-exception v-else type="empty">此服务下暂无分组</bk-exception>
        </div>
        <CreateGroup :category-list="categoryList" v-model:show="isCreateDialogShow" />
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
      padding: 6px 14px;
      font-size: 12px;
      line-height: 14px;
      color: #63656e;
      border-radius: 4px;
      cursor: pointer;
      &.active {
        color: #3a84ff;
        background: #ffffff;
        .count {
          background: #a3c5fd;
          color: #ffffff;
          line-height: 16px;
        }
      }
      .count {
        padding: 0 8px;
        border-radius: 2px;
        color: #979ba5;
      }
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
  }
</style>
