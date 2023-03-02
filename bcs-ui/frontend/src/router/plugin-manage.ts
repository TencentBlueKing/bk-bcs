// 插件管理
const Index = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/index.vue');
const DBList = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/db_list.vue');
const LogList = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/log_list.vue');
const Detail = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/detail.vue');
const BcsPolaris = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/polaris_list.vue');
const MetricManage = () => import(/* webpackChunkName: 'plugin' */'@/views/metric/index.vue');
const LoadBalance = () => import(/* webpackChunkName: 'plugin' */'@/views/network/loadbalance.vue');
const LoadBalanceDetail = () => import(/* webpackChunkName: 'plugin' */'@/views/network/loadbalance-detail.vue');

// 新版日志采集
const NewLogIndex = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/new-log/log.vue');
const NewLogList = () => import(/* webpackChunkName: 'plugin' */'@/views/crdcontroller/new-log/log-list.vue');

export default [
  // 组件库
  {
    path: 'tools',
    name: 'dbCrdcontroller',
    component: Index,
    meta: {
      crdKind: 'DbPrivilege',
    },
  },
  // 日志采集
  {
    path: 'log-collector',
    name: 'logCrdcontroller',
    component: Index,
    meta: {
      crdKind: 'BcsLog',
    },
  },
  // 新版日志采集
  {
    path: 'log',
    name: 'newLogController',
    component: NewLogIndex,
  },
  // 新版日志采集 - 配置
  {
    path: 'clusters/:clusterId/tools/log',
    props: true,
    name: 'newLogList',
    component: NewLogList,
    meta: {
      menuId: 'NEW_LOG',
    },
  },
  // DB授权配置
  {
    path: 'clusters/:clusterId/tools/db',
    name: 'crdcontrollerDBInstances',
    component: DBList,
    meta: {
      menuId: 'TOOLS',
    },
  },
  // polaris配置
  {
    path: 'clusters/:clusterId/tools/polaris',
    name: 'crdcontrollerPolarisInstances',
    component: BcsPolaris,
    meta: {
      menuId: 'TOOLS',
    },
  },
  // 日志配置
  {
    path: 'clusters/:clusterId/tools/log-collector',
    props: true,
    name: 'crdcontrollerLogInstances',
    component: LogList,
    meta: {
      menuId: 'LOG',
    },
  },
  // 更新组件
  {
    path: 'clusters/:clusterId/tools/:chartName/:id',
    name: 'crdcontrollerInstanceDetail',
    component: Detail,
    meta: {
      menuId: 'TOOLS',
    },
  },
  // metric
  {
    path: 'metric',
    name: 'metricManage',
    component: MetricManage,
  },
  // loadBalance
  {
    path: 'loadbalance',
    name: 'loadBalance',
    component: LoadBalance,
  },
  // loadBalance 详情
  {
    path: 'clusters/:clusterId/namespaces/:namespace/loadbalance/:lbId',
    name: 'loadBalanceDetail',
    component: LoadBalanceDetail,
  },
];
