/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/

// 集群首页
const Cluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster');
// 创建集群
const ClusterCreate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-cluster');
// 表单模式
const CreateFormCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-form-cluster');
// ee版本创建集群流程
const CreateFormClusterEE = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-form-cluster-ee');
// import模式
const CreateImportCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-import-cluster');
// 集群模板创建
const CreateClusterTemplate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-template-cluster');
// 集群详情
const ClusterDetail = () => import(/* webpackChunkName: 'cluster-detail' */'@/views/node/cluster-detail');
const ClusterNodeOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/node-overview');
const ContainerDetailForNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/container');

const childRoutes = [
  {
    path: ':projectCode',
    redirect: {
      name: 'clusterMain',
    },
  },
  {
    path: ':projectCode/',
    redirect: {
      name: 'clusterMain',
    },
  },
  // domain/bcs/projectId => domain/bcs/projectCode/cluster 容器服务的首页是具体项目的集群页面
  // 集群首页
  {
    path: ':projectCode/cluster',
    name: 'clusterMain',
    component: Cluster,
  },
  // 创建集群
  {
    path: ':projectCode/cluster/create',
    name: 'clusterCreate',
    component: ClusterCreate,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('新建集群'),
    },
  },
  // 创建集群 - 表单模式
  {
    path: ':projectCode/cluster/create/form',
    name: 'createFormCluster',
    component: CreateFormCluster,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('自建集群'),
    },
  },
  // 创建集群 - 表单模式（ee版本）
  {
    path: ':projectCode/cluster/create/form-ee',
    name: 'createFormClusterEE',
    component: CreateFormClusterEE,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('自建集群'),
    },
  },
  // 创建集群 - import导入模式
  {
    path: ':projectCode/cluster/create/import',
    name: 'createImportCluster',
    component: CreateImportCluster,
    meta: {
      menuId: 'CLUSTER',
      title: window.i18n.t('导入集群'),
    },
  },
  // 创建集群模板
  {
    path: ':projectCode/cluster/create/template',
    name: 'createClusterTemplate',
    component: CreateClusterTemplate,
    meta: {
      title: window.i18n.t('新建集群模板'),
    },
  },
  {
    path: ':projectCode/cluster/:clusterId/detail',
    name: 'clusterDetail',
    props: route => ({ ...route.params, ...route.query }),
    component: ClusterDetail,
    meta: {
      menuId: 'CLUSTER',
    },
  },
  // 集群总览
  {
    path: ':projectCode/cluster/:clusterId/overview',
    name: 'clusterOverview',
    redirect: {
      name: 'clusterDetail',
      query: {
        active: 'overview',
      },
    },
    alias: ':projectCode/cluster/:clusterId',
  },
  // 集群里的节点列表
  {
    path: ':projectCode/cluster/:clusterId/node',
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
    path: ':projectCode/cluster/:clusterId/info',
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
    path: ':projectCode/cluster/:clusterId/node/:nodeName/detail',
    name: 'clusterNodeOverview',
    component: ClusterNodeOverview,
    meta: {
      menuId: 'NODE',
    },
  },
  // 节点详情页面跳转的容器详情页面
  {
    path: ':projectCode/cluster/:clusterId/node/:nodeId/container/:containerId',
    name: 'containerDetailForNode',
    component: ContainerDetailForNode,
    meta: {
      menuId: 'OVERVIEW',
    },
  },
];

export default childRoutes;
