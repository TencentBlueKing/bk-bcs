<template>
  <BaseLayout
    title="PersistentVolumeClaims"
    kind="PersistentVolumeClaim" category="persistent_volume_claims" type="storages">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleSortChange,
        handleUpdateResource, handleDeleteResource,
        handleShowDetail, webAnnotations,
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
            <bk-button
              class="bcs-button-ellipsis" text
              @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
        <bk-table-column label="Status">
          <template #default="{ row }">
            <span>{{ row.status.phase || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Volume">
          <template #default="{ row }">
            <span>{{ row.spec.volumeName || '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Capacity">
          <template #default="{ row }">
            <span>{{ row.status.capacity ? row.status.capacity.storage : '--' }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Access Modes">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'accessModes').join(', ') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="StorageClass">
          <template #default="{ row }">
            <span>{{ row.spec.storageClassName || '--' }}</span>
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
            <bk-button text @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
            <bk-button
              class="ml10"
              text
              v-authority="{
                clickable: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.clickable : true,
                content: webAnnotations.perms.items[row.metadata.uid]
                  ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.tip : '',
                disablePerms: true
              }"
              @click="handleDeleteResource(row)">
              {{ $t('删除') }}
            </bk-button>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus :type="nameValue ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
        </template>
      </bk-table>
    </template>
    <template #detail="{ data, extData, clusterId }">
      <PvcDetail :data="data" :ext-data="extData" :cluster-id="clusterId"></PvcDetail>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';
import BaseLayout from '@/views/resource-view/common/base-layout';
import PvcDetail from './pvc-detail.vue';

export default defineComponent({
  components: { BaseLayout, PvcDetail },
});
</script>
