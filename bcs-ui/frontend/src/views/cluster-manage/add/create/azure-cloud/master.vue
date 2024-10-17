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
            <span class="ml-[8px]">{{ $t('bcs.cluster.managed') }}</span>
          </div>
        </bk-button>
      </div>
      <div class="text-[12px] leading-[20px] mt-[4px]">
        <span>{{ $t('cluster.create.label.manageType.managed.desc') }}</span>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('tke.label.apiServerCLB.text')"
      :desc="$t('tke.label.apiServerCLB.desc')"
      property="clusterAdvanceSettings.clusterConnectSetting.securityGroup"
      error-display-type="normal"
      key="securityGroup"
      required>
      <NetworkSelector
        class="max-w-[600px]"
        :value="masterConfig.clusterAdvanceSettings.clusterConnectSetting"
        :region="region"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :value-list="[
          { label: $t('tke.label.apiServerCLB.internet'), value: 'internet' },
          { label: $t('tke.label.apiServerCLB.intranet'), value: 'intranet' }
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
import { computed, inject, PropType, ref, watch } from 'vue';

import { ClusterDataInjectKey, IHostNode, IInstanceItem } from '../../../types/types';

import { cloudInstanceTypesByLevel } from '@/api/modules/cluster-manager';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
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

const emits = defineEmits(['next', 'cancel', 'pre', 'resource-type', 'instances-change']);

const clusterData = inject(ClusterDataInjectKey);
const subnetZoneList = computed(() => {
  if (!clusterData?.value) return [];

  return clusterData.value?.clusterAdvanceSettings?.networkType === 'VPC-CNI'
    ? clusterData.value?.networkSettings?.subnetSource?.new?.map(item => item.zone) || []
    : [];
});

// master配置
const masterConfig = ref({
  manageType: 'MANAGED_CLUSTER',
  autoGenerateMasterNodes: true,
  master: [],
  clusterBasicSettings: {
    clusterLevel: 'L20',
    isAutoUpgradeClusterLevel: true,
    module: {
      masterModuleID: '',
    },
  },
  clusterAdvanceSettings: {
    clusterConnectSetting: {
      isExtranet: true,
      subnetId: '',
      securityGroup: [],
    },
  },
  nodeSettings: {
    masterLogin: {
      initLoginUsername: '',
      initLoginPassword: '',
      keyPair: {
        keyID: '',
        keySecret: '',
        keyPublic: '',
      },
    },
  },
});

// 登录方式
const loginType = ref<'password'|'ssh'>('password');
const confirmPassword = ref('');
// watch([
//   () => masterConfig.value.nodeSettings.masterLogin.initLoginPassword,
//   () => confirmPassword.value,
//   () => masterConfig.value.nodeSettings.masterLogin.keyPair.keyID,
//   () => masterConfig.value.nodeSettings.masterLogin.keyPair.keySecret,
// ], () => {
//   formRef.value?.$refs?.loginTypeRef?.validate();
// });
// 动态 i18n 问题，这里使用computed
const masterConfigRules = computed(() => ({
  'clusterBasicSettings.clusterLevel': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
    },
  ],
  'clusterAdvanceSettings.clusterConnectSetting.securityGroup': [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      validator() {
        // todo ip校验
        if (masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.isExtranet) {
          return !!masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.securityGroup;
        }
        return true;
      },
    },
  ],
  master: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
      validator: () => {
        if (masterConfig.value.autoGenerateMasterNodes) {
          return true;
        }
        return !!masterConfig.value.master.length;
      },
    },
    {
      message: $i18n.t('cluster.create.validate.masterNum35'),
      trigger: 'custom',
      validator: () => {
        if (masterConfig.value.autoGenerateMasterNodes) {
          return true;
        }
        const maxMasterNum = [3, 5];
        return masterConfig.value.master.length && maxMasterNum.includes(masterConfig.value.master.length);
      },
    },
  ],
  'clusterBasicSettings.module.masterModuleID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
    },
  ],
  masterLogin: [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        if (loginType.value === 'password') {
          return !!masterConfig.value.nodeSettings.masterLogin.initLoginPassword;
        }
        return !!masterConfig.value.nodeSettings.masterLogin.keyPair.keyID
         && !!masterConfig.value.nodeSettings.masterLogin.keyPair.keySecret;
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('cluster.ca.nodePool.create.validate.password'),
      validator() {
        if (loginType.value === 'password') {
          const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/;
          return regex.test(masterConfig.value.nodeSettings.masterLogin.initLoginPassword);
        }
        return true;
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('tke.validate.passwordNotSame'),
      validator() {
        if (loginType.value === 'password' && confirmPassword.value) {
          return masterConfig.value.nodeSettings.masterLogin.initLoginPassword === confirmPassword.value;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.passwordNotSame'),
      validator() {
        if (loginType.value === 'password') {
          return masterConfig.value.nodeSettings.masterLogin.initLoginPassword === confirmPassword.value;
        }
        return true;
      },
    },
  ],
  // bk-form-item 在apply-host-resource组件中
  instanceChargeType: [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        return !!instanceCommonConfig.value.instanceChargeType;
      },
    },
  ],
  // bk-form-item 在apply-host-resource组件中
  instances: [{
    trigger: 'custom',
    message: $i18n.t('cluster.create.validate.masterNum35'),
    validator() {
      const count = instanceConfigList.value.reduce((pre, item) => {
        pre += item.applyNum;
        return pre;
      }, 0);
      return [3, 5].includes(count);
    },
  }],
  // bk-form-item 在apply-host-resource组件中
  securityGroupIDs: [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        return !!instanceCommonConfig.value.securityGroupIDs?.length;
      },
    },
  ],
}));

// 机型配置
const instanceLoading = ref(false);
const instanceCommonConfig = ref<Partial<IInstanceItem>>({});

const instanceConfigList = ref<Array<IInstanceItem>>([]);
const instances = computed(() => instanceConfigList.value.map(item => ({
  ...item,
  ...instanceCommonConfig.value,
  region: props.region,
  vpcID: props.vpcID,
  dockerGraphPath: '',
})));
watch(instances, () => {
  formRef.value?.$refs?.applyHostResourceRef?.$refs?.instancesRef?.validate();
  emits('instances-change', instances.value);
});

const level = ref('');
const handleGetCloudInstanceTypesByLevel = async () => {
  if (!props.region || !level.value || !props.cloudAccountID) return;

  instanceLoading.value = true;
  instanceConfigList.value = await cloudInstanceTypesByLevel({
    $cloudId: props.cloudID,
    $region: props.region,
    $level: level.value,
    vpcID: props.vpcID,
    accountID: props.cloudAccountID,
    zones: subnetZoneList.value.join(','),
  }).catch(() => []);
  if (!instanceConfigList.value.length) {
    $bkInfo({
      type: 'warning',
      clsName: 'custom-info-confirm',
      title: $i18n.t('tke.title.noAvailableSubnets'),
      defaultInfo: true,
      okText: $i18n.t('tke.button.addSubnets'),
      confirmFn: async () => {
        window.open(`https://console.cloud.tencent.com/vpc/vpc/detail?rid=1&id=${props.vpcID}`);
      },
    });
  }
  instanceLoading.value = false;
};

// master配置
const handleChangeManageType = (type: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER') => {
  masterConfig.value.manageType = type;
  masterConfig.value.master = [];
  // validate();
};
watch([
  () => props.region,
  () => props.cloudAccountID,
  () => props.vpcID,
  () => subnetZoneList.value,
], () => {
  handleGetCloudInstanceTypesByLevel();
});

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
const nextStep = async () => {
  const result = await validate();
  if (result) {
    emits('next', {
      ...masterConfig.value,
      master: (masterConfig.value.master as IHostNode[]).map(item => item.ip),
    });
  } else {
    // 自动滚动到第一个错误的位置
    const errDom = document.getElementsByClassName('form-error-tip');
    errDom[0]?.scrollIntoView({
      block: 'center',
      behavior: 'smooth',
    });
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
