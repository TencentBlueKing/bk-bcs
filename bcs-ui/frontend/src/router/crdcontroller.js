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

const Index = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/index.vue');
const DBList = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/db_list.vue');
const LogList = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/log_list.vue');
const Detail = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/detail.vue');
const BcsPolaris = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/polaris_list.vue');
const NewLogList = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/new-log-list.vue');

const childRoutes = [
  {
    path: ':projectCode/tools',
    name: 'dbCrdcontroller',
    component: Index,
    meta: {
      crdKind: 'DbPrivilege',
    },
  },

  {
    path: ':projectCode/tools/log',
    name: 'logCrdcontroller',
    component: Index,
    meta: {
      crdKind: 'BcsLog',
    },
  },

  {
    path: ':projectCode/cluster/:clusterId/crdcontroller/DbPrivilege/instances',
    name: 'crdcontrollerDBInstances',
    component: DBList,
    meta: {
      menuId: 'COMPONENTS',
    },
  },

  {
    path: ':projectCode/cluster/:clusterId/crdcontroller/BcsPolaris/instances',
    name: 'crdcontrollerPolarisInstances',
    component: BcsPolaris,
  },

  {
    path: ':projectCode/cluster/:clusterId/crdcontroller/BcsLog/instances',
    props: true,
    name: 'crdcontrollerLogInstances',
    component: window.REGION === 'ieod' ? LogList : NewLogList,
  },

  {
    path: ':projectCode/cluster/:clusterId/crdcontroller/:chartName/instances/:id',
    name: 'crdcontrollerInstanceDetail',
    component: Detail,
    meta: {
      menuId: 'COMPONENTS',
    },
  },
];

export default childRoutes;
