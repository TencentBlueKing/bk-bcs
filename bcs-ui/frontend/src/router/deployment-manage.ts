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

// 部署管理
const ChartList = () => import(/* webpackChunkName: 'deploy' */'@/views/deploy-manage/helm/chart-list.vue');
const ChartRelease = () => import(/* webpackChunkName: 'deploy' */'@/views/deploy-manage/helm/release-chart.vue');
const ReleaseList = () => import(/* webpackChunkName: 'deploy' */'@/views/deploy-manage/helm/release-list.vue');

const Variable = () => import(/* webpackChunkName: 'deploy' */'@/views/deploy-manage/variable/variable.vue');

// 首页
const Configuration = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/index.vue');

// 模板集
const Templateset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/templateset.vue');

// 模板实例化
const Instantiation = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/instantiation.vue');

// 创建 k8s 资源
const K8sConfigurationCreate = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8sCreate.vue');

// 添加模板集 - deployment
const K8sCreateDeployment = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/deployment.vue');

// 添加模板集 - service
const K8sCreateService = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/service.vue');

// 添加模板集 - configmap
const K8sCreateConfigmap = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/configmap.vue');

// 添加模板集 - secret
const K8sCreateSecret = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/secret.vue');

// 添加模板集 - daemonset
const K8sCreateDaemonset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/daemonset.vue');

// 添加模板集 - job
const K8sCreateJob = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/job.vue');

// 添加模板集 - statefulset
const K8sCreateStatefulset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/statefulset.vue');

// 添加模板集 - ingress
const K8sCreateIngress = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/ingress.vue');

// 添加模板集 - HPA
const K8sCreateHPA = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/hpa.vue');

// 添加yaml模板集 - yaml templateset
const K8sYamlTemplateset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/configuration/k8s-create/yaml-mode/index.vue');

const App = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/app/index.vue');

const Service = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/network/service.vue');
const Resource = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/resource/index.vue');
const ResourceConfigmap = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/resource/configmap.vue');
const ResourceSecret = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/resource/secret.vue');
const ResourceIngress = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/resource/ingress.vue');
const HPAIndex = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/deploy-manage/templateset/hpa/index.vue');

const childRoutes = [
  // helm
  {
    path: 'repos',
    props: route => ({ ...route.params, ...route.query }),
    name: 'chartList',
    component: ChartList,
  },
  {
    path: 'repos/:repoName/charts/:chartName/releases',
    props: true,
    name: 'releaseChart',
    component: ChartRelease,
    meta: {
      menuId: 'CHARTLIST',
    },
  },
  {
    path: 'releases',
    props: route => ({ ...route.params, ...route.query }),
    name: 'releaseList',
    component: ReleaseList,
  },
  {
    path: 'clusters/:cluster/repos/:repoName/charts/:chartName/releases/:namespace/:releaseName',
    props: true,
    name: 'updateRelease',
    component: ChartRelease,
    meta: {
      menuId: 'RELEASELIST',
    },
  },
  // 变量管理
  {
    path: 'variable',
    name: 'variable',
    component: Variable,
  },
  // 模板集
  {
    path: 'configuration',
    name: 'configurationMain',
    component: Configuration,
    children: [
      {
        path: 'templateset',
        name: 'templateset',
        component: Templateset,
      },
      {
        path: 'templateset/:templateId/instantiation',
        name: 'instantiation',
        component: Instantiation,
        meta: {
          menuId: 'TEMPLATESET',
        },
      },
      {
        path: 'k8s',
        name: 'k8sConfigurationCreate',
        component: K8sConfigurationCreate,
        children: [
          {
            path: 'templateset/deployment/:templateId',
            name: 'k8sTemplatesetDeployment',
            component: K8sCreateDeployment,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/service/:templateId',
            name: 'k8sTemplatesetService',
            component: K8sCreateService,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/configmap/:templateId',
            name: 'k8sTemplatesetConfigmap',
            component: K8sCreateConfigmap,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/secret/:templateId',
            name: 'k8sTemplatesetSecret',
            component: K8sCreateSecret,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/daemonset/:templateId',
            name: 'k8sTemplatesetDaemonset',
            component: K8sCreateDaemonset,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/job/:templateId',
            name: 'k8sTemplatesetJob',
            component: K8sCreateJob,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/statefulset/:templateId',
            name: 'k8sTemplatesetStatefulset',
            component: K8sCreateStatefulset,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/ingress/:templateId',
            name: 'k8sTemplatesetIngress',
            component: K8sCreateIngress,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'templateset/hpa/:templateId',
            name: 'k8sTemplatesetHPA',
            component: K8sCreateHPA,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
          {
            path: 'yaml-templateset/:templateId',
            name: 'K8sYamlTemplateset',
            component: K8sYamlTemplateset,
            meta: {
              menuId: 'TEMPLATESET',
            },
          },
        ],
      },
    ],
  },
  // 模板集应用列表
  {
    path: 'app',
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
              menuId: 'TEMPLATESET_DEPLOYMENT',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'deploymentsInstanceDetail2',
            meta: {
              menuId: 'TEMPLATESET_DEPLOYMENT',
            },
          },
          // k8s deployments 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'deploymentsContainerDetail',
            meta: {
              menuId: 'TEMPLATESET_DEPLOYMENT',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'deploymentsContainerDetail2',
            meta: {
              menuId: 'TEMPLATESET_DEPLOYMENT',
            },
          },
          // k8s deployments 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'deploymentsInstantiation',
            meta: {
              menuId: 'TEMPLATESET_DEPLOYMENT',
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
              menuId: 'TEMPLATESET_DAEMONSET',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'daemonsetInstanceDetail2',
            meta: {
              menuId: 'TEMPLATESET_DAEMONSET',
            },
          },
          // k8s daemonset 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'daemonsetContainerDetail',
            meta: {
              menuId: 'TEMPLATESET_DAEMONSET',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'daemonsetContainerDetail2',
            meta: {
              menuId: 'TEMPLATESET_DAEMONSET',
            },
          },
          // k8s daemonset 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'daemonsetInstantiation',
            meta: {
              menuId: 'TEMPLATESET_DAEMONSET',
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
              menuId: 'TEMPLATESET_JOB',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'jobInstanceDetail2',
            meta: {
              menuId: 'TEMPLATESET_JOB',
            },
          },
          // k8s job 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'jobContainerDetail',
            meta: {
              menuId: 'TEMPLATESET_JOB',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'jobContainerDetail2',
            meta: {
              menuId: 'TEMPLATESET_JOB',
            },
          },
          // k8s job 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'jobInstantiation',
            meta: {
              menuId: 'TEMPLATESET_JOB',
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
              menuId: 'TEMPLATESET_STATEFULSET',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory',
            name: 'statefulsetInstanceDetail2',
            meta: {
              menuId: 'TEMPLATESET_STATEFULSET',
            },
          },
          // k8s statefulset 应用里的容器详情页面
          {
            path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
            name: 'statefulsetContainerDetail',
            meta: {
              menuId: 'TEMPLATESET_STATEFULSET',
            },
          },
          {
            path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
            name: 'statefulsetContainerDetail2',
            meta: {
              menuId: 'TEMPLATESET_STATEFULSET',
            },
          },
          // k8s statefulset 应用里的应用实例化页面
          {
            path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
            name: 'statefulsetInstantiation',
            meta: {
              menuId: 'TEMPLATESET_STATEFULSET',
            },
          },
        ],
      },
      // k8s gamestatefulset 应用
      {
        path: 'gamestatefulset',
        name: 'gamestatefulset',
        children: [
          {
            path: ':category/detail',
            name: 'gamestatefulSetsInstanceDetail',
            meta: {
              menuId: 'TEMPLATESET_GAMESTATEFULSET',
            },
          },
        ],
      },
      // k8s gamedeployments 应用
      {
        path: 'gamedeployments',
        name: 'gamedeployments',
        children: [
          {
            path: ':category/detail',
            name: 'gamedeploymentsInstanceDetail',
            meta: {
              menuId: 'TEMPLATESET_GAMEDEPLOYMENT',
            },
          },
        ],
      },
      // k8s gamestatefulset 应用
      {
        path: 'customobjects',
        name: 'customobjects',
      },
    ],
  },
  // 模板集资源
  {
    path: 'service',
    name: 'service',
    component: Service,
  },
  {
    path: 'resource',
    name: 'resourceMain',
    component: Resource,
    children: [
      {
        path: 'configmap',
        component: ResourceConfigmap,
        name: 'resourceConfigmap',
      },
      {
        path: 'ingress',
        component: ResourceIngress,
        name: 'resourceIngress',
      },
      {
        path: 'secret',
        component: ResourceSecret,
        name: 'resourceSecret',
      },
    ],
  },
  {
    path: 'hpa',
    name: 'hpa',
    component: HPAIndex,
  },
];

export default childRoutes;
