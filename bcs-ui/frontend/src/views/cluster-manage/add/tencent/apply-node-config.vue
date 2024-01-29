<template>
  <div class="h-[calc(100vh-60px)]" ref="nodeConfigRef">
    <bk-form
      class="py-[20px] px-[40px] max-h-[calc(100vh-108px)] overflow-auto"
      form-type="vertical"
      :model="instanceItem"
      :rules="rules"
      ref="formRef">
      <bk-form-item
        :label="$t('tke.label.zone')"
        property="zone"
        error-display-type="normal"
        required>
        <Zone
          v-model="instanceItem.zone"
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :vpc-id="vpcId"
          :disabled-tips="$t('tke.tips.zone')"
          :enabled-zone-list="enabledZoneList"
          :init-data="true" />
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.subnet')"
        property="subnetID"
        error-display-type="normal"
        required>
        <bk-select
          searchable
          :clearable="false"
          :loading="subnetLoading && !!instanceItem.zone"
          v-model="instanceItem.subnetID">
          <bk-option
            v-for="net in subnets"
            :key="net.subnetID"
            :id="net.subnetID"
            :name="`${net.subnetName}(${net.subnetID})`"
            :disabled="!net.availableIPAddressCount || !!Object.keys(net.cluster || {}).length">
            <div
              class="flex items-center justify-between"
              v-bk-tooltips="{
                content: Object.keys(net.cluster || {}).length
                  ? $t('tke.tips.subnetInUsed', [net.cluster ? net.cluster.clusterName : ''])
                  : $t('tke.tips.noAvailableIp'),
                disabled: net.availableIPAddressCount && !Object.keys(net.cluster || {}).length,
                placement: 'left'
              }">
              <span>{{ `${net.subnetName}(${net.subnetID})` }}</span>
              <span
                :class="(!net.availableIPAddressCount || Object.keys(net.cluster || {}).length) ? '':'text-[#979BA5]'">
                {{ `${$t('tke.label.availableIpNum')}: ${net.availableIPAddressCount}` }}
              </span>
            </div>
          </bk-option>
          <template slot="extension">
            <SelectExtension
              :link-text="$t('tke.link.subnet')"
              :link="`https://console.cloud.tencent.com/vpc/subnet?unVpcId=${vpcId}`"
              @refresh="handleGetSubnets" />
          </template>
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.instanceType')"
        property="instanceType"
        error-display-type="normal"
        required>
        <div class="flex items-center">
          <span class="prefix">CPU</span>
          <bcs-select v-model="cpu" searchable class="ml-[-1px] w-[140px] mr-[16px]" @change="handleValidateInstance">
            <bcs-option v-for="item in cpuList" :key="item" :id="item" :name="item"></bcs-option>
          </bcs-select>
          <span class="prefix">{{ $t('generic.label.mem') }}</span>
          <bcs-select v-model="mem" searchable class="ml-[-1px] w-[140px]" @change="handleValidateInstance">
            <bcs-option v-for="item in memList" :key="item" :id="item" :name="item"></bcs-option>
          </bcs-select>
        </div>
        <bcs-table
          :data="instanceItem.zone && instanceItem.subnetID ? instanceList : []"
          v-bkloading="{ isLoading: instanceLoading && instanceItem.zone }"
          :pagination="pagination"
          :row-class-name="instanceRowClass"
          class="mt-[16px]"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange"
          @row-click="handleCheckInstanceType"
          @filter-change="handleFilterChange"
          @sort-change="handleSortChange">
          <bcs-table-column
            :label="$t('generic.ipSelector.label.serverModel')"
            :filters="nodeTypeFilters"
            :key="nodeTypeFilters.length"
            column-key="typeName"
            filter-multiple
            prop="typeName">
            <template #default="{ row }">
              <span
                v-bk-tooltips="{
                  disabled: row.status !== 'SOLD_OUT',
                  content: $t('cluster.ca.nodePool.create.instanceTypeConfig.status.soldOut')
                }">
                <bcs-radio
                  class="flex items-center node-config-radio"
                  :value="instanceItem.instanceType === row.nodeType"
                  :disabled="row.status === 'SOLD_OUT'">
                  <span class="bcs-ellipsis">{{row.typeName}}</span>
                </bcs-radio>
              </span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.label.specifications')"
            min-width="90"
            show-overflow-tooltip
            prop="nodeType">
          </bcs-table-column>
          <bcs-table-column label="CPU" prop="cpu" width="90" align="right" sortable>
            <template #default="{ row }">
              <span>{{ `${row.cpu}${$t('units.suffix.cores')}` }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('generic.label.mem')" prop="memory" width="80" align="right" sortable>
            <template #default="{ row }">
              <span>{{ row.memory }}G</span>
            </template>
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.text')"
            prop="unitPrice"
            sortable>
            <template #default="{ row }">
              <span>
                {{ $t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.unit',
                      { price: row.unitPrice })
                }}
              </span>
            </template>
          </bcs-table-column>
        </bcs-table>
        <p
          class="bcs-form-error-tip text-[#ea3636] text-[12px] h-[20px] leading-[20px] mt-[4px]"
          v-if="!instanceItem.instanceType && !firstTrigger">{{ $t('generic.validate.required') }}</p>
      </bk-form-item>
      <!-- 系统盘 -->
      <SystemDisk
        class="mt-[24px]"
        :value="instanceItem.systemDisk"
        @change="(v) => instanceItem.systemDisk = v" />
      <!-- 数据盘 -->
      <bk-form-item
        :label-width="0.1"
        property="cloudDataDisks"
        error-display-type="normal">
        <DataDisk
          class="mt-[20px]"
          :value="instanceItem.cloudDataDisks"
          :disabled="disableDataDisk"
          @change="(v) => instanceItem.cloudDataDisks = v" />
      </bk-form-item>
      <!-- 带宽包 -->
      <bk-form-item
        :label-width="0.1"
        property="internetAccess"
        error-display-type="normal"
        v-if="!disableInternetAccess">
        <InternetAccess
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :value="instanceItem.internetAccess"
          class="mb-[20px]"
          @change="(v) => instanceItem.internetAccess = v"
          @account-type-change="(v) => accountType = v" />
      </bk-form-item>

      <bk-form-item :label="$t('tke.label.count')">
        <bcs-input
          type="number"
          class="max-w-[120px]"
          :min="1"
          :max="maxNodes"
          v-model="instanceItem.applyNum">
        </bcs-input>
      </bk-form-item>
      <p class="text-[#979BA5] leading-[16px] mt-[8px] text-[12px]">
        {{ $t('tke.tips.maxNodeNum', [maxNodes]) }}
      </p>
    </bk-form>
    <div class="flex items-center px-[40px] absolute bottom-0 left-0 bg-[#FAFBFD] w-full h-[48px] bcs-border-top">
      <bcs-button theme="primary" class="min-w-[88px]" @click="handleConfirm">
        {{ isEdit ? $t('generic.button.confirm') : $t('generic.button.add') }}
      </bcs-button>
      <bcs-button class="min-w-[88px]" @click="handleCancel">{{ $t('generic.button.cancel') }}</bcs-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { merge } from 'lodash';
import { computed, inject, PropType, ref, watch } from 'vue';

import { ClusterDataInjectKey, IInstanceItem, IInstanceType, ISubnet } from './types';

import { cloudInstanceTypes, cloudSubnets } from '@/api/modules/cluster-manager';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import DataDisk from '@/views/cluster-manage/add/form/data-disk.vue';
import InternetAccess from '@/views/cluster-manage/add/form/internet-access.vue';
import SystemDisk from '@/views/cluster-manage/add/form/system-disk.vue';
import Zone from '@/views/cluster-manage/add/form/zone.vue';

const cloudID = 'tencentPublicCloud';

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
  accountId: {
    type: String,
    default: '',
  },
  vpcId: {
    type: String,
    default: '',
  },
  instance: {
    type: Object as PropType<IInstanceItem|null>,
    default: () => ({}),
  },
  nodeRole: {
    type: String,
    default: '',
  },
  disableDataDisk: {
    type: Boolean,
    default: true,
  },
  disableInternetAccess: {
    type: Boolean,
    default: true,
  },
  maxNodes: {
    type: Number,
    default: 5,
  },
});
const isEdit = computed(() => !!props.instance && !!Object.keys(props.instance).length);
const emits = defineEmits(['cancel', 'confirm']);

const clusterData = inject(ClusterDataInjectKey);
const enabledZoneList = computed(() => {
  if (!clusterData || !clusterData.value) return [];

  return clusterData.value?.clusterAdvanceSettings?.networkType === 'VPC-CNI'
    ? clusterData.value?.networkSettings?.subnetSource?.new?.map(item => item.zone) || []
    : [];
});

const initData = ref({
  subnetID: '',
  applyNum: 1,
  zone: '',
  instanceType: '',
  systemDisk: {
    diskType: 'CLOUD_PREMIUM',
    diskSize: '50',
  },
  cloudDataDisks: [
    {
      diskType: 'CLOUD_PREMIUM', // 类型
      diskSize: '100', // 大小
      fileSystem: 'ext4', // 文件系统
      autoFormatAndMount: true, // 是否格式化
      mountTarget: '/data', // 挂载路径
    },
  ],
  internetAccess: {
    internetChargeType: 'TRAFFIC_POSTPAID_BY_HOUR',
    internetMaxBandwidth: '0',
    publicIPAssigned: false,
    bandwidthPackageId: '',
  },
});
const instanceItem = ref(initData.value);
const accountType = ref<'STANDARD'|'LEGACY'>();
const rules = ref({
  zone: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  subnetID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  cloudDataDisks: [
    {
      trigger: 'custom',
      message: '',
      validator() {
        let preMountTarget = '';
        const repeatMountTarget = instanceItem.value.cloudDataDisks.some((item) => {
          if (preMountTarget === item.mountTarget) return true;
          preMountTarget = item.mountTarget;
          return false;
        });
        return !repeatMountTarget;
      },
    },
  ],
  internetAccess: [
    {
      trigger: 'custom',
      message: $i18n.t('ca.tips.requiredBandwidthPackage'),
      validator() {
        if (instanceItem.value.internetAccess.publicIPAssigned
        && instanceItem.value.internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE'
        && accountType.value === 'STANDARD') {
          return !!instanceItem.value.internetAccess.bandwidthPackageId;
        }
        return true;
      },
    },
  ],
});

// 子网
const subnets = ref<Array<ISubnet>>([]);
const subnetLoading = ref(false);
const handleGetSubnets = async () => {
  if (!props.accountId || !props.region || !props.vpcId || !instanceItem.value.zone) return;
  subnetLoading.value = true;
  subnets.value = await cloudSubnets({
    $cloudId: cloudID,
    region: props.region,
    accountID: props.accountId,
    vpcID: props.vpcId,
    zone: instanceItem.value.zone,
    injectCluster: true,
  }).catch(() => []);
  subnetLoading.value = false;
};

// 机型
const instanceTypesList = ref<Array<IInstanceType>>([]);
const instanceTypesListByZone = computed<IInstanceType[]>(() => instanceTypesList.value
  .filter(item => !instanceItem.value.zone || item.zones?.includes(instanceItem.value.zone)));
const cpu = ref();
const cpuList = computed(() => instanceTypesListByZone.value.reduce<number[]>((pre, item) => {
  if (pre.find(cpu => cpu === item.cpu)) return pre;

  pre.push(item.cpu);
  return pre;
}, []).sort((a, b) => a - b));
const mem = ref();
const memList = computed(() => instanceTypesListByZone.value.reduce<number[]>((pre, item) => {
  if (pre.find(mem => mem === item.memory)) return pre;

  pre.push(item.memory);
  return pre;
}, []).sort((a, b) => a - b));

const filters = ref<Record<string, string[]>>({});
const sortData = ref({
  prop: '',
  order: '',
});
const filterInstanceList = computed(() => instanceTypesListByZone.value
  .filter(item => (!cpu.value || item.cpu === cpu.value)
     && (!mem.value || item.memory === mem.value))
  .sort((a, b) => {
    // 排序
    if (sortData.value.prop === 'cpu') {
      return sortData.value.order === 'ascending' ? a.cpu - b.cpu : b.cpu - a.cpu;
    }
    if (sortData.value.prop === 'memory') {
      return sortData.value.order === 'ascending' ? a.memory - b.memory : b.memory - a.memory;
    }
    if (sortData.value.prop === 'unitPrice') {
      return sortData.value.order === 'ascending' ? a.unitPrice - b.unitPrice : b.unitPrice - a.unitPrice;
    }
    return 0;
  }));
const nodeTypeFilters = computed(() => filterInstanceList.value
  .reduce<Array<{text: string, value: string}>>((pre, item) => {
  const exist = pre.find(data => data.value === item.typeName);
  if (!exist) {
    pre.push({
      text: item.typeName,
      value: item.typeName,
    });
  }
  return pre;
}, []));

const handleFilterChange = (data) => {
  pageChange(1);
  filters.value = data;
};
const handleSortChange = ({ prop, order  }) => {
  sortData.value = {
    prop,
    order,
  };
};

// 过滤机型
const filterTableData = computed(() => filterInstanceList.value.filter(item => Object.keys(filters.value)
  .every(key => !filters.value[key]?.length || filters.value[key]?.includes(item[key]))));

const {
  pagination,
  curPageData: instanceList,
  pageChange,
  pageSizeChange,
} = usePage<IInstanceType>(filterTableData);


const instanceLoading = ref(false);
const handleGetInstanceType = async () => {
  if (!props.accountId || !props.region) return;
  instanceLoading.value = true;
  instanceTypesList.value = await cloudInstanceTypes({
    $cloudId: cloudID,
    region: props.region,
    accountID: props.accountId,
  }).catch(() => []);
  instanceLoading.value = false;
};
const instanceRowClass = ({ row }) => {
  // SELL 表示售卖，SOLD_OUT 表示售罄
  if (row.status === 'SELL') {
    return 'table-row-enable';
  }
  return 'table-row-disabled';
};
const handleCheckInstanceType = (row) => {
  if (row.status === 'SOLD_OUT') return;
  instanceItem.value.instanceType = row.nodeType;
};

// 取消
const handleCancel = () => {
  emits('cancel');
};
// 校验
const firstTrigger = ref(true);
const formRef = ref();
const validate = async () => {
  firstTrigger.value = false;
  const result = await formRef.value?.validate().catch(() => false);
  const validateSystemDiskSize = Number(instanceItem.value.systemDisk.diskSize) % 10 === 0;
  const validateDataDiskSize = instanceItem.value.cloudDataDisks.every(item => Number(item.diskSize) % 10 === 0);
  return result && !!instanceItem.value.instanceType && validateSystemDiskSize && validateDataDiskSize;
};
// 确定
const nodeConfigRef = ref();
const handleConfirm = async () => {
  const instance = instanceTypesList.value.find(item => item.nodeType === instanceItem.value.instanceType);
  const result = await validate();
  if (result) {
    emits('confirm', {
      ...instanceItem.value,
      applyNum: Number(instanceItem.value.applyNum), // 类型转换
      CPU: instance?.cpu || props.instance?.CPU,
      Mem: instance?.memory || props.instance?.Mem,
      systemDisk: {
        diskType: instanceItem.value.systemDisk.diskType,
        diskSize: String(instanceItem.value.systemDisk.diskSize), // 类型转换
      },
      cloudDataDisks: instanceItem.value.cloudDataDisks.map(item => ({
        ...item,
        diskSize: String(item.diskSize), // 类型转换
      })),
    });
  } else {
    // 自动滚动到第一个错误的位置
    const errDom = nodeConfigRef.value?.querySelectorAll('.form-error-tip');
    const bcsErrDom = nodeConfigRef.value?.querySelectorAll('.bcs-form-error-tip');
    const firstErrDom = errDom[0] || bcsErrDom[0];
    firstErrDom?.scrollIntoView({
      block: 'center',
      behavior: 'smooth',
    });
  }
};

// 校验机型是否在当前页中
const handleValidateInstance = () => {
  setTimeout(() => {
    const exitInstance = filterInstanceList.value.find(item => item.nodeType === instanceItem.value.instanceType);
    if (!exitInstance) {
      instanceItem.value.instanceType = '';
    }
  });
};

watch(
  [
    () => cpu.value,
    () => mem.value,
  ],
  () => {
    pageChange(1);
  },
);
watch(
  [
    () => props.accountId,
    () => props.region,
    () => props.vpcId,
  ],
  () => {
    handleGetSubnets();
  },
  { immediate: true },
);
watch(
  [
    () => props.accountId,
    () => props.region,
  ],
  () => {
    handleGetInstanceType();
  },
  { immediate: true },
);
watch(
  () => props.instance,
  () => {
    instanceItem.value = merge({}, initData.value, props.instance);
    cpu.value = props.instance?.CPU;
    mem.value = props.instance?.Mem;
    handleGetSubnets();
  },
  { deep: true, immediate: true },
);
watch(() => instanceItem.value.zone, () => {
  instanceItem.value.subnetID = '';
  instanceItem.value.instanceType = '';
  handleGetSubnets();
});
</script>
<style lang="postcss" scoped>
.prefix {
  display: inline-block;
  height: 32px;
  line-height: 32px;
  background: #F0F1F5;
  border: 1px solid #C4C6CC;
  border-radius: 2px 0 0 2px;
  padding: 0 8px;
  font-size: 12px;
  &.disabled {
    border-color: #dcdee5;
  }
}
.company {
  font-size: 12px;
  display: inline-block;
  min-width: 30px;
  padding: 0 4px 0 4px;
  height: 32px;
  border: 1px solid #C4C6CC;
  text-align: center;
  border-left: none;
  background-color: #fafbfd;
  &.disabled {
    border-color: #dcdee5;
  }
}

>>> .node-config-radio .bk-radio-text {
  flex: 1;
}
</style>
