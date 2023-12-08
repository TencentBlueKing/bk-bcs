<template>
  <div>
    <div class="flex flex-col items-center justify-center">
      <i
        :class="[
          'bk-icon icon-exclamation',
          'flex items-center justify-center rounded-full',
          ' text-[26px] bg-[#ffe8c3] text-[#ff9c01] leading-[42px] w-[42px] h-[42px]'
        ]">
      </i>
      <div class="text-[#313238] text-[20px] mt-[20px] leading-[32px]">
        {{ $t('tke.title.connectFailure.text') }}
      </div>
    </div>
    <div class="leading-[22px] mt-[8px]">
      {{ $t('tke.title.connectFailure.reason') }}
    </div>
    <div class="leading-[22px] mt-[2px]">
      <i18n path="tke.title.connectFailure.ref">
        <bk-link theme="primary">
          {{ $t('tke.title.connectFailure.refButton') }}
        </bk-link>
      </i18n>
    </div>
    <div class="bg-[#F5F7FA] p-[16px] mt-[16px]">
      <bk-radio-group v-model="isExtranet">
        <div class="flex items-center mb-[8px] w-full max-w-[500px]">
          <bk-radio :value="true">
            {{ $t('tke.label.apiServerCLB.internet') }}
          </bk-radio>
          <div class="flex items-center flex-1 ml-[16px]">
            <span
              class="prefix"
              v-bk-tooltips="$t('tke.label.securityGroup.desc')">
              <span class="bcs-border-tips">{{ $t('tke.label.securityGroup.text') }}</span>
            </span>
            <bcs-select
              class="ml-[-1px] flex-1 bg-[#fff]"
              searchable
              :clearable="false"
              :loading="securityGroupLoading"
              :disabled="!isExtranet"
              v-model="securityGroup">
              <bcs-option
                v-for="item in securityGroupList"
                :key="item.securityGroupID"
                :id="item.securityGroupID"
                :name="item.securityGroupName">
              </bcs-option>
            </bcs-select>
          </div>
        </div>
        <bk-radio :value="false">
          {{ $t('tke.label.apiServerCLB.intranet') }}
        </bk-radio>
      </bk-radio-group>
    </div>
    <div class="flex items-center justify-center mt-[24px]">
      <bk-button
        theme="primary"
        class="min-w-[88px]"
        :loading="isLoading"
        @click="handleConfirm">{{ $t('generic.button.confirm') }}</bk-button>
      <bk-button class="min-w-[88px]" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { merge } from 'lodash';
import { PropType, ref, watch } from 'vue';

import { ISecurityGroup } from '../add/tencent/types';

import { cloudSecurityGroups, modifyCluster } from '@/api/modules/cluster-manager';
import { ICluster } from '@/composables/use-app';

const props = defineProps({
  cluster: {
    type: Object as PropType<ICluster>,
    default: () => ({}),
  },
});
const emits = defineEmits(['cancel', 'confirm']);

const isExtranet = ref(true);
const securityGroup = ref('');

// 安全组
const securityGroupLoading = ref(false);
const securityGroupList = ref<Array<ISecurityGroup>>([]);
const handleGetSecurityGroups = async () => {
  const { provider, cloudAccountID, region } = props.cluster || {};
  if (!provider || !cloudAccountID || !region) return;
  securityGroupLoading.value = true;
  securityGroupList.value = await cloudSecurityGroups({
    $cloudId: provider,
    accountID: cloudAccountID,
    region,
  }).catch(() => []);
  securityGroupLoading.value = false;
};

const isLoading = ref(false);
const handleConfirm = async () => {
  isLoading.value = true;
  const result = await modifyCluster({
    $clusterId: props.cluster.clusterID,
    clusterAdvanceSettings: merge(
      props.cluster.clusterAdvanceSettings,
      {
        clusterConnectSetting: {
          isExtranet: isExtranet.value,
          securityGroup: securityGroup.value,
        },
      },
    ),
  }).then(() => true)
    .catch(() => false);

  isLoading.value = false;
  if (result) {
    emits('confirm', props.cluster);
    emits('cancel');
  }
};
const handleCancel = () => {
  emits('cancel');
};

watch(
  () => props.cluster,
  () => {
    handleGetSecurityGroups();
  },
  { immediate: true, deep: true },
);
</script>
<style scoped lang="postcss">
>>> .prefix {
  display: inline-block;
  height: 32px;
  line-height: 32px;
  background: #F0F1F5;
  border: 1px solid #C4C6CC;
  border-radius: 2px 0 0 2px;
  padding: 0 8px;
  font-size: 12px;
  &.disabled {
    border-color: #dcdee5;
  }
}
</style>
