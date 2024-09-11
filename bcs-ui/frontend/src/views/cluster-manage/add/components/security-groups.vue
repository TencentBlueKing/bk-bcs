<template>
  <bk-select
    searchable
    :multiple="multiple"
    :disabled="disabled"
    :clearable="clearable"
    :value="value"
    :loading="securityGroupLoading"
    @change="handleSecurityGroupsChange">
    <bk-option
      v-for="item in securityGroups"
      :key="item.securityGroupID"
      :id="item.securityGroupID"
      :name="`${item.securityGroupName}(${item.securityGroupID})`">
    </bk-option>
    <SelectExtension
      slot="extension"
      :link-text="$t('tke.link.securityGroup')"
      link="https://console.cloud.tencent.com/vpc/security-group"
      @refresh="handleGetSecurityGroups" />
  </bk-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import { cloudSecurityGroups } from '@/api/modules/cluster-manager';
import SelectExtension from '@/components/select-extension.vue';
import $store from '@/store';
import { ISecurityGroup } from '@/views/cluster-manage/types/types';;

const props = defineProps({
  value: {
    type: [String, Array],
    default: '',
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
  multiple: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  clearable: {
    type: Boolean,
    default: false,
  },
  initData: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['input', 'change']);

// 安全组
const securityGroupLoading = ref(false);
const securityGroups = ref<Array<ISecurityGroup>>($store.state.cloudMetadata.securityGroupsList);
const handleGetSecurityGroups = async () => {
  if (!props.region || !props.cloudAccountID || !props.cloudID) return;
  securityGroupLoading.value = true;
  securityGroups.value = await cloudSecurityGroups({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  $store.commit('cloudMetadata/updateSecurityGroupsList', securityGroups.value);
  securityGroupLoading.value = false;
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetSecurityGroups();
});

const handleSecurityGroupsChange = (value) => {
  emits('input', value);
  emits('change', value);
};

onBeforeMount(() => {
  props.initData && handleGetSecurityGroups();
});
</script>
