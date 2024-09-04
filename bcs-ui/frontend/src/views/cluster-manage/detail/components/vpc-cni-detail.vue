<template>
  <div>
    <bk-form-item :label="$t('cluster.create.label.networkMode.text')">
      {{ networkModeTextMap[clusterData.networkSettings?.networkMode] || '--' }}
    </bk-form-item>
    <bk-form-item
      :label="$t('tke.label.staticIpMode')">
      {{ clusterData.networkSettings && clusterData.networkSettings.isStaticIpMode
        ? $t('generic.status.support')
        : $t('generic.status.unSupport') }}
    </bk-form-item>
    <div class="px-[30px]">
      <bk-table
        :data="subnets"
        :span-method="objectSpanMethod"
        col-border
        v-bkloading="{ isLoading: subnetLoading }"
        class="network-table">
        <bk-table-column :label="$t('tke.label.zone')" prop="zoneName"></bk-table-column>
        <bk-table-column :label="$t('tke.label.subnetID')" prop="subnetID"></bk-table-column>
        <bk-table-column :label="$t('tke.label.subnetName')" prop="subnetName"></bk-table-column>
        <bk-table-column label="CIDR" prop="cidrRange"></bk-table-column>
        <bk-table-column :label="$t('tke.label.ipNum')" prop="counts">
          <template #default="{ row }">
            <span>{{ getIpNumber(row) }}</span>
          </template>
        </bk-table-column>
      </bk-table>
      <div
        :class="[
          'h-[36px] bg-[#FAFBFD] bcs-border mt-[-1px]',
          'flex items-center justify-center cursor-pointer'
        ]"
        v-if="!subnetLoading"
        @click="handleShowSubnetsDialog">
        <i class="bk-icon icon-plus-circle-shape mr5 text-[#979BA5] text-[14px]"></i>
        <span>{{ $t('tke.button.addSubnets') }}</span>
      </div>
    </div>
    <!-- 添加子网 -->
    <AddSubnetDialog
      :model-value="showSubnets"
      :cluster-data="clusterData"
      @cancel="showSubnets = false"
      @confirm="handleConfirmAddSubnet" />
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import AddSubnetDialog from './add-subnet-dialog.vue';

import { cloudSubnets } from '@/api/modules/cluster-manager';
import { countIPsInCIDR } from '@/common/util';
import { ICluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';
import { ISubnet } from '@/views/cluster-manage/types/types';

interface Props {
  data: ICluster
}

const props = defineProps<Props>();

const { clusterData, getClusterDetail } = useClusterInfo();

// 网络模式
const networkModeTextMap = {
  'tke-route-eni': $i18n.t('cluster.create.label.networkMode.cni.route'),
  'tke-direct-eni': $i18n.t('cluster.create.label.networkMode.cni.direct'),
};

// 子网
const subnets = ref<Array<ISubnet>>([]);
const subnetLoading = ref(false);
const handleGetSubnets = async () => {
  const { provider, cloudAccountID, region, vpcID } = clusterData.value;
  const eniSubnetIDs = clusterData.value?.networkSettings?.eniSubnetIDs || [];
  if (!provider || !region || !vpcID || !eniSubnetIDs.length) return;

  subnetLoading.value = true;
  subnets.value = await cloudSubnets({
    $cloudId: provider,
    region,
    accountID: cloudAccountID,
    vpcID,
    subnetID: eniSubnetIDs?.join(','),
    resourceGroupName: clusterData.value?.extraInfo?.nodeResourceGroup,
  }).catch(() => []);
  firstRowspan.value = {};
  // 排序子网，将相同子网放在一起
  subnets.value = subnets.value.sort((a, b) => {
    if (a.zone < b.zone) {
      return -1;
    }
    if (a.zone > b.zone) {
      return 1;
    }
    return 0;
  });
  subnetLoading.value = false;
};

// 可用区展示
const zoneCounts = computed<Record<string, ISubnet[]>>(() => {
  const counts = {};
  subnets.value.forEach((item) => {
    if (!counts[item.zone]) {
      counts[item.zone] = [item];
    } else {
      counts[item.zone].push(item);
    }
  });
  return counts;
});

// 获取IP数量
const getIpNumber = (row: ISubnet) => {
  const zones = zoneCounts.value[row.zone];
  const availableIPAddressCounts = zones.reduce((pre, item) => {
    pre += Number(item.availableIPAddressCount);
    return pre;
  }, 0);
  const cidrs = zones.reduce((pre, item) => {
    pre += Number(countIPsInCIDR(item.cidrRange));
    return pre;
  }, 0);
  return `${availableIPAddressCounts}/${cidrs}`;
};

// 添加子网
const showSubnets = ref(false);
const handleShowSubnetsDialog = () => {
  showSubnets.value = true;
};
const handleConfirmAddSubnet = async () => {
  showSubnets.value = false;
  subnetLoading.value = true;
  await getClusterDetail(props.data?.clusterID, true);// 获取最新的子网ID
  await handleGetSubnets();
  subnetLoading.value = false;
};

// 合并单元格
const firstRowspan = ref({});
const objectSpanMethod = ({ row, column }) => {
  if (column.property === 'zoneName' || column.property === 'counts') {
    if (zoneCounts.value[row.zone]?.length === 1) return {
      rowspan: 1,
      colspan: 1,
    };
    // 合并相同可用区
    if (zoneCounts.value[row.zone]?.length > 1 && !firstRowspan.value[`${column.property}-${row.zone}`]) {
      firstRowspan.value[`${column.property}-${row.zone}`] = true;
      return {
        rowspan: zoneCounts.value[row.zone]?.length,
        colspan: 1,
      };
    }
    return {
      rowspan: 0,
      colspan: 1,
    };
  }
};

watch(() => props.data?.clusterID, async () => {
  if (!props.data?.clusterID) return;

  clusterData.value = props.data;// 初始化值，不用再请求一次详情
  subnetLoading.value = true;
  await handleGetSubnets();
  subnetLoading.value = false;
}, { immediate: true });
</script>
