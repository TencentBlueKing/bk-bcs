<!-- eslint-disable max-len -->
<template>
  <BcsContent :padding="0" :title="$t('cluster.button.addCluster')" :desc="$t('cluster.tips.createK8SCluster')">
    <div class="h-full pt-[8px] bg-[#f0f1f5]">
      <bcs-tab
        :label-height="42"
        :validate-active="false"
        :active.sync="activeTabName"
        type="card-tab"
        class="h-full">
        <!-- 基本信息 -->
        <bcs-tab-panel :name="steps[0].name">
          <template #label>
            <StepTabLabel :title="$t('generic.title.basicInfo1')" :step-num="1" :active="activeTabName === steps[0].name" />
          </template>
          <bk-form
            :ref="steps[0].formRef"
            :model="basicInfo"
            :rules="basicInfoRules"
            class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]">
            <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
              <bk-form-item :label="$t('cluster.create.label.kubernetesProvider')">
                {{ $t('cluster.create.label.kubernetesCommunity') }}
              </bk-form-item>
              <bk-form-item :label="$t('cluster.labels.name')" property="clusterName" error-display-type="normal" required>
                <bk-input
                  :maxlength="64"
                  :placeholder="$t('cluster.create.validate.name')"
                  v-model.trim="basicInfo.clusterName">
                </bk-input>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.clusterVersion')"
                property="clusterBasicSettings.version"
                error-display-type="normal"
                required>
                <bcs-select
                  :loading="templateLoading"
                  v-model="basicInfo.clusterBasicSettings.version"
                  searchable
                  :clearable="false"
                  class="max-w-[calc(50%-46px)]">
                  <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
                </bcs-select>
              </bk-form-item>
              <bk-form-item
                :label="$t('k8s.label')"
                property="labels"
                error-display-type="normal">
                <KeyValue
                  v-model="basicInfo.labels"
                  :key-rules="[
                    {
                      message: $t('generic.validate.label'),
                      validator: LABEL_KEY_REGEXP,
                    }
                  ]"
                  :value-rules="[
                    {
                      message: $t('generic.validate.label'),
                      validator: LABEL_KEY_REGEXP,
                    }
                  ]" />
              </bk-form-item>
              <bk-form-item :label="$t('cluster.create.label.desc')">
                <bk-input maxlength="100" v-model="basicInfo.description" type="textarea"></bk-input>
              </bk-form-item>
            </DescList>
            <div>
              <DescList size="middle" :title="$t('cluster.labels.env')">
                <bk-form-item :label="$t('cluster.create.label.region')" property="region" error-display-type="normal" required>
                  <bcs-select
                    v-model="basicInfo.region"
                    searchable
                    :loading="regionLoading"
                    :clearable="false">
                    <bcs-option
                      v-for="item in regionList"
                      :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
                  </bcs-select>
                </bk-form-item>
                <bk-form-item :label="$t('cluster.labels.env')" property="environment" error-display-type="normal" required>
                  <bk-radio-group v-model="basicInfo.environment">
                    <bk-radio value="debug">
                      {{ $t('cluster.env.debug') }}
                    </bk-radio>
                    <bk-radio value="prod">
                      {{ $t('cluster.env.prod') }}
                    </bk-radio>
                  </bk-radio-group>
                </bk-form-item>
              </DescList>
              <DescList class="mt-[24px]" size="middle" :title="$t('cluster.title.clusterConfig')">
                <bk-form-item
                  :label="$t('cluster.create.label.system')"
                  property="clusterBasicSettings.OS"
                  error-display-type="normal"
                  required>
                  <bcs-select searchable :clearable="false" v-model="basicInfo.clusterBasicSettings.OS">
                    <bcs-option
                      v-for="item in osList"
                      :key="item"
                      :id="item"
                      :name="item" />
                  </bcs-select>
                </bk-form-item>
                <bk-form-item
                  :label="$t('cluster.create.label.containerRuntime')"
                  property="clusterBasicSettings.containerRuntime"
                  error-display-type="normal"
                  required>
                  <bk-radio-group
                    v-model="basicInfo.clusterAdvanceSettings.containerRuntime"
                    @change="handleRuntimeChange">
                    <bk-radio value="containerd" :disabled="!runtimeModuleParamsMap['containerd']">containerd</bk-radio>
                    <bk-radio value="docker" :disabled="!runtimeModuleParamsMap['docker']">docker</bk-radio>
                  </bk-radio-group>
                </bk-form-item>
                <bk-form-item
                  :label="$t('cluster.create.label.runtimeVersion')"
                  property="clusterBasicSettings.runtimeVersion"
                  error-display-type="normal"
                  required>
                  <bcs-select
                    searchable
                    :clearable="false"
                    v-model="basicInfo.clusterAdvanceSettings.runtimeVersion"
                    class="max-w-[50%]">
                    <bcs-option v-for="item in runtimeVersionList" :key="item" :id="item" :name="item"></bcs-option>
                  </bcs-select>
                </bk-form-item>
              </DescList>
            </div>
          </bk-form>
        </bcs-tab-panel>
        <!-- 网络配置 -->
        <bcs-tab-panel :name="steps[1].name" :disabled="steps[1].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.detail.title.network')"
              :step-num="2"
              :active="activeTabName === steps[1].name"
              :disabled="steps[1].disabled" />
          </template>
          <bk-form
            class="k8s-form"
            :label-width="160"
            :ref="steps[1].formRef"
            :model="networkConfig"
            :rules="networkConfigRules">
            <DescList size="middle" :title="$t('cluster.create.label.basicConfig')">
              <bk-form-item :label="$t('cluster.create.label.region')" required>
                <bcs-select
                  v-model="basicInfo.region"
                  searchable
                  :clearable="false"
                  disabled>
                  <bcs-option
                    v-for="item in regionList"
                    :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
                </bcs-select>
              </bk-form-item>
              <bk-form-item :label="$t('cluster.create.label.privateNet.text')" required>
                <bcs-select :clearable="false" searchable v-model="networkConfig.vpcID">
                  <bcs-option id="default" :name="$t('cluster.create.label.privateNet.customNet')"></bcs-option>
                </bcs-select>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.defaultNetPlugin')"
                property="clusterAdvanceSettings.networkType"
                error-display-type="normal"
                required>
                <bcs-select :clearable="false" searchable v-model="networkConfig.clusterAdvanceSettings.networkType">
                  <bcs-option id="flannel" name="Flannel(Overlay)"></bcs-option>
                </bcs-select>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.clusterIPType.text')"
                property="networkSettings.clusterIpType"
                error-display-type="normal"
                required>
                <bk-radio-group v-model="networkConfig.networkSettings.clusterIpType">
                  <bk-radio value="ipv4" :disabled="!supportNetworkType.includes('ipv4')">
                    IPv4
                  </bk-radio>
                  <bk-radio value="ipv6" :disabled="!supportNetworkType.includes('ipv6')">
                    IPv6
                  </bk-radio>
                  <bk-radio value="dual" :disabled="!supportNetworkType.includes('dual')">
                    {{ $t('cluster.create.label.clusterIPType.dual') }}
                  </bk-radio>
                  <!-- <bk-radio>
                    IPv4 或 IPv6 单栈 (由平台自动分配)
                  </bk-radio> -->
                </bk-radio-group>
              </bk-form-item>
            </DescList>
            <DescList class="mt-[24px]" size="middle" :title="$t('cluster.create.label.netSetting')">
              <bk-form-item
                label="Service IP"
                property="networkSettings.serviceCIDR"
                error-display-type="normal"
                required>
                <div class="flex">
                  <bk-input
                    class="mr-[8px] max-w-[50%]"
                    v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)"
                    v-model.trim="networkConfig.networkSettings.serviceIPv4CIDR">
                    <div
                      slot="prepend"
                      class="text-[12px] px-[12px] leading-[28px] whitespace-nowrap">
                      IPv4 CIDR
                    </div>
                  </bk-input>
                  <bk-input
                    class="max-w-[50%]"
                    v-if="['ipv6', 'dual'].includes(networkConfig.networkSettings.clusterIpType)"
                    v-model.trim="networkConfig.networkSettings.serviceIPv6CIDR">
                    <div
                      slot="prepend"
                      class="text-[12px] px-[12px] leading-[28px] whitespace-nowrap">
                      IPv6 CIDR
                    </div>
                  </bk-input>
                </div>
              </bk-form-item>
              <bk-form-item
                label="Pod IP"
                property="networkSettings.clusterCIDR"
                error-display-type="normal"
                required>
                <div class="flex">
                  <bk-input
                    class="mr-[8px] max-w-[50%]"
                    v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)"
                    v-model.trim="networkConfig.networkSettings.clusterIPv4CIDR">
                    <div
                      slot="prepend"
                      class="text-[12px] px-[12px] leading-[28px] whitespace-nowrap">
                      IPv4 CIDR
                    </div>
                  </bk-input>
                  <bk-input
                    class="max-w-[50%]"
                    v-model.trim="networkConfig.networkSettings.clusterIPv6CIDR"
                    v-if="['ipv6', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
                    <div
                      slot="prepend"
                      class="text-[12px] px-[12px] leading-[28px] whitespace-nowrap">
                      IPv6 CIDR
                    </div>
                  </bk-input>
                </div>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.maxNodePodNum')"
                property="networkSettings.maxNodePodNum"
                error-display-type="normal"
                required>
                <bcs-select searchable class="mr-[8px] max-w-[50%]" v-model="networkConfig.networkSettings.maxNodePodNum">
                  <bcs-option v-for="item in nodePodNumList" :id="item" :key="item" :name="item"></bcs-option>
                </bcs-select>
              </bk-form-item>
            </DescList>
          </bk-form>
        </bcs-tab-panel>
        <!-- Master配置 -->
        <bcs-tab-panel :name="steps[2].name" :disabled="steps[2].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.detail.title.controlConfig')"
              :step-num="3"
              :active="activeTabName === steps[2].name"
              :disabled="steps[2].disabled" />
          </template>
          <bk-form :ref="steps[2].formRef" :model="masterConfig" :rules="masterConfigRules">
            <bk-form-item
              key="master"
              class="tips-offset"
              :label="$t('cluster.create.label.hostResource')"
              property="master"
              error-display-type="normal"
              required>
              <IpSelector
                :region="basicInfo.region"
                :cloud-id="basicInfo.provider"
                :disabled-ip-list="nodesConfig.nodes.map(item => ({
                  ip: item.bk_host_innerip,
                  tips: $t('cluster.create.validate.ipExitInNode')
                }))"
                :region-list="regionList"
                v-model="masterConfig.master"
                class="max-w-[80%]"
                :validate-vpc-and-region="false"
                @change="validateMaster" />
            </bk-form-item>
            <bk-form-item label="Kube-apiserver">
              <KubeApiServer :enabled="masterConfig.master.length >= 3" />
            </bk-form-item>
          </bk-form>
        </bcs-tab-panel>
        <!-- 添加节点 -->
        <bcs-tab-panel :name="steps[3].name" :disabled="steps[3].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.nodeList.create.text')"
              :step-num="4"
              :active="activeTabName === steps[3].name"
              :disabled="steps[3].disabled" />
          </template>
          <bk-alert type="info">
            <template #title>
              <div>{{ $t('cluster.create.msg.independentClusterInfo.text') }}</div>
              <bk-checkbox v-model="skipAddNodes" class="text-[12px] mt5">
                <span class="text-[12px]">{{ $t('cluster.create.msg.independentClusterInfo.skip') }}</span>
              </bk-checkbox>
            </template>
          </bk-alert>
          <bk-form :ref="steps[3].formRef" :model="nodesConfig" class="mt-[16px]">
            <bk-form-item
              :label="$t('cluster.create.label.hostResource')"
              class="tips-offset"
              property="nodes"
              error-display-type="normal">
              <IpSelector
                :region="basicInfo.region"
                :cloud-id="basicInfo.provider"
                :disabled="skipAddNodes"
                :disabled-ip-list="masterConfig.master.map(item => ({
                  ip: item.bk_host_innerip,
                  tips: $t('cluster.create.validate.ipExitInMaster')
                }))"
                :region-list="regionList"
                v-model="nodesConfig.nodes"
                class="max-w-[80%]"
                :validate-vpc-and-region="false"
                @change="validateNodes" />
            </bk-form-item>
          </bk-form>
        </bcs-tab-panel>
      </bcs-tab>
      <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] absolute bottom-0 w-full bcs-border-top">
        <bk-button v-if="activeTabName !== steps[0].name" @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
        <bk-button
          theme="primary"
          class="ml10"
          v-if="activeTabName === steps[steps.length - 1].name"
          @click="handleShowConfirmDialog">
          {{ $t('cluster.create.button.createCluster') }}
        </bk-button>
        <bk-button theme="primary" class="ml10" v-else @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
        <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts" setup>
import { merge } from 'lodash';
import { computed, getCurrentInstance, onMounted, ref, watch } from 'vue';

import KeyValue from '../components/key-value.vue';

import IpSelector from './common/ip-selector.vue';
import StepTabLabel from './common/step-tab-label.vue';
import KubeApiServer from './form/kube-api-server.vue';
import { ICloudRegion } from './tencent/types';

import { cloudDetail, cloudRegionByAccount, cloudVersionModules, createCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import { validateCIDR } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import DescList from '@/components/desc-list.vue';
import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

const $cloudId = 'bluekingCloud';
const steps = ref([
  { name: 'basicInfo', formRef: 'basicInfoRef', disabled: false },
  { name: 'network', formRef: 'networkRef', disabled: true },
  { name: 'master', formRef: 'masterRef', disabled: true },
  { name: 'nodes', formRef: 'nodesRef', disabled: true },
]);
const nodePodNumList = ref([32, 64, 128, 256]);
const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');

// 基本信息
const basicInfo = ref({
  clusterName: '',
  environment: '',
  provider: '',
  clusterBasicSettings: {
    version: '',
    OS: '',
  },
  description: '',
  region: '',
  labels: {},
  clusterAdvanceSettings: {
    containerRuntime: '', // 运行时
    runtimeVersion: '', // 运行时版本
    enableHa: false, // 是否开启高可用
  },
});
const basicInfoRules = ref({
  clusterName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterBasicSettings.version': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  region: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  environment: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterBasicSettings.OS': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterAdvanceSettings.containerRuntime': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterAdvanceSettings.runtimeVersion': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  labels: [
    {
      message: $i18n.t('generic.validate.label'),
      trigger: 'custom',
      validator: () => {
        const { labels } = basicInfo.value;
        const rule = new RegExp(LABEL_KEY_REGEXP);
        return Object.keys(labels).every(key => rule.test(key) && rule.test(labels[key]));
      },
    },
  ],
});
// 网络配置
const networkConfig = ref({
  vpcID: 'default',
  networkType: 'overlay',
  networkSettings: {
    clusterIPv4CIDR: '10.244.0.0/16',
    clusterIPv6CIDR: 'fd00::1234:5678:100:0/104',
    serviceIPv4CIDR: '10.96.0.0/12',
    serviceIPv6CIDR: 'fd00::1234:5678:1:0/112',
    maxNodePodNum: 64, // 单节点pod数量上限
    clusterIpType: 'ipv4', // ipv4/ipv6/dual
  },
  clusterAdvanceSettings: {
    networkType: 'flannel',   // 网络插件
  },
});
const networkConfigRules = ref({
  'clusterAdvanceSettings.networkType': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'networkSettings.clusterIpType': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'networkSettings.clusterCIDR': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
      validator: () => {
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv4') {
          return !!networkConfig.value.networkSettings.clusterIPv4CIDR;
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv6') {
          return !!networkConfig.value.networkSettings.clusterIPv6CIDR;
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'dual') {
          return !!networkConfig.value.networkSettings.clusterIPv4CIDR
            && !!networkConfig.value.networkSettings.clusterIPv6CIDR;
        }
        return false;
      },
    },
    {
      message: $i18n.t('generic.validate.cidr'),
      trigger: 'custom',
      validator: () => {
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv4') {
          return validateCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR);
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv6') {
          return validateCIDR(networkConfig.value.networkSettings.clusterIPv6CIDR, true);
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'dual') {
          return validateCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR)
            && validateCIDR(networkConfig.value.networkSettings.clusterIPv6CIDR, true);
        }
        return false;
      },
    },
  ],
  'networkSettings.serviceCIDR': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
      validator: () => {
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv4') {
          return !!networkConfig.value.networkSettings.serviceIPv4CIDR;
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv6') {
          return !!networkConfig.value.networkSettings.serviceIPv6CIDR;
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'dual') {
          return !!networkConfig.value.networkSettings.serviceIPv4CIDR
            && !!networkConfig.value.networkSettings.serviceIPv6CIDR;
        }
        return false;
      },
    },
    {
      message: $i18n.t('generic.validate.cidr'),
      trigger: 'custom',
      validator: () => {
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv4') {
          return validateCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR);
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'ipv6') {
          return validateCIDR(networkConfig.value.networkSettings.serviceIPv6CIDR, true);
        }
        if (networkConfig.value.networkSettings.clusterIpType === 'dual') {
          return validateCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR)
            && validateCIDR(networkConfig.value.networkSettings.serviceIPv6CIDR, true);
        }
        return false;
      },
    },
  ],
  'networkSettings.maxNodePodNum': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
});
// master配置
const masterConfig = ref<{
  master: any[]
}>({
  master: [],
});
// 动态 i18n 问题，这里使用computed
const masterConfigRules = computed(() => ({
  master: [{
    message: basicInfo.value.environment === 'debug'
      ? $i18n.t('cluster.create.validate.masterNum135')
      : $i18n.t('cluster.create.validate.masterNum35'),
    trigger: 'custom',
    validator: () => {
      const maxMasterNum = basicInfo.value.environment === 'debug' ? [1, 3, 5] : [3, 5];
      return masterConfig.value.master.length && maxMasterNum.includes(masterConfig.value.master.length);
    },
  }],
}));
// 节点配置
const nodesConfig = ref<{
  nodes: any[]
}>({
  nodes: [],
});
// 动态 i18n 问题，这里使用computed
// const nodesConfigRules = computed(() => ({
//   nodes: [{
//     message: $i18n.t('generic.validate.required'),
//     trigger: 'custom',
//     validator: () => !!nodesConfig.value.nodes.length,
//   }],
// }));

const skipAddNodes = ref(false);

// 集群模板
const templateDetail = ref<Partial<ICloudTemplateDetail>>({});
// 操作系统
const osList = computed(() => templateDetail.value?.osManagement?.availableVersion || []);

const templateLoading = ref(false);
const handleGetTemplateList = async () => {
  templateLoading.value = true;
  templateDetail.value = await cloudDetail({
    $cloudId,
  }).catch(() => []);
  basicInfo.value.provider = templateDetail.value?.cloudID || '';
  // 初始化默认集群版本
  basicInfo.value.clusterBasicSettings.version = templateDetail.value?.clusterManagement?.availableVersion?.[0] || '';
  // 初始化默认操作系统
  basicInfo.value.clusterBasicSettings.OS = osList.value[0];
  templateLoading.value = false;
};

// 版本列表
const versionList = computed(() => templateDetail.value?.clusterManagement?.availableVersion || []);

// 区域列表
const regionLoading = ref(false);
const regionList = ref<Array<ICloudRegion>>([]);
const handleGetRegionList = async () => {
  regionLoading.value = true;
  regionList.value = await cloudRegionByAccount({
    $cloudId,
  }).catch(() => []);
  regionLoading.value = false;
};

// 运行时组件变更
const handleRuntimeChange = (flagName: string) => {
  basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParamsMap.value[flagName]?.defaultValue;
};

// 运行时组件参数
watch(() => basicInfo.value.clusterBasicSettings.version, () => {
  getRuntimeModuleParams();
});
const moduleLoading = ref(false);
const runtimeModuleParams = ref<IRuntimeModuleParams[]>([]);
const runtimeModuleParamsMap = computed<Record<string, IRuntimeModuleParams>>(() => runtimeModuleParams.value
  .reduce((pre, item) => {
    pre[item.flagName] = item;
    return pre;
  }, {}));
const runtimeVersionList = computed(() => {
  // 运行时版本
  const params = runtimeModuleParams.value
    .find(item => item.flagName === basicInfo.value.clusterAdvanceSettings.containerRuntime);
  return params?.flagValueList;
});
const getRuntimeModuleParams = async () => {
  if (!basicInfo.value.clusterBasicSettings.version) return;
  moduleLoading.value = true;
  const data = await cloudVersionModules({
    $cloudId,
    $version: basicInfo.value.clusterBasicSettings.version,
    $module: 'runtime',
  });
  runtimeModuleParams.value = data.filter(item => item.enable);
  // 初始化默认运行时
  basicInfo.value.clusterAdvanceSettings.containerRuntime = runtimeModuleParams.value?.[0]?.flagName || '';
  basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParams.value?.[0]?.defaultValue || '';
  moduleLoading.value = false;
};

// 当前集群版本支持的IP类型
const supportNetworkType = computed(() => {
  const type = runtimeModuleParamsMap.value[basicInfo.value.clusterAdvanceSettings.containerRuntime]?.networkType;
  return type ? type?.split(',') : ['ipv4', 'ipv6', 'dual'];
});

watch(supportNetworkType, () => {
  networkConfig.value.networkSettings.clusterIpType = 'ipv4';
});

// 上一步
const preStep = async () => {
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index - 1 >= 0) {
    activeTabName.value = steps.value[index - 1]?.name;
  }
};
// 下一步
const nextStep = async () => {
  const $refs = proxy?.$refs || {};
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  const validate = await ($refs[steps.value[index]?.formRef] as any)?.validate().catch(() => false);
  if (!validate) return;
  if (index > -1 && index + 1 < steps.value.length) {
    steps.value[index + 1].disabled = false;
    activeTabName.value = steps.value[index + 1]?.name;
  }
};
const handleCancel = () => {
  $router.back();
};
// 创建集群
const { proxy } = getCurrentInstance() || { proxy: null };
const handleShowConfirmDialog = async () => {
  const $refs = proxy?.$refs || {};
  const validateList = steps.value.map(step => ($refs[step.formRef] as any)?.validate().catch(() => {
    activeTabName.value = step.name;
    return false;
  }));
  const validateResults = await Promise.all(validateList);
  if (validateResults.some(result => !result)) return;

  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('cluster.create.button.confirmCreateCluster.text'),
    defaultInfo: true,
    confirmFn: async () => {
      await handleCreateCluster();
    },
  });
};
const { curProject } = useProject();
const user = computed(() => $store.state.user);
const handleCreateCluster = async () => {
  // 大于3台自动开启高可用
  basicInfo.value.clusterAdvanceSettings.enableHa = masterConfig.value.master.length >= 3;
  const params: Record<string, any> = merge({
    projectID: curProject.value.projectID,
    businessID: String(curProject.value.businessID),
    engineType: 'k8s',
    isExclusive: true,
    clusterType: 'single',
    creator: user.value.username,
    manageType: 'INDEPENDENT_CLUSTER',
    master: masterConfig.value.master.map(item => item.bk_host_innerip),
    nodes: !skipAddNodes.value ? nodesConfig.value.nodes.map(item => item.bk_host_innerip) : [],
  }, basicInfo.value, networkConfig.value);

  const result = await createCluster(params).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.deliveryTask'),
    });
    $router.push({ name: 'clusterMain' });
  }
};

// 校验master节点
const validateMaster = async () => {
  const $refs = proxy?.$refs || {};
  proxy?.$forceUpdate();
  return await ($refs.masterRef as any)?.validate().catch(() => false);
};
// 校验node节点
const validateNodes = async () => {
  const $refs = proxy?.$refs || {};
  return await ($refs.nodesRef as any)?.validate().catch(() => false);
};

// 环境变更后重新校验master节点数量
watch(() => basicInfo.value.environment, () => {
  validateMaster();
});

onMounted(() => {
  handleGetTemplateList();
  handleGetRegionList();
});
</script>
<style lang="postcss" scoped>
>>> .bk-tab-header {
  padding: 0 8px;
}
>>> .bk-tab-section {
  overflow: auto;
  height: calc(100% - 80px);
}

>>> .k8s-form .bk-form-content {
  max-width: 600px;
  padding-right: 24px;
}
</style>
