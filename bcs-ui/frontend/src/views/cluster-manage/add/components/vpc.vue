<template>
  <bk-select
    :value="value"
    :loading="vpcLoading"
    searchable
    :clearable="false"
    :disabled="disabled"
    @change="handleVpcChange">
    <!-- VPC可用容器网络IP数量最低限制 -->
    <bk-option
      v-for="item in vpcList"
      :key="item.vpcId"
      :id="item.vpcId"
      :name="`${item.name}(${item.vpcId})`">
      <div class="flex items-center place-content-between">
        <span>
          {{item.name}}
          <span class="vpc-id">{{`(${item.vpcId})`}}</span>
        </span>
      </div>
    </bk-option>
    <SelectExtension
      slot="extension"
      :link-text="$t('tke.link.vpc')"
      link="https://console.cloud.tencent.com/vpc/vpc"
      @refresh="handleGetVPCList" />
  </bk-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import { cloudVPC } from '@/api/modules/cluster-manager';
import SelectExtension from '@/components/select-extension.vue';
import $store from '@/store';
import { IVpcItem } from '@/views/cluster-manage/types/types';

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
  disabled: {
    type: Boolean,
    default: false,
  },
  initData: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['input', 'change']);

// vpc列表
const vpcLoading = ref(false);
const vpcList = ref<Array<IVpcItem>>($store.state.cloudMetadata.vpcList);
const handleGetVPCList = async () => {
  if (!props.region || !props.cloudAccountID || !props.cloudID) return;
  vpcLoading.value = true;
  vpcList.value = await cloudVPC({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  $store.commit('cloudMetadata/updateVpcList', vpcList.value);
  vpcLoading.value = false;
};
const handleVpcChange = (vpc: string) => {
  emits('change', vpc);
  emits('input', vpc);
};


watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetVPCList();
});

onBeforeMount(() => {
  props.initData &&  handleGetVPCList();
});
</script>
