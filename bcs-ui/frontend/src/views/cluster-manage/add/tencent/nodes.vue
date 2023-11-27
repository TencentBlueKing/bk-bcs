<template>
  <bk-form :model="nodeConfig" :rules="nodeConfigRules" ref="formRef">
    <bk-form-item :label="$t('cluster.create.label.initNodeTemplate')">
      <TemplateSelector v-model="nodeConfig.nodeTemplateID" is-tke-cluster />
    </bk-form-item>
    <bk-form-item
      key="master"
      class="tips-offset"
      :label="$t('cluster.create.label.hostResource')"
      property="nodes"
      error-display-type="normal"
      required>
      <bk-radio-group class="flex items-center mb-[6px] h-[32px]" :value="autoGenerateMasterNodes">
        <bk-radio disabled :value="true">
          {{ $t('cluster.create.label.applyResource') }}
        </bk-radio>
        <bk-radio disabled :value="false">
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
        v-if="!autoGenerateMasterNodes"
        @change="validate" />
    </bk-form-item>
    <!-- 申请主机资源 -->
    <template v-if="autoGenerateMasterNodes">
      <bcs-divider></bcs-divider>
      <ApplyHostResource
        class="mt-[8px] mb-[20px]"
        :show-recommended-config="false"
        :region="region"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :vpc-i-d="vpcID"
        :os="os"
        :region-list="regionList"
        :vpc-list="vpcList"
        :image-list-by-group="imageListByGroup"
        :zone-list="zoneList"
        :security-groups="securityGroups"
        :instances="instanceConfigList"
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
      <TopoSelector v-model="nodeConfig.clusterBasicSettings.module.workerModuleID" class="max-w-[500px]" />
    </bk-form-item>
    <!-- 用户名和密码 -->
    <bk-form-item
      :label="$t('tke.label.loginType.text')"
      property="workerLogin"
      error-display-type="normal"
      required
      ref="loginTypeRef">
      <div class="bk-button-group">
        <bk-button
          :class="[{ 'is-selected': loginType === 'password' }]"
          @click="loginType = 'password'">
          {{ $t('tke.label.loginType.password') }}
        </bk-button>
        <bk-button
          :class="[{ 'is-selected': loginType === 'ssh' }]"
          @click="loginType = 'ssh'">
          {{ $t('tke.label.loginType.ssh') }}
        </bk-button>
      </div>
      <div class="bg-[#F5F7FA] mt-[16px] max-w-[500px] py-[16px] pr-[16px]">
        <template v-if="loginType === 'password'">
          <bk-form-item :label="$t('tke.label.setPassword')" required>
            <bcs-input
              type="password"
              autocomplete="new-password"
              v-model="nodeConfig.nodeSettings.workerLogin.initLoginPassword">
            </bcs-input>
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.confirmPassword')" class="!mt-[16px]" required>
            <bcs-input type="password" autocomplete="new-password" v-model="confirmPassword"></bcs-input>
          </bk-form-item>
        </template>
        <template v-else-if="loginType === 'ssh'">
          <bk-form-item :label="$t('tke.label.publicKey')" :label-width="100" required>
            <bcs-select
              class="bg-[#fff]"
              :clearable="false"
              searchable
              v-model="nodeConfig.nodeSettings.workerLogin.keyPair.keyID">
              <bcs-option
                v-for="item in keyPairs"
                :key="item.KeyID"
                :id="item.KeyID"
                :name="item.KeyName">
              </bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.secretKey')" :label-width="100" class="!mt-[16px]" required>
            <bk-input type="textarea" v-model="nodeConfig.nodeSettings.workerLogin.keyPair.keySecret"></bk-input>
          </bk-form-item>
        </template>
      </div>
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
import { ClusterDataInjectKey, ICloudRegion, IHostNode, IInstanceItem, IKeyItem, ISecurityGroup, IZoneItem } from './types';

import { cloudKeyPairs, cloudSecurityGroups } from '@/api/modules/cluster-manager';
import $i18n from '@/i18n/i18n-setup';
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
  os: {
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
  regionList: {
    type: Array as PropType<ICloudRegion[]>,
    default: () => [],
  },
  vpcList: {
    type: Array as PropType<{
      name: string
      vpcId: string
    }[]>,
    default: () => [],
  },
  imageListByGroup: {
    type: Object,
    default: () => ({}),
  },
  zoneList: {
    type: Array as PropType<IZoneItem[]>,
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
});

const emits = defineEmits(['next', 'cancel', 'pre', 'confirm', 'instances-change']);

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
const watchOnce = watch(() => props.masterLogin, () => {
  // 默认跟master密码保持一致
  nodeConfig.value.nodeSettings.workerLogin = JSON.parse(JSON.stringify(props.masterLogin));
  confirmPassword.value = nodeConfig.value.nodeSettings.workerLogin.initLoginPassword;
  if (confirmPassword.value) {
    loginType.value = 'password';
  } else {
    loginType.value = 'ssh';
  }
  watchOnce();
});

const confirmPassword = ref('');

watch([
  () => nodeConfig.value.nodeSettings.workerLogin.initLoginPassword,
  () => confirmPassword.value,
  () => nodeConfig.value.nodeSettings.workerLogin.keyPair.keyID,
  () => nodeConfig.value.nodeSettings.workerLogin.keyPair.keySecret,
], () => {
  formRef.value?.$refs?.loginTypeRef?.validate();
});
// 动态 i18n 问题，这里使用computed
const nodeConfigRules = computed(() => ({
  nodes: [{
    message: $i18n.t('generic.validate.required'),
    trigger: 'custom',
    validator: () => {
      if (props.autoGenerateMasterNodes) {
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

// 登录方式
const loginType = ref<'password'|'ssh'>('password');
watch(loginType, () => {
  if (loginType.value === 'password') {
    nodeConfig.value.nodeSettings.workerLogin.keyPair = {
      keyID: '',
      keySecret: '',
      keyPublic: '',
    };
  } else if (loginType.value === 'ssh') {
    nodeConfig.value.nodeSettings.workerLogin.initLoginPassword = '';
    confirmPassword.value = '';
  }
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

// 密钥
const keyPairs = ref<Array<IKeyItem>>([]);
const cloudKeyPairsLoading = ref(false);
const handleGetCloudKeyPairs = async () => {
  if (!props.region || !props.cloudAccountID) return;
  cloudKeyPairsLoading.value = true;
  keyPairs.value = await cloudKeyPairs({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  cloudKeyPairsLoading.value = false;
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetSecurityGroups();
  handleGetCloudKeyPairs();
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
  }
};
// 取消
const handleCancel = () => {
  emits('cancel');
};
</script>
