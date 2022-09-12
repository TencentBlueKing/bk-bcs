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

// 变量管理
const Variable = () => import(/* webpackChunkName: 'variable' */'@/views/variable');

// 首页
const Configuration = () => import(/* webpackChunkName: 'configuration' */'@/views/configuration');

// 命名空间
const Namespace = () => import(/* webpackChunkName: 'namespace' */'@/views/configuration/namespace');

// 模板集
const Templateset = () => import(/* webpackChunkName: 'templateset' */'@/views/configuration/templateset');

// 模板实例化
const Instantiation = () => import(/* webpackChunkName: 'templateset' */'@/views/configuration/instantiation');

// 创建 k8s 资源
const K8sConfigurationCreate = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8sCreate');

// 添加模板集 - deployment
const K8sCreateDeployment = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/deployment');

// 添加模板集 - service
const K8sCreateService = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/service');

// 添加模板集 - configmap
const K8sCreateConfigmap = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/configmap');

// 添加模板集 - secret
const K8sCreateSecret = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/secret');

// 添加模板集 - daemonset
const K8sCreateDaemonset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/daemonset');

// 添加模板集 - job
const K8sCreateJob = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/job');

// 添加模板集 - statefulset
const K8sCreateStatefulset = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/statefulset');

// 添加模板集 - ingress
const K8sCreateIngress = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/ingress');

// 添加模板集 - HPA
const K8sCreateHPA = () => import(/* webpackChunkName: 'k8sTemplateset' */'@/views/configuration/k8s-create/hpa');

// 添加yaml模板集 - yaml templateset
const K8sYamlTemplateset = () => import(/* webpackChunkName: 'K8sYamlTemplateset' */'@/views/configuration/k8s-create/yaml-mode');

const childRoutes = [
  {
    path: ':projectCode/configuration',
    name: 'configurationMain',
    component: Configuration,
    children: [
      {
        path: 'namespace',
        component: Namespace,
        name: 'namespace',
        alias: '',
      },
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
          },
          {
            path: 'templateset/configmap/:templateId',
            name: 'k8sTemplatesetConfigmap',
            component: K8sCreateConfigmap,
          },
          {
            path: 'templateset/secret/:templateId',
            name: 'k8sTemplatesetSecret',
            component: K8sCreateSecret,
          },
          {
            path: 'templateset/daemonset/:templateId',
            name: 'k8sTemplatesetDaemonset',
            component: K8sCreateDaemonset,
          },
          {
            path: 'templateset/job/:templateId',
            name: 'k8sTemplatesetJob',
            component: K8sCreateJob,
          },
          {
            path: 'templateset/statefulset/:templateId',
            name: 'k8sTemplatesetStatefulset',
            component: K8sCreateStatefulset,
          },
          {
            path: 'templateset/ingress/:templateId',
            name: 'k8sTemplatesetIngress',
            component: K8sCreateIngress,
          },
          {
            path: 'templateset/hpa/:templateId',
            name: 'k8sTemplatesetHPA',
            component: K8sCreateHPA,
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
      {
        path: 'var',
        name: 'var',
        component: Variable,
      },
    ],
  },
];

export default childRoutes;
