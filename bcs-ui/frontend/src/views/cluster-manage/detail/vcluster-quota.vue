<template>
  <bk-form class="bcs-small-form px-[60px] py-[24px]">
    <bk-form-item :label="$t('cluster.labels.cpuQuota')">
      {{ quota.cpuLimits ? `${quota.cpuLimits}${$t('units.suffix.cores')}` : '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('cluster.labels.memQuota')">
      {{ quota.memoryLimits ? `${quota.memoryLimits}B` : '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('cluster.labels.maxNodePodNum')">
      {{ networkSettings.maxNodePodNum || '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('cluster.create.label.networkSetting1.maxServiceNum')">
      {{ networkSettings.maxServiceNum || '--' }}
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts" setup>
import { computed } from 'vue';

import { useClusterList } from '../cluster/use-cluster';

const props = defineProps({
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
});

const {  clusterList } = useClusterList();
const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId));
const networkSettings = computed(() => curCluster.value?.networkSettings || {
  maxNodePodNum: '',
  maxServiceNum: '',
});
const quota = computed<{
  cpuLimits: string
  memoryLimits: string
}>(() => {
  let data = {
    cpuLimits: '',
    memoryLimits: '',
  };
  try {
    data = JSON.parse(curCluster.value?.extraInfo?.namespaceInfo || {})?.quota || {};
  } catch (_) {
    data = {
      cpuLimits: '',
      memoryLimits: '',
    };
  }
  return data;
});

</script>
