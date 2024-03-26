import { RouteConfig } from 'vue-router';

import { ICluster } from '@/composables/use-app';
import $store from '@/store';
// 集群首页
const Cluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/index.vue');
// 创建集群
const ClusterCreate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/index.vue');
// 创建腾讯云集群
const CreateTencentCloudCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/add-cluster.vue');
// VCluster集群
const CreateVCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/vcluster/add-vcluster.vue');
// ee版本创建集群流程
const CreateK8SCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create-k8s.vue');
// const CreateCluster = () => import('@/views/cluster-manage/add/create-cluster.vue');
// import模式
const ImportCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import-cluster.vue');
const ImportGoogleCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/google-cloud.vue');
const ImportAzureCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/azure-cloud.vue');
const ClusterNodeOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node-overview.vue');
// const Node = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node.vue');
const NodeTemplate = () => import(/* webpackChunkName: 'cluster'  */'@/views/cluster-manage/node-template/node-template.vue');
const EditNodeTemplate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-template/edit-node-template.vue');
const AddClusterNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/add-nodes.vue');
const NodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/node-pool.vue');
const NodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/node-pool-detail.vue');
const EditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/edit-node-pool.vue');
const AutoScalerConfig = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/autoscaler-config.vue');
const InternalNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/node-pool.vue');
const InternalNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/node-pool-detail.vue');
const InternalEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/edit-node-pool.vue');
const PodDetail = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/detail/index.vue');
const CreateTKECluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/tencent/index.vue');

// google ca
const GoogleNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/node-pool.vue');
const GoogleNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/node-pool-detail.vue');
const GoogleEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/edit-node-pool.vue');

// 集群管理
export default [
  {
    path: 'clusters',
    name: 'clusterMain',
    props: route => ({ ...route.query, ...route.params }),
    component: Cluster,
    meta: {
      hideMenu: true,
    },
  },
  // 创建集群
  {
    path: 'clusters/create',
    name: 'clusterCreate',
    component: ClusterCreate,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // 创建腾讯云集群
  {
    path: 'clusters/tencent',
    name: 'createTencentCloudCluster',
    component: CreateTencentCloudCluster,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // 创建腾讯云集群
  {
    path: 'clusters/cloud/tencent',
    name: 'createTKECluster',
    component: CreateTKECluster,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // 创建k8s原生集群
  {
    path: 'clusters/k8s',
    name: 'createK8SCluster',
    component: CreateK8SCluster,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // {
  //   path: 'clusters/create',
  //   name: 'createCluster',
  //   component: CreateCluster,
  //   meta: {
  //     menuId: 'CLUSTER',
  //     title: window.i18n.t('cluster.button.addCluster'),
  //   },
  // },
  // 创建VCluster集群
  {
    path: 'clusters/vcluster',
    name: 'createVCluster',
    component: CreateVCluster,
    meta: {
      menuId: 'CLUSTER',
      id: 'VCLUSTER',
      hideMenu: true,
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
      hideMenu: true,
    },
  },
  // 导入集群 - 谷歌云
  {
    path: 'clusters/import/google-cloud',
    name: 'importGoogleCluster',
    component: ImportGoogleCluster,
    props: true,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.create.title.import'),
      hideMenu: true,
    },
  },
  // 导入集群 - 微软云
  {
    path: 'clusters/import/azure-cloud',
    name: 'importAzureCluster',
    component: ImportAzureCluster,
    props: true,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.create.title.import'),
      hideMenu: true,
    },
  },
  // 集群里的具体节点
  {
    path: 'clusters/:clusterId/nodes/:nodeName/detail',
    name: 'clusterNodeOverview',
    props: true,
    component: ClusterNodeOverview,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // Pods详情
  {
    path: 'clusters/:clusterId/nodes/:nodeName/:category/namespaces/:namespace/:name',
    name: 'nodePodDetail',
    props: route => ({ ...route.params, kind: route.query.kind, crd: route.query.crd }),
    component: PodDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // {
  //   path: 'nodes',
  //   name: 'nodeMain',
  //   component: Node,
  //   meta: {
  //     title: window.i18n.t('nav.nodeList'),
  //     hideBack: true,
  //   },
  // },
  {
    path: 'node-template',
    name: 'nodeTemplate',
    component: NodeTemplate,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'node-template/create',
    name: 'addNodeTemplate',
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('cluster.nodeTemplate.title.create'),
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'node-template/:nodeTemplateID',
    name: 'editNodeTemplate',
    props: true,
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('cluster.nodeTemplate.title.update'),
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'clusters/:clusterId/nodes/add',
    name: 'addClusterNode',
    props: route => ({ ...route.query, ...route.params }),
    component: AddClusterNode,
    meta: {
      title: window.i18n.t('cluster.nodeList.create.text'),
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'clusters/:clusterId/autoscaler',
    name: 'autoScalerConfig',
    props: true,
    component: AutoScalerConfig,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/nodepools',
    name: 'nodePool',
    props: true,
    component: window.REGION === 'ieod' ? InternalNodePool : NodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
    beforeEnter(to, from, next) {
      const clusterList = $store.state.cluster.clusterList as ICluster[];
      const cluster = clusterList.find(item => item.clusterID === to.params.clusterId);
      if (cluster?.provider === 'gcpCloud') {
        next({
          name: 'googleNodePool',
          params: {
            ...to.params,
          },
          query: {
            ...to.query,
          },
        });
      } else {
        next();
      }
    },
  },
  {
    path: 'cluster/:clusterId/nodepools/:nodeGroupID',
    name: 'editNodePool',
    props: true,
    component: window.REGION === 'ieod' ? InternalEditNodePool : EditNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
    beforeEnter(to, from, next) {
      const clusterList = $store.state.cluster.clusterList as ICluster[];
      const cluster = clusterList.find(item => item.clusterID === to.params.clusterId);
      if (cluster?.provider === 'gcpCloud') {
        next({
          name: 'googleEditNodePool',
          params: {
            ...to.params,
          },
          query: {
            ...to.query,
          },
        });
      } else {
        next();
      }
    },
  },
  {
    path: 'cluster/:clusterId/nodepools/:nodeGroupID/detail',
    name: 'nodePoolDetail',
    props: true,
    component: window.REGION === 'ieod' ? InternalNodePoolDetail : NodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
    beforeEnter(to, from, next) {
      const clusterList = $store.state.cluster.clusterList as ICluster[];
      const cluster = clusterList.find(item => item.clusterID === to.params.clusterId);
      if (cluster?.provider === 'gcpCloud') {
        next({
          name: 'googleNodePoolDetail',
          params: {
            ...to.params,
          },
          query: {
            ...to.query,
          },
        });
      } else {
        next();
      }
    },
  },
  // google ca
  {
    path: 'cluster/:clusterId/google/nodepools',
    name: 'googleNodePool',
    props: true,
    component: GoogleNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/google/nodepools/:nodeGroupID',
    name: 'googleEditNodePool',
    props: true,
    component: GoogleEditNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/google/nodepools/:nodeGroupID/detail',
    name: 'googleNodePoolDetail',
    props: true,
    component: GoogleNodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
] as RouteConfig[];
