<template>
  <bk-form class="bcs-small-form">
    <!-- 独立集群 -->
    <bk-form-item :label="$t('cluster.labels.clusterType')">
      <span class="text-[#313238]">
        {{ $t('bcs.cluster.selfDeployed') }}
      </span>
    </bk-form-item>
    <bk-form-item :label="$t('cluster.labels.masterInfo')">
      <MasterInfo :cluster-id="clusterId" />
    </bk-form-item>
    <bk-form-item label="Kube-apiserver">
      <KubeApiServer :enabled="enableHa" />
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';

import MasterInfo from './master-info.vue';

import { ICluster, useCluster } from '@/composables/use-app';
import KubeApiServer from '@/views/cluster-manage/add/components/kube-api-server.vue';

export default defineComponent({
  name: 'ClusterMaster',
  components: { KubeApiServer, MasterInfo },
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

    // Kube-apiserver
    const enableHa = computed(() => curCluster.value?.clusterAdvanceSettings?.enableHa);

    return {
      enableHa,
      curCluster,
    };
  },
});
</script>
