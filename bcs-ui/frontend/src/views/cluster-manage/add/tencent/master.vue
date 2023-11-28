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
        <bk-button
          :class="['min-w-[136px]', { 'is-selected': masterConfig.manageType === 'INDEPENDENT_CLUSTER' }]"
          @click="handleChangeManageType('INDEPENDENT_CLUSTER')">
          {{ $t('bcs.cluster.selfDeployed') }}
        </bk-button>
      </div>
      <div class="text-[12px]">
        <span
          v-if="masterConfig.manageType === 'MANAGED_CLUSTER'">
          {{ $t('cluster.create.label.manageType.managed.desc') }}
        </span>
        <span
          v-else-if="masterConfig.manageType === 'INDEPENDENT_CLUSTER'">
          {{ $t('cluster.create.label.manageType.independent.desc') }}
        </span>
      </div>
    </bk-form-item>
    <!-- 托管集群 -->
    <template v-if="masterConfig.manageType === 'MANAGED_CLUSTER'">
      <bk-form-item
        key="level"
        :label="$t('cluster.create.label.manageType.managed.clusterLevel.text')"
        property="clusterBasicSettings.clusterLevel"
        error-display-type="normal"
        required>
        <div class="bk-button-group">
          <bk-button
            :class="[
              'min-w-[48px]',
              { 'is-selected': item.level === masterConfig.clusterBasicSettings.clusterLevel }
            ]"
            v-for="item in clusterScale"
            :key="item.level"
            @click="handleChangeClusterScale(item.level)">
            {{ item.level }}
          </bk-button>
          <bk-checkbox disabled v-model="masterConfig.clusterBasicSettings.isAutoUpgradeClusterLevel" class="ml-[24px]">
            <span class="flex items-center">
              <span class="text-[12px]">{{ $t('cluster.create.label.manageType.managed.automatic.text') }}</span>
              <span
                class="ml5"
                v-bk-tooltips="{ content: $t('cluster.create.label.manageType.managed.automatic.tips') }">
                <i class="bcs-icon bcs-icon-question-circle"></i>
              </span>
            </span>
          </bk-checkbox>
        </div>
        <div class="text-[12px] leading-[20px] mt-[4px]">
          <i18n path="cluster.create.label.manageType.managed.clusterLevel.desc">
            <span place="nodes" class="text-[#313238]">{{ curClusterScale.level.split('L')[1] }}</span>
            <span place="pods" class="text-[#313238]">{{ curClusterScale.scale.maxNodePodNum }}</span>
            <span place="service" class="text-[#313238]">{{ curClusterScale.scale.maxServiceNum }}</span>
            <span place="crd" class="text-[#313238]">{{ curClusterScale.scale.cidrStep }}</span>
          </i18n>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.apiServerCLB.text')"
        :desc="$t('tke.label.apiServerCLB.desc')"
        property="clusterAdvanceSettings.clusterConnectSetting.securityGroup"
        error-display-type="normal"
        required>
        <bk-radio-group v-model="masterConfig.clusterAdvanceSettings.clusterConnectSetting.isExtranet">
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
              <bk-select
                class="ml-[-1px] flex-1"
                searchable
                :loading="securityGroupLoading"
                :disabled="!masterConfig.clusterAdvanceSettings.clusterConnectSetting.isExtranet"
                v-model="masterConfig.clusterAdvanceSettings.clusterConnectSetting.securityGroup">
                <bk-option
                  v-for="item in securityGroups"
                  :key="item.securityGroupID"
                  :id="item.securityGroupID"
                  :name="`${item.securityGroupName}(${item.securityGroupID})`">
                </bk-option>
                <div slot="extension">
                  <SelectExtension
                    :link-text="$t('tke.link.securityGroup')"
                    link="https://console.cloud.tencent.com/vpc/security-group"
                    @refresh="handleGetSecurityGroups" />
                </div>
              </bk-select>
            </div>
          </div>
          <bk-radio :value="false">
            {{ $t('tke.label.apiServerCLB.intranet') }}
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
    </template>
    <!-- 独立集群 -->
    <template v-else-if="masterConfig.manageType === 'INDEPENDENT_CLUSTER'">
      <bk-form-item
        key="master"
        class="tips-offset"
        :label="$t('cluster.create.label.hostResource')"
        property="master"
        error-display-type="normal"
        required>
        <bk-radio-group class="flex items-center mb-[6px] h-[32px]" v-model="masterConfig.autoGenerateMasterNodes">
          <bk-radio :value="true">
            {{ $t('cluster.create.label.applyResource') }}
          </bk-radio>
          <bk-radio :value="false">
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
          v-model="masterConfig.master"
          class="max-w-[80%]"
          v-if="!masterConfig.autoGenerateMasterNodes"
          @change="validate" />
      </bk-form-item>
      <!-- 申请主机资源 -->
      <template v-if="masterConfig.autoGenerateMasterNodes">
        <bcs-divider></bcs-divider>
        <ApplyHostResource
          class="mt-[8px] mb-[20px]"
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
          :level="level"
          node-role="MASTER_ETCD"
          v-bkloading="{ isLoading: instanceLoading }"
          ref="applyHostResourceRef"
          @instance-list-change="handleInstanceListChange"
          @common-config-change="handleCommonConfigChange"
          @level-change="handleLevelChange"
          @delete-instance="handleDeleteInstance"
          @refresh-security-groups="handleGetSecurityGroups" />
      </template>
      <bk-form-item
        :label="$t('tke.label.masterModule.text')"
        :desc="$t('tke.label.masterModule.desc')"
        property="clusterBasicSettings.module.masterModuleID"
        error-display-type="normal"
        required>
        <TopoSelector v-model="masterConfig.clusterBasicSettings.module.masterModuleID" class="max-w-[500px]" />
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.apiServerCLB.text')"
        :desc="$t('tke.label.apiServerCLB.desc')"
        property="clusterAdvanceSettings.clusterConnectSetting.securityGroup"
        error-display-type="normal"
        required>
        <bk-radio-group v-model="masterConfig.clusterAdvanceSettings.clusterConnectSetting.isExtranet">
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
              <bk-select
                class="ml-[-1px] flex-1"
                searchable
                :clearable="false"
                :loading="securityGroupLoading"
                v-model="masterConfig.clusterAdvanceSettings.clusterConnectSetting.securityGroup">
                <bk-option
                  v-for="item in securityGroups"
                  :key="item.securityGroupID"
                  :id="item.securityGroupID"
                  :name="`${item.securityGroupName}(${item.securityGroupID})`">
                </bk-option>
                <div slot="extension">
                  <SelectExtension
                    :link-text="$t('tke.link.securityGroup')"
                    link="https://console.cloud.tencent.com/vpc/security-group"
                    @refresh="handleGetSecurityGroups" />
                </div>
              </bk-select>
            </div>
          </div>
          <bk-radio :value="false">
            {{ $t('tke.label.apiServerCLB.intranet') }}
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <!-- 用户名和密码 -->
      <bk-form-item
        :label="$t('tke.label.loginType.text')"
        :desc="$t('tke.label.loginType.desc')"
        property="masterLogin"
        error-display-type="normal"
        required
        ref="loginTypeRef">
        <LoginType
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :value="masterConfig.nodeSettings.masterLogin"
          @change="handleLoginValueChange"
          @type-change="(v) => loginType = v"
          @pass-change="(v) => confirmPassword = v" />
      </bk-form-item>
    </template>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button theme="primary" class="ml10" @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { computed, inject, PropType, ref, watch } from 'vue';

import clusterScaleData from '../common/cluster-scale.json';
import IpSelector from '../common/ip-selector.vue';

import ApplyHostResource from './apply-host-resource.vue';
import TopoSelector from './topo-selector.vue';
import { ClusterDataInjectKey, ICloudRegion, IHostNode, IInstanceItem, IScale, ISecurityGroup, IZoneItem } from './types';

import { cloudInstanceTypesByLevel, cloudSecurityGroups } from '@/api/modules/cluster-manager';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $i18n from '@/i18n/i18n-setup';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import LoginType from '@/views/cluster-manage/add/form/login-type.vue';

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
  nodes: {
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
});

const emits = defineEmits(['next', 'cancel', 'pre', 'resource-type', 'instances-change']);
const clusterData = inject(ClusterDataInjectKey);
const subnetZoneList = computed(() => {
  if (!clusterData || !clusterData.value) return [];

  return clusterData.value?.clusterAdvanceSettings?.networkType === 'VPC-CNI'
    ? clusterData.value?.networkSettings?.subnetSource?.new?.map(item => item.zone) || []
    : [];
});

const disabledIpList = computed(() => props.nodes.map(ip => ({
  ip,
  tips: $i18n.t('cluster.create.validate.ipExitInNode'),
})));
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
      securityGroup: '',
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
const handleLoginValueChange = (value) => {
  masterConfig.value.nodeSettings.masterLogin = value;
};
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
        if (masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.isExtranet) {
          return !!masterConfig.value.clusterAdvanceSettings.clusterConnectSetting.securityGroup;
        }
        return true;
      },
    },
  ],
  master: [{
    message: props.environment === 'debug'
      ? $i18n.t('cluster.create.validate.masterNum35')
      : $i18n.t('cluster.create.validate.masterNum135'),
    trigger: 'custom',
    validator: () => {
      if (masterConfig.value.autoGenerateMasterNodes) {
        return true;
      }
      const maxMasterNum = props.environment === 'debug' ? [1, 3, 5] : [3, 5];
      return masterConfig.value.master.length && maxMasterNum.includes(masterConfig.value.master.length);
    },
  }],
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

const level = ref('');
const handleLevelChange = (data) => {
  level.value = data;
  handleGetCloudInstanceTypesByLevel();
};
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
// 托管集群
const clusterScale = ref<IScale[]>(clusterScaleData.data);
const curClusterScale = computed<IScale>(() => clusterScale.value
  .find(item => item.level === masterConfig.value.clusterBasicSettings.clusterLevel)
      || { level: 'L5', scale: { maxNodePodNum: 0, maxServiceNum: 0, cidrStep: 0 } });
// 托管集群规格
const handleChangeClusterScale = (scale) => {
  masterConfig.value.clusterBasicSettings.clusterLevel = scale;
};

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
  result && emits('next', {
    ...masterConfig.value,
    master: (masterConfig.value.master as IHostNode[]).map(item => item.ip),
  });
};
// 取消
const handleCancel = () => {
  emits('cancel');
};
</script>
