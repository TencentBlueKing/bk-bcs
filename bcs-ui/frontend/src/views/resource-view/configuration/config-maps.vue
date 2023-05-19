<template>
  <BaseLayout title="ConfigMaps" kind="ConfigMap" category="configmaps" type="configs">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleSortChange,
        handleShowDetail, handleUpdateResource,
        handleDeleteResource, webAnnotations,
        nameValue, handleClearSearchData
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('名称')" prop="metadata.name" sortable>
          <template #default="{ row }">
            <bk-button class="bcs-button-ellipsis" text @click="handleShowDetail(row)">
              {{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
        <bk-table-column label="Data">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'data').join(', ') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('编辑模式')" width="100">
          <template slot-scope="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('表单') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('操作')" :resizable="false" width="150">
          <template #default="{ row }">
            <bk-button
              text
              v-authority="{
                clickable: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].updateBtn.clickable : true,
                content: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].updateBtn.tip : '',
                disablePerms: true
              }"
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
      <ConfigMapsDetail :data="data" :ext-data="extData"></ConfigMapsDetail>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';
import ConfigMapsDetail from './config-maps-detail.vue';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  components: { BaseLayout, ConfigMapsDetail },
});
</script>
