<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { VERSION_STATUS_MAP } from '../../../../constants/config'
  import { IScriptCiteQuery, IScriptCitedItem } from '../../../../../types/script'
  import { getScriptCiteList } from '../../../../api/script'
  import SearchInput from '../../../../components/search-input.vue'

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
    const params: IScriptCiteQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.searchKey = searchStr.value
    }
    if (props.versionId) {
      params.release_id = props.versionId
    }
    const res = await getScriptCiteList(spaceId.value, props.id, params)
    list.value = res.details
    pagination.value.count = res.count
    loading.value = false
  }

  const getHref = (id: number) => {
    const { href } = router.resolve({ name: 'service-config', params: { spaceId: spaceId.value, appId: id } })
    return href
  }

  const refreshList = (current: number = 1) => {
    pagination.value.current = current
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
      <SearchInput v-model="searchStr" placeholder="服务名称/版本名称/被引用的版本" @search="refreshList()" />
    </div>
    <div class="cited-data-table">
      <bk-table
        :border="['outer']"
        :data="list"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-change="refreshList">
        <bk-table-column label="脚本名称" prop="hook_release_name"></bk-table-column>
        <bk-table-column label="服务名称" prop="app_name"></bk-table-column>
        <bk-table-column label="配置文件版本">
          <template #default="{ row }">
            <bk-link v-if="row.config_release_id" class="link-btn" theme="primary" target="_blank" :href="getHref(row.app_id)">{{ row.config_release_name }}</bk-link>
          </template>
        </bk-table-column>
        <bk-table-column label="配置文件版本状态">
          <template #default="{ row }">
            <span v-if="row.state">{{ VERSION_STATUS_MAP[row.state as keyof typeof VERSION_STATUS_MAP] }}</span>
          </template>
        </bk-table-column>
      </bk-table>
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
  .action-btn {
    padding: 8px 24px;
    background: #fafbfd;
    box-shadow: 0 -1px 0 0 #dcdee5;
    .bk-button {
      min-width: 88px;
    }
  }
</style>
