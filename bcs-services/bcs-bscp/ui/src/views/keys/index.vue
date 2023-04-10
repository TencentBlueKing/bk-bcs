<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'

  const keys = ref([])
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  const refreshList = () => {}

  const handlePageLimitChange = () => {}

</script>
<template>
  <section class="keys-management-page">
    <div class="operate-area">
      <bk-button theme="primary"><Plus class="button-icon" />新建密钥</bk-button>
      <div class="filter-actions">
        <bk-input class="search-group-input" placeholder="状态/说明/更新人/更新时间">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
    </div>
    <bk-table :border="['outer']">
      <bk-table-column label="密钥"></bk-table-column>
      <bk-table-column label="说明"></bk-table-column>
      <bk-table-column label="更新人"></bk-table-column>
      <bk-table-column label="更新时间"></bk-table-column>
      <bk-table-column label="状态"></bk-table-column>
      <bk-table-column label="操作"></bk-table-column>
    </bk-table>
    <bk-pagination
      class="table-list-pagination"
      v-model="pagination.current"
      location="left"
      :layout="['total', 'limit', 'list']"
      :count="pagination.count"
      :limit="pagination.limit"
      @change="refreshList()"
      @limit-change="handlePageLimitChange"/>
  </section>
</template>
<style lang="scss" scoped>
  .keys-management-page {
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
  .search-group-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
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
