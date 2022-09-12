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

import Vue from 'vue';
import VueRouter from 'vue-router';
import store from '@/store';
import http from '@/api';
import resourceRoutes from '@/router/resource';
import nodeRoutes from '@/router/node';
import mcRoutes from '@/router/mc';
import depotRoutes from './depot';
import metricRoutes from './metric';
import clusterRoutes from './cluster';
import appRoutes from './app';
import configurationRoutes from './configuration';
import networkRoutes from './network';
import helmRoutes from './helm';
import HPARoutes from './hpa';
import crdController from './crdcontroller.js';
import storageRoutes from './storage';
import dashboardRoutes from './dashboard';
import menuConfig from '@/store/menu';
import cloudtokenRoutes from './cloudtoken';
import i18n from '@/i18n/i18n-setup';

const originalPush = VueRouter.prototype.push;
const originalReplace = VueRouter.prototype.replace;
// push
VueRouter.prototype.push = function push(location, onResolve, onReject) {
  if (onResolve || onReject) return originalPush.call(this, location, onResolve, onReject);
  return originalPush.call(this, location).catch(err => err);
};
// replace
VueRouter.prototype.replace = function push(location, onResolve, onReject) {
  if (onResolve || onReject) return originalReplace.call(this, location, onResolve, onReject);
  return originalReplace.call(this, location).catch(err => err);
};
Vue.use(VueRouter);

const Entry = () => import(/* webpackChunkName: entry */'@/views/index');
const NotFound = () => import(/* webpackChunkName: 'none' */'@/components/exception');
const ProjectManage = () => import(/* webpackChunkName: 'projectmanage' */'@/views/project/project.vue');
const userToken = () => import(/* webpackChunkName: 'token' */'@/views/token/token.vue');
const Forbidden = () => import(/* webpackChunkName: 'none' */'@/components/exception/403.vue');

const router = new VueRouter({
  mode: 'history',
  routes: [
    {
      path: `${SITE_URL}`,
      name: 'entry',
      component: Entry,
      children: [
        ...clusterRoutes,
        ...nodeRoutes,
        ...appRoutes,
        ...configurationRoutes,
        ...networkRoutes,
        ...resourceRoutes,
        ...depotRoutes,
        ...metricRoutes,
        ...mcRoutes,
        ...helmRoutes,
        ...HPARoutes,
        ...crdController,
        ...storageRoutes,
        ...dashboardRoutes,
        ...cloudtokenRoutes,
      ],
    },
    {
      path: '/:projectCode/api-key',
      name: 'token',
      component: userToken,
    },
    {
      path: '/project/manage',
      name: 'projectManage',
      component: ProjectManage,
    },
    {
      path: '/exception/403',
      name: '403',
      props: route => ({ ...route.params, ...route.query }),
      component: Forbidden,
    },
    // 404
    {
      path: '*',
      name: '404',
      component: NotFound,
    },
  ],
});

const cancelRequest = async () => {
  const allRequest = http.queue.get();
  const requestQueue = allRequest.filter(request => request.cancelWhenRouteChange);
  await http.cancel(requestQueue.map(request => request.requestId));
};

router.beforeEach(async (to, from, next) => {
  // 设置必填路由参数
  if (!to.params.projectId && store.state.curProjectId) {
    to.params.projectId = store.state.curProjectId;
  }
  if (!to.params.projectCode && store.state.curProjectCode) {
    to.params.projectCode = store.state.curProjectCode;
  }
  if (!to.params.clusterId && store.state.cluster.curCluster) {
    to.params.clusterId = store.state.cluster.curCluster.cluster_id;
  }

  await cancelRequest();

  // 路由切换二次确认
  if (from.meta?.backConfirm
        && !from.meta?.backConfirmExcludeRoutes?.includes(to.name)
        && !to.params.skipBackConfirm) {
    Vue.prototype.$bkInfo({
      type: 'warning',
      clsName: 'custom-info-confirm',
      title: i18n.t('确认退出当前编辑状态'),
      subTitle: i18n.t('退出后，你修改的内容将丢失'),
      defaultInfo: true,
      confirmFn: () => {
        next();
      },
    });
  } else {
    next();
  }
});

let containerEle = null;
router.afterEach((to) => {
  if (!containerEle) {
    containerEle = document.getElementsByClassName('container-content');
  }
  // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
  if (containerEle && containerEle[0] && containerEle[0].scrollTop !== 0) {
    containerEle[0].scrollTop = 0;
  }

  // 设置左侧菜单栏选中项
  let activeMenuId = to.meta?.menuId; // 1. 是否指定了菜单ID
  if (!activeMenuId) { // 2. 在菜单配置中查找当前路由对应的菜单ID
    const menuList = to.meta.isDashboard ? menuConfig.dashboardMenuList : menuConfig.k8sMenuList;
    menuList.find((menu) => {
      if (menu?.routeName === to.name) {
        activeMenuId = menu?.id;
        return true;
      } if (menu.children) {
        const child = menu.children.find(child => child.routeName === to.name);
        activeMenuId = child?.id;
        return !!activeMenuId;
      }
      return false;
    });
  }
  if (activeMenuId) {
    store.commit('updateCurMenuId', activeMenuId);
  } else {
    console.warn('找不到当前路由对应的菜单项，请检查', to);
  }
});

export default router;
