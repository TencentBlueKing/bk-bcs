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

const Node = () => import(/* webpackChunkName: 'node' */'@/views/node/node.vue');
const NodeTemplate = () => import(/* webpackChunkName: 'node'  */'@/views/node/node-template.vue');
const EditNodeTemplate = () => import(/* webpackChunkName: 'node' */'@/views/node/edit-node-template.vue');
const AddClusterNode = () => import(/* webpackChunkName: 'node' */'@/views/node/add-cluster-node.vue');

const childRoutes = [
  // domain/bcs/projectCode/node 节点页面
  {
    path: ':projectCode/node',
    name: 'nodeMain',
    component: Node,
    meta: {
      title: window.i18n.t('节点'),
      hideBack: true,
    },
  },
  {
    path: ':projectCode/node-template',
    name: 'nodeTemplate',
    component: NodeTemplate,
    meta: {
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: ':projectCode/node-template/create',
    name: 'addNodeTemplate',
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('新建节点模板'),
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: ':projectCode/node-template/edit/:nodeTemplateID',
    name: 'editNodeTemplate',
    props: true,
    component: EditNodeTemplate,
    meta: {
      title: window.i18n.t('编辑节点模板'),
      menuId: 'NODETEMPLATE',
    },
  },
  {
    path: ':projectCode/cluster/:clusterId/node/add',
    name: 'addClusterNode',
    props: true,
    component: AddClusterNode,
    meta: {
      title: window.i18n.t('添加节点'),
      menuId: 'CLUSTER',
    },
  },
];

export default childRoutes;
