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
    </DescList>
    <DescList class="mt-[16px]" :title="$t('k8s.containerNetwork')">
      <!-- 底层网络插件: VPC-CNI模式 -->
      <bk-form-item
        :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
        {{ maxServiceNum }}
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.serviceCidr')">
        {{ clusterData.networkSettings &&clusterData.networkSettings.serviceIPv4CIDR
          ? clusterData.networkSettings.serviceIPv4CIDR : '--' }}
      </bk-form-item>
    </DescList>
  </bk-form>
</template>
<script lang="ts" setup>
import { computed, onMounted } from 'vue';

import DescList from '@/components/desc-list.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';
import useCloud from '@/views/cluster-manage/use-cloud';

const props = defineProps({
  clusterId: {
    type: String,
    required: true,
  },
});

const { clusterData, isLoading, getClusterDetail } = useClusterInfo();

// service网段 IP数量
const maxServiceNum = computed(() => clusterData.value?.networkSettings?.maxServiceNum || 0);

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
