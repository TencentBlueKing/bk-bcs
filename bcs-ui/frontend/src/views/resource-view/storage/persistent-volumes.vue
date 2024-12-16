<template>
  <BaseLayout
    title="PersistentVolumes"
    kind="PersistentVolume" category="persistent_volumes" type="storages" :show-create="false">
    <template
      #default="{
        curPageData, pageConf, handleShowViewConfig, clusterNameMap, isClusterMode,
        handlePageChange, handlePageSizeChange, handleGetExtData, handleSortChange,
        sourceTypeMap
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" sortable></bk-table-column>
        <bk-table-column :label="$t('cluster.labels.nameAndId')" v-if="!isClusterMode">
          <template #default="{ row }">
            <div class="flex flex-col py-[6px] h-[50px]">
              <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
              <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column label="Capacity">
          <template #default="{ row }">
            <span>{{ row.spec.capacity.storage || 'null' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Access modes">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'accessModes').join(', ') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Reclaim policy">
          <template #default="{ row }">
            <span>{{ row.spec.persistentVolumeReclaimPolicy || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Status">
          <template #default="{ row }">
            <span>{{ row.status.phase || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Claim">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'claim') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="StorageClass">
          <template #default="{ row }">
            <span>{{ row.spec.storageClassName || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Reason">
          <template #default="{ row }">
            <span>{{ row.status.reason || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="VolumeMode">
          <template #default="{ row }">
            <span>{{ row.spec.volumeMode || '--' }}</span>
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
          <template slot-scope="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
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
import { defineComponent } from 'vue';

import sourceTableCell from '../common/source-table-cell.vue';

import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  components: { BaseLayout, sourceTableCell },
});
</script>
