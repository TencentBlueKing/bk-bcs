<template>
  <div class="h-[calc(100vh-60px)]">
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
        <bcs-select
          searchable
          :clearable="false"
          v-model="instanceItem.zone">
          <bcs-option
            v-for="zone in zoneList"
            :key="zone.zoneID"
            :id="zone.zone"
            :name="zone.zoneName"
            :disabled="enabledZoneList.length&& !enabledZoneList.includes(zone.zone)"
            v-bk-tooltips="{
              content: $t('tke.tips.zone'),
              disabled: !enabledZoneList.length || enabledZoneList.includes(zone.zone),
              placement: 'left'
            }">
          </bcs-option>
        </bcs-select>
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.subnet')"
        property="subnetID"
        error-display-type="normal"
        required>
        <bk-select
          searchable
          :clearable="false"
          :loading="subnetLoading"
          v-model="instanceItem.subnetID">
          <bk-option
            v-for="net in subnets"
            :key="net.subnetID"
            :id="net.subnetID"
            :name="net.subnetName"
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
              <span>{{ net.subnetName }}</span>
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
          <bcs-select v-model="cpu" searchable class="ml-[-1px] w-[140px] mr-[16px]">
            <bcs-option v-for="item in cpuList" :key="item" :id="item" :name="item"></bcs-option>
          </bcs-select>
          <span class="prefix">{{ $t('generic.label.mem') }}</span>
          <bcs-select v-model="mem" searchable class="ml-[-1px] w-[140px]">
            <bcs-option v-for="item in memList" :key="item" :id="item" :name="item"></bcs-option>
          </bcs-select>
        </div>
        <bcs-table
          :data="instanceList"
          v-bkloading="{ isLoading: instanceLoading }"
          :pagination="pagination"
          :row-class-name="instanceRowClass"
          class="mt-[16px]"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange"
          @row-click="handleCheckInstanceType">
          <bcs-table-column :label="$t('generic.ipSelector.label.serverModel')" prop="typeName" show-overflow-tooltip>
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
          <bcs-table-column label="CPU" prop="cpu" width="80" align="right">
            <template #default="{ row }">
              <span>{{ `${row.cpu}${$t('units.suffix.cores')}` }}</span>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('generic.label.mem')" prop="memory" width="80" align="right">
            <template #default="{ row }">
              <span>{{ row.memory }}G</span>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.text')">
            <template #default="{ row }">
              <span>
                {{ $t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.unit',
                      { price: row.unitPrice })
                }}
              </span>
            </template>
          </bcs-table-column>
        </bcs-table>
        <div class="h-[24px] flex items-center">
          <span
            class="text-[#ea3636] text-[12px] h-[24px] leading-[24px]"
            v-show="!instanceItem.instanceType">{{ $t('generic.validate.required') }}</span>
        </div>
        <div class="flex items-center">
          <span class="prefix">{{ $t('tke.label.systemDisk') }}</span>
          <bcs-select :clearable="false" class="ml-[-1px] w-[140px]" v-model="instanceItem.systemDisk.diskType">
            <bcs-option
              v-for="diskItem in diskEnum"
              :key="diskItem.id"
              :id="diskItem.id"
              :name="diskItem.name">
            </bcs-option>
          </bcs-select>
          <bcs-select
            class="w-[88px] bg-[#fff] ml10"
            :clearable="false"
            v-model="instanceItem.systemDisk.diskSize">
            <bcs-option id="50" name="50"></bcs-option>
            <bcs-option id="100" name="100"></bcs-option>
          </bcs-select>
          <span class="company">GB</span>
        </div>
        <div class="mt-[20px]">
          <bk-checkbox v-model="showDataDisk">{{ $t('tke.button.purchaseDataDisk') }}</bk-checkbox>
          <template v-if="showDataDisk">
            <div
              class="bg-[#F5F7FA] py-[16px] px-[24px] mt-[10px]"
              v-for="item, index in instanceItem.cloudDataDisks"
              :key="index">
              <div class="flex items-center">
                <span class="prefix">{{ $t('tke.label.dataDisk') }}</span>
                <bcs-select :clearable="false" class="ml-[-1px] w-[140px] mr-[16px] bg-[#fff]" v-model="item.diskType">
                  <bcs-option
                    v-for="diskItem in diskEnum"
                    :key="diskItem.id"
                    :id="diskItem.id"
                    :name="diskItem.name">
                  </bcs-option>
                </bcs-select>
                <bcs-input class="max-w-[120px]" type="number" v-model="item.diskSize">
                  <span slot="append" class="group-text !px-[4px]">GB</span>
                </bcs-input>
              </div>
              <div class="flex items-center mt-[16px]">
                <bk-checkbox v-model="item.autoFormatAndMount" class="mr-[8px]">
                  {{ $t('tke.button.autoFormatAndMount') }}
                </bk-checkbox>
                <template v-if="item.autoFormatAndMount">
                  <bcs-select :clearable="false" class="w-[80px] mr-[8px] bg-[#fff]" v-model="item.fileSystem">
                    <bcs-option v-for="file in fileSystem" :key="file" :name="file" :id="file"></bcs-option>
                  </bcs-select>
                  <bk-input class="flex-1" v-model="item.mountTarget"></bk-input>
                </template>
              </div>
            </div>
          </template>
        </div>
      </bk-form-item>
      <bk-form-item :label="$t('tke.label.count')">
        <bcs-input type="number" class="max-w-[120px]" :min="1" :max="5" v-model="instanceItem.applyNum"></bcs-input>
      </bk-form-item>
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

import SelectExtension from './select-extension.vue';
import { ClusterDataInjectKey, IInstanceItem, IInstanceType, ISubnet } from './types';

import { cloudInstanceTypes, cloudSubnets } from '@/api/modules/cluster-manager';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';

const cloudID = 'tencentPublicCloud';

const props = defineProps({
  region: {
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
  zoneList: {
    type: Array as PropType<any[]>,
    default: () => [],
  },
  instance: {
    type: Object as PropType<IInstanceItem|null>,
    default: () => ({}),
  },
  nodeRole: {
    type: String,
    default: '',
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
});
const instanceItem = ref(initData.value);
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
});

// 子网
const subnets = ref<Array<ISubnet>>([]);
const subnetLoading = ref(false);
const handleGetSubnets = async () => {
  if (!props.accountId || !props.region || !props.vpcId) return;
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

const filterInstanceList = computed(() => instanceTypesListByZone.value
  .filter(item => (!cpu.value || item.cpu === cpu.value)
     && (!mem.value || item.memory === mem.value)));

const {
  pagination,
  curPageData: instanceList,
  pageChange,
  pageSizeChange,
} = usePage<IInstanceType>(filterInstanceList);


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

// 磁盘类型
const diskEnum = ref([
  {
    id: 'CLOUD_PREMIUM',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.premium'),
  },
  {
    id: 'CLOUD_SSD',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
  },
  {
    id: 'CLOUD_HSSD',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.hssd'),
  },
]);

// 数据盘
const showDataDisk = ref(true);
const fileSystem = ref(['ext3', 'ext4', 'xfs']);

// 取消
const handleCancel = () => {
  emits('cancel');
};
// 校验
const formRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  return result && !!instanceItem.value.instanceType;
};
// 确定
const handleConfirm = async () => {
  const result = await validate();
  result && emits('confirm', {
    ...instanceItem.value,
    systemDisk: {
      diskType: instanceItem.value.systemDisk.diskType,
      diskSize: String(instanceItem.value.systemDisk.diskSize), // 类型转换
    },
    cloudDataDisks: instanceItem.value.cloudDataDisks.map(item => ({
      ...item,
      diskSize: String(item.diskSize), // 类型转换
    })),
  });
};

watch(
  props,
  () => {
    handleGetSubnets();
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
  },
  { deep: true, immediate: true },
);
watch(() => instanceItem.value.zone, () => {
  instanceItem.value.subnetID = '';
  instanceItem.value.instanceType = '';
  // cpu.value = '';
  // mem.value = '';
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
