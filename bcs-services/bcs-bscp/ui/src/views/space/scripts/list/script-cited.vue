<script setup lang="ts">
  import { ref } from 'vue'
  import { Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { IScriptCiteQuery } from '../../../../../types/script'
  import { getScriptCiteList } from '../../../../api/script'

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    show: boolean;
    id: number;
  }>()

  const emits = defineEmits(['update:show'])

  const loading = ref(false)
  const list = ref([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10
  })

  const getScriptCiteData = async () => {
    loading.value = true
    const params: IScriptCiteQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.searchKey = searchStr.value
    }
    const res = await getScriptCiteList(spaceId.value, props.id, params)
    list.value = res.detail
    loading.value = false
  }

  const refreshList = (val: number = 1) => {
    pagination.value.current = val
    getScriptCiteData()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    getScriptCiteData()
  }

  const handleClose = () => {
    emits('update:show', false)
  }
</script>
<template>
  <bk-sideslider
    title="关联配置项"
    width="640"
    :is-show="props.show"
    @closed="handleClose">
    <div class="search-area">
      <bk-input class="search-input" placeholder="服务名称/版本名称/被引用的版本">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
      </bk-input>
    </div>
    <div class="cited-data-table">
      <bk-table :border="['outer']">
        <bk-table-column label="脚本名称"></bk-table-column>
        <bk-table-column label="服务名称"></bk-table-column>
        <bk-table-column label="配置文件版本"></bk-table-column>
        <bk-table-column label="配置文件版本状态"></bk-table-column>
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
    </div>
    <div class="action-btn">
      <bk-button @click="handleClose">关闭</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .search-area {
    padding: 24px 24px 16px;
    text-align: right;
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .cited-data-table {
    padding: 0 24px;
    height: calc(100vh - 172px);
    overflow: auto;
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
  .action-btn {
    padding: 8px 24px;
    background: #fafbfd;
    box-shadow: 0 -1px 0 0 #dcdee5;
    .bk-button {
      min-width: 88px;
    }
  }
</style>