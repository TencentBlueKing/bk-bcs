<template>
  <BaseLayout title="Services" kind="Service" category="services" type="networks">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleShowDetail,
        handleSortChange,handleUpdateResource, isClusterMode,
        handleDeleteResource, handleShowViewConfig,clusterNameMap, goNamespace, isViewEditable
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
        <bk-table-column label="Type" :resizable="false">
          <template #default="{ row }">
            <span>{{ row.spec.type || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="ClusterIPv4" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'clusterIPv4') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="ClusterIPv6" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'clusterIPv6') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="External-ip" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'externalIP', []).join(', ') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Ports" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'ports', []).join(', ') || '*' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
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
      <ServiceDetail :data="data" :ext-data="extData" :cluster-id="extData.clusterID"></ServiceDetail>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';

import ServiceDetail from './service-detail.vue';

import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'NetworkService',
  components: { BaseLayout, ServiceDetail },
});
</script>
