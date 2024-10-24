<template>
  <bk-form
    class="k8s-form"
    :label-width="160"
    :model="networkConfig"
    :rules="networkConfigRules"
    ref="formRef">
    <DescList size="middle" :title="$t('cluster.create.label.basicConfig')">
      <bk-form-item :label="$t('cluster.create.label.region')" required>
        <Region :value="region" :cloud-account-i-d="cloudAccountID" :cloud-i-d="cloudID" disabled />
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.privateNet.text')"
        property="vpcID"
        error-display-type="normal"
        required
        class="private-net-form-item">
        <div class="flex items-center">
          <Vpc
            class="max-w-[576px] flex-1"
            :cloud-account-i-d="cloudAccountID"
            :cloud-i-d="cloudID"
            :region="region"
            init-data
            v-model="networkConfig.vpcID"
            @change="handleGetVpcDetail" />
          <span
            :class="[
              'inline-flex items-center',
              'px-[16px] h-[24px] rounded-full bg-[#F0F1F5] text-[12px] ml-[8px]'
            ]"
            v-if="networkConfig.clusterAdvanceSettings.networkType === 'VPC-CNI'">
            {{ $t('tke.tips.totalIpNum', [vpcDetail.allocateIpNum || '--']) }}
          </span>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.defaultNetPlugin')"
        :desc="{
          allowHTML: true,
          content: '#netPlugin',
        }"
        property="clusterAdvanceSettings.networkType"
        error-display-type="normal"
        required>
        <bcs-select :clearable="false" searchable v-model="networkConfig.clusterAdvanceSettings.networkType">
          <bcs-option id="VPC-CNI" name="VPC-CNI"></bcs-option>
        </bcs-select>
        <div id="netPlugin">
          <div>{{ $t('cluster.create.aws.tips.awsDesc') }}</div>
          <div>
            <i18n path="tke.button.goDetail">
              <span
                class="text-[12px] text-[#699DF4] cursor-pointer"
                @click="openLink('https://docs.aws.amazon.com/zh_cn/eks/latest/userguide/managing-vpc-cni.html')">
                {{ $t('cluster.create.aws.tips.awsLink') }}
              </span>
            </i18n>
          </div>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.clusterIPType.text')"
        property="networkSettings.clusterIpType"
        error-display-type="normal"
        required
        class="unset-form-content-width">
        <bk-radio-group v-model="networkConfig.networkSettings.clusterIpType">
          <bk-radio value="ipv4">
            <span>IPv4</span>
          </bk-radio>
          <bk-radio value="ipv6" disabled>
            <span v-bk-tooltips="$t('tke.tips.notSupport')">IPv6</span>
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
    </DescList>
    <DescList class="mt-[24px]" size="middle" :title="$t('cluster.create.label.netSetting')">
      <!-- VPC-CNI模式 -->
      <bk-form-item
        label="Service"
        :desc="$t('tke.tips.ipCannotBeAdjustedWhenCreated', ['Service'])"
        property="serviceIP"
        error-display-type="normal"
        required
        key="VPC-CNI-Service">
        <div class="flex items-center">
          <template v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
            <div class="flex-1 flex max-w-[50%]">
              <bk-input
                class="mr-[24px] flex-1"
                :placeholder="$t('tke.placeholder.example', ['172.16.0.0/20'])"
                v-model.trim="networkConfig.networkSettings.serviceIPv4CIDR">
                <div
                  slot="prepend"
                  class="text-[12px] px-[12px] leading-[28px] whitespace-nowrap">
                  CIDR
                </div>
              </bk-input>
            </div>
            <span
              class="inline-flex items-center px-[16px] h-[24px] rounded-full bg-[#F0F1F5] text-[12px] ml-[-8px]">
              {{ $t('tke.tips.totalIpNum', [countIPsInCIDR(networkConfig.networkSettings.serviceIPv4CIDR) || 0]) }}
            </span>
          </template>
        </div>
      </bk-form-item>
      <bk-form-item
        label="Pod IP"
        :desc="$t('tke.tips.vpcCniPodIp')"
        property="podIP"
        error-display-type="normal"
        required
        key="VPC-CNI-Pod-IP">
        <template v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
          <VpcCni
            :subnets="networkConfig.networkSettings.subnetSource.new"
            :cloud-account-i-d="cloudAccountID"
            :cloud-i-d="cloudID"
            :region="region"
            value-id="zoneName"
            @change="handleSetSubnetSourceNew" />
        </template>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.aws.securityGroup.text')"
        :desc="{
          allowHTML: true,
          content: '#secPlugin',
        }"
        error-display-type="normal"
        property="clusterAdvanceSettings.clusterConnectSetting.securityGroup"
        required>
        <SecurityGroups
          class="max-w-[600px]"
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          v-model="networkConfig.clusterAdvanceSettings.clusterConnectSetting.securityGroup" />
        <div id="secPlugin">
          <div>
            <i18n path="cluster.create.aws.securityGroup.vpcLink">
              <span
                class="text-[12px] text-[#699DF4] cursor-pointer"
                @click="openLink('https://us-west-1.console.aws.amazon.com/vpcconsole/home?region=us-west-1#SecurityGroups')">
                {{ $t('cluster.create.aws.securityGroup.vpcMaster') }}
              </span>
            </i18n>
          </div>
        </div>
      </bk-form-item>
    </DescList>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button
        :loading="validating"
        theme="primary"
        class="ml10"
        @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import { IVpcItem } from '../../../types/types';

import VpcCni from './vpc-cni.vue';

import { cloudCidrconflict, cloudVPC } from '@/api/modules/cluster-manager';
import { cidrContains, countIPsInCIDR, validateCIDR } from '@/common/util';
import DescList from '@/components/desc-list.vue';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';
import Region from '@/views/cluster-manage/add/components/region.vue';
import SecurityGroups from '@/views/cluster-manage/add/components/security-groups.vue';
import Vpc from '@/views/cluster-manage/add/components/vpc.vue';

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
});

const emits = defineEmits(['next', 'cancel', 'pre']);

// 网络配置
const networkConfig = ref({
  vpcID: '',
  networkType: '', // overlay underlay
  networkSettings: {
    serviceIPv4CIDR: '',
    maxNodePodNum: 64, // 单节点pod数量上限
    maxServiceNum: 128,
    clusterIpType: 'ipv4', // ipv4
    isStaticIpMode: false,
    subnetSource: {
      new: [],
    },
    securityGroupIDs: [], // 安全组
  },
  clusterAdvanceSettings: {
    networkType: 'VPC-CNI',   // 网络插件
    clusterConnectSetting: {
      securityGroup: '',
    },
  },
});

watch(() => props.region, () => {
  networkConfig.value.networkSettings.subnetSource.new = [];
});
watch(() => networkConfig.value.clusterAdvanceSettings.networkType, () => {
  networkConfig.value.networkSettings.maxServiceNum = 0;
});
const conflictCIDR = ref('');
const networkConfigRules = ref({
  vpcID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
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
  'clusterAdvanceSettings.clusterConnectSetting.securityGroup': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  // service IP
  serviceIP: [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
    },
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.cidr'),
      async validator() {
        return networkConfig.value.networkSettings.serviceIPv4CIDR
          && validateCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR);
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minIpNum2', [128]),
      async validator() {
        const counts = countIPsInCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR) || 0;
        return counts >= 128;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.cidrMaskLen', ['[10, 24]']),
      async validator() {
        const cidr = networkConfig.value.networkSettings.serviceIPv4CIDR;
        const mask = Number(cidr?.split('/')?.[1] || 0);
        return mask >= 10 && mask <= 24;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.supportCidrList', ['10.0.0.0/8, 172.16.0.0/16 ~ 172.31.0.0/16, 192.168.0.0/16']),
      async validator() {
        const cidr = networkConfig.value.networkSettings.serviceIPv4CIDR;
        return cidrContains(cidr, '10.0.0.0/8')
          || cidrContains(cidr, ['172.16.0.0/16', '172.31.0.0/16'])
          || cidrContains(cidr, '192.168.0.0/16');
      },
    },
    {
      trigger: 'blur',
      message: () => $i18n.t('generic.validate.cidrConflict', { cidr: conflictCIDR.value }),
      async validator() {
        return networkConfig.value.networkSettings.serviceIPv4CIDR
          && await cidrConflict(networkConfig.value.networkSettings.serviceIPv4CIDR);
      },
    },
  ],
  // POD IP数量
  podIP: [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      validator() {
        const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
          mask: number
          zone: string
        }>;
        return subnetSource.length && subnetSource.every(item => item.mask && item.zone);
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minPodIP'),
    },
    {
      trigger: 'custom',
      message: $i18n.t('tke.validate.podIPsNeedLessThanVpcIPs'),
      validator() {
        const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
          mask: number
          zone: string
        }>;
        const counts = subnetSource.reduce((counts, item) => {
          counts += item.mask;
          return counts;
        }, 0);

        return counts <= (vpcDetail.value.allocateIpNum || 0);
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('tke.validate.leastTwoDifferentAZs'),
      validator() {
        const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
          mask: number
          zone: string
        }>;
        const zones = subnetSource.map(item => item.zone);

        return zones.length > 1 && new Set(zones).size > 1;
      },
    },
  ],
});
// VPC详情
const vpcDetail = ref<Partial<IVpcItem>>({});
const handleGetVpcDetail = async (vpcID: string) => {
  validating.value = true;
  // 取数组第 1 个
  const [detail] = await cloudVPC({
    $cloudId: props.cloudID,
    region: props.region,
    accountID: props.cloudAccountID,
    vpcID,
  });
  vpcDetail.value = detail;
  validating.value = false;
};

// 设置vpc-cni子网
const handleSetSubnetSourceNew = (data) => {
  networkConfig.value.networkSettings.subnetSource.new = data;
};

// CIDR 冲突检测
const cidrConflict = async (cidr) => {
  if (!networkConfig.value.vpcID || !props.region || !props.cloudAccountID) return true;
  const { cidrs = [] } =  await cloudCidrconflict({
    $cloudId: props.cloudID,
    $vpc: networkConfig.value.vpcID,
    region: props.region,
    accountID: props.cloudAccountID,
    cidr,
  }).catch(() => false);
  conflictCIDR.value = cidrs.join(',');
  return !cidrs.length;
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  networkConfig.value.vpcID = '';
}, { immediate: true });

// 跳转链接
const openLink = (link: string) => {
  if (!link) return;

  window.open(link);
};

// 校验
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
const validating = ref(false);
const nextStep = async () => {
  validating.value = true;
  const result = await validate();
  validating.value = false;
  if (result) {
    emits('next', {
      ...networkConfig.value,
      networkType: 'underlay',
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
<style scoped lang="postcss">
>>> .private-net-form-item .bk-form-content {
  max-width: 800px !important;
}

>>> .unset-form-content-width .bk-form-content {
  max-width: none !important;
}
</style>
