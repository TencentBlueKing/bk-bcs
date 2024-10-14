import {
  hostCheck as hostCheckAdapter,
  hostsDetails as hostsDetailsAdapter,
  topologyHostCount as topologyHostCountAdapter,
  topolopyHostIdsNodes as topologyHostIdsNodesAdapter } from '@blueking/ip-selector/dist/adapter';
import createFactory from '@blueking/ip-selector/dist/vue2.6.x.esm';

import '@blueking/ip-selector/dist/styles/vue2.6.x.css';
import {
  hostCheck,
  hostInfoByHostId,
  topologyHostCount,
  topologyHostIdList,
} from '@/api/modules/cluster-manager';
import $store from '@/store';

const $biz = $store.state.curProject.businessID;
const $scope = 'biz';
const Service = {
  // panel顺序和表格默认列配置
  fetchCustomSettings() {
    return new Promise(resolve => resolve({}));
  },
  updateCustomSettings() {
    return new Promise(resolve => resolve({}));
  },
  // 获取topo树（带主机数量）
  async fetchTopologyHostCount(params) {
    const data = await topologyHostCount({
      ...params,
      $biz,
      $scope,
      scopeList: [
        {
          scopeType: $scope,
          scopeId: $biz,
        },
      ],
    }).catch(() => []);
    return topologyHostCountAdapter(data);
  },
  // 获取topo树当前页的主机列表（在单文件组件里面配）
  // async fetchTopologyHostsNodes(params) {
  //   const data = await topologyHostsNodes({
  //     ...params,
  //     $biz,
  //     $scope,
  //   }).catch(() => []);
  //   return topologyHostsNodesAdapter(data);
  // },
  // 获取topo树下所以主机ID
  async fetchTopologyHostIdList(params) {
    const data = await topologyHostIdList({
      ...params,
      $biz,
      $scope,
    }).catch(() => []);
    return topologyHostIdsNodesAdapter(data);
  },
  // 获取主机ID对应的主机详情
  async fetchHostInfoByHostId(params) {
    const data = await hostInfoByHostId({
      ...params,
      $biz,
      $scope,
    }).catch(() => []);
    return hostsDetailsAdapter(data);
  },
  // 手动输入
  async fetchHostCheck(params) {
    const data = await hostCheck({
      ...params,
      $biz,
      $scope,
    }).catch(() => []);
    return hostCheckAdapter(data);
  },
};
// 初始化配置，创建组件
const IpSelector = createFactory({
  // 组件版本（改变版本重置用户自定义配置）
  version: '',
  // eslint-disable-next-line max-len
  // 需要支持的面板（'staticTopo', 'dynamicTopo', 'dynamicGroup', 'serviceTemplate', 'setTemplate', 'serviceInstance', 'manualInput'）
  panelList: ['staticTopo', 'manualInput'],
  // 面板选项的值是否唯一
  unqiuePanelValue: false,
  // 字段命名风格（'camelCase', 'kebabCase'）
  nameStyle: 'camelCase',
  // 主机列表全选模式，false: 本页全选；true: 跨页全选
  hostTableDefaultSelectAllMode: false,
  hostOnlyValid: true,
  // 自定义主机列表列
  hostTableCustomColumnList: [
    // {
    //   key: 'collectStatus',
    //   index: 5,
    //   width: '100px',
    //   label: '采集状态',
    //   renderHead: h => h('span', '采集状态'),
    //   field: 'collect_status',
    //   renderCell: (h, row) => h('span', row.collect_status || '--'),
    // }
  ],
  nodeTableCustomColumnList: [],  // 自定义动态拓扑列表列 同上
  serviceTemplateTableCustomColumnList: [],  // 自定义服务模板列表列 同上
  setTemplateCustomColumnList: [],  // 自定义集群模板列表列 同上
  hostMemuExtends: [
    // {
    //   name: '复制采集状态异常',
    //   action: () => {
    //     console.log('复制成功');
    //   },
    // },
  ],
  // 主机预览字段显示
  hostViewFieldRender: host => host.host_id,
  // 主机列表显示列（默认值：['ip', 'ipv6', 'alive', 'osName']），按配置顺序显示列
  // 内置所有列的 key ['ip', 'ipv6', 'cloudArea', 'alive', 'hostName', 'osName', 'coludVerdor', 'osType', 'hostId', 'agentId']
  hostTableRenderColumnList: [],

  // 创建时是否提示 service 信息
  serviceConfigError: false,

  // 需要的数据源配置（返回 Promise）
  // 主机拓扑
  fetchTopologyHostCount: Service.fetchTopologyHostCount,
  // fetchTopologyHostsNodes: Service.fetchTopologyHostsNodes,
  fetchTopologyHostIdsNodes: Service.fetchTopologyHostIdList,
  fetchHostsDetails: Service.fetchHostInfoByHostId,
  fetchHostCheck: Service.fetchHostCheck,

  // 自定义配置
  fetchCustomSettings: Service.fetchCustomSettings,
  // 更新配置
  updateCustomSettings: Service.updateCustomSettings,
  // 系统相关配置
  fetchConfig: () => Promise.resolve()
    .then(() => ({
    })),
});

export default IpSelector;
