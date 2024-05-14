// 插件管理
const Index = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/index.vue');
const DBList = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/db_list.vue');
const LogList = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/log_list.vue');
const Detail = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/detail.vue');
const BcsPolaris = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/polaris_list.vue');
const MetricManage = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/metric/metric-manage.vue');
const LoadBalanceDetail = () => import(/* webpackChunkName: 'plugin' */'@/views/deploy-manage/templateset/network/loadbalance-detail.vue');

// 日志采集
const logCollector = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/log-collector/log-collector.vue');

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
    component: logCollector,
  },
  // DB授权配置
  {
    path: 'clusters/:clusterId/tools/db',
    props: true,
    name: 'crdcontrollerDBInstances',
    component: DBList,
    meta: {
      menuId: 'TOOLS',
    },
  },
  // polaris配置
  {
    path: 'clusters/:clusterId/tools/polaris',
    props: true,
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
  // // loadBalance
  // {
  //   path: 'loadbalance',
  //   name: 'loadBalance',
  //   component: LoadBalance,
  // },
  // loadBalance 详情
  {
    path: 'clusters/:clusterId/namespaces/:namespace/loadbalance/:lbId',
    name: 'loadBalanceDetail',
    component: LoadBalanceDetail,
  },
];
