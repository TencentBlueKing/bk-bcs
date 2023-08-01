<template>
  <bk-form class="bcs-small-form px-[60px] py-[24px]" :label-width="160" v-bkloading="{ isLoading }">
    <bk-form-item :label="$t('cluster.labels.region')">
      {{ clusterData.region || '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('cluster.labels.networkType')">{{ clusterData.networkType || '--' }}</bk-form-item>
    <bk-form-item :label="$t('cluster.labels.cidr')">{{ cidr }}</bk-form-item>
    <bk-form-item label="VPC">
      {{ clusterData.vpcID || '--' }}
    </bk-form-item>
    <bk-form-item label="IPVS">
      {{ IPVS }}
    </bk-form-item>
    <bk-form-item :label="$t('cluster.create.label.networkSetting.cidrStep.text')">{{ cidrStep }}</bk-form-item>
    <bk-form-item :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">{{ maxServiceNum }}</bk-form-item>
    <bk-form-item :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">{{ maxNodePodNum }}</bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted } from 'vue';
import { useClusterInfo } from '../use-cluster';

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
    const { clusterData, isLoading, getClusterDetail } = useClusterInfo();
    const cidr = computed(() => {
      const { multiClusterCIDR = [], clusterIPv4CIDR = '' } = clusterData.value.networkSettings || {};
      return [...multiClusterCIDR, clusterIPv4CIDR].filter(cidr => !!cidr).join(', ') || '--';
    });
    const IPVS = computed(() => clusterData.value?.clusterAdvanceSettings?.IPVS || '--');
    const maxServiceNum = computed(() => clusterData.value?.networkSettings?.maxServiceNum || '--');
    const maxNodePodNum = computed(() => clusterData.value?.networkSettings?.maxNodePodNum || '--');
    const cidrStep = computed(() => clusterData.value?.networkSettings?.cidrStep);

    onMounted(async () => {
      await getClusterDetail(props.clusterId);
    });

    return {
      isLoading,
      clusterData,
      cidr,
      IPVS,
      cidrStep,
      maxServiceNum,
      maxNodePodNum,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
