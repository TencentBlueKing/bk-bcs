<template>
  <bk-select
    searchable
    :clearable="false"
    :loading="subnetLoading"
    :value="value"
    @change="handleSubnetChange">
    <bk-option
      v-for="net in subnets"
      :key="net.subnetID"
      :id="net.subnetID"
      :name="`${net.subnetName}(${net.subnetID})`"
      :disabled="!net.availableIPAddressCount || !!Object.keys(net.cluster || {}).length">
      <div
        class="flex items-center justify-between"
        v-bk-tooltips="{
          content: Object.keys(net.cluster || {}).length
            ? $t('tke.tips.subnetInUsed', [net.cluster ? net.cluster.clusterName : ''])
            : $t('tke.tips.noAvailableIp'),
          disabled: net.availableIPAddressCount && !Object.keys(net.cluster || {}).length,
          placement: 'left'
        }">
        <span>{{ `${net.subnetName}(${net.subnetID})` }}</span>
        <span
          :class="(!net.availableIPAddressCount || Object.keys(net.cluster || {}).length) ? '':'text-[#979BA5]'">
          {{ `${$t('tke.label.availableIpNum')}: ${net.availableIPAddressCount}` }}
        </span>
      </div>
    </bk-option>
  </bk-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import { ISubnet } from '../../types/types';

import { cloudSubnets } from '@/api/modules/cluster-manager';

const props = defineProps({
  value: {
    type: String,
  },
  region: {
    type: String,
    default: '',
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  vpcId: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['input', 'change']);

// 子网
const subnets = ref<Array<ISubnet>>([]);
const subnetLoading = ref(false);
const handleGetSubnets = async () => {
  if (!props.cloudAccountID || !props.region || !props.vpcId) return;
  subnetLoading.value = true;
  subnets.value = await cloudSubnets({
    $cloudId: props.cloudID,
    region: props.region,
    accountID: props.cloudAccountID,
    vpcID: props.vpcId,
    injectCluster: true,
  }).catch(() => []);
  subnetLoading.value = false;
};

const handleSubnetChange = (subnetID: string) => {
  emits('change', subnetID);
  emits('input', subnetID);
};

watch([
  () => props.region,
  () => props.cloudAccountID,
  () => props.vpcId,
], () => {
  handleGetSubnets();
});

onBeforeMount(() => {
  handleGetSubnets();
});
</script>
