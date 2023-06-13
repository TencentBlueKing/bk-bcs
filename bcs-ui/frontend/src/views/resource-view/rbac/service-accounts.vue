<template>
  <BaseLayout title="ServiceAccounts" kind="ServiceAccount" category="service_accounts" type="rbac">
    <template
      #default="{ curPageData, pageConf, nameValue, handleClearSearchData,
                  handlePageChange, handlePageSizeChange, handleGetExtData,
                  handleShowDetail, handleSortChange,handleUpdateResource,handleDeleteResource }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('名称')" prop="metadata.name" sortable>
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis" text
              @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
        <bk-table-column label="Secrets">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'secrets') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('操作')" :resizable="false" width="150">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleDeleteResource(row)">{{ $t('删除') }}</bk-button>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus :type="nameValue ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
        </template>
      </bk-table>
    </template>
    <template #detail="{ data, extData }">
      <ServiceAccountsDetail :data="data" :ext-data="extData"></ServiceAccountsDetail>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';
import ServiceAccountsDetail from './service-accounts-detail.vue';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  components: { BaseLayout, ServiceAccountsDetail },
});
</script>
