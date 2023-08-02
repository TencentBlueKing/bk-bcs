<template>
  <BaseLayout title="DaemonSets" kind="DaemonSet" category="daemonsets" type="workloads">
    <template
      #default="{
        curPageData,
        pageConf,
        handlePageChange,
        handlePageSizeChange,
        handleGetExtData,
        gotoDetail,
        handleSortChange,
        handleUpdateResource,
        handleDeleteResource,
        nameValue, handleClearSearchData
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" min-width="100" sortable>
          <template #default="{ row }">
            <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" min-width="100" sortable></bk-table-column>
        <bk-table-column :label="$t('k8s.image')" min-width="280" :show-overflow-tooltip="false">
          <template slot-scope="{ row }">
            <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
              {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="Desired" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.desiredNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Current" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.currentNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Ready" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.numberReady || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Up-to-date" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.updatedNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Available" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.numberAvailable || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :resizable="false">
          <template #default="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.createdBy')">
          <template slot-scope="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'creator') || '--'}}</span>
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
        <bk-table-column :label="$t('generic.label.action')" :resizable="false" width="150">
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
          <BcsEmptyTableStatus :type="nameValue ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
        </template>
      </bk-table>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'WorkloadDaemonsets',
  components: { BaseLayout },
});
</script>
