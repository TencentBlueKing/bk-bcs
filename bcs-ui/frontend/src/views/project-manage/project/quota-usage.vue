<template>
  <div class="p-[20px]">
    <bk-table :data="data.nodeGroups || []" :outer-border="false">
      <bk-table-column :label="$t('generic.label.cluster')" prop="clusterId">
        <template #default="{ row }">
          <bk-button :disabled="!projectCode" text @click="handleCluster(row.clusterId)">
            {{ row.clusterId }}
          </bk-button>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('cluster.ca.nodePool.label.ID')" prop="nodeGroupId" />
      <bk-table-column :label="$t('cluster.ca.nodePool.label.nodeQuota')" prop="quotaUsed" />
      <bk-table-column :label="$t('cluster.ca.nodePool.label.nodeCounts')" prop="quotaNum" />
    </bk-table>
  </div>
</template>
<script lang="ts" setup>
import { computed } from 'vue';

import $store from '@/store';

defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
});

const projectCode = computed(() => $store.getters.curProjectCode);

function handleCluster(clusterId: string) {
  const url = new URL(`/projects/${projectCode.value}/clusters/${clusterId}/workloads/deployments`, window.location.origin);
  window.open(url.toString(), '_blank');
}
</script>
