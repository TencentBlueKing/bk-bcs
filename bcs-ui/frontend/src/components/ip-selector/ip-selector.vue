<template>
  <IpSelector
    :show-dialog="showDialog"
    :disable-host-method="disableHostMethod"
    :service="{
      fetchTopologyHostsNodes
    }"
    :value="{
      hostList: ipList
    }"
    :key="selectorKey"
    @change="confirm"
    @close-dialog="cancel" />
</template>
<script setup lang="ts">
import { computed, PropType, ref, watch } from 'vue';

import { topolopyHostsNodes as topologyHostsNodesAdapter } from '@blueking/ip-selector/dist/adapter';

import IpSelector from './ipv6-selector';

import { cloudNodes, nodeAvailable, topologyHostsNodes } from '@/api/modules/cluster-manager';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

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
const nodeList = ref<IHostData[]>([]);
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
}> = {};
// 获取topo树当前页的主机列表
const fetchTopologyHostsNodes = async (params) => {
  const data: {data: IHostData[]} = await topologyHostsNodes({
    ...params,
    $biz: $biz.value,
    $scope,
  }).catch(() => []);
  nodeList.value = data.data;

  const ipList = data.data.filter(item => !!item.ip).map(item => item.ip);
  // 查询主机是否可用
  const nodeAvailableData = await nodeAvailable({
    innerIPs: ipList,
  });
  cacheNodeAvailableMap = nodeAvailableData;
  // 查询当前主机云上信息
  if (props.cloudId && props.region && data.data.length && props.validateVpcAndRegion) {
    const cloudData = await cloudNodes({
      $cloudId: props.cloudId,
      region: props.region,
      ipList: ipList.join(','),
    });
    const cloudDataMap = cloudData.reduce((pre, item) => {
      if (item.innerIP) {
        pre[item.innerIP] = item;
      }
      return pre;
    }, {});
    cacheNodeListCloudDataMap = cloudDataMap;
  }
  return topologyHostsNodesAdapter(data);
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
    } else if (cacheNodeListCloudDataMap[row.ip]?.vpc !== props.vpc?.vpcID) {
      tips = $i18n.t('generic.ipSelector.tips.ipVpcNotMatched', [cacheNodeListCloudDataMap[row.ip]?.vpc, props.vpc?.vpcID]);
    }
  }
  return tips;
};

const confirm = ({ hostList }) => {
  // 兼容以前数据
  const data = hostList.map((item) => {
    const host = nodeList.value.find(node => node.ip === item.ip) || item;
    return {
      ...host,
      ...(cacheNodeListCloudDataMap[host.ip] || {}),
      // 兼容以前数据结构（新UI不要使用这些字段）
      bk_host_innerip: host.ip,
      bk_cloud_id: host?.cloudArea?.id,
      agent_alive: host.alive, // agent状态
    };
  });
  emits('confirm', data);
};
const cancel = () => {
  emits('cancel');
};
</script>
