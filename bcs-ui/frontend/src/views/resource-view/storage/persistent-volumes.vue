<template>
  <BaseLayout
    title="PersistentVolumes"
    kind="PersistentVolume" category="persistent_volumes" type="storages" :show-name-space="false" :show-create="false">
    <template
      #default="{ curPageData, pageConf, nameValue, handleClearSearchData,
                  handlePageChange, handlePageSizeChange, handleGetExtData, handleSortChange }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('名称')" prop="metadata.name" sortable></bk-table-column>
        <bk-table-column label="Capacity">
          <template #default="{ row }">
            <span>{{ row.spec.capacity.storage || 'null' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Access Modes">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'accessModes').join(', ') || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Reclaim Policy">
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
        <!-- <bk-table-column :label="$t('操作')" :resizable="false" width="150">
                    <template #default="{ row }">
                        <bk-button text @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
                        <bk-button class="ml10" text @click="handleDeleteResource(row)">{{ $t('删除') }}</bk-button>
                    </template>
                </bk-table-column> -->
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
  components: { BaseLayout },
});
</script>
