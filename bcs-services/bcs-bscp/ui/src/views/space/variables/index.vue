<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { InfoBox } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../store/global'
  import { ICommonQuery, IPagination } from '../../../../types/index';
  import { IVariableEditParams, IVariableItem } from '../../../../types/variable'
  import { getVariableList, deleteVariable } from '../../../api/variable'
  import VariableCreate from './variable-create.vue';
  import VariableEdit from './variable-edit.vue';
  import SearchInput from '../../../components/search-input.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const loading = ref(false)
  const list = ref<IVariableItem[]>([])
  const searchStr = ref('')
  const pagination = ref<IPagination>({
    current: 1,
    count: 0,
    limit: 10,
  })

  const isCreateSliderShow = ref(false)
  const editSliderData = ref<{ open: boolean; id: number; data: IVariableEditParams; }>({
    open: false,
    id: 0,
    data: {
      name: '',
      type: '',
      default_val: '',
      memo: ''
    }
  })

  watch(() => spaceId.value, () => {
    refreshList()
  })

  onMounted(() => {
    getVariables()
  })

  const getVariables = async() => {
    loading.value = true
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.search_fields = 'name'
      params.search_value = searchStr.value
    }
    const res = await getVariableList(spaceId.value, params)
    list.value = res.details
    pagination.value.count = res.count
    loading.value = false
  }

  const handleSearchInputChange = () => {
    if (searchStr.value === '') {
      refreshList()
    }
  }

  const handleEditVar = (variable: IVariableItem) => {
    const { id, spec} = variable
    editSliderData.value = {
      open: true,
      id: id,
      data: { ...spec }
    }
  }

  const handleDeleteVar = (variable: IVariableItem) => {
    InfoBox({
      title: `确定删除变量[${variable.spec.name}]?`,
      confirmText: '删除',
      infoType: 'warning',
      onConfirm: async () => {
        await deleteVariable(spaceId.value, variable.id)
        if (list.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        getVariables()
      },
    } as any)
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }

  const refreshList = (current: number = 1) => {
    pagination.value.current = current
    getVariables()
  }
</script>
<template>
  <section class="variables-management-page">
    <bk-alert theme="info">
      可为配置文件内的变量预先定义默认值。默认变量值只在变量首次使用时生效，之后更新配置文件，变量值将使用上一次生成版本是的变量值填充。
    </bk-alert>
    <div class="operation-area">
      <bk-button theme="primary" @click="isCreateSliderShow = true"><Plus class="button-icon" />新增变量</bk-button>
      <SearchInput v-model="searchStr" placeholder="请输入变量名称" :width="320" @search="refreshList()" />
    </div>
    <div class="variable-table">
      <bk-table
        empty-text="暂无变量"
        :border="['outer']"
        :data="list"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-change="refreshList()">
        <bk-table-column label="变量名称">
          <template #default="{ row }">
            <bk-button v-if="row.spec" text theme="primary" @click="handleEditVar(row)">{{ row.spec.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column label="类型" prop="spec.type" width="120"></bk-table-column>
        <bk-table-column label="默认值" prop="spec.default_val"></bk-table-column>
        <bk-table-column label="描述">
          <template #default="{ row }">
            <span v-if="row.spec">{{ row.spec.memo || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="操作" width="200">
          <template #default="{ row }">
            <div class="action-btns">
              <bk-button text theme="primary" @click="handleEditVar(row)">编辑</bk-button>
              <bk-button text theme="primary" @click="handleDeleteVar(row)">删除</bk-button>
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </div>
    <VariableCreate v-model:show="isCreateSliderShow" @created="refreshList" />
    <VariableEdit
      v-model:show="editSliderData.open"
      :id="editSliderData.id"
      :data="editSliderData.data"
      @edited="refreshList" />
  </section>
</template>
<style lang="scss" scoped>
  .variables-management-page {
    height: 100%;
    background: #f5f7fa;
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    padding: 0 24px;
    .button-icon {
      font-size: 18px;
    }
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .variable-table {
    padding: 16px 24px 24px;
  }
  .action-btns {
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
