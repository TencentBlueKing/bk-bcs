<template>
  <div class="text-[14px]">
    <span class="inline-flex">
      <bk-checkbox
        v-model="internetAccess.publicIPAssigned">
        {{$t('tke.label.publicIPAssigned.text')}}
      </bk-checkbox>
    </span>
    <div class="bg-[#F5F7FA] py-[16px] px-[24px] mt-[10px]" v-if="internetAccess.publicIPAssigned">
      <div class="flex items-center h-[32px]">
        <label class="inline-flex w-[80px]">
          {{$t('tke.label.publicIPAssigned.chargeMode.text')}}
        </label>
        <bk-radio-group
          v-model="internetAccess.internetChargeType"
          @change="handleChargeTypeChange">
          <bk-radio :disabled="accountType === 'LEGACY'" value="TRAFFIC_POSTPAID_BY_HOUR">
            <span
              v-bk-tooltips="{
                content: $t('ca.internetAccess.tips.accountNotAvailable'),
                disabled: accountType !== 'LEGACY'
              }">
              {{$t('tke.label.publicIPAssigned.chargeMode.traffic_postpaid_by_hour')}}
            </span>
          </bk-radio>
          <bk-radio :disabled="accountType === 'LEGACY'" value="BANDWIDTH_PREPAID">
            <span
              v-bk-tooltips="{
                content: $t('ca.internetAccess.tips.accountNotAvailable'),
                disabled: accountType !== 'LEGACY'
              }">
              {{$t('tke.label.publicIPAssigned.chargeMode.bandwidth_prepaid')}}
            </span>
          </bk-radio>
          <bk-radio value="BANDWIDTH_PACKAGE">
            {{ $t('ca.internetAccess.bandwidthPackage') }}
          </bk-radio>
        </bk-radio-group>
      </div>
      <div
        class="flex items-center h-[32px] mt10"
        v-if="accountType === 'STANDARD'
          && internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE'">
        <label class="inline-flex w-[80px]">{{ $t('ca.internetAccess.label.selectBandwidthPackage') }}</label>
        <bcs-select
          :loading="bandwidthLoading"
          class="w-[200px] bg-[#fff]"
          v-model="internetAccess.bandwidthPackageId">
          <bcs-option
            v-for="item in bandwidthList"
            :key="item.id"
            :id="item.id"
            :name="`${item.name}(${item.id})`" />
          <template #extension>
            <SelectExtension
              :link-text="$t('tke.link.package')"
              link="https://console.cloud.tencent.com/vpc/package"
              @refresh="getBandwidthList" />
          </template>
        </bcs-select>
      </div>
      <div class="flex items-center h-[32px] mt10">
        <label class="inline-flex w-[80px]">{{$t('tke.label.publicIPAssigned.maxBandWidth')}}</label>
        <bk-radio-group v-model="maxBandwidthType">
          <bk-radio value="limit">
            <div class="flex items-center">
              <span class="mr-[8px]">{{ $t('ca.internetAccess.label.limit') }}</span>
              <bcs-input
                class="min-w-[80px] flex-1"
                type="number"
                :min="1"
                :max="internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE' ? 2000 : 100"
                v-model="internetAccess.internetMaxBandwidth">
                <span slot="append" class="group-text !px-[4px]">Mbps</span>
              </bcs-input>
            </div>
          </bk-radio>
          <bk-radio
            value="un-limit"
            v-if="internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE'">
            <span>{{ $t('ca.internetAccess.label.unLimit') }}</span>
          </bk-radio>
        </bk-radio-group>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, PropType, ref, watch } from 'vue';

import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import { IInternetAccess, InternetChargeType } from '@/views/cluster-manage/add/tencent/types';
import { useCloud } from '@/views/cluster-manage/cluster/use-cluster';

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
  value: {
    type: Object as PropType<IInternetAccess>,
    default: () => ({}),
  },
});
const emits = defineEmits(['change', 'account-type-change']);

const internetAccess = ref<IInternetAccess>({
  publicIPAssigned: false,
  internetChargeType: 'TRAFFIC_POSTPAID_BY_HOUR',
  bandwidthPackageId: '',
  internetMaxBandwidth: '0',
});
watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  internetAccess.value = Object.assign({
    publicIPAssigned: false,
    internetChargeType: 'TRAFFIC_POSTPAID_BY_HOUR',
    bandwidthPackageId: '',
    internetMaxBandwidth: '0',
  }, props.value);
}, { immediate: true });

watch(internetAccess, () => {
  // 后端需要字符串类型
  internetAccess.value.internetMaxBandwidth = String(internetAccess.value.internetMaxBandwidth);
  emits('change', internetAccess.value);
}, { deep: true });

// 账户类型
const { accountType, getCloudAccountType, getCloudBwps } = useCloud();
watch(accountType, () => {
  emits('account-type-change', accountType.value);
});

// 免费分配公网IP
watch(() => internetAccess.value.publicIPAssigned, (publicIPAssigned) => {
  internetAccess.value.internetMaxBandwidth = publicIPAssigned ? '10' : '0';
});

// 带宽包
const maxBandwidth = 65535;
const maxBandwidthType = ref<'limit'|'un-limit'>(Number(internetAccess.value.internetMaxBandwidth) === maxBandwidth
  ? 'un-limit'
  : 'limit');
const bandwidthLoading = ref(false);
const bandwidthList = ref<any[]>([]);
const getBandwidthList = async () => {
  if (!props.cloudAccountID || !props.cloudID || !props.region) return;
  bandwidthLoading.value = true;
  bandwidthList.value = await getCloudBwps({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  });
  bandwidthLoading.value = false;
};
// 账户类型
const handleGetCloudAccountType = async () => {
  if (!props.cloudAccountID || !props.cloudID) return;
  await getCloudAccountType({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
  });
  if (accountType.value === 'LEGACY') { // 传统账户只能选择带宽包
    internetAccess.value.internetChargeType = 'BANDWIDTH_PACKAGE';
  }
};

const handleChargeTypeChange = (value: InternetChargeType) => {
  if (value !== 'BANDWIDTH_PACKAGE' && Number(internetAccess.value.internetMaxBandwidth) > 100) {
    // 'TRAFFIC_POSTPAID_BY_HOUR' | 'BANDWIDTH_PREPAID' 类型最大带宽为 100
    internetAccess.value.internetMaxBandwidth = '100';
  }
  if (value !== 'BANDWIDTH_PACKAGE' && maxBandwidthType.value === 'un-limit') {
    maxBandwidthType.value = 'limit';
  }
};

watch([
  () => props.cloudAccountID,
  () => props.cloudID,
  () => props.region,
], () => {
  handleGetCloudAccountType();
  getBandwidthList();
});

onBeforeMount(() => {
  handleGetCloudAccountType();
  getBandwidthList();
});
</script>
