<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'

  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  const refreshList = () => {}

  const handlePageLimitChange = () => {}
</script>
<template>
  <section class="scripts-manange-page">
    <div class="side-menu">
      <div class="group-wrapper">
        <li><i class="bk-bscp-icon icon-app-store"></i>全部脚本</li>
        <li><i class="bk-bscp-icon icon-app-store"></i>未分类</li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <bk-button theme="primary"><Plus class="button-icon" />新建脚本</bk-button>
        <bk-input class="search-group-input" placeholder="脚本名称">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
      <bk-table :border="['outer']">
        <bk-table-column label="脚本名称"></bk-table-column>
        <bk-table-column label="脚本语言"></bk-table-column>
        <bk-table-column label="分类标签"></bk-table-column>
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
        @change="refreshList()"
        @limit-change="handlePageLimitChange"/>
    </div>
  </section>
</template>
<style lang="scss" scoped>
  .scripts-manange-page {
    display: flex;
    align-items: center;
    height: 100%;
    background: #ffffff;
  }
  .side-menu {
    padding: 16px 0;
    width: 280px;
    height: 100%;
    background: #f5f7fa;
    box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.15);
    z-index: 1;
    .group-wrapper {
      padding-bottom: 16px;
      border-bottom: 1px solid #dcdee5;
      & > li {
        padding: 8px 22px;
        color: #313238;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          color: #348aff;
        }
        i {
          margin-right: 8px;
          font-size: 14px;
          color: #979ba5;
        }
      }
    }
  }
  .script-list-wrapper {
    padding: 24px;
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
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
