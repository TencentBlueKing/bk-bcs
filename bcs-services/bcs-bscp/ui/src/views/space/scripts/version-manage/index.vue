<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import DetailLayout from '../components/detail-layout.vue'

  const emits = defineEmits(['update:show'])

  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  const refreshList = (val: number = 1) => {
    pagination.value.current = 1
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <DetailLayout name="版本管理" :show-footer="false" @close="handleClose">
    <template #content>
      <div class="script-version-manage">
        <div class="operation-area">
          <bk-button theme="primary"><Plus class="button-icon" />新建版本</bk-button>
          <bk-input class="search-input" placeholder="版本号/版本说明/更新人">
              <template #suffix>
                <Search class="search-input-icon" />
              </template>
          </bk-input>
        </div>
        <bk-table :border="['outer']">
          <bk-table-column label="版本号"></bk-table-column>
          <bk-table-column label="版本说明"></bk-table-column>
          <bk-table-column label="被引用"></bk-table-column>
          <bk-table-column label="更新人"></bk-table-column>
          <bk-table-column label="更新时间"></bk-table-column>
          <bk-table-column label="操作"></bk-table-column>
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