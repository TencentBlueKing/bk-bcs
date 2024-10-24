import { RouteConfig } from 'vue-router';

import { ICluster } from '@/composables/use-app';
import $store from '@/store';
// 集群首页
const Cluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/cluster/index.vue');
// 创建集群
const ClusterCreate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/index.vue');
// 创建腾讯云集群
const CreateTencentCloudCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/tencent-cloud.vue');
// VCluster集群
const CreateVCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/vcluster/vcluster.vue');
// ee版本创建集群流程
const CreateK8SCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/k8s.vue');
// 创建aws云集群
const CreateAWSCloudCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/aws-cloud/index.vue');
// 创建aws云集群
const CreateAzureCloudCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/azure-cloud/index.vue');

// import模式
const ImportCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/import-cluster.vue');
const ImportGoogleCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/google-cloud.vue');
const importBkSopsCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/bk-sops.vue');
const ImportAzureCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/azure-cloud.vue');
const ImportHuaweiCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/huawei-cloud.vue');
const ImportAwsCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/import/amazon-cloud.vue');
const ClusterNodeOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node-overview.vue');
// const Node = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/node.vue');
const NodeTemplate = () => import(/* webpackChunkName: 'cluster'  */'@/views/cluster-manage/node-template/node-template.vue');
const EditNodeTemplate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-template/edit-node-template.vue');
const AddClusterNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/add-nodes.vue');
const batchSettingNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/node-list/batch-settings.vue');
const NodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/node-pool.vue');
const NodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/node-pool-detail.vue');
const EditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/tencent/edit-node-pool.vue');
const AutoScalerConfig = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/autoscaler-config.vue');
const InternalNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/node-pool.vue');
const InternalNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/node-pool-detail.vue');
const InternalEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/internal/edit-node-pool.vue');
const PodDetail = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/detail/index.vue');
const CreateTKECluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/add/create/tencent-public-cloud/index.vue');

// google ca
const GoogleNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/node-pool.vue');
const GoogleNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/node-pool-detail.vue');
const GoogleEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/google/edit-node-pool.vue');

const NamespaceCreate = () => import(/* webpackChunkName: 'dashboard' */'@/views/cluster-manage/namespace/create.vue');
// azure ca
// 新建节点池
const AzureNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/azure/node-pool.vue');
// 扩缩容记录
const AzureNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/azure/node-pool-detail.vue');
// 编辑配置
const AzureEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/azure/edit-node-pool.vue');

// huawei ca
// 新建节点池
const HuaweiNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/huawei/node-pool.vue');
// 扩缩容记录
const HuaweiNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/huawei/node-pool-detail.vue');
// 编辑配置
const HuaweiEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/huawei/edit-node-pool.vue');

// aws ca
// 新建节点池
const AwsNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/aws/node-pool.vue');
// 扩缩容记录
const AwsNodePoolDetail = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/aws/node-pool-detail.vue');
// 编辑配置
const AwsEditNodePool = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster-manage/autoscaler/aws/edit-node-pool.vue');

// 集群管理
export default [
  {
    path: 'clusters',
    name: 'clusterMain',
    props: route => ({ ...route.query, ...route.params }),
    component: Cluster,
    meta: {
      hideMenu: true,
      keepAlive: true,
    },
  },
  {
    path: 'clusters/:clusterId/namespaces/create',
    name: 'createNamespace',
    props: true,
    component: NamespaceCreate,
    meta: {
      menuId: 'CLUSTER',
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
  // 创建亚马逊云集群
  {
    path: 'clusters/cloud/aws',
    name: 'CreateAWSCloudCluster',
    component: CreateAWSCloudCluster,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // 创建微软云集群
  {
    path: 'clusters/cloud/azure',
    name: 'CreateAzureCloudCluster',
    component: CreateAzureCloudCluster,
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
  // 导入集群 - bk-sops
  {
    path: 'clusters/import/bk-sops',
    name: 'importBkSopsCluster',
    component: importBkSopsCluster,
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
  // 导入集群 - 华为云
  {
    path: 'clusters/import/huawei-cloud',
    name: 'importHuaweiCluster',
    component: ImportHuaweiCluster,
    props: true,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('cluster.create.title.import'),
      hideMenu: true,
    },
  },
  // 导入集群 - 亚马逊云
  {
    path: 'clusters/import/amazon-cloud',
    name: 'importAwsCluster',
    component: ImportAwsCluster,
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
    props: route => ({ ...route.params, ...route.query }),
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
      showClusterName: true,
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
      let name = '';
      // 优化,增加azureCA,huawei节点新建
      switch (cluster?.provider) {
        case 'gcpCloud':
          name = 'googleNodePool';
          break;
        case 'azureCloud':
          name = 'azureNodePool';
          break;
        case 'huaweiCloud':
          name = 'huaweiNodePool';
          break;
        case 'awsCloud':
          name = 'awsNodePool';
          break;
      }
      name ? next({
        name,
        params: { ...to.params },
        query: { ...to.query },
      }) : next();
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
      let name = '';
      // 优化，增加azureCA,huaweiCA节点配置
      switch (cluster?.provider) {
        case 'gcpCloud':
          name = 'googleEditNodePool';
          break;
        case 'azureCloud':
          name = 'azureEditNodePool';
          break;
        case 'huaweiCloud':
          name = 'huaweiEditNodePool';
          break;
        case 'awsCloud':
          name = 'awsEditNodePool';
          break;
      }
      name ? next({
        name,
        params: { ...to.params },
        query: { ...to.query },
      }) : next();
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
      let name = '';
      // 优化，增加azureCA,huaweiCA节点池详情
      switch (cluster?.provider) {
        case 'gcpCloud':
          name = 'googleNodePoolDetail';
          break;
        case 'azureCloud':
          name = 'azureNodePoolDetail';
          break;
        case 'huaweiCloud':
          name = 'huaweiNodePoolDetail';
          break;
        case 'awsCloud':
          name = 'awsNodePoolDetail';
          break;
      }
      name ? next({
        name,
        params: { ...to.params },
        query: { ...to.query },
      }) : next();
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
  // azure ca
  {
    path: 'cluster/:clusterId/azure/nodepools',
    name: 'azureNodePool',
    props: true,
    component: AzureNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/azure/nodepools/:nodeGroupID',
    name: 'azureEditNodePool',
    props: true,
    component: AzureEditNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/azure/nodepools/:nodeGroupID/detail',
    name: 'azureNodePoolDetail',
    props: true,
    component: AzureNodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // huawei ca
  {
    path: 'cluster/:clusterId/huawei/nodepools',
    name: 'huaweiNodePool',
    props: true,
    component: HuaweiNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/huawei/nodepools/:nodeGroupID',
    name: 'huaweiEditNodePool',
    props: true,
    component: HuaweiEditNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/huawei/nodepools/:nodeGroupID/detail',
    name: 'huaweiNodePoolDetail',
    props: true,
    component: HuaweiNodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // 批量设置（标签或污点）
  {
    path: 'clusters/:clusterId/nodes/setting/:type',
    name: 'batchSettingNode',
    props: route => ({ ...route.query, ...route.params }),
    component: batchSettingNode,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  // aws ca
  {
    path: 'cluster/:clusterId/aws/nodepools',
    name: 'awsNodePool',
    props: true,
    component: AwsNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/aws/nodepools/:nodeGroupID',
    name: 'awsEditNodePool',
    props: true,
    component: AwsEditNodePool,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
  {
    path: 'cluster/:clusterId/aws/nodepools/:nodeGroupID/detail',
    name: 'awsNodePoolDetail',
    props: true,
    component: AwsNodePoolDetail,
    meta: {
      menuId: 'CLUSTER',
      hideMenu: true,
    },
  },
] as RouteConfig[];
