<template>
  <bk-radio-group v-model="clusterConnectSetting.isExtranet">
    <div class="flex items-center mb-[8px] w-full">
      <bk-radio :value="true">
        {{ $t('tke.label.apiServerCLB.internet') }}
      </bk-radio>
      <div class="flex items-center flex-1 ml-[16px]">
        <span
          class="prefix"
          v-bk-tooltips="$t('tke.label.securityGroup.desc')">
          <span class="bcs-border-tips">{{ $t('tke.label.securityGroup.text') }}</span>
        </span>
        <SecurityGroups
          :class="[
            'ml-[-1px] flex-1 w-0',
            clusterConnectSetting.isExtranet ? 'bg-[#fff]' : ''
          ]"
          :disabled="!clusterConnectSetting.isExtranet"
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          init-data
          v-model="clusterConnectSetting.securityGroup" />
      </div>
    </div>
    <bk-radio :value="false">
      {{ $t('tke.label.apiServerCLB.intranet') }}
    </bk-radio>
  </bk-radio-group>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import SecurityGroups from './security-groups.vue';

const props = defineProps({
  value: {
    type: Object,
    default: () => ({}),
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
});
const emits = defineEmits(['change']);

const clusterConnectSetting = ref({
  isExtranet: true,
  securityGroup: '',
});

watch(() => props.value, () => {
  if (JSON.stringify(props.value) === JSON.stringify(clusterConnectSetting.value)) return;

  clusterConnectSetting.value = JSON.parse(JSON.stringify(props.value));
}, { immediate: true });

watch(clusterConnectSetting, () => {
  emits('change', clusterConnectSetting.value);
}, { deep: true });
</script>
