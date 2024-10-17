<template>
  <component :is="networkProviderMap[provider]" :cluster-id="clusterId" v-if="provider"></component>
</template>
<script lang="ts">
import { Component, computed, defineComponent } from 'vue';

import awsNet from './aws.vue';
import azureNet from './azure.vue';
import bluekingNet from './blueking.vue';
import gcpNet from './gcp.vue';
import tkeNet from './tke.vue';
import tkePublicNet from './tke-public.vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'ClusterNetwork',
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

    const networkProviderMap: Record<CloudID, Component> = {
      tencentCloud: tkeNet,
      tencentPublicCloud: tkePublicNet,
      gcpCloud: gcpNet,
      bluekingCloud: bluekingNet,
      azureCloud: azureNet,
      huaweiCloud: bluekingNet,
      awsCloud: awsNet,
    };
    return {
      provider,
      networkProviderMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .bk-table .cell {
  padding-left: 15px !important;
}
</style>
