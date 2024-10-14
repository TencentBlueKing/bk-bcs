<template>
  <IpSelector
    :show-dialog="showDialog"
    :disable-host-method="disableHostMethod"
    :service="{
      fetchTopologyHostsNodes,
      fetchHostCheck
    }"
    :value="{
      hostList: ipList
    }"
    :key="selectorKey"
    keep-host-field-output
    @change="confirm"
    @close-dialog="cancel" />
</template>
<script lang="ts">
import { computed, PropType, ref, watch } from 'vue';

import {
  hostCheck as hostCheckAdapter,
  topolopyHostsNodes as topologyHostsNodesAdapter,
} from '@blueking/ip-selector/dist/adapter';

import IpSelector from './ipv6-selector';

import { cloudNodes, hostCheck, nodeAvailable, topologyHostsNodes  } from '@/api/modules/cluster-manager';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

// 节点是否被占用Map
let cacheNodeAvailableMap: Record<string, {
  clusterID: string
  clusterName: string
  isExist: false
}> = {};
// 节点云上信息
let cacheNodeListCloudDataMap: Record<string, {
  region: string
  innerIP: string
  vpc: string
  zone: string
  zoneName: string
}> = {};
</script>
<script setup lang="ts">


interface IHostData {
  ip: string
}

const props = defineProps({
  showDialog: {
    type: Boolean,
    default: false,
  },
  // 回显IP列表
  ipList: {
    type: Array,
    default: () => ([]),
  },
  disabledIpList: {
    type: Array as PropType<Array<string|{ip: string, tips: string}>>,
    default: () => [],
  },
  cloudId: {
    type: String,
    default: '',
  },
  region: {
    type: String,
    default: '',
  },
  // 集群VPC
  vpc: {
    type: Object,
    default: () => ({}),
  },
  // 是否校验集群vpc和区域
  validateVpcAndRegion: {
    type: Boolean,
    default: true,
  },
  accountID: {
    type: String,
    default: '',
  },
  availableZoneList: {
    type: Array,
    default: () => [],
  },
  // 支持选不同的vpc，默认不支持
  validateVpc: {
    type: Boolean,
    default: false,
  },
});

const selectorKey = ref('');
watch(() => props.showDialog, () => {
  if (props.showDialog) return;

  selectorKey.value = `${Math.random() * 10}`;
});

const emits = defineEmits(['confirm', 'cancel']);

// 区域信息
const regionList = ref<any[]>([]);
const getRegionList = async () => {
  if (!props.cloudId) return;
  regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
    $cloudId: props.cloudId,
  });
};
watch(() => props.cloudId, () => {
  if (!props.validateVpcAndRegion) return;

  getRegionList();
}, { immediate: true, deep: true });

const getRegionName = (region) => {
  const name = regionList.value.find(item => item.region === region)?.regionName;

  return name ? `${name}(${region})` : region;
};

// 获取topo树当前页的主机列表
const $biz = computed(() => $store.state.curProject.businessID);
const $scope = 'biz';

// 是否显示不可用主机
const isHostOnlyValid = ref(false);

// 获取主机云和占有信息
const handleGetHostAvailableAndCloudInfo = async (hostData: IHostData[]) => {
  const ipList = hostData.filter(item => !!item.ip).map(item => item.ip);
  // 查询主机是否可用
  const nodeAvailableData = await nodeAvailable({
    innerIPs: ipList,
  });
  cacheNodeAvailableMap = nodeAvailableData;
  // 查询当前主机云上信息
  if (props.cloudId && props.region && ipList.length && props.validateVpcAndRegion) {
    const cloudData = await cloudNodes({
      $cloudId: props.cloudId,
      region: props.region,
      ipList: ipList.join(','),
      accountID: props.accountID,
    });
    const cloudDataMap = cloudData.reduce((pre, item) => {
      if (item.innerIP) {
        pre[item.innerIP] = item;
      }
      return pre;
    }, {});
    cacheNodeListCloudDataMap = {
      ...cacheNodeListCloudDataMap,
      ...cloudDataMap,
    };
  }
};
// 获取topo树当前页的主机列表
const fetchTopologyHostsNodes = async (params, hostOnlyValid) => {
  isHostOnlyValid.value = hostOnlyValid;
  const data: {data: IHostData[]} = await topologyHostsNodes({
    ...params,
    $biz: $biz.value,
    $scope,
    showAvailableNode: hostOnlyValid,
  }).catch(() => []);
  await handleGetHostAvailableAndCloudInfo(data.data);
  return topologyHostsNodesAdapter(data);
};
// 手动输入IP获取主机列表
const fetchHostCheck = async (params) => {
  const data = await hostCheck({
    ...params,
    $biz: $biz.value,
    $scope,
    showAvailableNode: isHostOnlyValid.value, // todo 手动输入IP场景切换 "仅显示可用" 不会触发fetchHostCheck函数
  }).catch(() => []);
  await handleGetHostAvailableAndCloudInfo(data);
  return hostCheckAdapter(data);
};

// 禁用ip列表
const disabledIpData = computed(() => props.disabledIpList.reduce((pre, item) => {
  if (typeof item === 'object') {
    pre[item.ip] = item.tips;
  } else {
    pre[item] = $i18n.t('generic.ipSelector.tips.ipNotAvailable');
  }
  return pre;
}, {}));

// 当前行是否可勾选
const disableHostMethod = (row: IHostData) => {
  let tips = '';
  if (cacheNodeAvailableMap[row.ip]?.isExist) {
    const { clusterName = '', clusterID = '' } = cacheNodeAvailableMap[row.ip];
    tips = $i18n.t('generic.ipSelector.tips.ipInUsed', {
      name: clusterName,
      id: clusterID ? ` (${clusterID}) ` : '',
    });
  } else if (disabledIpData.value[row.ip]) {
    tips = disabledIpData.value[row.ip];
  } else if (!!props.cloudId && !!props.region && props.validateVpcAndRegion) {
    if (cacheNodeListCloudDataMap[row.ip]?.region !== props.region) {
      tips = $i18n.t('generic.ipSelector.tips.ipRegionNotMatched', [getRegionName(props.region)]);
    } else if (cacheNodeListCloudDataMap[row.ip]?.vpc !== props.vpc?.vpcID && !props.validateVpc) { // 增加支持选不同的vpc
      tips = $i18n.t('generic.ipSelector.tips.ipVpcNotMatched', [cacheNodeListCloudDataMap[row.ip]?.vpc, props.vpc?.vpcID]);
    } else if (!!props.availableZoneList?.length
      && !props.availableZoneList.includes(cacheNodeListCloudDataMap[row.ip]?.zone)) {
      tips = $i18n.t('tke.tips.nodeNotInSubnetZone', [cacheNodeListCloudDataMap[row.ip]?.zoneName, props.availableZoneList.join(',')]);
    }
  }
  return tips;
};

const confirm = ({ hostList }) => {
  // 兼容以前数据
  const data = hostList.map(item => ({
    ...item,
    ...(cacheNodeListCloudDataMap[item.ip] || {}),
    // 兼容以前数据结构（新UI不要使用这些字段）
    bk_host_innerip: item.ip,
    bk_cloud_id: item?.cloudArea?.id,
    agent_alive: item.alive, // agent状态
  }));
  emits('confirm', data);
};
const cancel = () => {
  emits('cancel');
};
</script>
