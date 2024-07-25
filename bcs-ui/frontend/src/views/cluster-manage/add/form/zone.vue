<template>
  <bcs-select
    searchable
    :clearable="false"
    :value="value"
    @change="handleZoneChange">
    <bcs-option
      v-for="zone in zoneList"
      :key="zone.zoneID"
      :id="zone.zone"
      :name="zone.zoneName"
      :disabled="disabledList.includes(zone.zone)"
      v-bk-tooltips="{
        content: disabledTips,
        disabled: !disabledList.includes(zone.zone),
        placement: 'left'
      }">
      <div class="flex items-center">
        <span class="flex-1 bcs-ellipsis">{{ zone.zoneName }}</span>
        <span v-if="vpcId" class="text-[#979BA5]">
          {{ $t('tke.tips.subnetNum', [zone.subnetNum || 0]) }}
        </span>
      </div>
    </bcs-option>
  </bcs-select>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';

import { IZoneItem } from '../tencent/types';

import { cloudsZones } from '@/api/modules/cluster-manager';
import $store from '@/store';

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
  disabledZoneList: {
    type: Array,
    default: () => [],
  },
  enabledZoneList: {
    type: Array,
    default: () => [],
  },
  disabledTips: {
    type: String,
    default: '',
  },
  initData: {
    type: Boolean,
    default: false,
  },
  vpcId: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['input', 'change']);

const disabledList = computed(() => {
  const parseDisabledList = props.enabledZoneList?.length
    ? zoneList.value.filter(item => !props.enabledZoneList?.includes((item.zone))).map(item => item.zone)
    : [];
  return [
    ...props.disabledZoneList,
    ...parseDisabledList,
  ];
});
// 可用区
const zoneList = ref<Array<IZoneItem>>($store.state.cloudMetadata.zoneList);
const zoneLoading = ref(false);
const handleGetZoneList = async () => {
  if (!props.region || !props.cloudID) return;
  zoneLoading.value = true;
  const data = await cloudsZones({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
    vpcId: props.vpcId,
  }).catch(() => []);
  zoneList.value = data.filter(item => item.zoneState === 'AVAILABLE');
  $store.commit('cloudMetadata/updateZoneList', zoneList.value);
  zoneLoading.value = false;
};

const handleZoneChange = (zone: string) => {
  emits('change', zone);
  emits('input', zone);
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetZoneList();
});

onBeforeMount(() => {
  props.initData && handleGetZoneList();
});
</script>
