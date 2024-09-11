<template>
  <bk-form class="bcs-small-form" :label-width="160" v-bkloading="{ isLoading }">
    <DescList :title="$t('generic.label.basicConfig')">
      <bk-form-item :label="$t('cluster.labels.region')">
        <LoadingIcon v-if="regionLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
        <span v-else>{{ region }}</span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.privateNet.text')">
        <LoadingIcon v-if="vpcLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
        <span v-else>{{ vpc || '--' }}</span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.networkType')">
        {{ networkType }}
      </bk-form-item>
      <bk-form-item label="IPVS">
        {{ IPVS }}
      </bk-form-item>
    </DescList>
    <DescList class="mt-[16px]" :title="$t('k8s.containerNetwork')">
      <!-- 底层网络插件: VPC-CNI模式 -->
      <template
        v-if="clusterData.clusterAdvanceSettings && clusterData.clusterAdvanceSettings.networkType === 'VPC-CNI'">
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
          {{ maxServiceNum }}
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.serviceCidr')">
          {{ clusterData.networkSettings &&clusterData.networkSettings.serviceIPv4CIDR
            ? clusterData.networkSettings.serviceIPv4CIDR : '--' }}
        </bk-form-item>
        <VpcCniDetail :data="clusterData" />
      </template>
      <!-- 其他网络插件: Global Route -->
      <template v-else>
        <!-- 腾讯云 -->
        <bk-form-item
          :label="$t('cluster.labels.cidr')">
          {{ clusterIPv4CIDR }}
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.cidrStep.text')">
          {{ clusterIPv4CIDRCount }}
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
          {{ maxServiceNum }}
        </bk-form-item>
        <!-- 单节点Pod数量上限 -->
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
          {{ maxNodePodNum }}
        </bk-form-item>
        <!-- GR模式下 是否启用vpc-cni -->
        <bk-form-item label="VPC-CNI">
          <LoadingIcon v-if="clusterData.networkSettings?.status === 'INITIALIZATION'">
            {{ clusterData.networkSettings?.enableVPCCni
              ? $t('generic.status.enable') : $t('generic.status.disable') }}...
          </LoadingIcon>
          <StatusIcon
            :status="clusterData.networkSettings?.status"
            v-else-if="clusterData.networkSettings?.status === 'FAILURE'">
            {{ clusterData.networkSettings?.enableVPCCni
              ? $t('generic.status.enableFailed') : $t('generic.status.disableFailed') }}
          </StatusIcon>
          <bcs-switcher
            :value="clusterData.networkSettings?.enableVPCCni"
            :pre-check="toggleVpcCNIStatus"
            v-else>
          </bcs-switcher>
        </bk-form-item>
        <VpcCniDetail
          :data="clusterData"
          v-if="clusterData.networkSettings?.enableVPCCni
            && (['RUNNING', ''].includes(clusterData.networkSettings?.status))" />
      </template>
    </DescList>
    <AddSubnetDialog
      :model-value="showSubnets"
      :cluster-data="clusterData"
      :width="600"
      :confirm-fn="handleConfirmEnableVpcCNI"
      @cancel="showSubnets = false" />
  </bk-form>
</template>
<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';

import AddSubnetDialog from '../components/add-subnet-dialog.vue';
import VpcCniDetail from '../components/vpc-cni-detail.vue';

import { underlayNetwork  } from '@/api/modules/cluster-manager';
import { countIPsInCIDR } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import DescList from '@/components/desc-list.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import useInterval from '@/composables/use-interval';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';
import { ISubnetItem } from '@/views/cluster-manage/types/types';
import useCloud from '@/views/cluster-manage/use-cloud';

const props = defineProps({
  clusterId: {
    type: String,
    required: true,
  },
});

const { clusterData, isLoading, getClusterDetail } = useClusterInfo();

// 容器网段
const clusterContainerCIDR = computed(() => {
  const { multiClusterCIDR = [], clusterIPv4CIDR = '' } = clusterData.value.networkSettings || {};
  return [...multiClusterCIDR, clusterIPv4CIDR];
});
const clusterIPv4CIDR = computed(() => clusterContainerCIDR.value.filter(cidr => !!cidr).join(', ') || '--');
// 容器网段 IP数量
const clusterIPv4CIDRCount = computed(() => clusterContainerCIDR.value.reduce((count, CIDR) => {
  count += countIPsInCIDR(CIDR) || 0;
  return count;
}, 0));

// service网段 IP数量
const maxServiceNum = computed(() => clusterData.value?.networkSettings?.maxServiceNum || 0);

// Pod IP数量
const maxNodePodNum = computed(() => clusterData.value?.networkSettings?.maxNodePodNum || 0);

const IPVS = computed(() => clusterData.value?.clusterAdvanceSettings?.IPVS);

// 网络插件类型
const netTypeMap = {
  GR: 'Global Router',
};
const networkType = computed(() => {
  const type = clusterData.value?.clusterAdvanceSettings?.networkType;
  return type
    ? `${netTypeMap[type] || type}(${clusterData.value.networkType})`
    : '--';
});

// 可用区 和 vpc
const {
  regionLoading,
  regionList,
  handleGetRegionList,
  vpcList,
  vpcLoading,
  handleGetVPCList,
} = useCloud();

const region = computed(() => {
  const data = regionList.value.find(item => item.region === clusterData.value.region);
  return data ? `${data.regionName}(${data.region})` : clusterData.value.region;
});
const vpc = computed(() => {
  const data = vpcList.value.find(item => item.vpcId === clusterData.value.vpcID);
  return data ? `${data.name}(${data.vpcId})` : (clusterData.value.vpcID || '--');
});

// 启用vpc-cni确认
const showSubnets = ref(false);
const toggleVpcCNIStatus = async (value: boolean) => new Promise(async (resolve, reject) => {
  if (value) {
    // 启用vpc-cni
    showSubnets.value = true;
    reject();// dialog处理手后续流程
    return;
  }
  // 禁用vpc-cni
  $bkInfo({
    type: 'warning',
    theme: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('cluster.detail.title.disableVpcCNI.text'),
    subTitle: `${$i18n.t('cluster.detail.title.disableVpcCNI.p1')}, ${$i18n.t('cluster.detail.title.disableVpcCNI.p2')}`,
    defaultInfo: true,
    confirmFn: async () => {
      const result = await underlayNetwork({
        $clusterId: clusterData.value.clusterID,
        disable: true,
        operator: $store.state.user?.username,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        getClusterDetail(props.clusterId, true);
        resolve(true);
      } else {
        reject();
      }
    },
    cancelFn: () => {
      reject(false);
    },
  });
});
const handleConfirmEnableVpcCNI = async (subnets: Array<ISubnetItem>) => {
  const result = await underlayNetwork({
    $clusterId: clusterData.value.clusterID,
    disable: false,
    isStaticIpMode: true, // 自研云默认为 true，暂时不允许用户传递
    claimExpiredSeconds: 300, // 默认为 300
    operator: $store.state.user?.username,
    subnet: {
      new: subnets,
      existed: {
        ids: [], // 暂时不需要了
      },
    },
  }).then(() => true)
    .catch(() => false);
  if (result) {
    getClusterDetail(props.clusterId, true);
    showSubnets.value = false;
  }
};
// 轮询vpc-cni状态
const { start, stop } = useInterval(async () => {
  await getClusterDetail(props.clusterId, true, false);
}, 5000, true);
watch(() => clusterData.value?.networkSettings?.status, () => {
  if (clusterData.value?.networkSettings?.status !== 'INITIALIZATION') {
    stop();
  } else {
    start();
  }
});

onMounted(async () => {
  await getClusterDetail(props.clusterId, true);
  await Promise.all([
    handleGetRegionList({
      cloudAccountID: clusterData.value.cloudAccountID,
      cloudID: clusterData.value.provider,
    }),
    handleGetVPCList({
      region: clusterData.value.region,
      cloudAccountID: clusterData.value.cloudAccountID,
      cloudID: clusterData.value.provider,
      resourceGroupName: clusterData.value.extraInfo?.nodeResourceGroup,
    }),
  ]);
});
</script>
<style lang="postcss" scoped>
>>> .bk-table .cell {
  padding-left: 15px !important;
}
</style>
