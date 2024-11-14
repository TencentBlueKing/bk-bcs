<template>
  <bk-form :model="masterConfig" :rules="masterConfigRules" ref="formRef">
    <bk-form-item
      :label="$t('tke.label.apiServerCLB.text')"
      :desc="$t('tke.label.apiServerCLB.desc')"
      property="cidrs"
      error-display-type="normal"
      class="max-w-[600px]"
      key="cidrs"
      :required="ITNType === 'intranet'">
      <bk-radio-group
        v-model="ITNType"
        @change="handleChange"
      >
        <div class="flex items-center mb-[8px] h-[32px]">
          <bk-radio value="internet">{{ $t('tke.label.apiServerCLB.internet') }}</bk-radio>
        </div>
        <div class="flex items-center mb-[8px] h-[32px]">
          <bk-radio value="intranet">{{ $t('tke.label.apiServerCLB.intranet') }}</bk-radio>
        </div>
        <combination-input
          :list="cidrs"
          @data-change="updateSecurityGroup"
          :ref="el => combinationInputRef = el"
          :key-required="!masterConfig.clusterAdvanceSettings.clusterConnectSetting.isExtranet"
          :key-rules="keyRules" />
      </bk-radio-group>
    </bk-form-item>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button theme="primary" class="ml10 min-w-[75px]" @click="nextStep">
        {{ needNode ? $t('generic.button.next') : $t('generic.button.confirm') }}
      </bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { computed, PropType, ref } from 'vue';

import { INTERNET_CIDR_REGEX, INTRANET_CIDR_REGEX } from '@/common/constant';
import combinationInput from '@/components/combination-input.vue';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';

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
  needNode: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['next', 'cancel', 'pre', 'confirm']);

// master配置
const masterConfig = ref({
  autoGenerateMasterNodes: true,
  clusterAdvanceSettings: {
    clusterConnectSetting: {
      isExtranet: true,
      internet: {
        publicIPAssigned: false,
        publicAccessCidrs: [],
      },
    },
  },
});
// 网络类型
const ITNType = ref<'internet' | 'intranet'>('internet');
// cidr白名单
const cidrs = ref([]);

// 动态 i18n 问题，这里使用computed
const masterConfigRules = computed(() => ({
  cidrs: [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      async validator() {
        return await combinationInputRef.value?.validateAll().catch(() => false);
      },
    },
  ],
}));

// 校验master节点
const formRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  const result2 = await combinationInputRef.value?.validateAll().catch(() => false);
  return result && result2;
};

const combinationInputRef = ref<any>(null);
// 将ip对象数组转为ip字符串数组
function updateSecurityGroup(value) {
  cidrs.value = value;
  // eslint-disable-next-line max-len
  masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.internet.publicAccessCidrs = value.map(item => item.key);
}
function handleChange(value) {
  masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.isExtranet = !!['internetAndIntranet', 'internet'].includes(value);
  validate();
}

// const internetIpRegex = new RegExp()
const keyRules = ref([
  {
    message: $i18n.t('cluster.create.aws.cidrTips.tips1'),
    validator: '^(?!0.0.0.0/0$).*$',
  },
  {
    message: $i18n.t('cluster.create.aws.cidrTips.tips2'),
    validator: (val) => {
      if (ITNType.value === 'internet') {
        return new RegExp(INTERNET_CIDR_REGEX).test(String(val));
      }
      return new RegExp(INTRANET_CIDR_REGEX).test(String(val));
    },
  },
]);

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
    !props.needNode && emits('confirm');
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
