<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { getUnNamedVersionAppsBoundByTemplate } from '../../../../../api/template';
  import { ICommonQuery } from '../../../../../../types/index';
  import { IAppBoundByTemplateDetailItem } from '../../../../../../types/template'
  import SearchInput from '../../../../../components/search-input.vue';

  const router = useRouter()

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    currentTemplateSpace: number;
    config: { id: number; name: string; };
  }>()

  const emits = defineEmits(['update:show'])

  const isShow = ref(false)
  const appList = ref<IAppBoundByTemplateDetailItem[]>([])
  const searchStr = ref('')
  const loading = ref(false)
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })

  watch(() => props.show, val => {
    isShow.value = val
    if (val) {
      searchStr.value = ''
      getList()
    }
  })

  const getList = async() => {
    loading.value = true
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.search_fields = 'name,id'
      params.search_value = searchStr.value
    }
    const res = await getUnNamedVersionAppsBoundByTemplate(props.spaceId, props.currentTemplateSpace, props.config.id, params)
    appList.value = res.details
    pagination.value.count = res.count
    loading.value = false
  }

  const getHref = (id: number) => {
    const { href } = router.resolve({ name: 'service-config', params: { spaceId: props.spaceId, appId: id } })
    return href
  }

  const handleSearch = (val: string) => {
    searchStr.value = val
    pagination.value.current = 1
    getList()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.current = 1
    pagination.value.limit = val
    getList()
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
  <bk-sideslider
    title="被引用"
    :width="640"
    :quick-close="true"
    :is-show="isShow"
    @closed="close">
    <div class="top-area">
      <div class="config-name">{{ props.config.name }}</div>
      <SearchInput placeholder="服务名称/配置项版本" v-model="searchStr" :width="320" @search="handleSearch" />
    </div>
    <div class="apps-table">
      <bk-table
        :border="['outer']"
        :data="appList"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-change="getList">
        <bk-table-column label="配置项版本" prop="template_revision_name"></bk-table-column>
        <bk-table-column label="引用此配置项的服务">
          <template #default="{ row }">
            <bk-link
              v-if="row.app_id"
              class="link-btn"
              theme="primary"
              target="_blank"
              :href="getHref(row.app_id)">
              {{ row.app_name }}
            </bk-link>
          </template>
        </bk-table-column>
      </bk-table>
    </div>
    <div class="action-btn">
      <bk-button @click="close">关闭</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .top-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 24px 0;
    padding: 0 24px;
    height: 32px;
    .config-name {
      margin-right: 8px;
      font-size: 14px;
      font-weight: 700;
      color: #63656e;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .search-input {
      flex-shrink: 0;
    }
  }
  .apps-table {
    padding: 0 24px;
    height: calc(100vh - 180px);
    overflow: auto;
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
