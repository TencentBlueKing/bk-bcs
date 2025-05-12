<template>
  <BaseLayout
    title="HookTemplates"
    kind="HookTemplate"
    type="crd"
    category="custom_objects"
    :crd="crd"
    default-active-detail-type="yaml"
    :show-detail-tab="false"
    scope="Namespaced">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleUpdateResource,
        handleDeleteResource,handleSortChange,
        handleShowDetail, renderCrdHeader,
        getJsonPathValue, additionalColumns,
        webAnnotations, handleShowViewConfig,
        clusterNameMap, goNamespace, isViewEditable,
        isClusterMode, sourceTypeMap
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        ref="tableRef"
        v-bk-column-memory="{
          instance: tableRef
        }"
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
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" min-width="100" sortable>
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
        <bk-table-column
          v-for="item in additionalColumns"
          :key="item.name"
          :label="item.name"
          :prop="item.jsonPath"
          :render-header="renderCrdHeader">
          <template #default="{ row }">
            <span>
              {{ typeof getJsonPathValue(row, item.jsonPath) !== 'undefined'
                ? getJsonPathValue(row, item.jsonPath) : '--' }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.source')" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <sourceTableCell
              :row="row"
              :source-type-map="sourceTypeMap" />
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template #default="{ row }">
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
              v-authority="{
                clickable: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.clickable : true,
                content: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.tip : '',
                disablePerms: true
              }"
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
  </BaseLayout>
</template>
<script>
import { defineComponent, ref } from 'vue';

import sourceTableCell from '../common/source-table-cell.vue';

import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'HookTemplates',
  components: { BaseLayout, sourceTableCell },
  props: {
    crd: {
      type: String,
      default: 'hooktemplates.tkex.tencent.com',
    },
  },
  setup() {
    const tableRef = ref(null);
    return {
      tableRef,
    };
  },
});
</script>
