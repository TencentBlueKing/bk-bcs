<template>
  <bk-form class="bcs-small-form">
    <!-- 独立集群 -->
    <bk-form-item :label="$t('cluster.labels.clusterType')">
      <span class="text-[#313238]">
        {{ $t('bcs.cluster.selfDeployed') }}
      </span>
    </bk-form-item>
    <bk-form-item :label="$t('cluster.labels.masterInfo')">
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
    </bk-form-item>
    <bk-form-item label="Kube-apiserver">
      <KubeApiServer :enabled="enableHa" />
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import { masterList } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import { useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import KubeApiServer from '@/views/cluster-manage/add/components/kube-api-server.vue';

export default defineComponent({
  name: 'ClusterMaster',
  components: { KubeApiServer },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId));

    // Kube-apiserver
    const enableHa = computed(() => curCluster.value?.clusterAdvanceSettings?.enableHa);

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

    // 复制IP
    const handleCopyIPv4 = () => {
      copyText(masterData.value.map(item => item.innerIP).join('\n'));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };

    onBeforeMount(() => {
      if (Object.keys(curCluster.value?.master || {}).length) {
        handleGetMasterData();
      }
    });

    return {
      enableHa,
      curCluster,
      isLoading,
      masterData,
      handleCopyIPv4,
    };
  },
});
</script>
