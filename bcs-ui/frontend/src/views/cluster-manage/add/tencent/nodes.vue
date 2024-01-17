<template>
  <bk-form :model="nodeConfig" :rules="nodeConfigRules" ref="formRef">
    <bk-form-item :label="$t('cluster.create.label.initNodeTemplate')">
      <TemplateSelector
        class="max-w-[600px]"
        v-model="nodeConfig.nodeTemplateID"
        :provider="cloudID"
        :show-preview="false" />
    </bk-form-item>
    <bk-form-item
      key="master"
      class="tips-offset"
      :label="$t('cluster.create.label.hostResource')"
      property="nodes"
      error-display-type="normal"
      required
      ref="masterItemRef">
      <bk-radio-group
        class="flex items-center mb-[6px] h-[32px] max-w-[600px]"
        v-model="nodeConfig.autoGenerateMasterNodes"
        v-bk-tooltips="{
          content: $t('tke.tips.needSameResourceType'),
          disabled: !disableChooseResourceType,
        }">
        <bk-radio :disabled="disableChooseResourceType" :value="true">
          {{ $t('cluster.create.label.applyResource') }}
        </bk-radio>
        <bk-radio :disabled="disableChooseResourceType" :value="false">
          {{ $t('cluster.create.label.useExitHost') }}
        </bk-radio>
      </bk-radio-group>
      <IpSelector
        :region="region"
        :cloud-id="provider"
        :disabled-ip-list="disabledIpList"
        :region-list="regionList"
        :vpc="{ vpcID: vpcID }"
        :account-i-d="cloudAccountID"
        :available-zone-list="subnetZoneList"
        v-model="nodeConfig.nodes"
        class="max-w-[80%]"
        v-if="!nodeConfig.autoGenerateMasterNodes"
        @change="handleValidateNode" />
    </bk-form-item>
    <!-- 申请主机资源 -->
    <template v-if="nodeConfig.autoGenerateMasterNodes">
      <bcs-divider></bcs-divider>
      <ApplyHostResource
        class="mt-[8px] mb-[20px]"
        :show-recommended-config="false"
        :region="region"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :vpc-i-d="vpcID"
        :security-groups="securityGroups"
        :instances="instanceConfigList"
        :disable-data-disk="false"
        :disable-internet-access="false"
        :max-nodes="100"
        node-role="WORKER"
        v-bkloading="{ isLoading: instanceLoading }"
        ref="applyHostResourceRef"
        @instance-list-change="handleInstanceListChange"
        @common-config-change="handleCommonConfigChange"
        @delete-instance="handleDeleteInstance" />
    </template>
    <bk-form-item
      :label="$t('tke.label.nodeModule.text')"
      :desc="$t('tke.label.nodeModule.desc')"
      property="clusterBasicSettings.module.workerModuleID"
      error-display-type="normal"
      required>
      <TopoSelector
        :placeholder="$t('generic.placeholder.select')"
        v-model="nodeConfig.clusterBasicSettings.module.workerModuleID"
        class="max-w-[600px]" />
    </bk-form-item>
    <!-- 用户名和密码 -->
    <bk-form-item
      :label="$t('tke.label.loginType.text')"
      :desc="$t('tke.label.loginType.desc')"
      property="workerLogin"
      error-display-type="normal"
      required
      ref="loginTypeRef">
      <LoginType
        :region="region"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :confirm-pass="confirmPassword"
        :value="nodeConfig.nodeSettings.workerLogin"
        :type="loginType"
        @pass-blur="validateLogin('custom')"
        @confirm-pass-blur="validateLogin('')"
        @key-secret-blur="validateLogin('')"
        @change="handleLoginValueChange"
        @type-change="(v) => loginType = v"
        @pass-change="handleConfirmPasswordChange" />
    </bk-form-item>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button class="min-w-[88px]" @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button
        theme="primary"
        class="ml10 min-w-[88px]"
        @click="handleConfirm">{{ $t('generic.button.confirm') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { computed, inject, PropType, ref, watch } from 'vue';

import IpSelector from '../common/ip-selector.vue';

import ApplyHostResource from './apply-host-resource.vue';
import TopoSelector from './topo-selector.vue';
import { ClusterDataInjectKey, IHostNode, IInstanceItem, ISecurityGroup } from './types';

import { cloudSecurityGroups } from '@/api/modules/cluster-manager';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import LoginType from '@/views/cluster-manage/add/form/login-type.vue';
import TemplateSelector from '@/views/cluster-manage/components/template-selector.vue';

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
  environment: {
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
  master: {
    type: Array as PropType<string[]>,
    default: () => [],
  },
  masterLogin: {
    type: Object,
    default: () => ({
      initLoginUsername: '',
      initLoginPassword: '',
      keyPair: {
        keyID: '',
        keySecret: '',
        keyPublic: '',
      },
    }),
  },
  autoGenerateMasterNodes: {
    type: Boolean,
    default: false,
  },
  manageType: {
    type: String as PropType<'MANAGED_CLUSTER'|'INDEPENDENT_CLUSTER'>,
    default: 'MANAGED_CLUSTER',
  },
});

const emits = defineEmits(['next', 'cancel', 'pre', 'confirm', 'instances-change']);

const regionList = computed(() => $store.state.cloudMetadata.regionList);

const disableChooseResourceType = computed(() => props.manageType === 'INDEPENDENT_CLUSTER');

const clusterData = inject(ClusterDataInjectKey);
const subnetZoneList = computed(() => {
  if (!clusterData || !clusterData.value) return [];

  return clusterData.value?.clusterAdvanceSettings?.networkType === 'VPC-CNI'
    ? clusterData.value?.networkSettings?.subnetSource?.new?.map(item => item.zone) || []
    : [];
});

const disabledIpList = computed(() => props.master.map(ip => ({
  ip,
  tips: $i18n.t('cluster.create.validate.ipExitInMaster'),
})));
// node配置
const nodeConfig = ref({
  autoGenerateMasterNodes: false,
  nodes: [],
  nodeTemplateID: '',
  clusterBasicSettings: {
    module: {
      workerModuleID: '',
    },
  },
  nodeSettings: {
    workerLogin: props.masterLogin,
  },
});
watch(
  [
    () => props.autoGenerateMasterNodes,
  ],
  () => {
    nodeConfig.value.autoGenerateMasterNodes = props.autoGenerateMasterNodes;
  },
  { immediate: true },
);

const watchOnce = watch(() => props.masterLogin, () => {
  // 默认跟master密码保持一致
  nodeConfig.value.nodeSettings.workerLogin = JSON.parse(JSON.stringify(props.masterLogin));
  confirmPassword.value = nodeConfig.value.nodeSettings.workerLogin.initLoginPassword;
  const { keyID, keySecret } = nodeConfig.value.nodeSettings.workerLogin?.keyPair || {};
  if (keyID && keySecret) {
    loginType.value = 'ssh';
  } else {
    loginType.value = 'password';
  }
  watchOnce();
});

// 登录方式
const loginType = ref<'password'|'ssh'>('password');
const confirmPassword = ref('');
const validateLogin = (trigger = '') => {
  formRef.value?.$refs?.loginTypeRef?.validate(trigger);
};
const handleLoginValueChange = (value) => {
  nodeConfig.value.nodeSettings.workerLogin = value;
};
const handleConfirmPasswordChange = (v) => {
  confirmPassword.value = v;
};

// watch([
//   () => nodeConfig.value.nodeSettings.workerLogin.initLoginPassword,
//   () => confirmPassword.value,
//   () => nodeConfig.value.nodeSettings.workerLogin.keyPair.keyID,
//   () => nodeConfig.value.nodeSettings.workerLogin.keyPair.keySecret,
// ], () => {
//   formRef.value?.$refs?.loginTypeRef?.validate();
// });
// 动态 i18n 问题，这里使用computed
const nodeConfigRules = computed(() => ({
  nodes: [{
    message: $i18n.t('generic.validate.required'),
    trigger: 'custom',
    validator: () => {
      if (nodeConfig.value.autoGenerateMasterNodes) {
        return true;
      }
      return !!nodeConfig.value.nodes.length;
    },
  }],
  'clusterBasicSettings.module.workerModuleID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
    },
  ],
  workerLogin: [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        if (loginType.value === 'password') {
          return !!nodeConfig.value.nodeSettings.workerLogin.initLoginPassword;
        }
        return !!nodeConfig.value.nodeSettings.workerLogin.keyPair.keyID
         && !!nodeConfig.value.nodeSettings.workerLogin.keyPair.keySecret;
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('cluster.ca.nodePool.create.validate.password'),
      validator() {
        if (loginType.value === 'password') {
          const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/;
          return regex.test(nodeConfig.value.nodeSettings.workerLogin.initLoginPassword);
        }
        return true;
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('tke.validate.passwordNotSame'),
      validator() {
        if (loginType.value === 'password' && confirmPassword.value) {
          return nodeConfig.value.nodeSettings.workerLogin.initLoginPassword === confirmPassword.value;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.passwordNotSame'),
      validator() {
        if (loginType.value === 'password') {
          return nodeConfig.value.nodeSettings.workerLogin.initLoginPassword === confirmPassword.value;
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
    message: $i18n.t('generic.validate.required'),
    validator() {
      const count = instanceConfigList.value.reduce((pre, item) => {
        pre += item.applyNum;
        return pre;
      }, 0);
      return count > 0;
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

const handleValidateNode = () => {
  formRef.value?.$refs?.masterItemRef?.validate();
};

// 机型配置
const instanceLoading = ref(false);
const instanceCommonConfig = ref<Partial<IInstanceItem>>({});
const handleCommonConfigChange = (config: Partial<IInstanceItem>) => {
  instanceCommonConfig.value = config;
};

const instanceConfigList = ref<Array<IInstanceItem>>([]);
const handleInstanceListChange = (data: IInstanceItem[]) => {
  instanceConfigList.value = data;
};
const handleDeleteInstance = (index: number) => {
  instanceConfigList.value.splice(index, 1);
};
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

// 安全组
const securityGroupLoading = ref(false);
const securityGroups = ref<Array<ISecurityGroup>>([]);
const handleGetSecurityGroups = async () => {
  if (!props.region || !props.cloudAccountID) return;
  securityGroupLoading.value = true;
  securityGroups.value = await cloudSecurityGroups({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  securityGroupLoading.value = false;
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetSecurityGroups();
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
const handleConfirm = async () => {
  const result = await validate();
  if (result) {
    emits('next', {
      ...nodeConfig.value,
      nodes: (nodeConfig.value.nodes as IHostNode[]).map(item => item.ip),
    });
    emits('confirm');
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
