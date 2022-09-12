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

const App = () => import(/* webpackChunkName: 'app-entry' */'@/views/app');

const childRoutes = [
  // domain/bcs/projectId/app 应用页面
  {
    path: ':projectCode/app',
    component: App,
    children: [
      // k8s deployments 应用
      {
        path: 'deployments',
        name: 'deployments',
        children: [
          // k8s deployments 应用里的实例详情页面
          {
            path: ':instanceId',
            name: 'deploymentsInstanceDetail',
            meta: {
              menuId: 'deployments',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'deploymentsInstanceDetail2',
            meta: {
              menuId: 'deployments',
            },
          },
          // k8s deployments 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'deploymentsContainerDetail',
            meta: {
              menuId: 'deployments',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'deploymentsContainerDetail2',
            meta: {
              menuId: 'deployments',
            },
          },
          // k8s deployments 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'deploymentsInstantiation',
            meta: {
              menuId: 'deployments',
            },
          },
        ],
      },
      // k8s daemonset 应用
      {
        path: 'daemonset',
        name: 'daemonset',
        children: [
          // k8s daemonset 应用里的实例详情页面
          {
            path: ':instanceId',
            name: 'daemonsetInstanceDetail',
            meta: {
              menuId: 'daemonset',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'daemonsetInstanceDetail2',
            meta: {
              menuId: 'daemonset',
            },
          },
          // k8s daemonset 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'daemonsetContainerDetail',
            meta: {
              menuId: 'daemonset',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'daemonsetContainerDetail2',
            meta: {
              menuId: 'daemonset',
            },
          },
          // k8s daemonset 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'daemonsetInstantiation',
            meta: {
              menuId: 'daemonset',
            },
          },
        ],
      },
      // k8s job 应用
      {
        path: 'job',
        name: 'job',
        children: [
          // k8s job 应用里的实例详情页面
          {
            path: ':instanceId',
            name: 'jobInstanceDetail',
            meta: {
              menuId: 'job',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'jobInstanceDetail2',
            meta: {
              menuId: 'job',
            },
          },
          // k8s job 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'jobContainerDetail',
            meta: {
              menuId: 'job',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'jobContainerDetail2',
            meta: {
              menuId: 'job',
            },
          },
          // k8s job 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'jobInstantiation',
            meta: {
              menuId: 'job',
            },
          },
        ],
      },
      // k8s statefulset 应用
      {
        path: 'statefulset',
        name: 'statefulset',
        children: [
          // k8s statefulset 应用里的实例详情页面
          {
            path: ':instanceId',
            name: 'statefulsetInstanceDetail',
            meta: {
              menuId: 'statefulset',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'statefulsetInstanceDetail2',
            meta: {
              menuId: 'statefulset',
            },
          },
          // k8s statefulset 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'statefulsetContainerDetail',
            meta: {
              menuId: 'statefulset',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'statefulsetContainerDetail2',
            meta: {
              menuId: 'statefulset',
            },
          },
          // k8s statefulset 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'statefulsetInstantiation',
            meta: {
              menuId: 'statefulset',
            },
          },
        ],
      },
      // k8s gamestatefulset 应用
      {
        path: 'gamestatefulset',
        name: 'gamestatefulset',
      },
      // k8s gamedeployments 应用
      {
        path: 'gamedeployments',
        name: 'gamedeployments',
      },
      // k8s gamestatefulset 应用
      {
        path: 'customobjects',
        name: 'customobjects',
      },
    ],
  },
];

export default childRoutes;
