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
          <bcs-option id="GR" name="Global Router"></bcs-option>
          <bcs-option id="VPC-CNI" name="VPC-CNI"></bcs-option>
        </bcs-select>
        <div id="netPlugin">
          <div>{{ $t('tke.tips.grDesc') }}</div>
          <div>{{ $t('tke.tips.vpcCniDesc') }}</div>
          <div>
            <i18n path="tke.button.goDetail">
              <span
                class="text-[12px] text-[#699DF4] cursor-pointer"
                @click="openLink('https://cloud.tencent.com/document/product/457/50353')">
                {{ $t('tke.link.netOverview') }}
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
          <bk-radio value="dual" disabled>
            <span v-bk-tooltips="$t('tke.tips.notSupport')">{{ $t('cluster.create.label.clusterIPType.dual') }}</span>
          </bk-radio>
          <bk-radio value="dual-single" disabled>
            <span v-bk-tooltips="$t('tke.tips.notSupport')">
              {{ $t('cluster.create.label.clusterIPType.dualSingle') }}
            </span>
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
    </DescList>
    <DescList class="mt-[24px]" size="middle" :title="$t('cluster.create.label.netSetting')">
      <!-- GR网络模式 -->
      <template v-if="networkConfig.clusterAdvanceSettings.networkType === 'GR'">
        <bk-form-item
          :label="$t('tke.label.containerNet')"
          :desc="{
            allowHTML: true,
            content: '#networkCIDR',
          }"
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
          <div id="networkCIDR">
            <div>{{ $t('tke.validate.minIpNum', [4096]) }}</div>
            <div>{{ $t('tke.tips.supportCidr') }}</div>
            <ul>
              <li>- 10.0.0.0/8</li>
              <li>- 172.16.0.0/16 ~ 172.31.0.0/16</li>
              <li>- 192.168.0.0/16</li>
            </ul>
          </div>
        </bk-form-item>
        <bk-form-item
          label="Service IP"
          :desc="$t('tke.tips.ipCannotBeAdjustedWhenCreated', ['Service'])"
          property="serviceIP"
          error-display-type="normal"
          required
          key="GR-Service">
          <div class="flex items-center">
            <template v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
              <div class="flex-1 flex max-w-[50%]">
                <span class="prefix">{{ $t('tke.label.ipNum') }}</span>
                <bcs-select
                  class="flex-1 ml-[-1px]"
                  :clearable="false"
                  searchable
                  v-model="networkConfig.networkSettings.maxServiceNum">
                  <bcs-option v-for="item in serviceIPList" :key="item" :id="item" :name="item"></bcs-option>
                </bcs-select>
              </div>
            </template>
          </div>
        </bk-form-item>
        <bk-form-item
          label="Pod IP"
          :desc="$t('tke.tips.podIpCalcFormula')"
          property="podIP"
          error-display-type="normal"
          required
          key="GR-Pod"
          ref="podIpRef">
          <template v-if="['ipv4', 'dual'].includes(networkConfig.networkSettings.clusterIpType)">
            <div class="flex flex-1 max-w-[50%]">
              <!-- GR模式时 pod ip = 网段总数 减去 service数量 -->
              <span class="prefix">{{ $t('tke.label.ipNum') }}</span>
              <bk-input
                class="ml-[-1px] flex-1"
                disabled
                :value="maxPodNum">
              </bk-input>
            </div>
          </template>
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
      </template>
      <!-- VPC-CNI模式 -->
      <template v-else-if="networkConfig.clusterAdvanceSettings.networkType === 'VPC-CNI'">
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
              @change="handleSetSubnetSourceNew" />
          </template>
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.staticIpMode')"
          :desc="{
            allowHTML: true,
            content: '#staticIpMode',
          }"
          key="staticIpMode">
          <bk-checkbox v-model="networkConfig.networkSettings.isStaticIpMode">
            {{ $t('tke.label.enableStaticIpMode') }}
          </bk-checkbox>
          <div id="staticIpMode">
            <div>{{ $t('tke.tips.staticIpMode.p1') }}</div>
            <i18n tag="div" path="tke.tips.staticIpMode.p2">
              <bk-link theme="primary" target="_blank" href="https://cloud.tencent.com/document/product/457/34994">
                <span class="text-[12px]">{{ $t('tke.link.staticIpMode') }}</span>
              </bk-link>
            </i18n>
          </div>
        </bk-form-item>
      </template>
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
import Vpc from '@/views/cluster-manage/add/components/vpc.vue';
import VpcCni from '@/views/cluster-manage/components/vpc-cni.vue';

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

const nodePodNumList = ref([32, 64, 128]);
const serviceIPList = ref([128, 256, 512, 1024, 2048, 4096]);

// 网络配置
const networkConfig = ref({
  vpcID: '',
  networkType: '', // overlay underlay
  networkSettings: {
    clusterIPv4CIDR: '',
    serviceIPv4CIDR: '',
    maxNodePodNum: 64, // 单节点pod数量上限
    maxServiceNum: 128,
    clusterIpType: 'ipv4', // ipv4/ipv6/dual
    isStaticIpMode: false,
    subnetSource: {
      new: [],
    },
  },
  clusterAdvanceSettings: {
    networkType: 'GR',   // 网络插件
  },
});
const countsClusterIPv4CIDR = computed(() => countIPsInCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR) || 0);
const maxPodNum = computed(() => {
  const counts = countsClusterIPv4CIDR.value - networkConfig.value.networkSettings.maxServiceNum;
  return counts >= 0 ? counts : 0;
});
watch(() => props.region, () => {
  networkConfig.value.networkSettings.subnetSource.new = [];
});
watch(() => networkConfig.value.clusterAdvanceSettings.networkType, () => {
  if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
    networkConfig.value.networkSettings.maxServiceNum = 128;
    networkConfig.value.networkSettings.serviceIPv4CIDR = '';
    networkConfig.value.networkSettings.subnetSource.new = [];
  } else if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
    networkConfig.value.networkSettings.maxServiceNum = 0;
    networkConfig.value.networkSettings.clusterIPv4CIDR = '';
  }
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
  'networkSettings.maxNodePodNum': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  // 容器网段（GR网络插件模式）
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
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          return networkConfig.value.networkSettings.clusterIPv4CIDR
              && validateCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR);
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minIpNum2', [4096]),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          const counts = countIPsInCIDR(networkConfig.value.networkSettings.clusterIPv4CIDR) || 0;
          return counts >= 4096;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.cidrMaskLen', ['[10, 20]']),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          const cidr = networkConfig.value.networkSettings.clusterIPv4CIDR;
          const mask = Number(cidr?.split('/')?.[1] || 0);
          return mask >= 10 && mask <= 20;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.supportCidrList', ['10.0.0.0/8, 172.16.0.0/16 ~ 172.31.0.0/16, 192.168.0.0/16']),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          const cidr = networkConfig.value.networkSettings.clusterIPv4CIDR;
          return cidrContains(cidr, '10.0.0.0/8')
            || cidrContains(cidr, ['172.16.0.0/16', '172.31.0.0/16'])
            || cidrContains(cidr, '192.168.0.0/16');
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: () => $i18n.t('generic.validate.cidrConflict', { cidr: conflictCIDR.value }),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          return networkConfig.value.networkSettings.clusterIPv4CIDR
              && await cidrConflict(networkConfig.value.networkSettings.clusterIPv4CIDR);
        }
        return true;
      },
    },
  ],
  // service IP
  serviceIP: [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'GR') {
          return !!networkConfig.value.networkSettings.maxServiceNum;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.cidr'),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          return networkConfig.value.networkSettings.serviceIPv4CIDR
            && validateCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR);
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minIpNum2', [128]),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          const counts = countIPsInCIDR(networkConfig.value.networkSettings.serviceIPv4CIDR) || 0;
          return counts >= 128;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.cidrMaskLen', ['[10, 24]']),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          const cidr = networkConfig.value.networkSettings.serviceIPv4CIDR;
          const mask = Number(cidr?.split('/')?.[1] || 0);
          return mask >= 10 && mask <= 24;
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.supportCidrList', ['10.0.0.0/8, 172.16.0.0/16 ~ 172.31.0.0/16, 192.168.0.0/16']),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          const cidr = networkConfig.value.networkSettings.serviceIPv4CIDR;
          return cidrContains(cidr, '10.0.0.0/8')
            || cidrContains(cidr, ['172.16.0.0/16', '172.31.0.0/16'])
            || cidrContains(cidr, '192.168.0.0/16');
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: () => $i18n.t('generic.validate.cidrConflict', { cidr: conflictCIDR.value }),
      async validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          return networkConfig.value.networkSettings.serviceIPv4CIDR
            && await cidrConflict(networkConfig.value.networkSettings.serviceIPv4CIDR);
        }
        return true;
      },
    },
  ],
  // POD IP数量
  podIP: [
    {
      trigger: 'blur',
      message: $i18n.t('generic.validate.required'),
      validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
            ipCnt: number
            zone: string
          }>;
          return subnetSource.length && subnetSource.every(item => item.ipCnt && item.zone);
        }
        return true;
      },
    },
    {
      trigger: 'blur',
      message: $i18n.t('tke.validate.minPodIP'),
      validator() {
        if (networkConfig.value.clusterAdvanceSettings.networkType === 'VPC-CNI') {
          return true;
        }
        return maxPodNum.value > 0;
      },
    },
    {
      trigger: 'custom',
      message: $i18n.t('tke.validate.podIPsNeedLessThanVpcIPs'),
      validator() {
        const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
          ipCnt: number
          zone: string
        }>;
        const counts = subnetSource.reduce((counts, item) => {
          counts += item.ipCnt;
          return counts;
        }, 0);

        return counts <= (vpcDetail.value.allocateIpNum || 0);
      },
    },
  ],
});
const podIpRef = ref();
const validatePodIP = () => {
  podIpRef.value?.validate('blur');
};
watch(maxPodNum, () => {
  validatePodIP();
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
const validating = ref(false);
const nextStep = async () => {
  validating.value = true;
  const result = await validate();
  validating.value = false;
  if (result) {
    emits('next', {
      ...networkConfig.value,
      networkType: networkConfig.value.clusterAdvanceSettings.networkType === 'GR' ? 'overlay' : 'underlay',
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
