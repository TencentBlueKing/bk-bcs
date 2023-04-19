<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { InfoBox } from 'bkui-vue/lib'
  import { getSpaceGroupList, deleteGroup } from '../../api/group'
  import { IGroupItem } from '../../../types/group'
  import CreateGroup from './create-group.vue'
  import EditGroup from './edit-group.vue'
  import RuleTag from './components/rule-tag.vue'
  
  const route = useRoute()

  const groupList = ref([])
  const listLoading = ref(false)
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

  onMounted(() => {
    loadGroupList()
  })

  const loadGroupList = async() => {
    try {
      listLoading.value = true
      const res = await getSpaceGroupList(<string>route.params.spaceId)
      groupList.value = res.details
      pagination.value.count = res.details.length
    } catch(e) {
      console.error(e)
    } finally {
      listLoading.value = false
    }
  }

  const openCreateGroupDialog = () => {
    isCreateGroupShow.value = true
  }

  const openEditGroupDialog = (group: IGroupItem) => {
    isEditGroupShow.value = true
    editingGroup.value = group
  }

  const handleDeleteGroup = (group: IGroupItem) => { 
    InfoBox({
      title: `确认是否删除分组【${group.name}?】`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteGroup(<string>route.params.spaceId, group.id)
        if (groupList.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        loadGroupList()
      },
    } as any)
   }

  const handlePageLimitChange = () => {}

</script>
<template>
  <section class="groups-management-page">
    <div class="operate-area">
      <bk-button theme="primary" @click="openCreateGroupDialog"><Plus class="button-icon" />新增分组</bk-button>
      <div class="filter-actions">
        <bk-checkbox class="rule-filter-checkbox" size="small">按规则分类查看</bk-checkbox>
        <bk-input class="search-group-input" placeholder="分组名称/分组规则">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
    </div>
    <div class="group-table-wrapper">
      <template v-if="groupList.length > 0">
        <bk-table :border="['outer']" :data="groupList">
          <bk-table-column label="分组名称" :width="210" prop="name"></bk-table-column>
          <bk-table-column label="分组规则">
            <template #default="{ row }">
              <template v-if="row.selector">
                <rule-tag
                  v-for="(rule, index) in (row.selector.labels_or || row.selector.labels_and)"
                  class="tag-item"
                  :key="index"
                  :rule="rule"/>
              </template>
              <span v-else>-</span>
            </template>
          </bk-table-column>
          <bk-table-column label="服务可见范围" :width="240">
            <template #default="{ row }">
              <span v-if="row.public">公开</span>
              <span v-else>{{ row.bind_apps && row.bind_apps.join(',') }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="上线服务数" :width="110" prop="released_apps_num"></bk-table-column>
          <bk-table-column label="操作" :width="120">
            <template #default="{ row }">
              <div class="action-btns">
                <bk-button text theme="primary" @click="openEditGroupDialog(row)">编辑分组</bk-button>
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
          @change="loadGroupList"
          @limit-change="handlePageLimitChange"/>
      </template>
      <bk-exception v-else class="group-data-empty" type="empty" scene="part">
        当前暂无数据
        <div class="create-group-text-btn">
          <bk-button text theme="primary" @click="openCreateGroupDialog">立即创建</bk-button>
        </div>
      </bk-exception>
    </div>
    <create-group v-model:show="isCreateGroupShow" @reload="loadGroupList"></create-group>
    <edit-group v-model:show="isEditGroupShow" :group="editingGroup" @reload="loadGroupList"></edit-group>
  </section>
</template>
<style lang="scss" scoped>
  .groups-management-page {
    height: 100%;
    padding: 24px;
    background: #f5f7fa;
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
  }
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
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
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>
