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
        <span
          class="text-[14px] text-[#699DF4] cursor-pointer"
          @click="openLink('https://bk.tencent.com/docs/markdown/ZH/BCS/1.29/UserGuide/FAQ/faq.md')">
          {{ $t('tke.title.connectFailure.refButton') }}
        </span>
      </i18n>
    </div>
    <div class="bg-[#F5F7FA] p-[16px] mt-[16px]">
      <ApiServer
        :value="clusterConnectSetting"
        :region="cluster.region"
        :cloud-account-i-d="cluster.cloudAccountID"
        :cloud-i-d="cluster.provider"
        @change="(v) => clusterConnectSetting = v" />
    </div>
    <div class="flex items-center justify-center mt-[24px]">
      <bk-button
        theme="primary"
        class="min-w-[88px]"
        :loading="isLoading"
        :disabled="clusterConnectSetting.isExtranet && !clusterConnectSetting.securityGroup"
        @click="handleConfirm">{{ $t('generic.button.confirm') }}</bk-button>
      <bk-button class="min-w-[88px]" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { merge } from 'lodash';
import { PropType, ref, watch } from 'vue';

import { modifyCluster } from '@/api/modules/cluster-manager';
import { ICluster } from '@/composables/use-app';
import ApiServer from '@/views/cluster-manage/add/components/api-server.vue';

const props = defineProps({
  cluster: {
    type: Object as PropType<ICluster>,
    default: () => ({}),
  },
});
const emits = defineEmits(['cancel', 'confirm']);

const clusterConnectSetting = ref({
  isExtranet: true,
  securityGroup: '',
});

watch(
  () => props.cluster,
  () => {
    const setting = props.cluster?.clusterAdvanceSettings?.clusterConnectSetting || {};
    clusterConnectSetting.value = JSON.parse(JSON.stringify(setting));
  },
  { immediate: true, deep: true },
);

const isLoading = ref(false);
const handleConfirm = async () => {
  isLoading.value = true;
  const result = await modifyCluster({
    $clusterId: props.cluster.clusterID,
    clusterAdvanceSettings: merge(
      props.cluster.clusterAdvanceSettings,
      {
        clusterConnectSetting: {
          isExtranet: clusterConnectSetting.value.isExtranet,
          securityGroup: clusterConnectSetting.value.securityGroup,
        },
      },
    ),
  }).then(() => true)
    .catch(() => false);

  isLoading.value = false;
  if (result) {
    emits('confirm', props.cluster.clusterID);
    emits('cancel');
  }
};
// 跳转链接
const openLink = (link: string) => {
  if (!link) return;

  window.open(link);
};
const handleCancel = () => {
  emits('cancel');
};

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
