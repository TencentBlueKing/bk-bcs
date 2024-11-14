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
            :show-link="false"
            :value-key="'name'"
            v-model="networkConfig.vpcID"
            @change="handleGetVpcDetail" />
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.subnet')"
        property="clusterAdvanceSettings.clusterConnectSetting.subnetId"
        error-display-type="normal"
        required>
        <Subnet
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :region="region"
          :vpc-id="networkConfig.vpcID"
          v-model="networkConfig.clusterAdvanceSettings.clusterConnectSetting.subnetId" />
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
        </bk-radio-group>
      </bk-form-item>
    </DescList>
    <DescList class="mt-[24px]" size="middle" :title="$t('cluster.create.label.netSetting')">
      <bk-form-item
        :label="$t('tke.label.containerNet')"
        :desc="$t('cluster.create.google.tips.containerNet')"
        property="networkSettings.clusterIPv4CIDR"
        error-display-type="normal"
        required>
        <div class="flex items-center">
          <div class="flex flex-1 max-w-[50%]">
            <span class="prefix">CIDR</span>
            <bk-input
              class="ml-[-1px] flex-1"
              :placeholder="$t('tke.placeholder.example', ['172.16.0.0/20'])"
              v-model.trim="networkConfig.networkSettings.clusterIPv4CIDR">
            </bk-input>
          </div>
          <span
            class="inline-flex items-center px-[16px] h-[24px] rounded-full bg-[#F0F1F5] text-[12px] ml-[16px]">
            {{ $t('tke.tips.totalIpNum', [countsClusterIPv4CIDR || '--']) }}
          </span>
        </div>
      </bk-form-item>
      <bk-form-item
        label="Service"
        :desc="$t('tke.tips.ipCannotBeAdjustedWhenCreated', ['Service'])"
        property="serviceIP"
        error-display-type="normal"
        required
        key="VPC-CNI-Service">
        <div class="flex items-center">
          <template v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
            <div class="flex flex-1 max-w-[50%]">
              <span class="prefix">CIDR</span>
              <bk-input
                class="ml-[-1px] flex-1"
                :placeholder="$t('tke.placeholder.example', ['172.16.0.0/20'])"
                v-model.trim="networkConfig.networkSettings.serviceIPv4CIDR">
              </bk-input>
            </div>
            <span
              class="inline-flex items-center px-[16px] h-[24px] rounded-full bg-[#F0F1F5] text-[12px] ml-[16px]">
              {{ $t('tke.tips.totalIpNum', [countIPsInCIDR(networkConfig.networkSettings.serviceIPv4CIDR) || 0]) }}
            </span>
          </template>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.maxNodePodNum')"
        :desc="$t('tke.tips.ipCannotBeAdjustedWhenCreated2', [$t('cluster.create.label.maxNodePodNum')])"
        property="networkSettings.maxNodePodNum"
        error-display-type="normal"
        required>
        <bcs-select
          searchable
          :clearable="false"
          class="mr-[8px] max-w-[50%]"
          v-model="networkConfig.networkSettings.maxNodePodNum">
          <bcs-option v-for="item in nodePodNumList" :id="item" :key="item" :name="item"></bcs-option>
        </bcs-select>
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
import { computed, ref, watch } from 'vue';

import { IVpcItem } from '../../../types/types';

import { cloudCidrconflict, cloudVPC } from '@/api/modules/cluster-manager';
import { cidrContains, countIPsInCIDR, validateCIDR } from '@/common/util';
import DescList from '@/components/desc-list.vue';
import $i18n from '@/i18n/i18n-setup';
import Region from '@/views/cluster-manage/add/components/region.vue';
import Subnet from '@/views/cluster-manage/add/components/subnet.vue';
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
  vpcId: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['next', 'cancel', 'pre']);

const nodePodNumList = ref([32, 64, 128]);

// 网络配置
const networkConfig = ref({
  vpcID: '',
  networkType: '', // overlay underlay
  networkSettings: {
    clusterIPv4CIDR: '',
    serviceIPv4CIDR: '',
    maxNodePodNum: 64, // 单节点pod数量上限
    clusterIpType: 'ipv4', // ipv4/ipv6/dual
  },
  clusterAdvanceSettings: {
    clusterConnectSetting: {
      subnetId: '',
    },
  },
});
const countsClusterIPv4CIDR = computed(() => countIPsInCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR) || 0);

const conflictCIDR = ref('');
const networkConfigRules = ref({
  vpcID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterAdvanceSettings.clusterConnectSetting.subnetId': [
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
  'networkSettings.maxNodePodNum': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'networkSettings.clusterIPv4CIDR': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.cidr'),
      async validator() {
        return networkConfig.value.networkSettings.clusterIPv4CIDR
            && validateCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR);
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minIpNum2', [4096]),
      async validator() {
        const counts = countIPsInCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR) || 0;
        return counts >= 4096;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.cidrMaskLen', ['[10, 20]']),
      async validator() {
        const cidr = networkConfig.value.networkSettings.clusterIPv4CIDR;
        const mask = Number(cidr?.split('/')?.[1] || 0);
        return mask >= 10 && mask <= 20;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.supportCidrList', ['10.0.0.0/8, 172.16.0.0/16 ~ 172.31.0.0/16, 192.168.0.0/16']),
      async validator() {
        const cidr = networkConfig.value.networkSettings.clusterIPv4CIDR;
        return cidrContains(cidr, '10.0.0.0/8')
          || cidrContains(cidr, ['172.16.0.0/16', '172.31.0.0/16'])
          || cidrContains(cidr, '192.168.0.0/16');
      },
    },
    {
      trigger: 'blur',
      message: () => $i18n.t('generic.validate.cidrConflict', { cidr: conflictCIDR.value }),
      async validator() {
        return networkConfig.value.networkSettings.clusterIPv4CIDR
            && await cidrConflict(networkConfig.value.networkSettings.clusterIPv4CIDR);
      },
    },
  ],
  serviceIP: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
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
      message: $i18n.t('tke.validate.minIpNum2', [4096]),
      async validator() {
        const counts = countIPsInCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR) || 0;
        return counts >= 4096;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.cidrMaskLen', ['[10, 20]']),
      async validator() {
        const cidr = networkConfig.value.networkSettings.serviceIPv4CIDR;
        const mask = Number(cidr?.split('/')?.[1] || 0);
        return mask >= 10 && mask <= 20;
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
});

// VPC详情
const vpcDetail = ref<Partial<IVpcItem>>({});
const handleGetVpcDetail = async (vpcID: string) => {
  // 取数组第 1 个
  const [detail] = await cloudVPC({
    $cloudId: props.cloudID,
    region: props.region,
    accountID: props.cloudAccountID,
    vpcID,
  });
  vpcDetail.value = detail;
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

watch(() => networkConfig.value.vpcID, () => {
  networkConfig.value.clusterAdvanceSettings.clusterConnectSetting.subnetId = '';
});

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
const validating = ref(false);
const nextStep = async () => {
  validating.value = true;
  const result = await validate();
  validating.value = false;
  if (result) {
    emits('next', {
      ...networkConfig.value,
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
<style scoped lang="postcss">
>>> .private-net-form-item .bk-form-content {
  max-width: 800px !important;
}

>>> .unset-form-content-width .bk-form-content {
  max-width: none !important;
}
</style>
