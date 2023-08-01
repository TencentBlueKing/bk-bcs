<template>
  <ipSelector
    ref="selectorRef"
    v-bkloading="{ isLoading }"
    :panels="panels"
    :height="height"
    :active.sync="active"
    :preview-data="previewData"
    :get-default-data="handleGetDefaultData"
    :get-search-table-data="handleGetSearchTableData"
    :static-table-config="staticTableConfig"
    :custom-input-table-config="staticTableConfig"
    :get-default-selections="getDefaultSelections"
    :get-row-disabled-status="getRowDisabledStatus"
    :get-row-tips-content="getRowTipsContent"
    :preview-operate-list="previewOperateList"
    :default-expand-level="1"
    :search-data-options="searchDataOptions"
    :tree-data-options="treeDataOptions"
    :preview-width="240"
    :left-panel-width="300"
    :preview-title="$t('generic.ipSelector.selected.text')"
    :default-active-name="['nodes']"
    :enable-search-panel="false"
    :enable-tree-filter="true"
    :static-table-placeholder="$t('generic.ipSelector.placeholder.searchIp')"
    :across-page="true"
    ip-key="bk_host_innerip"
    ellipsis-direction="ltr"
    :default-accurate="true"
    :default-selected-node="defaultSelectedNode"
    @check-change="handleCheckChange"
    @remove-node="handleRemoveNode"
    @menu-click="handleMenuClick"
    @search-selection-change="handleCheckChange">
    <template #collapse-title="{ item }">
      <i18n path="generic.ipSelector.selected.suffix">
        <span place="count" class="preview-count">{{ item.data.length }}</span>
      </i18n>
    </template>
  </ipSelector>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { defineComponent, reactive, toRefs, h, ref, watch, computed, PropType } from 'vue';
import { ipSelector, AgentStatus } from './ip-selector';
import './ip-selector.css';
import { fetchBizTopo, fetchBizHosts, nodeAvailable } from '@/api/base';
import { copyText } from '@/common/util';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';
import $store from '@/store';
import { useClusterOperate } from '@/views/cluster-manage/cluster/use-cluster';

export interface ISelectorState {
  isLoading: boolean;
  panels: any[];
  active: string;
  previewData: any[];
  staticTableConfig: any[];
  previewOperateList: any[];
  searchDataOptions: any;
  treeDataOptions: any;
  curTreeNode: any;
}
export default defineComponent({
  name: 'IpSelectorBcs',
  components: {
    ipSelector,
  },
  props: {
    // 回显IP列表
    ipList: {
      type: Array,
      default: () => ([]),
    },
    height: {
      type: Number,
      default: 600,
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
  },
  setup(props, ctx) {
    const statusMap = {
      0: 'terminated',
      1: 'running',
    };
    const textMap = {
      0: $i18n.t('generic.status.error'),
      1: $i18n.t('generic.status.ready'),
    };
    const renderIpAgentStatus = row => h(AgentStatus, {
      props: {
        type: 2,
        data: [
          {
            status: statusMap[row.agent_alive],
            display: textMap[row.agent_alive],
          },
        ],
      },
    });
    const state = reactive<ISelectorState>({
      isLoading: false,
      panels: [
        {
          name: 'static-topo',
          label: $i18n.t('generic.ipSelector.button.staticTop'),
        },
        {
          name: 'custom-input',
          label: $i18n.t('generic.ipSelector.button.customInput'),
        },
      ],
      active: 'static-topo',
      previewData: [],
      staticTableConfig: [
        {
          prop: 'bk_host_innerip',
          label: $i18n.t('generic.ipSelector.label.innerIp'),
        },
        {
          prop: 'agent_alive',
          label: $i18n.t('generic.ipSelector.label.agentStatus'),
          render: renderIpAgentStatus,
        },
        {
          prop: 'idc_unit_name',
          label: $i18n.t('generic.ipSelector.label.idc'),
        },
        {
          prop: 'svr_device_class',
          label: $i18n.t('generic.ipSelector.label.serverModel'),
        },
      ],
      previewOperateList: [
        {
          id: 'removeAll',
          label: $i18n.t('generic.ipSelector.selected.action.removeAll'),
        },
        {
          id: 'copyIp',
          label: $i18n.t('generic.ipSelector.selected.action.copyIp.text'),
        },
      ],
      searchDataOptions: {},
      treeDataOptions: {
        idKey: 'id',
        nameKey: 'bk_inst_name',
        childrenKey: 'child',
      },
      curTreeNode: null,
    });
    const selectorRef = ref<any>(null);

    // 初始化回显列表
    const { ipList, disabledIpList, cloudId } = toRefs(props);
    watch(ipList, () => {
      const groups = state.previewData.find(item => item.id === 'nodes');
      if (groups) {
        ipList.value.forEach((item) => {
          const index = groups.data.find(data => identityIp(data, item));
          index === -1 && groups.data.push(item);
        });
      } else {
        state.previewData.push({
          id: 'nodes',
          data: [...ipList.value],
          dataNameKey: 'bk_host_innerip',
        });
      }
    }, { immediate: true });

    // 获取左侧Tree数据
    let treeData: any[] = [];
    const defaultSelectedNode = ref();
    const handleSetTreeId = (nodes: any[] = []) => {
      nodes.forEach((node) => {
        node.id = `${node.bk_inst_id}-${node.bk_obj_id}`;
        if (node.child) {
          handleSetTreeId(node.child);
        }
      });
    };
    const handleGetDefaultData = async () => {
      if (!treeData.length) {
        treeData = await fetchBizTopo().catch(() => []);
        defaultSelectedNode.value = `${treeData[0]?.bk_inst_id}-${treeData[0]?.bk_obj_id}`;
        handleSetTreeId(treeData);
      }
      return treeData;
    };
    const nodeAvailableMap = {};
    const { getCloudNodes } = useClusterOperate();
    // 静态表格数据处理
    const handleGetStaticTableData = async (params) => {
      const { selections = [], current, limit, tableKeyword, accurate } = params;
      const bizHostsParams: any = {
        limit,
        offset: (current - 1) * limit,
        fuzzy: !accurate,
        ip_list: tableKeyword.split(' ').filter(ip => !!ip),
      };
      const [node] = selections;
      state.curTreeNode = node;
      if (!node) return { total: 0, data: [] };

      if (node.bk_obj_id === 'set') {
        bizHostsParams.set_id = node.bk_inst_id;
      } else if (node.bk_obj_id === 'module') {
        bizHostsParams.module_id = node.bk_inst_id;
      }
      const data = await fetchBizHosts(bizHostsParams).catch(() => ({ results: [] }));
      const nodeAvailableData = await nodeAvailable({
        innerIPs: data.results.map(item => item.bk_host_innerip),
      });
      // 合并节点对应的云数据信息
      if (props.cloudId && props.region && data.results.length) {
        const nodeCloudData = await getCloudNodes({
          $cloudId: props.cloudId,
          region: props.region,
          ipList: data.results.map(item => item.bk_host_innerip).join(','),
        });
        data.results = data.results.map((item) => {
          const extraData = nodeCloudData.find(node => node.innerIP === item.bk_host_innerip) || {};
          return {
            ...item,
            ...extraData,
          };
        });
      }
      Object.assign(nodeAvailableMap, nodeAvailableData);
      return {
        total: data.count || 0,
        data: data.results,
      };
    };
    // 自定义输入表格数据处理
    const handleGetCustomInputTableData = async (params) => {
      const { accurate, ipList = [] } = params;
      const bizHostsParams: any = {
        desire_all_data: true,
        fuzzy: !accurate,
        ip_list: ipList,
      };
      const data = await fetchBizHosts(bizHostsParams).catch(() => ({ results: [] }));
      const nodeAvailableData = await nodeAvailable({
        innerIPs: data.results.map(item => item.bk_host_innerip),
      });
      // 合并节点对应的云数据信息
      if (props.cloudId && props.region && ipList.length) {
        const nodeCloudData = await getCloudNodes({
          $cloudId: props.cloudId,
          region: props.region,
          ipList: ipList.join(','),
        });
        data.results = data.results.map((item) => {
          const extraData = nodeCloudData.find(node => node.innerIP === item.bk_host_innerip) || {};
          return {
            ...item,
            ...extraData,
          };
        });
      }
      Object.assign(nodeAvailableMap, nodeAvailableData);
      return {
        total: data.count || 0,
        data: data.results,
      };
    };
    const handleGetSearchTableData = async (params) => {
      if (state.active === 'static-topo') {
        return handleGetStaticTableData(params);
      } if (state.active === 'custom-input') {
        return handleGetCustomInputTableData(params);
      }
    };
    // 判断两个IP节点是否相同
    const identityIp = (current, origin) => current.bk_cloud_id === origin.bk_cloud_id
      && current.bk_host_innerip === origin.bk_host_innerip;
    // 重新获取表格勾选状态
    const resetTableCheckedStatus = () => {
      selectorRef.value?.handleGetDefaultSelections();
    };
    // 预览菜单点击事件
    const handleMenuClick = ({ menu }) => {
      if (menu.id === 'removeAll') {
        state.previewData = [];
        resetTableCheckedStatus();
        handleChange();
      } else if (menu.id === 'copyIp') {
        const group = state.previewData.find(data => data.id === 'nodes');
        const ipList = group?.data.map(item => item.bk_host_innerip) || [];
        copyText(ipList.join('\n'));
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.ipSelector.selected.action.copyIp.msg', { number: ipList.length }),
        });
      }
    };
    // 预览面板单个节点移除事件
    const handleRemoveNode = ({ child, item }) => {
      const group = state.previewData.find(data => data.id === item.id);
      const index = group?.data.findIndex(data => identityIp(data, child));

      index > -1 && group.data.splice(index, 1);
      resetTableCheckedStatus();
      handleChange();
    };
    // 本页选择
    const handleCurrentPageChecked = (data) => {
      const { selections = [], excludeData = [] } = data;
      const index = state.previewData.findIndex((item: any) => item.id === 'nodes');
      if (index > -1) {
        const { data } = state.previewData[index];
        selections.forEach((select) => {
          const index = data.findIndex(data => identityIp(data, select));

          index === -1 && data.push(select);
        });
        excludeData.forEach((exclude) => {
          const index = data.findIndex(data => identityIp(data, exclude));

          index > -1 && data.splice(index, 1);
        });
      } else {
        state.previewData.push({
          id: 'nodes',
          data: [...selections],
          dataNameKey: 'bk_host_innerip',
        });
      }
    };
    // 静态选择跨页全选
    const handleStaticTopoAllChecked = async (data) => {
      const { excludeData = [], checkValue } = data;
      if (checkValue === 1) {
        excludeData.forEach((exclude) => {
          const index = state.previewData.findIndex(data => identityIp(data, exclude));

          index > -1 && state.previewData.splice(index, 1);
        });
      } else if (checkValue === 2) {
        state.isLoading = true;
        const params: any = {
          desire_all_data: true,
        };
        if (state.curTreeNode?.bk_obj_id === 'set') {
          params.set_id = state.curTreeNode.bk_inst_id;
        } else if (state.curTreeNode?.bk_obj_id === 'module') {
          params.module_id = state.curTreeNode.bk_inst_id;
        }
        const data = await fetchBizHosts(params).catch(() => ({ results: [] }));
        const ipList = data.results.filter(item => !getRowDisabledStatus(item));
        state.previewData = [
          {
            id: 'nodes',
            data: ipList,
            dataNameKey: 'bk_host_innerip',
          },
        ];
        state.isLoading = false;
      } else if (checkValue === 0) {
        state.previewData = [];
      }
    };
    // 表格勾选事件
    const handleCheckChange = async (data) => {
      if (data?.checkType === 'current') {
        handleCurrentPageChecked(data);
      } else if (state.active === 'custom-input') {
        data.selections = data.selections?.filter(row => !getRowDisabledStatus(row));
        handleCurrentPageChecked(data);
      } else if (data?.checkType === 'all') {
        await handleStaticTopoAllChecked(data);
      }
      // 统一抛出change事件
      handleChange();
    };
    // 获取当前行的勾选状态
    const getDefaultSelections = (row) => {
      const group = state.previewData.find(data => data.id === 'nodes');
      return group?.data.some(data => identityIp(data, row));
    };
    // 禁用ip列表
    const disabledIpData = computed(() => disabledIpList.value.reduce((pre, item) => {
      if (typeof item === 'object') {
        pre[item.ip] = item.tips;
      } else {
        pre[item] = $i18n.t('generic.ipSelector.tips.ipNotAvailable');
      }
      return pre;
    }, {}));
    // 表格表格当前行禁用状态
    const getRowDisabledStatus = row => !row.is_valid
      || nodeAvailableMap[row.bk_host_innerip]?.isExist
      || disabledIpData.value[row.bk_host_innerip]
      || (!!props.cloudId && !!props.region && (row.vpc !== props.vpc?.vpcID || row.region !== props.region));
    // 获取表格当前行tips内容
    const getRowTipsContent = (row) => {
      let tips: any = '';
      if (!row.is_valid) {
        tips = $i18n.t('generic.ipSelector.tips.dockerIpNotAvailable');
      } else if (nodeAvailableMap[row.bk_host_innerip]?.isExist) {
        const { clusterName = '', clusterID = '' } = nodeAvailableMap[row.bk_host_innerip];
        tips = $i18n.t('generic.ipSelector.tips.ipInUsed', {
          name: clusterName,
          id: clusterID ? ` (${clusterID}) ` : '',
        });
      } else if (disabledIpData.value[row.bk_host_innerip]) {
        tips = disabledIpData.value[row.bk_host_innerip];
      } else if (!!props.cloudId && !!props.region) {
        if (row.region !== props.region) {
          tips = $i18n.t('generic.ipSelector.tips.ipRegionNotMatched', [getRegionName(props.region)]);
        } else if (row.vpc !== props.vpc?.vpcID) {
          tips = $i18n.t('generic.ipSelector.tips.ipVpcNotMatched', [row.vpc, props.vpc?.vpcID]);
        }
      }
      return tips;
    };
    // 统一抛出change事件
    const handleChange = () => {
      const group = state.previewData.find(data => data.id === 'nodes') || {};
      group.data = (group.data || []).filter(row => !nodeAvailableMap[row.bk_host_innerip]?.isExist);
      ctx.emit('change', group?.data || []);
    };
    // 获取IP节点数据
    const handleGetData = () => {
      const group = state.previewData.find(data => data.id === 'nodes');
      return group?.data || [];
    };

    const regionList = ref<any[]>([]);
    const getRegionList = async () => {
      if (!cloudId.value) return;
      regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
        $cloudId: cloudId.value,
      });
    };
    watch(cloudId, () => {
      getRegionList();
    }, { immediate: true, deep: true });

    const getRegionName = (region) => {
      const name = regionList.value.find(item => item.region === region)?.regionName;

      return name ? `${name}(${region})` : region;
    };

    return {
      ...toRefs(state),
      selectorRef,
      handleGetDefaultData,
      handleGetSearchTableData,
      handleCheckChange,
      handleMenuClick,
      getDefaultSelections,
      handleRemoveNode,
      handleChange,
      getRowDisabledStatus,
      getRowTipsContent,
      handleGetData,
      defaultSelectedNode,
    };
  },
});
</script>
<style lang="postcss" scoped>
/deep/ .preview-count {
    color: #3a84ff;
    font-weight: 700;
    padding: 0 2px;
    font-size: 12px;
}
</style>
