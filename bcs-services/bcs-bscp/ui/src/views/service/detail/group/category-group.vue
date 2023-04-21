<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { RightShape, Del } from 'bkui-vue/lib/icon'
  import { InfoBox } from "bkui-vue/lib";
  import { IGroupItem, ICategoryItem, ECategoryType, IAllCategoryGroupItem } from '../../../../../types/group'
  import { getCategoryGroupList, delCategory, deleteGroup } from '../../../../api/group'
  import RuleTag from '../../../groups/components/rule-tag.vue'

  const props = defineProps<{
    appId: number,
    mode: ECategoryType,
    categoryGroup: IAllCategoryGroupItem
  }>()

  const emits = defineEmits(['edit'])

  const folded = ref(true)
  const listData = ref<IGroupItem[]>([])
  const listLoading = ref(false)
  const count = ref(props.categoryGroup.groups.length)
  const pagination = ref({
    count: 0,
    limit: 10,
    current: 1
  })

  watch(() => folded.value, (val) => {
    if (!val) {
      getListData()
    }
  })

  const getListData = async() => {
    listLoading.value = true
    const params = {
      mode: props.mode,
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    const res = await getCategoryGroupList(props.appId, props.categoryGroup.group_category_id, params)
    const { count, details } = res
    listData.value = details
    listLoading.value = false
    pagination.value.count = count
  }

  const refreshList = () => {
    pagination.value.current = 1
    getListData()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }

  const handleToggleFold = () => {
    folded.value = !folded.value
  }

  const handleDeleteCategory = () => {
    if (listData.value.length > 0) {
      InfoBox({
        title: `暂无法删除【${props.categoryGroup.group_category_name}】`,
        subTitle: '请先删除此分类下所有分组',
        type: "warning",
        headerAlign: "center" as const,
        footerAlign: "center" as const
      } as any)
    } else {
      InfoBox({
        title: `确认是否删除分类【${props.categoryGroup.group_category_name}?】`,
        type: "danger",
        headerAlign: "center" as const,
        footerAlign: "center" as const,
        onConfirm: async () => {
          await delCategory(props.appId, props.categoryGroup.group_category_id)
          return true
        },
      } as any)
    }
  }

  // 编辑分组
  const handleEditGroup = (group: IGroupItem) => { 
    emits('edit', group)
   }

  // 删除分组
  const handleDeleteGroup = (group: IGroupItem) => { 
    InfoBox({
      title: `确认是否删除分组【${group.spec.name}?】`,
      type: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteGroup(props.appId, group.id)
        if (listData.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        getListData()
      },
    } as any)
   }

</script>
<template>
  <section class="category-group">
    <div :class="['header-area', { 'expanded': !folded }]" @click="handleToggleFold">
      <div class="category-content">
        <RightShape class="arrow-icon" />
        <span class="name">{{ categoryGroup.group_category_name }}</span>
        <span>（{{ count }}）</span>
      </div>
      <Del class="delete-icon" @click.stop="handleDeleteCategory" />
    </div>
    <template v-if="!folded">
      <bk-loading :loading="listLoading">
        <bk-table class="group-table" :border="['outer']" :data="listData">
          <bk-table-column label="分组名称" prop="spec.name"></bk-table-column>
          <bk-table-column label="分组规则">
            <template #default="{ row }">
              <template v-if="row.spec && row.spec.mode === ECategoryType.Custom">
                <rule-tag
                  v-for="(rule, index) in (row.spec.selector.labels_or || row.spec.selector.labels_and)"
                  class="tag-item"
                  :key="index"
                  :rule="rule"/>
              </template>
              <span v-else>-</span>
            </template>
          </bk-table-column>
          <bk-table-column label="当前上线版本"></bk-table-column>
          <bk-table-column label="操作" :width="180">
            <template #default="{ row }">
              <div class="action-btns">
                <bk-button text theme="primary" @click="handleEditGroup(row)">编辑分组</bk-button>
                <bk-button text theme="primary" @click="handleDeleteGroup(row)">删除</bk-button>
              </div>
            </template>
          </bk-table-column>
        </bk-table>
        <bk-pagination
          class="table-list-pagination"
          v-model="pagination.current"
          location="left"
          :layout="['total', 'limit', 'list']"
          :count="pagination.count"
          :limit="pagination.limit"
          @change="refreshList"
          @limit-change="handlePageLimitChange"/>
      </bk-loading>
    </template>
  </section>
</template>
<style lang="scss" scoped>
  .header-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 18px 8px 8px;
    background: #dcdee5;
    border-radius: 2px;
    cursor: pointer;
    &.expanded {
      .arrow-icon {
        transform: rotate(90deg);
      }
    }
    .category-content {
      display: flex;
      align-items: center;
      color: #313238;
      font-size: 12px;
      line-height: 16px;
    }
    .arrow-icon {
      display: inline-block;
      font-size: 12px;
      color: #63656e;
    }
    .name {
      margin-left: 9px;
    }
    .delete-icon {
      font-size: 13px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .group-table {
    margin-top: 8px
  }
  .tag-item:not(:first-of-type) {
    margin-left: 8px;
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