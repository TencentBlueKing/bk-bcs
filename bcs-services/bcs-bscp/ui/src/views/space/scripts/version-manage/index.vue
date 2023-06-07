<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { IScriptVersion } from '../../../../../types/script'
  import { getScriptDetail, getScriptVersionList } from '../../../../api/script'
  import DetailLayout from '../components/detail-layout.vue'
  import CreateVersion from './create-version.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const router = useRouter()
  const route = useRoute()

  const scriptId = ref(Number(route.params.spaceId))
  const detailLoading = ref(true)
  const scriptDetail = ref({ spec: { name: '' } })
  const versionLoading = ref(true)
  const versionList = ref<IScriptVersion[]>([])
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  onMounted(() => {
    getScriptDetailData()
    getVersionList()
  })

  // 获取脚本详情
  const getScriptDetailData = async() => {
    detailLoading.value = true
    scriptDetail.value = await getScriptDetail(spaceId.value, scriptId.value)
    detailLoading.value = false
  }

  // 获取版本列表
  const getVersionList = async() => {
    versionLoading.value = true
    const params: { start: number; limit: number; searchKey?: string } = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.searchKey = searchStr.value
    }
    const res = await getScriptVersionList(spaceId.value, scriptId.value, params)
    versionList.value = res.details
    pagination.value.count = res.count
  }

  const handleOpenScriptPanel = () => {}

  const refreshList = (val: number = 1) => {
    pagination.value.current = val
    getVersionList()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }

  const handleClose = () => {
    router.push({ name: 'script-list', params: { spaceId: spaceId.value } })
  }

</script>
<template>
  <DetailLayout
    v-if="!detailLoading"
    :name="`版本管理 - ${scriptDetail.spec.name}`"
    :show-footer="false"
    @close="handleClose">
    <template #content>
      <div class="script-version-manage">
        <div class="operation-area">
          <CreateVersion :script-id="scriptId" @create="handleOpenScriptPanel" />
          <bk-input class="search-input" placeholder="版本号/版本说明/更新人">
              <template #suffix>
                <Search class="search-input-icon" />
              </template>
          </bk-input>
        </div>
        <bk-table :border="['outer']" :data="versionList">
          <bk-table-column label="版本号" prop="spec.name" show-overflow-tooltip></bk-table-column>
          <bk-table-column label="版本说明">
            <template #default="{ row }">
              <span>{{ (row.spec && row.spec.memo) || '--' }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="被引用" prop="spec.publish_num"></bk-table-column>
          <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
          <bk-table-column label="更新时间" prop="revision.update_at"></bk-table-column>
          <bk-table-column label="状态">
            <template #default="{ row }">
              <span v-if="row.spec">{{ row.spec.pub_state }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="操作" width="280">
            <template #default="{ row }">
              <div class="action-btns">
                <bk-button text theme="primary">上线</bk-button>
                <bk-button text theme="primary">编辑</bk-button>
                <bk-button text theme="primary">版本对比</bk-button>
                <bk-button text theme="primary">复制并新建</bk-button>
                <bk-button text theme="primary">删除</bk-button>
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
      </div>
    </template>
  </DetailLayout>
</template>
<style lang="scss" scoped>
  .script-version-manage {
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .action-btns {
    .bk-button {
      margin-right: 8px;
    }
  }
  .table-list-pagination {
    padding: 12px;
    background: #ffffff;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>