<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import { Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { IScriptCiteQuery, IScriptCitedItem } from '../../../../../types/script'
  import { getScriptCiteList, getScriptVersionCiteList } from '../../../../api/script'

  const { spaceId } = storeToRefs(useGlobalStore())

  const router = useRouter()

  const props = defineProps<{
    show: boolean;
    id: number;
    versionId?: number;
  }>()

  const emits = defineEmits(['update:show'])

  const loading = ref(false)
  const list = ref<IScriptCitedItem[]>([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10
  })

  watch(()=> props.show, val => {
    if (val) {
      getCitedData()
    }
  })

  const getCitedData = async () => {
    loading.value = true
    let res
    const params: IScriptCiteQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.searchKey = searchStr.value
    }
    if (props.versionId) {
      res = await getScriptVersionCiteList(spaceId.value, props.id, props.versionId, params)
    } else {
      res = await getScriptCiteList(spaceId.value, props.id, params)
    }
    list.value = res.details
    pagination.value.count = res.count
    loading.value = false
  }

  const getHref = (id: number) => {
    const { href } = router.resolve({ name: 'service-config', params: { spaceId: spaceId.value, appId: id } })
    return href
  }

  const handleNameInputChange = (val: string) => {
    if (!val) {
      refreshList()
    }
  }

  const refreshList = () => {
    pagination.value.current = 1
    getCitedData()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    getCitedData()
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
      <bk-input
        v-model="searchStr"
        class="search-input"
        placeholder="服务名称/版本名称/被引用的版本"
        :clearable="true"
        @enter="refreshList"
        @clear="refreshList"
        @change="handleNameInputChange">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
      </bk-input>
    </div>
    <div class="cited-data-table">
      <bk-table :border="['outer']" :data="list">
        <bk-table-column label="脚本版本">
          <template #default="{ row }">
            <template v-if="row.hook_revision_name || row.revision_name">{{ row.hook_revision_name || row.revision_name }}</template>
          </template>
        </bk-table-column>
        <bk-table-column label="服务名称" prop="app_name"></bk-table-column>
        <bk-table-column label="配置文件版本">
          <template #default="{ row }">
            <bk-link v-if="row.release_name" class="link-btn" theme="primary" target="_blank" :href="getHref(row.app_id)">{{ row.release_name }}</bk-link>
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
  .link-btn {
    font-size: 12px;
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