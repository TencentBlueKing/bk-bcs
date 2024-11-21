<template>
  <bk-select
    :value="value"
    searchable
    :clearable="false"
    :loading="regionLoading"
    :disabled="disabled"
    @change="handleRegionChange">
    <bk-option
      v-for="item in regionList"
      :key="item.region"
      :id="item.region"
      :name="item.regionName">
    </bk-option>
  </bk-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import { cloudRegionByAccount } from '@/api/modules/cluster-manager';
import $store from '@/store';
import { ICloudRegion } from '@/views/cluster-manage/types/types';

const props = defineProps({
  value: {
    type: String,
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  initData: {
    type: Boolean,
    default: false,
  },
  filter: {
    type: Function,
    default: null,
  },
});
const emits = defineEmits(['input', 'change']);
// 区域列表
const regionLoading = ref(false);
const regionList = ref<Array<ICloudRegion>>($store.state.cloudMetadata.regionList);
const handleGetRegionList = async () => {
  if (!props.cloudAccountID || !props.cloudID) return;

  regionLoading.value = true;
  const data = await cloudRegionByAccount({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
  }).catch(() => []);
  $store.commit('cloudMetadata/updateRegionList', data);
  regionList.value = data.filter(item => item.regionState === 'AVAILABLE');
  if (props.filter) {
    regionList.value = props.filter(regionList.value, handleGetRegionList);
  }
  regionLoading.value = false;
};

const handleRegionChange = (region: string) => {
  emits('change', region);
  emits('input', region);
};


watch(() => props.cloudAccountID, () => {
  handleGetRegionList();
});

onBeforeMount(() => {
  props.initData && handleGetRegionList();
});
</script>
