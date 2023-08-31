import { RouteConfig } from 'vue-router';
// 集群首页
const Cluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/index.vue');
// 创建集群
const ClusterCreate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/create/cluster-type.vue');
// 创建腾讯云集群
const CreateTencentCloudCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/create/add-cluster.vue');
// VCluster集群
const CreateVCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/create/add-vcluster.vue');
// ee版本创建集群流程
const CreateCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/create/create-cluster.vue');
// import模式
const ImportCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/create/import-cluster.vue');
// 集群详情
const ClusterDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/cluster-detail.vue');
const ClusterNodeOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node-overview.vue');
const Node = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node.vue');
const NodeTemplate = () => import(/* webpackChunkName: 'cluster'  */'@/views/cluster-manage/node-template/node-template.vue');
const EditNodeTemplate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-template/edit-node-template.vue');
const AddClusterNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/add-cluster-node.vue');
const NodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/tencent/node-pool.vue');
const NodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/tencent/node-pool-detail.vue');
const EditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/tencent/edit-node-pool.vue');
const AutoScalerConfig = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/autoscaler-config.vue');
const InternalNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/internal/node-pool.vue');
const InternalNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/internal/node-pool-detail.vue');
const InternalEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/autoscaler/internal/edit-node-pool.vue');
const PodDetail = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/detail/index.vue');

// 云凭证
const tencentCloud = () => import(/* webpackChunkName: 'project' */'@/views/cluster-manage/cloudtoken/tencentCloud.vue');

// 集群管理
export default [
  {
    path: 'clusters',
    name: 'clusterMain',
    component: Cluster,
  },
  // 创建集群
  {
    path: 'clusters/create',
    name: 'clusterCreate',
    component: ClusterCreate,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.button.addCluster'),
    },
  },
  // 创建集群
  {
    path: 'clusters/tencent',
    name: 'createTencentCloudCluster',
    component: CreateTencentCloudCluster,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'clusters/create',
    name: 'createCluster',
    component: CreateCluster,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.button.addCluster'),
    },
  },
  // 创建VCluster集群
  {
    path: 'clusters/vcluster',
    name: 'createVCluster',
    component: CreateVCluster,
    meta: {
      menuId: 'CLUSTER',
      id: 'VCLUSTER',
    },
  },
  // 导入集群 - import导入模式
  {
    path: 'clusters/:importType/import',
    name: 'importCluster',
    component: ImportCluster,
    props: true,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.create.title.import'),
    },
  },
  // 集群详情
  {
    path: 'clusters/:clusterId',
    name: 'clusterDetail',
    props: route => ({ ...route.params, ...route.query }),
    component: ClusterDetail,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  // 集群总览
  {
    path: 'clusters/:clusterId',
    name: 'clusterOverview',
    redirect: {
      name: 'clusterDetail',
      query: {
        active: 'overview',
      },
    },
  },
  // 节点列表
  {
    path: 'clusters/:clusterId',
    name: 'clusterNode',
    redirect: {
      name: 'clusterDetail',
      query: {
        active: 'node',
      },
    },
    meta: {
      menuId: 'OVERVIEW',
    },
  },
  // 集群里的集群信息
  {
    path: 'clusters/:clusterId',
    name: 'clusterInfo',
    redirect: {
      name: 'clusterDetail',
      query: {
        active: 'info',
      },
    },
    meta: {
      menuId: 'OVERVIEW',
    },
  },
  // 集群里的具体节点
  {
    path: 'clusters/:clusterId/nodes/:nodeName/detail',
    name: 'clusterNodeOverview',
    props: true,
    component: ClusterNodeOverview,
    meta: {
      menuId: 'NODE',
    },
  },
  // Pods详情
  {
    path: 'clusters/:clusterId/nodes/:nodeName/:category/namespaces/:namespace/:name',
    name: 'nodePodDetail',
    props: route => ({ ...route.params, kind: route.query.kind, crd: route.query.crd }),
    component: PodDetail,
    meta: {
      menuId: 'NODE',
    },
  },
  {
    path: 'nodes',
    name: 'nodeMain',
    component: Node,
    meta: {
      title: window.i18n.t('nav.nodeList'),
      hideBack: true,
    },
  },
  {
    path: 'node-template',
    name: 'nodeTemplate',
    component: NodeTemplate,
    meta: {
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: 'node-template/create',
    name: 'addNodeTemplate',
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('cluster.nodeTemplate.title.create'),
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: 'node-template/:nodeTemplateID',
    name: 'editNodeTemplate',
    props: true,
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('cluster.nodeTemplate.title.update'),
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: 'clusters/:clusterId/nodes/add',
    name: 'addClusterNode',
    props: true,
    component: AddClusterNode,
    meta: {
      title: window.i18n.t('cluster.nodeList.create.text'),
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'clusters/:clusterId/autoscaler',
    name: 'autoScalerConfig',
    props: true,
    component: AutoScalerConfig,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'cluster/:clusterId/nodepools',
    name: 'nodePool',
    props: true,
    component: window.REGION === 'ieod' ? InternalNodePool : NodePool,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'cluster/:clusterId/nodepools/:nodeGroupID',
    name: 'editNodePool',
    props: true,
    component: window.REGION === 'ieod' ? InternalEditNodePool : EditNodePool,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'cluster/:clusterId/nodepools/:nodeGroupID/detail',
    name: 'nodePoolDetail',
    props: true,
    component: window.REGION === 'ieod' ? InternalNodePoolDetail : NodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  {
    path: 'cluster/tencent-cloud',
    name: 'tencentCloud',
    component: tencentCloud,
    meta: {
      title: 'Tencent Cloud',
      hideBack: true,
    },
  },
] as RouteConfig[];
