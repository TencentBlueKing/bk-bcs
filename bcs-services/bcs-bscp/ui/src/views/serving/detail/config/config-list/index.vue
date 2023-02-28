<script setup lang="ts">
  import { defineProps, ref, computed, watch, onMounted } from 'vue'
  import InfoBox from "bkui-vue/lib/info-box";
  import { FilterOp, IPageFilter } from '../../../../../types'
  import { getServingConfigList } from '../../../../../api/config'
  import { deleteServingConfigItem } from '../../../../../api/config'
  import EditConfig from './edit-config.vue'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const loading = ref(false)
  const configList = ref([])
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })
  const editPanelShow = ref(false)
  const activeConfig = ref(0)

  const pageFilter = computed(():IPageFilter => {
    return {
      count: false,
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    }
  })

  watch(() => props.appId, (val) => {
    getConfigList()
  })

  onMounted(() => {
    getConfigList()
  })

  const getConfigList = async () => {
    loading.value = true
    try {
      const resp = await getServingConfigList(props.bkBizId, props.appId, { op: FilterOp.AND, rules: [] }, pageFilter.value)
      // @ts-ignore
      configList.value = resp.details
      pagination.value.count = 4
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const refreshConfigList = (current: number = 1) => {
    pagination.value.current = current
    getConfigList()
  }

  const handleEdit = (config: any) => {
    activeConfig.value = config.id
    editPanelShow.value = true
  }

  const handleDiff = () => {}

  const handleDel = (config: any) => {
    InfoBox({
      title: `确认是否删除配置项 ${config.spec.name}?`,
      type: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteServingConfigItem(config.id, config.attachment.biz_id, config.attachment.app_id)
        if (configList.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current -= 1
        }
        getConfigList();
      },
    } as any);
  }

  const handlePageLimitChange = (limit: number) => {
    pagination.value.limit = limit
    refreshConfigList()
  }

  const handlePageChange = (val: number) => {
    pagination.value.current = val
    getConfigList()
  }

  defineExpose({
    refreshConfigList
  })
</script>
<template>
  <section class="config-list-table">
    <bk-loading :loading="loading">
      <bk-table :border="['outer']" :data="configList">
        <bk-table-column label="配置项名称" prop="spec.name" :sort="true"></bk-table-column>
        <bk-table-column label="配置预览">-</bk-table-column>
        <bk-table-column label="配置格式" prop="spec.file_type"></bk-table-column>
        <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
        <bk-table-column label="修改人" prop="revision.reviser"></bk-table-column>
        <bk-table-column label="修改时间" prop="revision.update_at" :sort="true"></bk-table-column>
        <bk-table-column label="变更状态">-</bk-table-column>
        <bk-table-column label="操作">
          <template #default="{ row }">
            <div class="operate-action-btns">
              <bk-button text theme="primary" @click="handleEdit(row)">编辑</bk-button>
              <bk-button text theme="primary" :disabled="true" @click="handleDiff()">对比</bk-button>
              <bk-button text theme="primary" @click="handleDel(row)">删除</bk-button>
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
        @change="refreshConfigList($event)"
        @limit-change="handlePageLimitChange"/>
    </bk-loading>
    <edit-config
      v-model:show="editPanelShow"
      :config-id="activeConfig"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      @confirm="refreshConfigList" />
  </section>
</template>
<style lang="scss" scoped>
  .config-list-table {
    :deep(.bk-pagination) {
      padding-left: 15px;
    }
  }
  .operate-action-btns {
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
