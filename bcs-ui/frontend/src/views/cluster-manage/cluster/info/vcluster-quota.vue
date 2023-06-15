<template>
  <bk-form class="bcs-small-form px-[60px] py-[24px]">
    <bk-form-item :label="$t('CPU 配额')">
      {{ quota.cpuLimits ? `${quota.cpuLimits}${$t('核')}` : '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('内存配额')">
      {{ quota.memoryLimits ? `${quota.memoryLimits}B` : '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('Pod IP  数量')">
      {{ networkSettings.maxNodePodNum || '--' }}
    </bk-form-item>
    <bk-form-item :label="$t('Service IP 数量')">
      {{ networkSettings.maxServiceNum || '--' }}
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts" setup>
import { computed } from 'vue';
import { useClusterList } from '../use-cluster';

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
    data = JSON.parse(curCluster.value?.extraInfo?.namespaceInfo || {})?.quota;
  } catch (_) {
    data = {
      cpuLimits: '',
      memoryLimits: '',
    };
  }
  return data;
});

</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
