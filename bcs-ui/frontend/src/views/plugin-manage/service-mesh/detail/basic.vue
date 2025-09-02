<template>
  <div>
    <div class="font-bold mb-[16px] text-[14px]">{{ $t('serviceMesh.label.info') }}</div>
    <div class="grid grid-cols-2 gap-4">
      <FieldItem :label="$t('serviceMesh.label.name')" :value="data?.name" />
      <FieldItem :label="$t('serviceMesh.label.ID')" :value="data?.meshID" />
      <FieldItem :label="$t('serviceMesh.label.version')" :value="data?.version" />
      <FieldItem :label="$t('serviceMesh.label.networkID')" :value="data?.networkID" />
      <FieldItem label="revision" :value="data?.revision" />
    </div>
    <FieldItem class="mt-[16px]" :label="$t('serviceMesh.label.description')" :value="data?.description" />
    <bcs-divider class="!mt-[24px] !mb-[8px]"></bcs-divider>
    <div class="font-bold mb-[16px] text-[14px]">{{ $t('serviceMesh.label.clusterConfig') }}</div>
    <FieldItem :label="$t('serviceMesh.label.primaryClusters')" :value="data?.primaryClusters?.join?.()" />
    <!-- 多集群 -->
    <ContentSwitcher
      v-if="data?.multiClusterEnabled"
      class="mt-[24px] text-[12px]"
      :label="$t('serviceMesh.label.multiCluster')"
      readonly>
      <div class="grid grid-cols-2">
        <FieldItem :label="$t('serviceMesh.label.differentNetwork')">
          <span v-if="data?.differentNetwork">{{ $t('serviceMesh.label.connected') }}</span>
          <span v-else>{{ $t('serviceMesh.label.notConnected') }}</span>
        </FieldItem>
        <FieldItem :label="$t('serviceMesh.label.clbID')" :value="data?.clbID" />
      </div>
      <p class="text-[#979BA5] mt-[14px]">{{ $t('serviceMesh.label.subClusters') }}</p>
      <bcs-table
        class="mt-[8px]"
        custom-header-color="#F5F7FA"
        empty-block-class-name="border-b border-[#DCDEE5]"
        cell-class-name="!h-[36px]"
        header-cell-class-name="!h-[36px]"
        :outer-border="false"
        :data="data?.remoteClusters || []">
        <bcs-table-column :label="$t('generic.label.cluster')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ `${row.clusterName || '--'}(${row.clusterID || '--'})` }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.clusterRegion')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.region || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.status')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.status || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column width="160" :label="$t('serviceMesh.label.joinTime')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.joinTime ? timeFormat(Number(row.joinTime)) : '--' }}
          </template>
        </bcs-table-column>
        <template #empty>
          {{ $t('serviceMesh.label.tableEmpty') }}
        </template>
      </bcs-table>
    </ContentSwitcher>
    <div v-else class="bg-[#F0F1F5] px-[8px] py-[4px] rounded-sm flex items-center mt-[24px] text-[12px]">
      <span>{{ $t('serviceMesh.label.multiCluster') }}</span>
      <span class="ml-[16px] text-[#979BA5]">{{ $t('generic.status.notEnable') }}</span>
    </div>
  </div>
</template>
<script setup lang="ts">
import ContentSwitcher from '../content-switcher.vue';

import FieldItem from './field-item.vue';

import { timeFormat } from '@/common/util';

defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
});
</script>
<style lang="postcss" scoped>
:deep(.bk-table-empty-text) {
  padding: 0px;
}
:deep(.bk-table th>.cell) {
  height: 36px;
}
</style>
