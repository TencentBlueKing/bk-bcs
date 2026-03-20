// 插件管理
const Index = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/index.vue');
const DBList = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/db_list.vue');
const Detail = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/detail.vue');
const BcsPolaris = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/polaris_list.vue');
const MetricManage = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/metric/metric-manage.vue');

// 日志采集
const logCollector = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/tools/log-collector/log-collector.vue');

// 服务网格
const ServiceMesh = () => import(/* webpackChunkName: 'plugin' */'@/views/plugin-manage/service-mesh/index.vue');

export default [
  // 组件库
  {
    path: 'tools',
    name: 'dbCrdcontroller',
    component: Index,
    meta: {
      crdKind: 'DbPrivilege',
      resource: window.i18n.t('nav.clusterTools'),
    },
  },
  // 日志采集
  {
    path: 'log-collector',
    name: 'logCrdcontroller',
    component: logCollector,
    meta: {
      resource: window.i18n.t('nav.log'),
    },
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
    meta: {
      resource: window.i18n.t('nav.metric'),
    },
  },
  // 服务网格
  {
    path: 'service-mesh',
    name: 'serviceMesh',
    component: ServiceMesh,
    meta: {
      resource: window.i18n.t('serviceMesh.title'),
    },
  },
];
