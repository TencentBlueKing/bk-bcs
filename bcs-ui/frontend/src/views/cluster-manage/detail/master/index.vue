<template>
  <component :is="masterProviderMap[provider]" :cluster-id="clusterId" v-if="provider"></component>
</template>
<script lang="ts">
import { Component, computed, defineComponent } from 'vue';

import AwsMaster from './master-aws.vue';
import AzureMaster from './master-azure.vue';
import BluekingMaster from './master-blueking.vue';
import GcpMaster from './master-gcp.vue';
import HuaweiMaster from './master-huawei.vue';
import TkeMaster from './master-tke.vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'ClusterMaster',
  components: {
    TkeMaster,
    GcpMaster,
    BluekingMaster,
    AzureMaster,
    AwsMaster,
    HuaweiMaster,
  },
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
    const provider = computed(() => curCluster.value?.provider);

    const masterProviderMap: Record<CloudID, Component> = {
      tencentCloud: TkeMaster,
      tencentPublicCloud: TkeMaster,
      gcpCloud: GcpMaster,
      bluekingCloud: BluekingMaster,
      azureCloud: AzureMaster,
      huaweiCloud: HuaweiMaster,
      awsCloud: AwsMaster,
    };
    return {
      provider,
      masterProviderMap,
    };
  },
});
</script>
