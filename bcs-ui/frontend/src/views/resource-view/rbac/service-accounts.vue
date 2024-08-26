<template>
  <BaseLayout title="ServiceAccounts" kind="ServiceAccount" category="service_accounts" type="rbac">
    <template
      #default="{
        curPageData, pageConf, handleShowViewConfig,
        handlePageChange, handlePageSizeChange, handleGetExtData,
        handleShowDetail, handleSortChange,handleUpdateResource,handleDeleteResource,
        clusterNameMap, goNamespace, isViewEditable, isClusterMode
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" sortable fixed="left">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable"
              @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('cluster.labels.nameAndId')" v-if="!isClusterMode">
          <template #default="{ row }">
            <div class="flex flex-col py-[6px] h-[50px]">
              <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
              <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" sortable>
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable"
              @click="goNamespace(row)">
              {{ row.metadata.namespace }}
            </bk-button>
          </template>
        </bk-table-column>
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
        <bk-table-column :label="$t('generic.label.source')" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
              {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
                handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
            </span>
            <span v-else-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Helm'">
              {{ handleGetExtData(row.metadata.uid, 'chart')
                ?`${handleGetExtData(row.metadata.uid, 'chart') || '--'}`
                : 'Helm' }}
            </span>
            <span v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template slot-scope="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.action')"
          :resizable="false"
          width="150"
          fixed="right"
          v-if="!isViewEditable">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleDeleteResource(row)">{{ $t('generic.button.delete') }}</bk-button>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus
            :button-text="$t('generic.button.resetSearch')"
            type="search-empty"
            @clear="handleShowViewConfig" />
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
