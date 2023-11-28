<template>
  <div class="max-w-[750px]">
    <bk-form-item :label="$t('tke.label.region')">
      <bcs-select
        class="max-w-[500px]"
        :value="region"
        searchable
        :clearable="false"
        disabled>
        <bcs-option
          v-for="item in regionList"
          :key="item.region"
          :id="item.region"
          :name="item.regionName">
        </bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item :label="$t('cluster.create.label.privateNet.text')">
      <bcs-select
        class="max-w-[500px]"
        :value="vpcID"
        disabled>
        <bcs-option
          v-for="item in vpcList"
          :key="item.vpcId"
          :id="item.vpcId"
          :name="`${item.name}(${item.vpcId})`">
          <div class="flex items-center place-content-between">
            <span>
              {{`${item.name}(${item.vpcId})`}}
              <span class="vpc-id">{{`(${item.vpcId})`}}</span>
            </span>
          </div>
        </bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item :label="$t('tke.label.system')">
      <bcs-select disabled :value="os" class="max-w-[500px]">
        <bcs-option-group v-for="group in imageListByGroup" :key="group.provider" :name="group.name">
          <template v-if="group.provider === 'PRIVATE_IMAGE'">
            <bcs-option
              v-for="item in group.children"
              :key="item.imageID"
              :id="item.imageID"
              :name="item.alias">
            </bcs-option>
          </template>
          <template v-else>
            <bcs-option
              v-for="item in group.children"
              :key="item.osName"
              :id="item.osName"
              :name="item.alias">
            </bcs-option>
          </template>
        </bcs-option-group>
      </bcs-select>
    </bk-form-item>
    <bk-form-item :label="$t('tke.label.configRec')" v-if="showRecommendedConfig">
      <bk-radio-group :value="level" class="flex flex-col" @change="handleLevelChange">
        <bk-radio
          v-for="item in masterConfigSchema"
          :key="item.id"
          :value="item.id"
          class="!ml-[0px] !h-[32px] leading-[32px]">
          {{ item.name }}
          <span class="text-[#979BA5] ml-[8px]">{{ item.desc }}</span>
        </bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('tke.label.chargeType.text')"
      :desc="chargeDesc"
      property="instanceChargeType"
      error-display-type="normal"
      required>
      <bk-radio-group
        class="inline-flex items-center h-[32px]"
        v-model="instanceCommonConfig.instanceChargeType"
        @change="handleChargeTypeChange">
        <bk-radio value="PREPAID">
          {{ $t('tke.label.chargeType.prepaid') }}
        </bk-radio>
        <bk-radio value="POSTPAID_BY_HOUR">
          {{ $t('tke.label.chargeType.postpaid_by_hour') }}
        </bk-radio>
        <bk-link
          theme="primary"
          class="ml-[30px]"
          href="https://cloud.tencent.com/document/product/213/2180"
          target="_blank">
          <i class="bcs-icon bcs-icon-fenxiang mr-[2px]"></i>
          {{ $t('tke.button.chargeTypeDiff') }}
        </bk-link>
      </bk-radio-group>
      <div id="chargeDesc">
        <div>{{ $t('tke.label.chargeType.prepaidDesc', [$t('tke.label.chargeType.prepaid')]) }}</div>
        <div>{{ $t('tke.label.chargeType.postpaid_by_hour_desc', [$t('tke.label.chargeType.postpaid_by_hour')]) }}</div>
      </div>
    </bk-form-item>
    <template v-if="instanceCommonConfig.instanceChargeType === 'PREPAID' && instanceCommonConfig.charge">
      <bk-form-item :label="$t('tke.label.period')">
        <bcs-select :clearable="false" searchable v-model="instanceCommonConfig.charge.period" class="max-w-[500px]">
          <bcs-option v-for="item in periodList" :key="item.id" :id="item.id" :name="item.name"></bcs-option>
        </bcs-select>
      </bk-form-item>
      <bk-form-item :label="$t('tke.label.autoRenewal.text')">
        <bcs-checkbox
          true-value="NOTIFY_AND_AUTO_RENEW"
          false-value="NOTIFY_AND_MANUAL_RENEW"
          v-model="instanceCommonConfig.charge.renewFlag">
          {{ $t('tke.label.autoRenewal.desc') }}
        </bcs-checkbox>
      </bk-form-item>
    </template>
    <bk-form-item
      :label="$t('tke.label.nodeConfig')"
      required
      property="instances"
      error-display-type="normal"
      ref="instancesRef">
      <bk-button theme="primary" icon="plus" outline @click="handleAddInstance">
        {{ $t('tke.button.addNodeConfig') }}
      </bk-button>
      <div
        class="bg-[#F5F7FA] mt-[16px] p-[8px] rounded text-[12px] relative"
        v-for="item, index in instances"
        :key="index">
        <div class="absolute top-[8px] right-[24px] text-[14px]">
          <i
            class="bk-icon icon-edit-line cursor-pointer mr-[16px] hover:text-[#3A84FF]"
            @click="handleModifyInstance(index)"></i>
          <i
            class="bk-icon icon-close3-shape cursor-pointer hover:text-[#3A84FF]"
            @click="handleDeleteInstance(index)">
          </i>
        </div>
        <div class="flex items-center h-[32px]">
          <label class="node-config-label">{{ $t('tke.label.zone') }}</label>
          {{ getZoneName(item.zone) }}
        </div>
        <div class="flex items-center h-[32px]">
          <label class="node-config-label">{{ $t('tke.label.subnet') }}</label>
          {{ item.subnetID }}
        </div>
        <div class="flex items-center h-[32px]">
          <label class="node-config-label">{{ $t('tke.label.instanceType') }}</label>
          {{ item.instanceType }}
        </div>
        <div class="flex items-center h-[32px]">
          <label class="node-config-label">{{ $t('tke.label.systemDisk') }}</label>
          {{ `${diskMap[item.systemDisk.diskType]} ${item.systemDisk.diskSize}G` }}
        </div>
        <div class="flex items-center h-[32px]">
          <label class="node-config-label">{{ $t('tke.label.dataDisk') }}</label>
          <div>
            <div v-for="disk, i in item.cloudDataDisks" :key="i">
              {{ `${diskMap[disk.diskType]} ${disk.diskSize}G ${disk.fileSystem} ${disk.mountTarget}` }}
            </div>
          </div>
        </div>
        <div
          :class="[
            'flex items-center justify-center rounded-tl-lg',
            'absolute right-0 bottom-0 min-w-[80px] h-[32px]',
            'text-[#fff] text-[14px] font-bold bg-[#699DF4]'
          ]">
          x {{ item.applyNum }}
        </div>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('tke.label.securityGroup.text')"
      property="securityGroupIDs"
      error-display-type="normal"
      required>
      <bk-select
        class="max-w-[500px]"
        searchable
        multiple
        :clearable="false"
        v-model="instanceCommonConfig.securityGroupIDs">
        <bk-option
          v-for="item in securityGroups"
          :key="item.securityGroupID"
          :id="item.securityGroupID"
          :name="item.securityGroupName">
        </bk-option>
        <template slot="extension">
          <SelectExtension
            :link-text="$t('tke.link.securityGroup')"
            link="https://console.cloud.tencent.com/vpc/security-group"
            @refresh="refreshSecurityGroups" />
        </template>
      </bk-select>
    </bk-form-item>
    <bcs-sideslider
      :is-show.sync="showNodeConfig"
      :quick-close="false"
      :title="editIndex === -1 ? $t('tke.title.addNodeConfig') : $t('tke.title.modifyNodeConfig')"
      :width="740">
      <ApplyNodeConfig
        :region="region"
        :account-id="cloudAccountID"
        :vpc-id="vpcID"
        :zone-list="zoneList"
        :node-role="nodeRole"
        :instance="editInstanceItem"
        :cloud-account-i-d="cloudAccountID"
        :cloud-i-d="cloudID"
        :disable-data-disk="disableDataDisk"
        :disable-internet-access="disableInternetAccess"
        slot="content"
        @cancel="showNodeConfig = false"
        @confirm="handleNodeConfigConfirm" />
    </bcs-sideslider>
  </div>
</template>
<script setup lang="ts">
import { computed, defineProps, PropType, ref, watch } from 'vue';

import ApplyNodeConfig from './apply-node-config.vue';
import { ICloudRegion, IInstanceItem, ISecurityGroup, IZoneItem } from './types';

import $i18n from '@/i18n/i18n-setup';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';

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
  os: {
    type: String,
    default: '',
  },
  vpcID: {
    type: String,
    default: '',
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
  securityGroups: {
    type: Array as PropType<ISecurityGroup[]>,
    default: () => [],
  },
  zoneList: {
    type: Array as PropType<IZoneItem[]>,
    default: () => [],
  },
  instances: {
    type: Array as PropType<IInstanceItem[]>,
    default: () => [],
  },
  level: {
    type: String,
    default: '',
  },
  nodeRole: {
    type: String,
    default: '',
  },
  showRecommendedConfig: {
    type: Boolean,
    default: true,
  },
  disableDataDisk: {
    type: Boolean,
    default: true,
  },
  disableInternetAccess: {
    type: Boolean,
    default: true,
  },
});

const emits = defineEmits(['instance-list-change', 'level-change', 'delete-instance', 'common-config-change', 'refresh-security-groups']);

const diskMap = ref({
  CLOUD_PREMIUM: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.premium'),
  CLOUD_SSD: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
  CLOUD_HSSD: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.hssd'),
});
const masterConfigSchema = ref([
  {
    id: 'L100',
    name: $i18n.t('tke.configRec.l100.text'),
    desc: $i18n.t('tke.configRec.l100.desc'),
  },
  {
    id: 'L500',
    name: $i18n.t('tke.configRec.l500.text'),
    desc: $i18n.t('tke.configRec.l500.desc'),
  },
  {
    id: 'L1000',
    name: $i18n.t('tke.configRec.l1000.text'),
    desc: $i18n.t('tke.configRec.l1000.desc'),
  },
  {
    id: '',
    name: $i18n.t('tke.configRec.custom.text'),
    desc: $i18n.t('tke.configRec.custom.desc'),
  },
]);
const handleLevelChange = (level) => {
  emits('level-change', level);
};

const periodList = ref([
  ...[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11].map(month => ({
    id: month,
    name: $i18n.t('units.time.nMonths', [month]),
  })),
  ...[1, 2, 3].map(year => ({
    id: year * 12,
    name: $i18n.t('units.time.nYears', [year]),
  })),
]);
const showNodeConfig = ref(false);

// 获取zone name
const getZoneName = zone => props.zoneList.find(item => item.zone === zone)?.zoneName;

// 节点公共配置
const instanceCommonConfig = ref<Partial<IInstanceItem>>({
  nodeRole: props.nodeRole, // MASTER_ETCD WORKER
  instanceChargeType: 'PREPAID',
  securityGroupIDs: [],
  isSecurityService: true, // 默认true
  isMonitorService: true, // 默认true
  charge: {
    period: 1,
    renewFlag: 'NOTIFY_AND_AUTO_RENEW', // NOTIFY_AND_AUTO_RENEW, NOTIFY_AND_MANUAL_RENEW, DISABLE_NOTIFY_AND_MANUAL_RENEW
  },
});
watch(
  instanceCommonConfig,
  () => {
    emits('common-config-change', instanceCommonConfig.value);
  },
  { deep: true, immediate: true },
);

// 计费模式
const chargeDesc = ref({
  allowHTML: true,
  content: '#chargeDesc',
});
const handleChargeTypeChange = (value) => {
  if (value === 'PREPAID') {
    instanceCommonConfig.value.charge = {
      period: 1,
      renewFlag: 'NOTIFY_AND_AUTO_RENEW',
    };
  } else if (value === 'POSTPAID_BY_HOUR') {
    instanceCommonConfig.value.charge = null;
  }
};

// 添加或者修改节点配置
const editIndex = ref(-1);
const editInstanceItem = computed<IInstanceItem|null>(() => {
  if (editIndex.value === -1) return null;

  return props.instances[editIndex.value];
});
const handleAddInstance = () => {
  editIndex.value = -1;
  showNodeConfig.value = true;
};
const handleModifyInstance = (index) => {
  editIndex.value = index;
  showNodeConfig.value = true;
};
// 删除节点配置
const handleDeleteInstance = (index) => {
  emits('delete-instance', index);
  handleLevelChange('');// 重置level
};
const handleNodeConfigConfirm = async (item: IInstanceItem) => {
  if (editIndex.value > -1) {
    const data: IInstanceItem[] = JSON.parse(JSON.stringify(props.instances));
    data.splice(editIndex.value, 1, item);
    emits('instance-list-change', data);
  } else {
    emits('instance-list-change', [
      ...props.instances,
      item,
    ]);
  }
  handleLevelChange('');// 重置level
  showNodeConfig.value = false;
};

const refreshSecurityGroups = () => {
  emits('refresh-security-groups');
};
</script>
<style scoped lang="postcss">
.node-config-label {
  display: flex;
  justify-content: end;
  width: 90px;
  margin-right: 4px;
  &::after {
    content: ':';
    margin-left: 4px;
  }
}
</style>
