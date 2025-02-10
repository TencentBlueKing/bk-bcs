<template>
  <div class="flex max-w-[800px]">
    <bk-table :data="masterData" v-bkloading="{ isLoading }">
      <bk-table-column :label="$t('cluster.labels.hostName')">
        <template #default="{ row }">
          {{ row.nodeName || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column label="IPv4">
        <template #default="{ row }">
          {{ row.innerIP || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column label="IPv6">
        <template #default="{ row }">
          {{ row.innerIPv6 || '--' }}
        </template>
      </bk-table-column>
      <template v-if="$INTERNAL">
        <bk-table-column :label="$t('generic.ipSelector.label.idc')" prop="idc"></bk-table-column>
        <bk-table-column :label="$t('cluster.labels.rack')" prop="rack"></bk-table-column>
        <bk-table-column
          :label="$t('generic.ipSelector.label.serverModel')"
          prop="deviceClass">
        </bk-table-column>
      </template>
    </bk-table>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import { masterList } from '@/api/modules/cluster-manager';
import { ICluster, useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'MasterInfo',
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useCluster();
    const curCluster = computed<Partial<ICluster>>(() => clusterList.value
      .find(item => item.clusterID === props.clusterId) || {});

    // 独立集群Master信息
    const isLoading = ref(false);
    const masterData = ref<any[]>([]);
    const handleGetMasterData = async () => {
      isLoading.value = true;
      masterData.value = await masterList({
        $clusterId: props.clusterId,
      }).catch(() => []);
      isLoading.value = false;
    };

    onBeforeMount(() => {
      if (Object.keys(curCluster.value.master || {}).length) {
        handleGetMasterData();
      }
    });

    return {
      isLoading,
      masterData,
    };
  },
});
</script>
