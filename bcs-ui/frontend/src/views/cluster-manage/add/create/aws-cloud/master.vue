<template>
  <bk-form :model="masterConfig" :rules="masterConfigRules" ref="formRef">
    <bk-form-item :label="$t('cluster.create.label.manageType.text')">
      <div class="bk-button-group">
        <bk-button
          :class="['min-w-[136px]', { 'is-selected': masterConfig.manageType === 'MANAGED_CLUSTER' }]"
          @click="handleChangeManageType('MANAGED_CLUSTER')">
          <div class="flex items-center">
            <span class="flex text-[16px] text-[#f85356]">
              <i class="bcs-icon bcs-icon-hot"></i>
            </span>
            <span class="ml-[8px]">{{ $t('bcs.cluster.managed1') }}</span>
          </div>
        </bk-button>
      </div>
      <div class="text-[12px] leading-[20px] mt-[4px]">
        <span>{{ $t('amazonCloud.label.aksDesc') }}</span>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('tke.label.apiServerCLB.text')"
      :desc="$t('tke.label.apiServerCLB.desc')"
      property="clusterAdvanceSettings.clusterConnectSetting.cidrs"
      error-display-type="normal"
      key="cidrs"
      required>
      <NetworkSelector
        :ref="el => networkSelectorRef = el"
        class="max-w-[600px]"
        :value="masterConfig.clusterAdvanceSettings.clusterConnectSetting"
        :region="region"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :value-list="[
          { label: $t('tke.label.apiServerCLB.internet'), value: 'internet' },
          { label: $t('tke.label.apiServerCLB.intranet'), value: 'intranet' },
        ]"
        @change="(v) => masterConfig.clusterAdvanceSettings.clusterConnectSetting = v" />
    </bk-form-item>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button theme="primary" class="ml10" @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { computed, PropType, ref } from 'vue';

import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';
import NetworkSelector from '@/views/cluster-manage/add/components/network-selector.vue';

const props = defineProps({
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
  provider: {
    type: String,
    default: '',
  },
  vpcID: {
    type: String,
    default: '',
  },
  nodes: {
    type: Array as PropType<string[]>,
    default: () => [],
  },
});

const emits = defineEmits(['next', 'cancel', 'pre']);

// master配置
const masterConfig = ref({
  manageType: 'MANAGED_CLUSTER',
  autoGenerateMasterNodes: true,
  clusterAdvanceSettings: {
    clusterConnectSetting: {
      ITNType: 'intranet',
      isExtranet: false,
      subnetId: '',
      cidrs: [],
      internet: {
        publicIPAssigned: false,
        publicAccessCidrs: [],
      },
    },
  },
});

const networkSelectorRef = ref<InstanceType<typeof NetworkSelector> | null>(null);
// 动态 i18n 问题，这里使用computed
const masterConfigRules = computed(() => ({
  'clusterAdvanceSettings.clusterConnectSetting.cidrs': [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      async validator() {
        if (masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.ITNType === 'internet') {
          return await networkSelectorRef.value?.validate();
        }
        return true;
      },
    },
  ],
}));

// master配置
const handleChangeManageType = (type: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER') => {
  masterConfig.value.manageType = type;
};

// 校验master节点
const formRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  return result;
};

// 上一步
const preStep = () => {
  emits('pre');
};
// 下一步
const { focusOnErrorField } = useFocusOnErrorField();
const nextStep = async () => {
  const result = await validate();
  if (result) {
    emits('next', {
      ...masterConfig.value,
    });
  } else {
    // 自动滚动到第一个错误的位置
    focusOnErrorField();
  }
};
// 取消
const handleCancel = () => {
  emits('cancel');
};

defineExpose({
  validate,
});
</script>
