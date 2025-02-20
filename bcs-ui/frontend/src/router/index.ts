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

import ClusterManage from './cluster-manage';
import DeployManage from './deployment-manage';
import PluginManage from './plugin-manage';
import ProjectManage from './project-manage';
import ResourceView from './resource-view';// todo 有循环依赖

import { clusterDetail } from '@/api/modules/cluster-manager';
import cancelRequest from '@/common/cancel-request';
import $store from '@/store';
import useMenu from '@/views/app/use-menu';

Vue.use(VueRouter);

const Entry = () => import(/* webpackChunkName: 'entry' */'@/views/index.vue');
const DefaultSideMenu = () => import(/* webpackChunkName: 'entry' */'@/views/app/side-menu.vue');
const NotFound = () => import(/* webpackChunkName: 'entry' */'@/views/app/404.vue');
const Forbidden = () => import(/* webpackChunkName: 'entry' */'@/views/app/403.vue');
const Token = () => import(/* webpackChunkName: 'entry' */'@/views/user-token/token.vue');
const ProjectList = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/project/project.vue');

const router = new VueRouter({
  mode: 'history',
  routes: [
    {
      path: `${SITE_URL}`,
      name: 'home',
      redirect: {
        name: 'dashboardHome',
        params: {
          projectCode: $store.getters.curProjectCode,
        },
      },
    },
    // 403和user-token路由优先级比 ${SITE_URL}/:projectCode 高
    {
      path: `${SITE_URL}/:projectCode/403`,
      name: '403',
      props: route => ({ ...route.params, ...route.query }),
      component: Forbidden,
    },
    {
      path: `${SITE_URL}/projects`,
      name: 'projectManage',
      component: ProjectList,
      meta: {
        menuId: 'PROJECT_LIST',
      },
    },
    {
      path: `${SITE_URL}/:projectCode`,
      redirect: {
        name: 'dashboardHome',
      },
    },
    {
      path: `${SITE_URL}/projects/:projectCode`,
      components: {
        default: Entry,
        sideMenu: DefaultSideMenu,
      },
      redirect: {
        name: 'clusterMain',
      },
      children: [
        {
          path: 'user-token',
          name: 'token',
          component: Token,
        },
        ...ClusterManage,
        ...DeployManage,
        ...ProjectManage,
        ...PluginManage,
      ],
    },
    // 资源视图
    {
      path: `${SITE_URL}/projects/:projectCode`,
      name: 'dashboardIndex',
      component: Entry,
      meta: {
        hideMenu: true,
      },
      children: [
        ...ResourceView,
      ],
    },
    // 404
    {
      path: '*',
      name: '404',
      component: NotFound,
    },
  ],
});

// 自定义back逻辑
VueRouter.prototype.back = () => {
  if (window.history.length <= 2) {
    router.push({
      name: $store.state.curNav?.route || 'home',
    });
  } else {
    router.go(-1);
  }
};

const clusterPathPattern = new RegExp(`^${SITE_URL}/clusters/([^/]+)`); // 匹配 ${SITE_URL}/clusters/:clusterId

router.beforeEach(async (to, from, next) => {
  // 资源视图兼容无projectCode的路径
  if (!to.name || to.name === '404') {
    const match = to.fullPath?.match(clusterPathPattern);
    // 如果路径匹配 ${SITE_URL}/clusters/:clusterId
    const clusterId = match?.[1]; // 从匹配结果中获取 clusterId
    if (clusterId && clusterId !== '-') {
      // 查找clusterId对应的项目
      const data = await clusterDetail({
        $clusterId: clusterId,
      });
      const projectCode = data?.extraInfo?.projectCode;

      if (projectCode) {
        // 如果路径匹配 ${SITE_URL}/clusters/:clusterId
        const newPath = `${SITE_URL}/projects/${projectCode}${to.fullPath?.replace(SITE_URL, '')}`;
        // 重定向到新路径
        next(newPath);
        return;
      }
    }
  }

  // 设置必填路由参数
  if (!to.params.projectId && $store.getters.curProjectId) {
    to.params.projectId = $store.getters.curProjectId;
  }
  if (!to.params.projectCode && $store.getters.curProjectCode) {
    to.params.projectCode = $store.getters.curProjectCode;
  }
  // 取消上一个页面的请求
  await cancelRequest();
  // 校验路由是否开启
  const { validateRouteEnable, getNavByRoute } = useMenu();
  const result = await validateRouteEnable(to);
  if (!result) {
    // 未开启菜单项
    next({ name: '404' });
  } else {
    // 数据上报
    window.BkTrace?.startReported({
      module: 'router',
      operation: 'router',
      desc: '路由跳转',
      username: $store.state.user.username,
      projectCode: to.params.projectCode,
      to: to.name,
      from: from.name,
      navID: getNavByRoute(to)?.id,
    }, 'router');
    next();
  }
});

export default router;
