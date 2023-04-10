<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'

  const groupList = ref([])
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  const refreshList = () => {}

  const handlePageLimitChange = () => {}

</script>
<template>
  <section class="groups-management-page">
    <div class="operate-area">
      <bk-button theme="primary"><Plus class="button-icon" />新增分组</bk-button>
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
        <bk-table :border="['outer']">
          <bk-table-column label="分组名称" :width="210"></bk-table-column>
          <bk-table-column label="分组规则"></bk-table-column>
          <bk-table-column label="服务可见范围" :width="240"></bk-table-column>
          <bk-table-column label="上线服务数" :width="110"></bk-table-column>
          <bk-table-column label="操作" :width="120"></bk-table-column>
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
      </template>
      <bk-exception v-else class="group-data-empty" type="empty" scene="part">
        当前暂无数据
        <div class="create-group-text-btn">
          <bk-button text theme="primary">立即创建</bk-button>
        </div>
      </bk-exception>
    </div>
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
